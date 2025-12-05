package client

import (
	"bytes"
	"context"
	"easyHR/pkg/llm"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudwego/eino/schema"
)

type GenericClient struct {
	config llm.Configuration
	client *http.Client
}

func NewGenericClient(config llm.Configuration) *GenericClient {
	return &GenericClient{
		config: config,
		client: &http.Client{},
	}
}

func (c *GenericClient) GetModelType() string {
	return c.config.ModelType
}

type openAIRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}

func (c *GenericClient) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	reqMessages := make([]message, len(messages))
	for i, m := range messages {
		role := "user"
		if m.Role == schema.System {
			role = "system"
		} else if m.Role == schema.Assistant {
			role = "assistant"
		}
		reqMessages[i] = message{
			Role:    role,
			Content: m.Content,
		}
	}

	reqBody := openAIRequest{
		Model:    c.config.ModelName,
		Messages: reqMessages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, err
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &schema.Message{
		Role:    schema.Assistant,
		Content: openAIResp.Choices[0].Message.Content,
	}, nil
}

func (c *GenericClient) StreamResponse(ctx context.Context, messages []*schema.Message, cb llm.StreamCallback) (string, error) {
	// TODO: Implement streaming
	return "", fmt.Errorf("streaming not implemented")
}
