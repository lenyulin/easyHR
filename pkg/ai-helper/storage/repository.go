package storage

import (
	"context"

	"easyHR/pkg/ai-helper"
)

// MessageRepository 定义消息存储的接口
// 提供保存消息、获取消息和删除消息的方法
// 是存储层的抽象，支持不同的存储实现

type MessageRepository interface {
	// SaveMessage 保存消息到存储
	// 接收上下文和消息，返回保存后的消息和错误
	SaveMessage(ctx context.Context, msg aihelper.Message) (aihelper.Message, error)

	// GetMessagesBySessionID 根据会话ID获取消息列表
	// 接收上下文和会话ID，返回消息列表和错误
	GetMessagesBySessionID(ctx context.Context, sessionID string) ([]aihelper.Message, error)

	// DeleteMessage 删除消息
	// 接收上下文和消息ID，返回删除结果和错误
	DeleteMessage(ctx context.Context, id uint) (bool, error)

	// DeleteMessagesBySessionID 根据会话ID删除所有消息
	// 接收上下文和会话ID，返回删除结果和错误
	DeleteMessagesBySessionID(ctx context.Context, sessionID string) (bool, error)

	// GetMessageCount 获取消息总数
	// 接收上下文，返回消息总数和错误
	GetMessageCount(ctx context.Context) (int64, error)

	// Close 关闭存储连接
	// 接收上下文，返回关闭结果和错误
	Close(ctx context.Context) error
}

// StorageType 定义存储类型的枚举
// 支持MongoDB、MySQL等不同存储类型
// 用于创建不同的存储实例
type StorageType int

const (
	// StorageTypeMongoDB MongoDB存储类型
	StorageTypeMongoDB StorageType = iota

	// StorageTypeMySQL MySQL存储类型
	StorageTypeMySQL

	// StorageTypeMemory 内存存储类型（用于测试）
	StorageTypeMemory
)

// NewMessageRepository 创建一个新的MessageRepository实例
// 接收存储类型和配置，返回MessageRepository实例和错误
// 目前只实现了MongoDB存储，可扩展支持其他存储类型
func NewMessageRepository(storageType StorageType, config map[string]interface{}) (MessageRepository, error) {
	switch storageType {
	case StorageTypeMongoDB:
		// 创建MongoDB存储实例
		return NewMongoDBRepository(config)
	case StorageTypeMySQL:
		// 目前未实现MySQL存储，可扩展
		return nil, ErrStorageTypeNotSupported
	case StorageTypeMemory:
		// 目前未实现内存存储，可扩展
		return nil, ErrStorageTypeNotSupported
	default:
		return nil, ErrStorageTypeNotSupported
	}
}

// 自定义错误类型，用于存储操作
var (
	ErrStorageTypeNotSupported = Error("storage type not supported")
	ErrMessageNotFound         = Error("message not found")
	ErrFailedToSaveMessage     = Error("failed to save message")
	ErrFailedToGetMessages     = Error("failed to get messages")
	ErrFailedToDeleteMessage   = Error("failed to delete message")
)

// Error 自定义错误类型，实现error接口
// 用于存储操作中的错误处理

type Error string

// Error 实现error接口的Error方法
func (e Error) Error() string {
	return string(e)
}