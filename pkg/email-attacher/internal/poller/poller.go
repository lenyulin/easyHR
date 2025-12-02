package poller

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"easyHR/pkg/email-attacher/domain"
	"easyHR/pkg/email-attacher/internal/config"
	"easyHR/pkg/email-attacher/internal/factory"
	"easyHR/pkg/email-attacher/internal/retry"
	"easyHR/pkg/email-attacher/internal/storage"

	emailattacher "easyHR/pkg/email-attacher"
)

// Poller 轮询调度器
type Poller struct {
	cfg                *config.InternalConfig           // 内部配置
	clients            []domain.EmailClient             // 邮箱客户端列表
	processedStorage   *storage.ProcessedStorage        // 已处理邮件存储
	attachmentStorage  *storage.AttachmentStorage       // 附件存储
	logger             *config.ZapLogger                // 日志器
	retryCfg           config.RetryConfig               // 重试配置
	attachmentCallback emailattacher.AttachmentCallback // 附件下载回调
	errorCallback      emailattacher.ErrorCallback      // 错误回调
}

// NewPoller 初始化轮询器
func NewPoller(internalCfg *config.InternalConfig) (*Poller, error) {
	// 初始化已处理邮件存储
	processedStorage, err := storage.NewProcessedStorage(internalCfg.StorageConfig.ProcessedEmailsPath)
	if err != nil {
		return nil, fmt.Errorf("已处理存储初始化失败：%w", err)
	}

	// 初始化附件存储
	attachmentStorage := storage.NewAttachmentStorage(internalCfg.StorageConfig.AttachmentSavePath)

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
		logger:            internalCfg.Logger,
		retryCfg:          internalCfg.RetryConfig,
	}, nil
}

// Run 启动轮询（阻塞，需在goroutine中运行）
func (p *Poller) Run(ctx context.Context) {
	p.logger.Info("轮询器启动", config.Zap.String("interval", p.cfg.PollerConfig.Interval.String()))

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
		p.logger.Debug("开始处理服务商", config.Zap.String("provider", provider))

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
		if err != nil {
			p.logger.Error("获取未读邮件失败", config.Zap.String("provider", provider), config.Zap.Error(err))
			if p.errorCallback != nil {
				p.errorCallback(err, provider)
			}
			continue
		}

		p.logger.Info("发现未读邮件", config.Zap.String("provider", provider), config.Zap.Int("count", len(unreadEmails)))

		// 处理每封邮件
		for _, email := range unreadEmails {
			// 过滤已处理邮件
			if p.processedStorage.IsProcessed(email.ID) {
				p.logger.Debug("邮件已处理，跳过", config.Zap.String("email_id", email.ID))
				continue
			}

			// 下载附件
			if len(email.Attachments) > 0 {
				p.logger.Debug("开始下载附件", config.Zap.String("email_id", email.ID), config.Zap.Int("attach_count", len(email.Attachments)))
				for _, att := range email.Attachments {
					// 构建附件保存路径（服务商/邮件ID/附件名）
					savePath := filepath.Join(p.attachmentStorage.GetBasePath(), provider, email.ID)
					err := p.attachmentStorage.SaveAttachment(client, att, savePath)
					if err != nil {
						p.logger.Error("附件下载失败", config.Zap.String("attach_name", att.Name), config.Zap.Error(err))
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
				p.logger.Error("标记已读失败", config.Zap.String("email_id", email.ID), config.Zap.Error(err))
				if p.errorCallback != nil {
					p.errorCallback(err, provider)
				}
				continue
			}

			// 记录已处理
			if err := p.processedStorage.MarkAsProcessed(email.ID); err != nil {
				p.logger.Error("记录已处理失败", config.Zap.String("email_id", email.ID), config.Zap.Error(err))
				if p.errorCallback != nil {
					p.errorCallback(err, provider)
				}
			}
		}
	}

	p.logger.Info("本轮轮询结束")
}

// SetAttachmentCallback 设置附件下载回调
func (p *Poller) SetAttachmentCallback(callback emailattacher.AttachmentCallback) {
	p.attachmentCallback = callback
}

// SetErrorCallback 设置错误回调
func (p *Poller) SetErrorCallback(callback emailattacher.ErrorCallback) {
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
