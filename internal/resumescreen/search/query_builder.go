package search

import "context"

// Query represents a parsed search query.
type Query struct {
	SQL         string
	Vector      []float32
	Keywords    []string
	Constraints map[string]interface{}
}

// QueryBuilder converts natural language to a structured Query.
type QueryBuilder interface {
	Build(ctx context.Context, nlQuery string) (*Query, error)
}

// LLMQueryBuilder uses an LLM to build queries.
type LLMQueryBuilder struct {
	// LLM client
}

func NewLLMQueryBuilder() *LLMQueryBuilder {
	return &LLMQueryBuilder{}
}

func (qb *LLMQueryBuilder) Build(ctx context.Context, nlQuery string) (*Query, error) {
	// Implementation placeholder
	return &Query{}, nil
}
