package emailreply

import (
	"errors"
	"fmt"
)

// Config 存储SMTP服务器连接参数和认证信息
type Config struct {
	SMTPServer string // SMTP服务器地址
	SMTPPort   int    // SMTP服务器端口
	Username   string // 邮箱用户名
	Password   string // 邮箱密码
	UseTLS     bool   // 是否使用TLS
	FromEmail  string // 发件人邮箱
	FromName   string // 发件人名称
}

// ReplyParams 存储邮件回复模板所需的动态数据
type ReplyParams struct {
	Name        string // 收件人姓名
	CompanyName string // 公司名称
	Year        string // 年份
	Semester    string // 学期
	JobName     string // 岗位名称
	Link        string // 投递状态查询链接
	HRName      string // HR姓名
	HRPhone     string // HR电话
}

var (
	// config 存储包的全局配置
	config *Config
	// ErrNotInitialized 表示包未初始化错误
	ErrNotInitialized = errors.New("email-reply package not initialized")
)

// Init 初始化email-reply包，设置必要的配置信息
// 参数：
//
//	cfg - 包含SMTP服务器连接参数和认证信息的配置对象
//
// 返回值：
//
//	error - 初始化过程中遇到的错误，如果成功则返回nil
func Init(cfg *Config) error {
	if cfg == nil {
		return errors.New("config cannot be nil")
	}

	// 验证必要的配置项
	if cfg.SMTPServer == "" {
		return errors.New("SMTPServer is required")
	}
	if cfg.SMTPPort == 0 {
		return errors.New("SMTPPort is required")
	}
	if cfg.Username == "" {
		return errors.New("Username is required")
	}
	if cfg.Password == "" {
		return errors.New("Password is required")
	}
	if cfg.FromEmail == "" {
		return errors.New("FromEmail is required")
	}
	if cfg.FromName == "" {
		return errors.New("FromName is required")
	}

	// 保存配置
	config = cfg
	return nil
}

// SendReply 发送邮件回复
// 参数：
//
//	toEmail - 收件人邮箱地址
//	params - 包含模板所需动态数据的参数对象
//	subject - 邮件主题
//
// 返回值：
//
//	error - 发送过程中遇到的错误，如果成功则返回nil
func SendReply(toEmail string, params *ReplyParams, subject string) error {
	if config == nil {
		return ErrNotInitialized
	}

	if toEmail == "" {
		return errors.New("toEmail is required")
	}

	if params == nil {
		return errors.New("params cannot be nil")
	}

	if subject == "" {
		return errors.New("subject is required")
	}

	// 渲染邮件模板
	body, err := renderTemplate(params)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// 发送邮件
	if err := sendEmail(toEmail, subject, body); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
