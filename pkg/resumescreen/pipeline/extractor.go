package pipeline

import (
	"context"
	"easyHR/pkg/resumescreen/types"
)

// Extractor defines the interface for extracting structured data from text.
type Extractor interface {
	Extract(ctx context.Context, text string) (*types.StructuredResume, error)
}

// LLMExtractor is an implementation of Extractor using an LLM.
type LLMExtractor struct {
	// LLM client or service would go here
}

func NewLLMExtractor() *LLMExtractor {
	return &LLMExtractor{}
}

func (e *LLMExtractor) Extract(ctx context.Context, text string) (*types.StructuredResume, error) {
	// Implementation placeholder
	return &types.StructuredResume{}, nil
}
