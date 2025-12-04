package cv

import (
	"context"
	"easyHR/pkg/logger"
	"time"
)

// Storage defines the interface for persisting data.
type Storage interface {
	Insert(ctx context.Context, collection string, doc interface{}) error
}

// Parser defines the interface for parsing CV documents.
type Parser interface {
	Parse(filePath string) (string, error)
}

// Service handles the processing of CV files.
type Service struct {
	parser     Parser
	storage    Storage
	collection string
	log        logger.LoggerV1
}

// NewService creates a new CV processing service.
func NewService(storage Storage, collection string, parser Parser, log logger.LoggerV1) *Service {
	return &Service{
		parser:     parser,
		storage:    storage,
		collection: collection,
		log:        log,
	}
}

// Run starts the service to process files from the channel.
// onIdle is called when the channel is empty.
func (s *Service) Run(ch <-chan string, onIdle func()) {
	go func() {
		for {
			select {
			case filePath, ok := <-ch:
				if !ok {
					return
				}
				s.log.Info("CV Service received file: " + filePath)
				if err := s.processFile(filePath); err != nil {
					s.log.Error("Failed to process file " + filePath + ": " + err.Error())
				} else {
					s.log.Info("Successfully processed file: " + filePath)
				}
			default:
				// Channel is empty
				if onIdle != nil {
					onIdle()
				}
				// Sleep briefly to avoid busy loop if channel remains empty
				time.Sleep(100 * time.Millisecond)

				// Check channel again (blocking wait to resume normal processing)
				select {
				case filePath, ok := <-ch:
					if !ok {
						return
					}
					s.log.Info("CV Service received file: " + filePath)
					if err := s.processFile(filePath); err != nil {
						s.log.Error("Failed to process file " + filePath + ": " + err.Error())
					} else {
						s.log.Info("Successfully processed file: " + filePath)
					}
				case <-time.After(time.Second):
					// Continue loop to trigger onIdle again if needed
				}
			}
		}
	}()
}

func (s *Service) processFile(filePath string) error {
	// Parse PDF
	content, err := s.parser.Parse(filePath)
	if err != nil {
		return err
	}

	// Create CV model
	cv := &CV{
		FilePath:  filePath,
		Content:   content,
		ParsedAt:  time.Now(),
		CreatedAt: time.Now(),
	}

	// Save to Store
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.storage.Insert(ctx, s.collection, cv)
}
