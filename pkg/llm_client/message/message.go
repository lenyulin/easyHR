package message

import (
	"context"
	"easyHR/pkg/llm"
	"easyHR/pkg/llm/storage"
	"sync"
	"time"
)

type MessageManager struct {
	repo storage.Repository
	mu   sync.RWMutex
}

func NewMessageManager(repo storage.Repository) *MessageManager {
	return &MessageManager{
		repo: repo,
	}
}

func (m *MessageManager) AddMessage(ctx context.Context, sessionID string, role string, content string) (llm.Message, error) {
	msg := llm.Message{
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}

	if err := m.repo.SaveMessage(ctx, msg); err != nil {
		return llm.Message{}, err
	}

	return msg, nil
}

func (m *MessageManager) GetHistory(ctx context.Context, sessionID string) ([]llm.Message, error) {
	return m.repo.GetMessages(ctx, sessionID)
}
