package agent

import (
	"context"
	"easyHR/internal/agent/llm"
	"easyHR/internal/agent/llm/gemini"

	"github.com/cloudwego/eino/schema"
)

// AIModel 定义AIModel的通用接口
type AIModel interface {
	GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error)
	GetModelType() string
}

// Configuration 定义LLM服务的配置
type Configuration struct {
	ModelType string `mapstructure:"model_type" yaml:"model_type"` // e.g., "doubao", "gpt4"
	ModelName string `mapstructure:"model_name" yaml:"model_name"` // e.g., "gemini-1.5-flash"
	APIKey    string `mapstructure:"api_key" yaml:"api_key"`
	BaseURL   string `mapstructure:"base_url" yaml:"base_url"`
}
type model struct {
	provider  llm.LLMProvider
	modelType string
	modelName string
}

func NewAiModel(config Configuration) AIModel {
	var provider llm.LLMProvider
	switch config.ModelType {
	case "gemini":
		provider = gemini.NewGeminiProvider(config.APIKey, config.ModelName)
	default:
		provider = gemini.NewGeminiProvider(config.APIKey, config.ModelName)
	}
	return &model{
		provider:  provider,
		modelType: config.ModelType,
		modelName: config.ModelName,
	}
}

// GenerateResponse 生成AI响应，非流式
func (a *model) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	//拼接系统prompt和用户消息
	var prompt string
	for _, msg := range messages {
		if msg.Role == "system" {
			prompt += msg.Content + "\n"
		}
	}
	return nil, nil
}

func (a *model) GetModelType() string {
	return a.modelType
}

func (a *model) GetModelName() string {
	return a.modelName
}
