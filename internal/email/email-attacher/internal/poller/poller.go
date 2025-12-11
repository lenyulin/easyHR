package poller

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"easyHR/internal/email/email-attacher/domain"
	"easyHR/internal/email/email-attacher/internal/config"
	"easyHR/internal/email/email-attacher/internal/factory"
	"easyHR/internal/email/email-attacher/internal/retry"
	"easyHR/internal/email/email-attacher/internal/storage"
	"easyHR/pkg/logger"
)

// Poller 轮询调度器
type Poller struct {
	cfg                *config.InternalConfig     // 内部配置
	clients            []domain.EmailClient       // 邮箱客户端列表
	processedStorage   *storage.ProcessedStorage  // 已处理邮件存储
	attachmentStorage  *storage.AttachmentStorage // 附件存储
	retryCfg           config.RetryConfig         // 重试配置
	attachmentCallback domain.AttachmentCallback  // 附件下载回调
	errorCallback      domain.ErrorCallback       // 错误回调
	logger             logger.LoggerV1            // 日志器
}

// NewPoller 初始化轮询器
func NewPoller(internalCfg *config.InternalConfig, logger logger.LoggerV1) (*Poller, error) {
	// 初始化已处理邮件存储
	processedStorage, err := storage.NewProcessedStorage(internalCfg.StorageConfig.ProcessedEmailsPath, logger)
	if err != nil {
		return nil, fmt.Errorf("已处理存储初始化失败：%w", err)
	}

	// 初始化附件存储
	attachmentStorage := storage.NewAttachmentStorage(internalCfg.StorageConfig.AttachmentSavePath, logger)

	// 创建所有服务商客户端
	clients, err := createClients(internalCfg.ProvidersConfig)
	if err != nil {
		return nil, fmt.Errorf("客户端创建失败：%w", err)
	}

	return &Poller{
		cfg:               internalCfg,
		clients:           clients,
		processedStorage:  processedStorage,
		attachmentStorage: attachmentStorage,
		logger:            logger,
		retryCfg:          internalCfg.RetryConfig,
	}, nil
}

// Run 启动轮询（阻塞，需在goroutine中运行）
func (p *Poller) Run(ctx context.Context) {
	p.logger.Info("轮训器启动", logger.Field{
		Key: "interval",
		Val: p.cfg.PollerConfig.Interval.String(),
	})

	// 首次执行一次
	p.doPoll()

	// 定时轮询
	ticker := time.NewTicker(p.cfg.PollerConfig.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("轮询器收到停止信号，正在退出...")
			return
		case <-ticker.C:
			p.doPoll()
		}
	}
}

// doPoll 单次轮询逻辑
func (p *Poller) doPoll() {
	p.logger.Info("开始新一轮轮询")

	for _, client := range p.clients {
		provider := client.GetProvider()
		p.logger.Info("开始处理服务商", logger.Field{
			Key: "provider",
			Val: provider,
		})

		// 获取未读邮件（带重试）
		var unreadEmails []domain.Email
		err := retry.Retry(context.Background(), p.retryCfg.MaxAttempts, p.retryCfg.Interval, func() error {
			emails, err := client.ListUnreadEmails()
			if err != nil {
				return err
			}
			unreadEmails = emails
			return nil
		})
		//emails, err := client.ListUnreadEmails()
		if err != nil {
			p.logger.Info("获取未读邮件失败", logger.Field{
				Key: "provider",
				Val: err.Error(),
			})
			if p.errorCallback != nil {
				p.errorCallback(err, provider)
			}
			continue
		}
		//unreadEmails = emails
		if len(unreadEmails) == 0 {
			p.logger.Info("发现未读邮件",
				logger.Field{
					Key: "provider",
					Val: provider},
				logger.Field{
					Key: "count",
					Val: len(unreadEmails),
				})
		}
		// 处理每封邮件
		for _, email := range unreadEmails {
			// 过滤已处理邮件
			if p.processedStorage.IsProcessed(email.ID) {
				p.logger.Debug("邮件已处理，跳过", logger.Field{
					Key: "email_id",
					Val: email.ID,
				})
				continue
			}

			// 下载附件
			if len(email.Attachments) > 0 {
				p.logger.Debug("开始下载附件",
					logger.Field{
						Key: "email_id",
						Val: email.ID,
					}, logger.Field{
						Key: "attach_count",
						Val: len(email.Attachments),
					})
				for _, att := range email.Attachments {
					// 构建附件保存路径（服务商/邮件ID/邮件主题/附件名）
					savePath := filepath.Join(p.attachmentStorage.GetBasePath(), provider, strconv.Itoa(int(email.ID)), email.Subject, att.Name)
					err := p.attachmentStorage.SaveAttachment(client, att, savePath)
					if err != nil {
						p.logger.Debug("附件下载失败",
							logger.Field{
								Key: "attach_name",
								Val: att.Name,
							}, logger.Field{
								Key: "err",
								Val: err.Error(),
							})
						if p.errorCallback != nil {
							p.errorCallback(err, provider)
						}
						continue
					}

					// 触发回调
					if p.attachmentCallback != nil {
						p.attachmentCallback(email, att, savePath)
					}
				}
			}

			// 标记邮件为已读（带重试）
			err = retry.Retry(context.Background(), p.retryCfg.MaxAttempts, p.retryCfg.Interval, func() error {
				return client.MarkAsRead(email.ID)
			})
			if err != nil {
				p.logger.Debug("标记已读失败",
					logger.Field{
						Key: "email_id",
						Val: email.ID,
					}, logger.Field{
						Key: "err",
						Val: err.Error(),
					})
				if p.errorCallback != nil {
					p.errorCallback(err, provider)
				}
				continue
			}

			// 记录已处理
			if err := p.processedStorage.MarkAsProcessed(email.ID); err != nil {
				p.logger.Debug("记录已处理失败",
					logger.Field{
						Key: "email_id",
						Val: email.ID,
					}, logger.Field{
						Key: "err",
						Val: err.Error(),
					})
				if p.errorCallback != nil {
					p.errorCallback(err, provider)
				}
			}
		}
	}

	p.logger.Info("本轮轮询结束")
}

// SetAttachmentCallback 设置附件下载回调
func (p *Poller) SetAttachmentCallback(callback domain.AttachmentCallback) {
	p.attachmentCallback = callback
}

// SetErrorCallback 设置错误回调
func (p *Poller) SetErrorCallback(callback domain.ErrorCallback) {
	p.errorCallback = callback
}

// createClients 创建所有服务商客户端
func createClients(providersConfig []config.ProviderConfig) ([]domain.EmailClient, error) {
	var clients []domain.EmailClient

	for _, provCfg := range providersConfig {
		// 通过工厂创建客户端实例
		client, err := factory.NewEmailClient(provCfg.Type)
		if err != nil {
			return nil, err
		}

		// 初始化客户端（传入强类型配置）
		if err := client.Init(provCfg); err != nil {
			return nil, fmt.Errorf("服务商%s初始化失败：%w", provCfg.Type, err)
		}

		clients = append(clients, client)
	}

	return clients, nil
}
