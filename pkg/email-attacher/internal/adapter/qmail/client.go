package qq

import (
	"bytes"
	"context"
	"easyHR/pkg/email-attacher/domain"
	"easyHR/pkg/email-attacher/internal/config"
	"easyHR/pkg/email-attacher/internal/retry"
	"encoding/base64"
	"fmt"
	"io"
	"mime/quotedprintable"
	"os"
	"path/filepath"
	"strings"
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
	uids, err := q.imapClient.UIDSearch(&imap.SearchCriteria{
		NotFlag: []imap.Flag{imap.FlagSeen},
	}, nil).Wait()

	if err != nil {
		return nil, err
	}
	if len(uids.AllUIDs()) == 0 {
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
	fetchCmd := q.imapClient.Fetch(imap.UIDSetNum(uids.AllUIDs()...), fetchOptions)
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

		// 解析附件（简化逻辑，实际需解析邮件正文结构）
		email.Attachments, _ = q.parseAttachments(msg)
		//email.Attachments = []domain.Attachment{}
		emails = append(emails, email)
	}

	return emails, nil
}

// DownloadAttachment 下载附件，只处理PDF文件
func (q *QQEmailClient) DownloadAttachment(att domain.Attachment, savePath string) error {
	// 检查文件后缀，只处理PDF文件
	if strings.ToLower(filepath.Ext(att.Name)) != ".pdf" {
		// 不是PDF文件，跳过下载
		return nil
	}

	// 检查att.Content是否为空
	if att.Content == nil {
		return fmt.Errorf("PDF附件内容为空，无法保存")
	}

	// 创建保存路径（如果不存在）
	if err := os.MkdirAll(savePath, 0755); err != nil {
		return fmt.Errorf("创建保存路径失败: %w", err)
	}

	// 构建完整的文件路径
	filePath := filepath.Join(savePath, att.Name)

	// 使用att.Content写入PDF文件
	if err := os.WriteFile(filePath, att.Content, 0644); err != nil {
		return fmt.Errorf("写入PDF文件失败: %w", err)
	}

	return nil
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

// parseAttachments 解析邮件附件
func (q *QQEmailClient) parseAttachments(msg *imapclient.FetchMessageBuffer) ([]domain.Attachment, error) {
	var attachments []domain.Attachment

	// 定义遍历函数，符合 BodyStructureWalkFunc 签名
	walkFunc := func(path []int, part imap.BodyStructure) (walkChildren bool) {
		// 总是继续遍历子部分
		walkChildren = true

		// 类型断言：检查是否是单个部分
		singlePart, ok := part.(*imap.BodyStructureSinglePart)
		if !ok {
			// 不是单个部分，继续遍历子部分
			return
		}

		// 获取文件名
		filename := singlePart.Filename()
		isAttachment := false

		// 检查是否是附件
		disposition := part.Disposition()
		if disposition != nil && disposition.Value == "attachment" {
			isAttachment = true
			// 如果 Filename() 返回空，尝试从 disposition params 获取
			if filename == "" && disposition.Params != nil {
				filename = disposition.Params["filename"]
			}
		}

		// 检查是否有文件名（可能是内联附件）
		if filename != "" {
			isAttachment = true
		}

		// 如果是附件，创建 Attachment 对象
		if isAttachment && filename != "" {
			// 生成唯一 ID（使用邮件 UID 和文件名生成）
			id := fmt.Sprintf("%d-%s", msg.UID, filename)

			// 初始化 Attachment 对象
			attachment := domain.Attachment{
				ID:   id,
				Name: filename,
				Size: int64(singlePart.Size),
				URL:  "",
			}

			// 检查是否为PDF文件，如果是，获取附件内容
			if strings.ToLower(filepath.Ext(filename)) == ".pdf" {
				// 创建 BodySection 请求，指定要获取的部分路径
				bodySection := &imap.FetchItemBodySection{
					Part: path,
				}

				// 配置fetch选项，获取该部分的内容
				fetchOptions := &imap.FetchOptions{
					BodySection: []*imap.FetchItemBodySection{bodySection},
				}

				// 发送fetch命令获取该附件内容
				fetchCmd := q.imapClient.Fetch(imap.UIDSetNum(msg.UID), fetchOptions)
				fetchMsgs, err := fetchCmd.Collect()
				if err != nil {
					// 获取附件内容失败，记录错误但继续处理其他附件
					fmt.Printf("获取PDF附件内容失败: %v\n", err)
				} else if len(fetchMsgs) > 0 {
					// 使用 FindBodySection 方法获取附件内容
					rawContent := fetchMsgs[0].FindBodySection(bodySection)
					if rawContent != nil {
						// 根据编码类型解码内容
						encoding := singlePart.Encoding
						var decodedContent []byte

						switch strings.ToLower(encoding) {
						case "base64":
							// 解码base64内容
							decodedContent = make([]byte, base64.StdEncoding.DecodedLen(len(rawContent)))
							n, err := base64.StdEncoding.Decode(decodedContent, rawContent)
							if err != nil {
								fmt.Printf("解码base64内容失败: %v\n", err)
								break
							}
							decodedContent = decodedContent[:n]
						case "quoted-printable":
							// 解码quoted-printable内容
							reader := quotedprintable.NewReader(bytes.NewReader(rawContent))
							decodedContent, err = io.ReadAll(reader)
							if err != nil {
								fmt.Printf("解码quoted-printable内容失败: %v\n", err)
								break
							}
						default:
							// 其他编码（如7BIT、8BIT等）直接使用原始内容
							decodedContent = rawContent
						}

						// 将解码后的内容保存到Attachment对象
						attachment.Content = decodedContent
					}
				}
			}

			attachments = append(attachments, attachment)
		}

		return
	}

	// 使用 Walk 方法遍历邮件结构树
	msg.BodyStructure.Walk(walkFunc)

	return attachments, nil
}
