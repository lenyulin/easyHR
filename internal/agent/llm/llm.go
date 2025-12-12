package llm

import (
	"context"
)

// LLMProvider 是所有 AI 模型的通用接口
type LLMProvider interface {
	// AnalyzeResume 接收文件流，返回结构化的 ResumeAnalysis
	AnalyzeResume(ctx context.Context, role string, sysPrompt string, usrPrompt string, file string, opt ...interface{}) (*ResumeAnalysis, error)
}
