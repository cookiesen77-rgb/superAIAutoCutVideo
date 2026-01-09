package ai

import (
	"context"
	"strings"
)

func GenerateChat(ctx context.Context, cfg ModelConfig, req ChatRequest) (string, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch provider {
	case "gemini", "google", "google_ai", "googleai", "vertex", "vertexai":
		client := NewGeminiClient(cfg.APIKey, cfg.BaseURL)
		return client.GenerateChat(ctx, req)
	default:
		client := NewOpenAIClient(cfg.APIKey, cfg.BaseURL)
		return client.GenerateChat(ctx, req)
	}
}
