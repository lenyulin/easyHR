package manager

import (
	"bytes"
	"context"
	"easyHR/pkg/llm"
	"easyHR/pkg/llm/message"
	"easyHR/pkg/llm/storage"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cloudwego/eino/schema"
	"github.com/ledongthuc/pdf"
)

type LLMManager struct {
	msgManager *message.MessageManager
	clients    map[string]llm.AIModel
	mu         sync.RWMutex
}

func NewLLMManager(repo storage.Repository) *LLMManager {
	return &LLMManager{
		msgManager: message.NewMessageManager(repo),
		clients:    make(map[string]llm.AIModel),
	}
}

// RegisterClient 注册一个LLM客户端
func (m *LLMManager) RegisterClient(name string, client llm.AIModel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.clients[name] = client
}

// GetClient 获取指定名称的LLM客户端
func (m *LLMManager) GetClient(name string) (llm.AIModel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	client, ok := m.clients[name]
	if !ok {
		return nil, fmt.Errorf("client %s not found", name)
	}
	return client, nil
}

// GenerateResponse 处理生成请求：保存用户消息 -> 调用模型 -> 保存AI响应
func (m *LLMManager) GenerateResponse(ctx context.Context, modelName string, sessionID string, content string) (string, error) {
	client, err := m.GetClient(modelName)
	if err != nil {
		return "", err
	}

	// 1. 保存用户消息
	if _, err := m.msgManager.AddMessage(ctx, sessionID, "user", content); err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	// 2. 获取历史消息用于构建上下文 (这里简化处理，实际可能需要截断或筛选)
	history, err := m.msgManager.GetHistory(ctx, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get history: %w", err)
	}

	// 转换消息格式为 Eino schema
	var einoMessages []*schema.Message
	for _, msg := range history {
		role := schema.User
		if msg.Role == "assistant" {
			role = schema.Assistant
		} else if msg.Role == "system" {
			role = schema.System
		}
		einoMessages = append(einoMessages, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	// 3. 调用模型
	resp, err := client.GenerateResponse(ctx, einoMessages)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	// 4. 保存AI响应
	if _, err := m.msgManager.AddMessage(ctx, sessionID, "assistant", resp.Content); err != nil {
		return "", fmt.Errorf("failed to save assistant message: %w", err)
	}

	return resp.Content, nil
}

// ProcessFileAndGenerate 读取文件内容，结合prompt，生成回复
func (m *LLMManager) ProcessFileAndGenerate(ctx context.Context, modelName string, sessionID string, filePath string, prompt string) (string, error) {
	// 1. 读取文件内容 (这里简单实现，实际可能需要根据文件类型解析，如PDF/Word等)
	// 假设是文本文件
	content, err := readFileContent(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// 2. 构建完整Prompt
	fullContent := fmt.Sprintf("%s\n\nFile Content:\n%s", prompt, content)

	// 3. 调用GenerateResponse
	return m.GenerateResponse(ctx, modelName, sessionID, fullContent)
}

func readFileContent(path string) (string, error) {
	ext := filepath.Ext(path)
	switch strings.ToLower(ext) {
	case ".pdf":
		return readPDF(path)
	default:
		return readTextFile(path)
	}
}

func readTextFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func readPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	totalPage := r.NumPage()

	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			return "", err
		}
		buf.WriteString(text)
	}
	return buf.String(), nil
}
