package resumescreen

import (
	"context"
	"easyHR/pkg/resumescreen/evaluation"
	"easyHR/pkg/resumescreen/pipeline"
	"easyHR/pkg/resumescreen/search"
	"easyHR/pkg/resumescreen/storage"
	"easyHR/pkg/resumescreen/types"
	"io"
)

// Re-export types from the types package for public API consistency
type (
	Candidate        = types.Candidate
	ResumeData       = types.ResumeData
	StructuredResume = types.StructuredResume
	SearchResult     = types.SearchResult
	Education        = types.Education
	Experience       = types.Experience
	Project          = types.Project
)

// Engine is the main interface for the ResumeScreen module.
type Engine interface {
	// Ingest processes a resume file and stores it.
	Ingest(ctx context.Context, input io.Reader, filename string) (*Candidate, error)

	// Search finds candidates matching a natural language query.
	Search(ctx context.Context, query string) ([]*SearchResult, error)

	// Evaluate scores a candidate.
	Evaluate(ctx context.Context, candidateID string) (float64, error)
}

// DefaultEngine is the default implementation of Engine.
type DefaultEngine struct {
	config    *EngineConfig
	processor *pipeline.Processor
	search    *search.Service
	repo      storage.ResumeRepository
	scorer    evaluation.Scorer
}

// NewEngine creates a new ResumeScreen Engine.
func NewEngine(
	repo storage.ResumeRepository,
	vectorRepo storage.VectorRepository,
	ocr pipeline.OCRAdapter,
	llmClient interface{}, // Placeholder for actual LLM client type
	opts ...Option,
) Engine {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Initialize subsystems
	// Note: In a real implementation, we would pass the LLM client to these constructors
	extractor := pipeline.NewLLMExtractor()
	cleaner := pipeline.NewBasicCleaner()
	processor := pipeline.NewProcessor(ocr, extractor, cleaner)

	qb := search.NewLLMQueryBuilder()
	ranker := search.NewHybridRanker()
	searchService := search.NewService(repo, vectorRepo, qb, ranker)

	scorer := &evaluation.DefaultScorer{}

	return &DefaultEngine{
		config:    cfg,
		processor: processor,
		search:    searchService,
		repo:      repo,
		scorer:    scorer,
	}
}

func (e *DefaultEngine) Ingest(ctx context.Context, input io.Reader, filename string) (*Candidate, error) {
	candidate, err := e.processor.Process(ctx, input, filename)
	if err != nil {
		return nil, err
	}

	if err := e.repo.Save(ctx, candidate); err != nil {
		return nil, err
	}

	// Indexing would happen here or asynchronously
	return candidate, nil
}

func (e *DefaultEngine) Search(ctx context.Context, query string) ([]*SearchResult, error) {
	return e.search.Search(ctx, query)
}

func (e *DefaultEngine) Evaluate(ctx context.Context, candidateID string) (float64, error) {
	candidate, err := e.repo.Get(ctx, candidateID)
	if err != nil {
		return 0, err
	}
	return e.scorer.Score(candidate), nil
}
