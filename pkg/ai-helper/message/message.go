package message

import (
	"time"

	"easyHR/pkg/ai-helper"
)

// MessageProcessor 消息处理器，负责消息的转换、验证和历史管理
// 提供将消息转换为不同格式、验证消息有效性和管理消息历史的功能
// 是消息处理的核心组件，连接AIHelper和存储层

type MessageProcessor struct {}

// NewMessageProcessor 创建一个新的MessageProcessor实例
// 初始化消息处理器，返回实例
func NewMessageProcessor() *MessageProcessor {
	return &MessageProcessor{}
}

// ConvertToSchemaMessage 将内部Message转换为schema.Message格式
// 接收内部Message实例，返回schema.Message实例
func (p *MessageProcessor) ConvertToSchemaMessage(msg aihelper.Message) *SchemaMessage {
	return &SchemaMessage{
		Role:    msg.Role,
		Content: msg.Content,
	}
}

// ConvertToInternalMessage 将外部消息转换为内部Message格式
// 接收角色、内容和会话ID，返回内部Message实例
func (p *MessageProcessor) ConvertToInternalMessage(role, content, sessionID string) aihelper.Message {
	return aihelper.Message{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
}

// ValidateMessage 验证消息的有效性
// 检查消息角色是否合法，内容是否为空
// 返回验证结果和错误信息
func (p *MessageProcessor) ValidateMessage(msg aihelper.Message) (bool, error) {
	// 检查角色是否合法
	if msg.Role != "user" && msg.Role != "system" && msg.Role != "assistant" {
		return false, ErrInvalidRole
	}

	// 检查内容是否为空
	if msg.Content == "" {
		return false, ErrEmptyContent
	}

	// 检查会话ID是否为空
	if msg.SessionID == "" {
		return false, ErrEmptySessionID
	}

	return true, nil
}

// FilterMessagesByRole 根据角色过滤消息
// 接收消息列表和角色，返回过滤后的消息列表
func (p *MessageProcessor) FilterMessagesByRole(messages []aihelper.Message, role string) []aihelper.Message {
	var filtered []aihelper.Message
	for _, msg := range messages {
		if msg.Role == role {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// FilterMessagesByTimeRange 根据时间范围过滤消息
// 接收消息列表、开始时间和结束时间，返回过滤后的消息列表
func (p *MessageProcessor) FilterMessagesByTimeRange(messages []aihelper.Message, startTime, endTime time.Time) []aihelper.Message {
	var filtered []aihelper.Message
	for _, msg := range messages {
		if msg.CreatedAt.After(startTime) && msg.CreatedAt.Before(endTime) {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// GetLatestMessages 获取最近的N条消息
// 接收消息列表和数量，返回最近的N条消息
func (p *MessageProcessor) GetLatestMessages(messages []aihelper.Message, count int) []aihelper.Message {
	// 如果消息数量小于等于count，返回所有消息
	if len(messages) <= count {
		return messages
	}

	// 返回最近的count条消息
	return messages[len(messages)-count:]
}

// 自定义错误类型，用于消息验证
var (
	ErrInvalidRole       = Error("invalid role")
	ErrEmptyContent      = Error("empty content")
	ErrEmptySessionID    = Error("empty session ID")
)

// Error 自定义错误类型，实现error接口
// 用于消息处理中的错误处理

type Error string

// Error 实现error接口的Error方法
func (e Error) Error() string {
	return string(e)
}

// SchemaMessage 定义外部消息格式的别名
// 方便与外部系统交互
// 是对github.com/cloudwego/eino/schema.Message的封装
type SchemaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}