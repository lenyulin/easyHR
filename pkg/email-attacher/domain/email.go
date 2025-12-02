package domain

import "time"

// Attachment 邮件附件模型（主项目可能需要引用）
type Attachment struct {
	ID      string // 附件唯一标识
	Name    string // 附件文件名
	Size    int64  // 附件大小（字节）
	URL     string // 附件下载地址（部分服务商提供）
	Content []byte // 附件二进制内容（直接获取时使用）
}

// Email 邮件模型（主项目可能需要引用）
type Email struct {
	ID          string       // 邮件唯一标识
	From        string       // 发件人（xxx@xxx.com）
	To          []string     // 收件人列表
	Subject     string       // 邮件主题
	SentAt      time.Time    // 发送时间
	Attachments []Attachment // 附件列表
	IsRead      bool         // 是否已读
}
