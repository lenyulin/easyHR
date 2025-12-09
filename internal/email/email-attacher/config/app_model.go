package config

import "time"

// AppConfig 子包核心配置（主项目需构造此结构体传入）
type AppConfig struct {
	PollInterval        int              `yaml:"poll_interval"`         // 轮询间隔（秒）
	AttachmentSavePath  string           `yaml:"attachment_save_path"`  // 附件保存路径（主项目指定）
	ProcessedEmailsPath string           `yaml:"processed_emails_path"` // 已处理邮件记录路径（主项目指定）
	Providers           []ProviderConfig `yaml:"providers"`             // 服务商列表
	RetryConfig         *RetryConfig     `yaml:"retry_config"`          // 重试配置（可选）
}

// ProviderConfig 单个服务商配置
type ProviderConfig struct {
	Type   string                 `yaml:"type"`   // 服务商类型（qq/netease/gmail/outlook）
	Config map[string]interface{} `yaml:"config"` // 服务商专属配置（imap地址、账号等）
}

// RetryConfig 重试配置（可选）
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts"` // 最大重试次数
	Interval    time.Duration `yaml:"interval"`     // 重试间隔
}
