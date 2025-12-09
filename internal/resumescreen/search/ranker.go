package search

import (
	"context"
	"easyHR/pkg/resumescreen/types"
)

// Ranker defines the interface for ranking search results.
type Ranker interface {
	Rank(ctx context.Context, results []*types.SearchResult, query *Query) ([]*types.SearchResult, error)
}

// HybridRanker combines vector scores and other metrics.
type HybridRanker struct{}

func NewHybridRanker() *HybridRanker {
	return &HybridRanker{}
}

func (r *HybridRanker) Rank(ctx context.Context, results []*types.SearchResult, query *Query) ([]*types.SearchResult, error) {
	// Implementation placeholder
	return results, nil
}
