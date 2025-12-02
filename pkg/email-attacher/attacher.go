package emailattacher

import (
	"context"

	"easyHR/pkg/email-attacher/config"
	"easyHR/pkg/email-attacher/domain"
	internalConfig "easyHR/pkg/email-attacher/internal/config"
	"easyHR/pkg/email-attacher/internal/config/validator"
	"easyHR/pkg/email-attacher/internal/poller"
)

// EmailAttacher 邮件附件下载器（对外暴露的核心结构体）
type EmailAttacher struct {
	poller *poller.Poller  // 内部轮询器（隐藏实现）
	ctx    context.Context // 上下文（用于优雅关闭）
	cancel context.CancelFunc
}

// 附件下载成功回调函数类型
type AttachmentCallback func(email domain.Email, att domain.Attachment, savePath string)

// 错误回调函数类型
type ErrorCallback func(err error, provider string)

// NewEmailAttacher 初始化下载器（主项目入口方法）
func NewEmailAttacher(cfg *config.AppConfig) (*EmailAttacher, error) {
	// 1. 校验外部配置合法性
	if err := validator.ValidateConfig(cfg); err != nil {
		return nil, err
	}

	// 2. 转换外部配置为内部配置
	internalCfg, err := internalConfig.InitInternalConfig(cfg)
	if err != nil {
		return nil, err
	}

	// 3. 初始化内部轮询器
	pollerInstance, err := poller.NewPoller(internalCfg)
	if err != nil {
		return nil, err
	}

	// 4. 创建上下文（用于停止轮询）
	ctx, cancel := context.WithCancel(context.Background())
	return &EmailAttacher{
		poller: pollerInstance,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Start 启动轮询（非阻塞，在goroutine中运行）
func (e *EmailAttacher) Start() error {
	go e.poller.Run(e.ctx)
	return nil
}

// Stop 停止轮询（优雅关闭）
func (e *EmailAttacher) Stop() {
	e.cancel()
}

// OnAttachmentDownloaded 注册附件下载成功回调（主项目可选）
func (e *EmailAttacher) OnAttachmentDownloaded(callback AttachmentCallback) {
	e.poller.SetAttachmentCallback(callback)
}

// OnError 注册错误回调（主项目可选）
func (e *EmailAttacher) OnError(callback ErrorCallback) {
	e.poller.SetErrorCallback(callback)
}
