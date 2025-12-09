package storage

import (
	"context"
	"easyHR/pkg/resumescreen/types"
)

// ResumeRepository defines the interface for storing and retrieving resumes.
type ResumeRepository interface {
	Save(ctx context.Context, candidate *types.Candidate) error
	Get(ctx context.Context, id string) (*types.Candidate, error)
	// Add other methods as needed
}

// VectorRepository defines the interface for vector operations.
type VectorRepository interface {
	Index(ctx context.Context, candidate *types.Candidate) error
	Search(ctx context.Context, queryVector []float32, limit int) ([]*types.SearchResult, error)
}
