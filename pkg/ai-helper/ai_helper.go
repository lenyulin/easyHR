package aihelper

import (
	"context"
	"os"
	"time"

	"github.com/cloudwego/eino/schema"
)

// NewAIHelper 创建一个新的AIHelper实例
func NewAIHelper(model AIModel, sessionID string, saveFunc func(Message) (Message, error)) *AIHelper {
	// 如果saveFunc为nil，使用默认实现
	if saveFunc == nil {
		// 默认实现：返回消息本身，不做持久化
		saveFunc = func(msg Message) (Message, error) {
			return msg, nil
		}
	}

	return &AIHelper{
		model:     model,
		messages:  []Message{},
		SessionID: sessionID,
		saveFunc:  saveFunc,
		SysMsg:    DefaultSystemPrompt, // 设置默认系统提示词
	}
}

// AddMessage 添加消息到历史记录，支持持久化
func (h *AIHelper) AddMessage(msg Message, save bool) (Message, error) {
	// 设置消息的创建时间和会话ID
	msg.CreatedAt = time.Now()
	if msg.SessionID == "" {
		msg.SessionID = h.SessionID
	}

	// 添加到内存中的消息历史
	h.mu.Lock()
	h.messages = append(h.messages, msg)
	h.mu.Unlock()

	// 如果需要持久化，调用saveFunc
	if save {
		return h.saveFunc(msg)
	}

	return msg, nil
}

// GenerateResponse 生成AI响应，非流式
func (h *AIHelper) GenerateResponse(ctx context.Context, userMsg string, save bool) (*Message, error) {
	// 创建用户消息
	userMessage := Message{
		SessionID: h.SessionID,
		Role:      "user",
		Content:   userMsg,
		CreatedAt: time.Now(),
	}

	// 添加用户消息到历史
	if _, err := h.AddMessage(userMessage, save); err != nil {
		return nil, err
	}

	// 转换为schema.Message格式
	var schemaMsgs []*schema.Message

	// 1. 添加系统提示词
	if h.SysMsg != "" {
		schemaMsgs = append(schemaMsgs, &schema.Message{
			Role:    schema.RoleType("system"),
			Content: h.SysMsg,
		})
	}

	// 2. 添加历史消息（排除系统提示词，避免重复）
	h.mu.RLock()
	for _, msg := range h.messages {
		// 跳过系统消息，因为我们已经单独添加了
		if msg.Role != "system" {
			schemaMsgs = append(schemaMsgs, &schema.Message{
				Role:    schema.RoleType(msg.Role),
				Content: msg.Content,
			})
		}
	}
	h.mu.RUnlock()

	// 调用模型生成响应
	aiResp, err := h.model.GenerateResponse(ctx, schemaMsgs)
	if err != nil {
		return nil, err
	}

	// 创建AI响应消息
	aiMessage := Message{
		SessionID: h.SessionID,
		Role:      "assistant",
		Content:   aiResp.Content,
		CreatedAt: time.Now(),
	}

	// 添加AI响应到历史
	if _, err := h.AddMessage(aiMessage, save); err != nil {
		return nil, err
	}

	return &aiMessage, nil
}

// StreamResponse 生成AI响应，流式
func (h *AIHelper) StreamResponse(ctx context.Context, userMsg string, save bool, cb StreamCallback) (*Message, error) {
	// 创建用户消息
	userMessage := Message{
		SessionID: h.SessionID,
		Role:      "user",
		Content:   userMsg,
		CreatedAt: time.Now(),
	}

	// 添加用户消息到历史
	if _, err := h.AddMessage(userMessage, save); err != nil {
		return nil, err
	}

	// 转换为schema.Message格式
	var schemaMsgs []*schema.Message

	// 1. 添加系统提示词
	if h.SysMsg != "" {
		schemaMsgs = append(schemaMsgs, &schema.Message{
			Role:    schema.RoleType("system"),
			Content: h.SysMsg,
		})
	}

	// 2. 添加历史消息（排除系统提示词，避免重复）
	h.mu.RLock()
	for _, msg := range h.messages {
		// 跳过系统消息，因为我们已经单独添加了
		if msg.Role != "system" {
			schemaMsgs = append(schemaMsgs, &schema.Message{
				Role:    schema.RoleType(msg.Role),
				Content: msg.Content,
			})
		}
	}
	h.mu.RUnlock()

	// 调用模型生成流式响应
	finalContent, err := h.model.StreamResponse(ctx, schemaMsgs, cb)
	if err != nil {
		return nil, err
	}

	// 创建AI响应消息
	aiMessage := Message{
		SessionID: h.SessionID,
		Role:      "assistant",
		Content:   finalContent,
		CreatedAt: time.Now(),
	}

	// 添加AI响应到历史
	if _, err := h.AddMessage(aiMessage, save); err != nil {
		return nil, err
	}

	return &aiMessage, nil
}

// GetMessages 获取当前会话的所有消息历史
func (h *AIHelper) GetMessages() []Message {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 返回消息的副本，避免外部修改
	msgs := make([]Message, len(h.messages))
	copy(msgs, h.messages)
	return msgs
}

// SetSaveFunc 设置消息存储回调函数
func (h *AIHelper) SetSaveFunc(saveFunc func(Message) (Message, error)) {
	if saveFunc != nil {
		h.saveFunc = saveFunc
	}
}

// SetSysMsg 设置系统提示词
func (h *AIHelper) SetSysMsg(sysMsg string) {
	h.mu.Lock()
	h.SysMsg = sysMsg
	h.mu.Unlock()
}

// GetSysMsg 获取当前系统提示词
func (h *AIHelper) GetSysMsg() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.SysMsg
}

// ClearMessages 清空消息历史
func (h *AIHelper) ClearMessages() {
	h.mu.Lock()
	h.messages = []Message{}
	h.mu.Unlock()
}

// ReadFile 读取文件内容
func ReadFile(filePath string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", err
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// GenerateResponseWithFile 结合文件内容生成AI响应，非流式
func (h *AIHelper) GenerateResponseWithFile(ctx context.Context, userMsg, filePath string, save bool) (*Message, error) {
	// 读取文件内容
	fileContent, err := ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 构建带有文件内容的完整提示词
	fullPrompt := "文件路径：" + filePath + "\n文件内容：" + fileContent + "\n\n" + userMsg

	// 调用普通生成方法
	return h.GenerateResponse(ctx, fullPrompt, save)
}

// StreamResponseWithFile 结合文件内容生成AI响应，流式
func (h *AIHelper) StreamResponseWithFile(ctx context.Context, userMsg, filePath string, save bool, cb StreamCallback) (*Message, error) {
	// 读取文件内容
	fileContent, err := ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 构建带有文件内容的完整提示词
	fullPrompt := "文件路径：" + filePath + "\n文件内容：" + fileContent + "\n\n" + userMsg

	// 调用流生成方法
	return h.StreamResponse(ctx, fullPrompt, save, cb)
}
