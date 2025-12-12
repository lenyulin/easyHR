package config

type AgentConfig struct {
	PromptDir string        `yaml:"promptDir"`
	Agents    []AgentDetail `yaml:"providers"`
}

type AgentDetail struct {
	Role      string `yaml:"role"`
	ModelType string `yaml:"model_type"`
	ModelName string `yaml:"model_name"`
	APIKey    string `yaml:"api_key"`
	BaseURL   string `yaml:"base_url"`
	SessionID string `yaml:"session_id"`
}
