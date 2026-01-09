package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/ai"
	"github.com/jackc/pgx/v5/pgxpool"
)

func loadActiveModelConfig(ctx context.Context, db *pgxpool.Pool, userID, modelType string) (string, map[string]any, bool) {
	rows, err := db.Query(ctx, `
		SELECT id, config
		FROM model_configs
		WHERE user_id=$1 AND type=$2
		ORDER BY updated_at DESC
	`, userID, modelType)
	if err != nil {
		return "", nil, false
	}
	defer rows.Close()

	var fallbackID string
	var fallbackCfg map[string]any
	for rows.Next() {
		var id string
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			continue
		}
		cfg := map[string]any{}
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &cfg); err != nil {
				continue
			}
		}
		if fallbackID == "" {
			fallbackID = id
			fallbackCfg = cfg
		}
		if enabled, ok := cfg["enabled"].(bool); ok && enabled {
			return id, cfg, true
		}
	}
	if fallbackID != "" {
		return fallbackID, fallbackCfg, true
	}
	return "", nil, false
}

func modelConfigFromMap(cfg map[string]any, fallbackProvider string) (ai.ModelConfig, error) {
	provider, _ := cfg["provider"].(string)
	if strings.TrimSpace(provider) == "" {
		provider = fallbackProvider
	}
	apiKey, _ := cfg["api_key"].(string)
	if strings.TrimSpace(apiKey) == "" || apiKey == "***" {
		return ai.ModelConfig{}, errors.New("api_key missing")
	}
	baseURL, _ := cfg["base_url"].(string)
	modelName, _ := cfg["model_name"].(string)
	if strings.TrimSpace(modelName) == "" {
		return ai.ModelConfig{}, errors.New("model_name missing")
	}
	extra := map[string]any{}
	if raw, ok := cfg["extra_params"]; ok {
		if m, ok := raw.(map[string]any); ok {
			extra = m
		}
	}
	return ai.ModelConfig{
		Provider:    provider,
		APIKey:      apiKey,
		BaseURL:     baseURL,
		ModelName:   modelName,
		ExtraParams: extra,
	}, nil
}

func parseJSONFromText(raw string) (map[string]any, string, error) {
	out := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &out); err == nil {
		return out, raw, nil
	} else {
		return parseJSONFromTextFallback(raw, err)
	}
}

func parseJSONFromTextFallback(raw string, lastErr error) (map[string]any, string, error) {
	candidate, ok := extractJSONObject(raw)
	if !ok {
		return nil, "", lastErr
	}
	out := map[string]any{}
	if err := json.Unmarshal([]byte(candidate), &out); err != nil {
		return nil, "", err
	}
	return out, candidate, nil
}

func extractJSONObject(raw string) (string, bool) {
	start := strings.Index(raw, "{")
	if start < 0 {
		return "", false
	}
	depth := 0
	inString := false
	escape := false
	for i := start; i < len(raw); i++ {
		ch := raw[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			if ch == '\\' {
				escape = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return raw[start : i+1], true
			}
		}
	}
	return "", false
}
