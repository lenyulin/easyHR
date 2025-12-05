package storage

import (
	"context"
	"easyHR/pkg/llm"
)

// Repository 定义消息存储接口
type Repository interface {
	SaveMessage(ctx context.Context, msg llm.Message) error
	GetMessages(ctx context.Context, sessionID string) ([]llm.Message, error)
}
