package agent

import (
	"context"
	"time"

	"github.com/cloudwego/eino/schema"
)

// StreamCallback 定义流式响应的回调函数
type StreamCallback func(msg string)

// AIModel 定义AIModel的通用接口
type AIModel interface {
	GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
	StreamResponse(ctx context.Context, messages []*schema.Message, cb StreamCallback) (string, error)
	GetModelType() string
}

// Configuration 定义LLM服务的配置
type Configuration struct {
	ModelType string `mapstructure:"model_type" yaml:"model_type"` // e.g., "doubao", "gpt4"
	APIKey    string `mapstructure:"api_key" yaml:"api_key"`
	BaseURL   string `mapstructure:"base_url" yaml:"base_url"`
	ModelName string `mapstructure:"model_name" yaml:"model_name"` // Specific model version
}
type model struct {
}

// GenerateResponse 生成AI响应，非流式
func (a *model) GenerateResponse(ctx context.Context, userMsg string, save bool) (*Message, error) {
	// 创建用户消息
	userMessage := Message{
		SessionID: a.SessionID,
		Role:      "user",
		Content:   userMsg,
		CreatedAt: time.Now(),
	}

	// 添加用户消息到历史
	if _, err := a.AddMessage(userMessage, save); err != nil {
		return nil, err
	}

	// 转换为schema.Message格式
	var schemaMsgs []*schema.Message

	// 1. 添加系统提示词
	if a.SysMsg != "" {
		schemaMsgs = append(schemaMsgs, &schema.Message{
			Role:    schema.RoleType("system"),
			Content: a.SysMsg,
		})
	}

	// 2. 添加历史消息（排除系统提示词，避免重复）
	a.mu.RLock()
	for _, msg := range a.messages {
		// 跳过系统消息，因为我们已经单独添加了
		if msg.Role != "system" {
			schemaMsgs = append(schemaMsgs, &schema.Message{
				Role:    schema.RoleType(msg.Role),
				Content: msg.Content,
			})
		}
	}
	a.mu.RUnlock()

	// 调用模型生成响应
	aiResp, err := a.model.GenerateResponse(ctx, schemaMsgs)
	if err != nil {
		return nil, err
	}

	// 创建AI响应消息
	aiMessage := Message{
		SessionID: a.SessionID,
		Role:      "assistant",
		Content:   aiResp.Content,
		CreatedAt: time.Now(),
	}

	// 添加AI响应到历史
	if _, err := a.AddMessage(aiMessage, save); err != nil {
		return nil, err
	}

	return &aiMessage, nil
}
