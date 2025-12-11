package config

type AgentConfig struct {
	Agents []AgentDetail `yaml:"agents"`
}

type AgentDetail struct {
	Role      string `yaml:"role"`
	ModelType string `yaml:"model_type"`
	APIKey    string `yaml:"api_key"`
	BaseURL   string `yaml:"base_url"`
	SessionID string `yaml:"session_id"`
}
