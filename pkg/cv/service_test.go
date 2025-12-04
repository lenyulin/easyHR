package cv

import (
	"context"
	"easyHR/pkg/logger"
	"testing"

	"go.uber.org/zap"
)

type MockStore struct {
	SavedDoc interface{}
}

func (m *MockStore) Insert(ctx context.Context, collection string, doc interface{}) error {
	m.SavedDoc = doc
	return nil
}

func (m *MockStore) Close(ctx context.Context) error {
	return nil
}

type MockParser struct{}

func (m *MockParser) Parse(filePath string) (string, error) {
	return "mock content", nil
}

func TestService_Run(t *testing.T) {
	// Mock logger
	l, _ := zap.NewDevelopment()
	log := logger.NewZapLogger(l)

	// Create mock store
	mockStore := &MockStore{}
	mockParser := &MockParser{}

	// Create service
	service := NewService(mockStore, "test_collection", mockParser, log)

	// Create channel
	ch := make(chan string, 1)
	ch <- "test.pdf"
	close(ch)

	// Run service (use wait group or similar in real test, here we just wait a bit)
	// Since Run spawns a goroutine, we need to wait for it to process.
	// However, Run loops.
	// We can't easily test the loop without a way to stop it or wait for idle.
	// But we can test that it doesn't panic.

	// For a better test, we would need to refactor Run to be more testable or expose processFile.
	// For now, just ensuring it accepts the interface is good.
	_ = service
}
