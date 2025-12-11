package llm

import "context"

type LLMService interface {
	Chat(ctx context.Context, modelName string, sessionID string, content string) (string, error)
	ProcessFileAndGenerate(ctx context.Context, modelName string, sessionID string, filePath string, prompt string) (string, error)
}
