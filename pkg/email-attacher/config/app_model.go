package config

import "time"

// AppConfig 子包核心配置（主项目需构造此结构体传入）
type AppConfig struct {
	PollInterval        int              // 轮询间隔（秒）
	AttachmentSavePath  string           // 附件保存路径（主项目指定）
	ProcessedEmailsPath string           // 已处理邮件记录路径（主项目指定）
	Providers           []ProviderConfig // 服务商列表
	LoggerConfig        *LoggerConfig    // 日志配置（可选）
	RetryConfig         *RetryConfig     // 重试配置（可选）
}

// ProviderConfig 单个服务商配置
type ProviderConfig struct {
	Type   string                 // 服务商类型（qq/netease/gmail/outlook）
	Config map[string]interface{} // 服务商专属配置（imap地址、账号等）
}

// LoggerConfig 日志配置（可选）
type LoggerConfig struct {
	Level string // 日志级别（debug/info/warn/error）
	Path  string // 日志保存路径（可选，为空则输出到控制台）
}

// RetryConfig 重试配置（可选）
type RetryConfig struct {
	MaxAttempts int           // 最大重试次数
	Interval    time.Duration // 重试间隔
}
