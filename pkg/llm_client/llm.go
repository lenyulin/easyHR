package llm

import (
	"context"
)

// LLMService 定义对外暴露的服务接口
type LLMService interface {
	// Chat 发送聊天消息并获取响应
	Chat(ctx context.Context, modelName string, sessionID string, content string) (string, error)
	// ProcessFileAndGenerate 读取文件并生成响应
	ProcessFileAndGenerate(ctx context.Context, modelName string, sessionID string, filePath string, prompt string) (string, error)
	// StreamChat 流式发送聊天消息 (暂未实现完全，接口预留)
	StreamChat(ctx context.Context, modelName string, sessionID string, content string, cb StreamCallback) (string, error)
}

// Ensure LLMManager implements LLMService (will be checked in main or wire)
// Note: LLMManager struct is in subpackage, so we might need an adapter or move LLMManager here if we want to expose it directly.
// For now, let's keep the interface here and implementation in manager.
