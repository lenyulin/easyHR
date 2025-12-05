package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"easyHR/pkg/ai-helper"
	"github.com/cloudwego/eino/schema"
)

// GPT4Model 实现了AIModel接口的GPT-4模型
// 注意：这是一个简化实现，实际使用时需要根据OpenAI API文档进行调整
// 当前使用的是OpenAI GPT-4 API的简化版本
// 注意：实际部署时，需要将硬编码的API地址替换为实际的OpenAI API地址
// 并确保API密钥的安全管理

// GPT4Config GPT-4模型的配置
// 包含API密钥、API基础URL和模型名称
// API密钥用于身份验证，基础URL用于替换默认的OpenAI API地址
// 模型名称用于指定使用的GPT-4模型版本
// 注意：实际使用时，需要根据OpenAI API文档调整配置项

type GPT4Config struct {
	APIKey    string
	BaseURL   string
	ModelName string
}

// GPT4Model GPT-4模型的实现
// 包含配置信息和HTTP客户端
// HTTP客户端用于发送请求到OpenAI API

type GPT4Model struct {
	config GPT4Config
	client *http.Client
}

// NewGPT4Model 创建一个新的GPT4Model实例
// 接收配置信息，返回GPT4Model实例
// 如果配置中没有指定基础URL，使用默认的OpenAI API地址
// 如果配置中没有指定模型名称，使用默认的gpt-4模型
// 注意：实际使用时，需要根据OpenAI API文档调整默认值

func NewGPT4Model(config GPT4Config) *GPT4Model {
	// 设置默认值
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.ModelName == "" {
		config.ModelName = "gpt-4"
	}

	return &GPT4Model{
		config: config,
		client: &http.Client{},
	}
}

// GenerateResponse 生成非流式响应
// 接收上下文和消息列表，返回生成的消息和错误
// 将消息转换为GPT API的请求格式，发送HTTP请求，然后解析响应
// 注意：实际使用时，需要根据OpenAI API文档调整请求和响应格式

func (m *GPT4Model) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	// 转换消息格式
	var gptMessages []map[string]interface{}
	for _, msg := range messages {
		gptMessages = append(gptMessages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// 构建请求体
	reqBody := map[string]interface{}{
		"model":     m.config.ModelName,
		"messages":  gptMessages,
		"max_tokens": 2048,
	}

	// 转换为JSON
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 构建请求
	url := fmt.Sprintf("%s/chat/completions", m.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))

	// 发送请求
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析响应
	var gptResp struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &gptResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 检查是否有响应
	if len(gptResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from GPT-4")
	}

	// 返回生成的消息
	return &schema.Message{
		Role:    schema.RoleType(gptResp.Choices[0].Message.Role),
		Content: gptResp.Choices[0].Message.Content,
	}, nil
}

// StreamResponse 生成流式响应
// 接收上下文、消息列表和回调函数，返回最终内容和错误
// 将消息转换为GPT API的请求格式，发送HTTP请求，处理流式响应
// 调用回调函数返回每个chunk的内容，最后返回完整内容
// 注意：实际使用时，需要根据OpenAI API文档调整请求和响应格式

func (m *GPT4Model) StreamResponse(ctx context.Context, messages []*schema.Message, cb aihelper.StreamCallback) (string, error) {
	// 转换消息格式
	var gptMessages []map[string]interface{}
	for _, msg := range messages {
		gptMessages = append(gptMessages, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// 构建请求体
	reqBody := map[string]interface{}{
		"model":     m.config.ModelName,
		"messages":  gptMessages,
		"max_tokens": 2048,
		"stream":    true,
	}

	// 转换为JSON
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 构建请求
	url := fmt.Sprintf("%s/chat/completions", m.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.config.APIKey))

	// 发送请求
	resp, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// 处理流式响应
	var finalContent string
	decoder := json.NewDecoder(resp.Body)

	for {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return finalContent, fmt.Errorf("failed to decode stream chunk: %w", err)
		}

		// 解析chunk
		if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok && content != "" {
						// 调用回调函数
						cb(content)
						// 拼接最终内容
						finalContent += content
					}
				}
			}
		}
	}

	return finalContent, nil
}

// GetModelType 返回模型类型
// 返回"gpt4"作为模型类型

func (m *GPT4Model) GetModelType() string {
	return "gpt4"
}