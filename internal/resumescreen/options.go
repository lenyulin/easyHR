package resumescreen

// Option defines a functional option for configuring the Engine.
type Option func(*EngineConfig)

// EngineConfig holds configuration for the ResumeScreen Engine.
type EngineConfig struct {
	MatchThreshold float64
	MaxResults     int
	// Add other configuration fields as needed
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *EngineConfig {
	return &EngineConfig{
		MatchThreshold: 0.7,
		MaxResults:     10,
	}
}

// WithMatchThreshold sets the minimum score threshold for matches.
func WithMatchThreshold(threshold float64) Option {
	return func(c *EngineConfig) {
		c.MatchThreshold = threshold
	}
}

// WithMaxResults sets the maximum number of results to return.
func WithMaxResults(max int) Option {
	return func(c *EngineConfig) {
		c.MaxResults = max
	}
}
