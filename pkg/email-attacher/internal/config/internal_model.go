package config

import (
	"time"

	"go.uber.org/zap"
)

// InternalConfig 子包内部核心配置（整合所有模块配置）
type InternalConfig struct {
	PollerConfig    PollerConfig     // 轮询器配置
	StorageConfig   StorageConfig    // 存储配置
	Logger          *zap.Logger      // 初始化后的日志器
	RetryConfig     RetryConfig      // 重试配置
	ProvidersConfig []ProviderConfig // 服务商配置（强类型）
}

// PollerConfig 轮询器专属配置
type PollerConfig struct {
	Interval time.Duration // 轮询间隔（time.Duration类型）
}

// StorageConfig 存储模块专属配置
type StorageConfig struct {
	AttachmentSavePath  string // 附件保存路径（绝对路径）
	ProcessedEmailsPath string // 已处理邮件记录路径（绝对路径）
}

// RetryConfig 重试模块专属配置
type RetryConfig struct {
	MaxAttempts int           // 最大重试次数
	Interval    time.Duration // 重试间隔
}

// ProviderConfig 内部服务商配置
type ProviderConfig struct {
	Type       string     // 服务商类型（qq/netease等）
	IMAPConfig IMAPConfig // IMAP协议配置（强类型）
}

// IMAPConfig IMAP协议专属配置
type IMAPConfig struct {
	Addr     string // 完整地址（host:port）
	Username string // 账号
	Password string // 授权码/密码
}
