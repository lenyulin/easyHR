package qq

import (
	"context"
	"easyHR/pkg/email-attacher/domain"
	"easyHR/pkg/email-attacher/internal/config"
	"easyHR/pkg/email-attacher/internal/retry"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

var searchCriteria = &imap.SearchCriteria{Or: [][2]imap.SearchCriteria{{
	{Body: []string{"校园招聘"}},
	{Body: []string{"Campus Recruitment"}},
}}}

type QQEmailClient struct {
	provider   string
	imapCfg    config.IMAPConfig
	imapClient *imapclient.Client
}

// NewQQEmailClient 创建QQ客户端实例（工厂调用）
func NewQQEmailClient() *QQEmailClient {
	return &QQEmailClient{
		provider: "qq",
	}
}

// Init 初始化客户端
func (q *QQEmailClient) Init(cfg interface{}) error {
	// 类型断言（内部配置已确保类型正确）
	internalCfg, ok := cfg.(config.ProviderConfig)
	if !ok {
		return fmt.Errorf("QQ客户端配置类型错误")
	}
	q.imapCfg = internalCfg.IMAPConfig

	// 建立IMAP连接（带重试）
	ctx := context.Background()
	return retry.Retry(ctx, 3, 2*time.Second, func() error {
		client, err := connectIMAP(q.imapCfg)
		if err != nil {
			return err
		}
		q.imapClient = client
		return nil
	})
}

// ListUnreadEmails 获取未读邮件列表
func (q *QQEmailClient) ListUnreadEmails() ([]domain.Email, error) {
	if q.imapClient == nil {
		return nil, fmt.Errorf("IMAP客户端未初始化")
	}

	// 选择收件箱
	_, err := q.imapClient.Select("INBOX", nil).Wait()
	if err != nil {
		return nil, err
	}

	// 搜索未读邮件
	data, err := q.imapClient.UIDSearch(searchCriteria, nil).Wait()
	if err != nil {
		return nil, err
	}
	uids := data.AllUIDs()
	if len(uids) == 0 {
		return []domain.Email{}, nil
	}

	// 获取邮件详情
	fetchOptions := &imap.FetchOptions{
		Envelope:      true,
		Flags:         true,
		InternalDate:  true,
		RFC822Size:    true,
		BodyStructure: &imap.FetchItemBodyStructure{Extended: true},
	}

	var emails []domain.Email
	// Fetch returns a command, we can stream responses
	fetchCmd := q.imapClient.Fetch(imap.UIDSetNum(uids...), fetchOptions)
	messages, err := fetchCmd.Collect()
	if err != nil {
		return nil, err
	}

	for _, msg := range messages {
		isRead := false
		for _, flag := range msg.Flags {
			if flag == imap.FlagSeen {
				isRead = true
				break
			}
		}

		email := domain.Email{
			ID:     msg.UID,
			IsRead: isRead,
		}

		// 解析信封信息
		if msg.Envelope != nil {
			if len(msg.Envelope.From) > 0 {
				email.From = msg.Envelope.From[0].Addr()
			}
			email.To = make([]string, len(msg.Envelope.To))
			for i, to := range msg.Envelope.To {
				email.To[i] = to.Addr()
			}
			email.Subject = msg.Envelope.Subject
			email.SentAt = msg.Envelope.Date
		}

		//TODO:解析附件未实现
		panic("not implete method")

		// 解析附件（简化逻辑，实际需解析邮件正文结构）
		// email.Attachments, _ = q.parseAttachments(msg)
		email.Attachments = []domain.Attachment{}
		emails = append(emails, email)
	}

	return emails, nil
}

// DownloadAttachment 下载附件
func (q *QQEmailClient) DownloadAttachment(att domain.Attachment, savePath string) error {
	// 实际实现：通过IMAP获取附件内容，写入savePath/att.Name
	// 简化示例：直接创建空文件（实际需解析邮件MIME结构获取附件内容）
	filePath := filepath.Join(savePath, att.Name)
	return os.WriteFile(filePath, att.Content, 0644)
}

// MarkAsRead 标记邮件为已读
func (q *QQEmailClient) MarkAsRead(emailID imap.UID) error {
	if q.imapClient == nil {
		return fmt.Errorf("IMAP客户端未初始化")
	}

	// 标记为已读（添加Seen标记）
	_, err := q.imapClient.Store(imap.UIDSetNum(emailID), &imap.StoreFlags{
		Op:    imap.StoreFlagsAdd,
		Flags: []imap.Flag{imap.FlagSeen},
	}, nil).Collect()
	return err
}

// GetProvider 获取服务商名称
func (q *QQEmailClient) GetProvider() string {
	return q.provider
}

// parseAttachments 解析邮件附件（简化实现）
func (q *QQEmailClient) parseAttachments(msg *imapclient.FetchMessageData) ([]domain.Attachment, error) {

	panic("not implete method")

	// var attachments []domain.Attachment

	// 实际需解析邮件MIME结构，这里简化模拟
	// if msg.Envelope != nil && strings.Contains(msg.Envelope.Subject, "附件") {
	// 	attachments = append(attachments, domain.Attachment{
	// 		ID:   fmt.Sprintf("qq-att-%d", msg.UID),
	// 		Name: fmt.Sprintf("attachment-%d.pdf", msg.UID),
	// 		Size: msg.Size,
	// 	})
	// }

	// return attachments, nil
}
