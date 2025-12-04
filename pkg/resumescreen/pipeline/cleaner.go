package pipeline

import (
	"context"
	"easyHR/pkg/resumescreen/types"
)

// Cleaner defines the interface for cleaning and normalizing resume data.
type Cleaner interface {
	Clean(ctx context.Context, data *types.StructuredResume) (*types.StructuredResume, error)
}

// BasicCleaner is a simple implementation of Cleaner.
type BasicCleaner struct{}

func NewBasicCleaner() *BasicCleaner {
	return &BasicCleaner{}
}

func (c *BasicCleaner) Clean(ctx context.Context, data *types.StructuredResume) (*types.StructuredResume, error) {
	// Implementation placeholder: e.g., normalize dates, trim whitespace
	return data, nil
}
