package search

import (
	"easyHR/pkg/resumescreen/types"
)

// Filter defines criteria for filtering candidates.
type Filter interface {
	Apply(candidates []*types.Candidate) []*types.Candidate
}

// BasicFilter implements basic filtering logic.
type BasicFilter struct {
	Criteria map[string]interface{}
}

func (f *BasicFilter) Apply(candidates []*types.Candidate) []*types.Candidate {
	// Implementation placeholder
	return candidates
}
