package domain

import "github.com/emersion/go-imap/v2"

// EmailClient 统一邮件客户端接口（所有服务商实现需遵循）
type EmailClient interface {
	// Init 初始化客户端（传入内部强类型配置）
	Init(cfg interface{}) error
	// ListUnreadEmails 获取未读邮件列表
	ListUnreadEmails() ([]Email, error)
	// DownloadAttachment 下载单个附件
	DownloadAttachment(att Attachment, savePath string) error
	// MarkAsRead 标记邮件为已读
	MarkAsRead(emailID imap.UID) error
	// GetProvider 获取服务商名称（如 "qq"、"netease"）
	GetProvider() string
}
