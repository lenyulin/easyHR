package evaluation

import "easyHR/pkg/resumescreen/types"

// Rubric defines the scoring standards.
type Rubric struct {
	Weights map[string]float64
}

// Scorer calculates a score for a candidate based on a rubric.
type Scorer interface {
	Score(candidate *types.Candidate) float64
}

type DefaultScorer struct {
	Rubric Rubric
}

func (s *DefaultScorer) Score(candidate *types.Candidate) float64 {
	// Implementation placeholder
	return 0.0
}
