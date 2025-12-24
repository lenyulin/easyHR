package doubao

import (
	"context"
	"easyHR/internal/agent/llm"
	"os"

	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
)

type doubao struct {
	client    *arkruntime.Client
	modelName string
}

func (d *doubao) AnalyzeResume(ctx context.Context, role string, sysPrompt string, usrPrompt string, f string, opt ...interface{}) (*llm.ResumeAnalysis, error) {

}

var baseUrl = "https://ark.cn-beijing.volces.com/api/v3"

func NewDoubaoProvider(key string, modelName string) llm.LLMProvider {
	// 1. 初始化客户端
	// 请将 "YOUR_API_KEY" 替换为你的实际 API Key
	client := arkruntime.NewClientWithApiKey(
		// 从环境变量中获取您的 API Key。此为默认方式，您可根据需要进行修改
		os.Getenv(key),
		// 此为默认路径，您可根据业务所在地域进行配置
		arkruntime.WithBaseUrl(baseUrl),
	)
	return &doubao{
		client:    client,
		modelName: modelName,
	}
}
