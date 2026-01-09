package ai

type ModelConfig struct {
	Provider    string
	APIKey      string
	BaseURL     string
	ModelName   string
	ExtraParams map[string]any
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model          string         `json:"model"`
	Messages       []ChatMessage  `json:"messages"`
	Temperature    *float64       `json:"temperature,omitempty"`
	TopP           *float64       `json:"top_p,omitempty"`
	MaxTokens      *int           `json:"max_tokens,omitempty"`
	ResponseFormat map[string]any `json:"response_format,omitempty"`
	Extra          map[string]any `json:"-"`
}
