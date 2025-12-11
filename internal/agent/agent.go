package agent

import (
	"context"
	"time"

	"github.com/cloudwego/eino/schema"
)

// Agent 封装了与Agent交互的核心逻辑
type Agent struct {
	model     AIModel // 审模型列表
	role      string  // 审模型角色
	SysMsg    string  // 系统默认提示词，用于指导Agent的行为
	SessionID string  // 会话ID
}

// NewAgent NewAgent 创建一个新的Agent实例
func NewAgent(model AIModel, sessionID string, role string, sysMsg string) *Agent {
	return &Agent{
		model:     model,
		SysMsg:    sysMsg, // 设置默认系统提示词
		SessionID: sessionID,
		role:      role,
	}
}

// AddTask 处理新的用户请求
func (a *Agent) AddTask(ctx context.Context, filePath string) (*Message, error) {
	// 构造Review的输入
	// 使用 a.SysMsg 作为 System prompt
	schemaMsgs := []*schema.Message{
		{Role: schema.System, Content: a.SysMsg},
		{Role: schema.User, Content: filePath},
	}

	// 调用模型生成响应
	resp, err := a.model.GenerateResponse(ctx, schemaMsgs)
	if err != nil {
		return nil, err
	}

	// 创建AI响应消息
	aiMessage := Message{
		SessionID: a.SessionID,
		Role:      a.role,
		Input:     filePath,
		Content:   resp.Content,
		CreatedAt: time.Now(),
	}
	return &aiMessage, nil
}
