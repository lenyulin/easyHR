package manager

import (
	"context"
	"easyHR/pkg/llm"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) SaveMessage(ctx context.Context, msg llm.Message) error {
	args := m.Called(ctx, msg)
	return args.Error(0)
}

func (m *MockRepository) GetMessages(ctx context.Context, sessionID string) ([]llm.Message, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).([]llm.Message), args.Error(1)
}

// MockAIModel
type MockAIModel struct {
	mock.Mock
}

func (m *MockAIModel) GenerateResponse(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	args := m.Called(ctx, messages)
	return args.Get(0).(*schema.Message), args.Error(1)
}

func (m *MockAIModel) StreamResponse(ctx context.Context, messages []*schema.Message, cb llm.StreamCallback) (string, error) {
	args := m.Called(ctx, messages, cb)
	return args.String(0), args.Error(1)
}

func (m *MockAIModel) GetModelType() string {
	args := m.Called()
	return args.String(0)
}

func TestLLMManager_GenerateResponse(t *testing.T) {
	repo := new(MockRepository)
	model := new(MockAIModel)
	mgr := NewLLMManager(repo)
	mgr.RegisterClient("test-model", model)

	ctx := context.Background()
	sessionID := "test-session"
	content := "hello"

	// Expect save user message
	repo.On("SaveMessage", ctx, mock.MatchedBy(func(msg llm.Message) bool {
		return msg.Role == "user" && msg.Content == content
	})).Return(nil)

	// Expect get history
	repo.On("GetMessages", ctx, sessionID).Return([]llm.Message{}, nil)

	// Expect generate response
	expectedResp := &schema.Message{Role: schema.Assistant, Content: "world"}
	model.On("GenerateResponse", ctx, mock.Anything).Return(expectedResp, nil)

	// Expect save assistant message
	repo.On("SaveMessage", ctx, mock.MatchedBy(func(msg llm.Message) bool {
		return msg.Role == "assistant" && msg.Content == "world"
	})).Return(nil)

	resp, err := mgr.GenerateResponse(ctx, "test-model", sessionID, content)
	assert.NoError(t, err)
	assert.Equal(t, "world", resp)

	repo.AssertExpectations(t)
	model.AssertExpectations(t)
}
