package search

import (
	"context"
	"easyHR/pkg/resumescreen/storage"
	"easyHR/pkg/resumescreen/types"
)

// Service handles search operations.
type Service struct {
	repo         storage.ResumeRepository
	vectorRepo   storage.VectorRepository
	queryBuilder QueryBuilder
	ranker       Ranker
}

func NewService(repo storage.ResumeRepository, vectorRepo storage.VectorRepository, qb QueryBuilder, ranker Ranker) *Service {
	return &Service{
		repo:         repo,
		vectorRepo:   vectorRepo,
		queryBuilder: qb,
		ranker:       ranker,
	}
}

func (s *Service) Search(ctx context.Context, nlQuery string) ([]*types.SearchResult, error) {
	// 1. Build Query
	query, err := s.queryBuilder.Build(ctx, nlQuery)
	if err != nil {
		return nil, err
	}

	// 2. Vector Search
	results, err := s.vectorRepo.Search(ctx, query.Vector, 50) // Default limit
	if err != nil {
		return nil, err
	}

	// 3. Rank
	rankedResults, err := s.ranker.Rank(ctx, results, query)
	if err != nil {
		return nil, err
	}

	return rankedResults, nil
}
