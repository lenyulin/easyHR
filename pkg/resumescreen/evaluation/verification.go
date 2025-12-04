package evaluation

import (
	"context"
	"easyHR/pkg/resumescreen/types"
)

// Verifier defines the interface for verifying candidate information.
type Verifier interface {
	Verify(ctx context.Context, candidate *types.Candidate) (bool, error)
}

// OnlineVerifier verifies information using online sources.
type OnlineVerifier struct{}

func (v *OnlineVerifier) Verify(ctx context.Context, candidate *types.Candidate) (bool, error) {
	// Implementation placeholder
	return true, nil
}
