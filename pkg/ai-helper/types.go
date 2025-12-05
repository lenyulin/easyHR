package aihelper

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// StreamCallback 定义流式响应的回调函数
type StreamCallback func(msg string)

// AIModel 定义AI模型的通用接口
type AIModel interface {
	GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
	StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error)
	GetModelType() string
}

// AIHelper 封装了与AI模型交互的核心逻辑
type AIHelper struct {
	model     AIModel                        // AI模型接口，支持不同模型实现
	messages  []Message                      // 消息历史列表，存储用户和AI的对话记录
	mu        sync.RWMutex                   // 读写锁，保护消息历史并发访问
	SessionID string                         // 会话唯一标识，用于绑定消息和上下文
	saveFunc  func(Message) (Message, error) // 消息存储回调函数，默认异步发布到RabbitMQ
	SysMsg    string                         // 系统默认提示词，用于指导AI模型的行为
}

// Message 定义消息持久化的结构
type Message struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id" bson:"_id,omitempty"`
	SessionID string    `gorm:"index;not null;type:varchar(36)" json:"session_id" bson:"session_id"`
	Role      string    `gorm:"type:varchar(20)" json:"role" bson:"role"` // user, system, assistant
	Content   string    `gorm:"type:text" json:"content" bson:"content"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// Configuration 定义LLM服务的配置
type Configuration struct {
	ModelType string `mapstructure:"model_type" yaml:"model_type"` // e.g., "doubao", "gpt4"
	APIKey    string `mapstructure:"api_key" yaml:"api_key"`
	BaseURL   string `mapstructure:"base_url" yaml:"base_url"`
	ModelName string `mapstructure:"model_name" yaml:"model_name"` // Specific model version
}
