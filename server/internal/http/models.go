package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/ai"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	modelTypeVideo   = "video_analysis"
	modelTypeContent = "content_generation"
)

type ModelConfigHandler struct {
	db *pgxpool.Pool
}

func NewModelConfigHandler(db *pgxpool.Pool) *ModelConfigHandler {
	return &ModelConfigHandler{db: db}
}

func (h *ModelConfigHandler) ListVideoConfigs(w http.ResponseWriter, r *http.Request) {
	h.listConfigs(w, r, modelTypeVideo)
}

func (h *ModelConfigHandler) ListContentConfigs(w http.ResponseWriter, r *http.Request) {
	h.listConfigs(w, r, modelTypeContent)
}

func (h *ModelConfigHandler) UpdateVideoConfig(w http.ResponseWriter, r *http.Request) {
	h.updateConfig(w, r, modelTypeVideo)
}

func (h *ModelConfigHandler) UpdateContentConfig(w http.ResponseWriter, r *http.Request) {
	h.updateConfig(w, r, modelTypeContent)
}

func (h *ModelConfigHandler) ActivateVideoConfig(w http.ResponseWriter, r *http.Request) {
	h.activateConfig(w, r, modelTypeVideo)
}

func (h *ModelConfigHandler) TestVideoConfig(w http.ResponseWriter, r *http.Request) {
	h.testConfig(w, r, modelTypeVideo)
}

func (h *ModelConfigHandler) TestContentConfig(w http.ResponseWriter, r *http.Request) {
	h.testConfig(w, r, modelTypeContent)
}

func (h *ModelConfigHandler) listConfigs(w http.ResponseWriter, r *http.Request, modelType string) {
	userID := UserIDFromContext(r.Context())
	rows, err := h.db.Query(r.Context(), `
		SELECT id, config
		FROM model_configs
		WHERE user_id=$1 AND type=$2
		ORDER BY updated_at DESC
	`, userID, modelType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	configs := make(map[string]any)
	for rows.Next() {
		var id string
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			continue
		}
		cfg := map[string]any{}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &cfg)
		}
		maskSecret(cfg, "api_key")
		configs[id] = cfg
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"configs": configs,
		},
		"message":   "获取配置成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ModelConfigHandler) updateConfig(w http.ResponseWriter, r *http.Request, modelType string) {
	userID := UserIDFromContext(r.Context())
	configID := chi.URLParam(r, "configID")
	if configID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "config id required"})
		return
	}

	payload, err := decodeConfigPayload(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	existing, _ := h.loadConfig(r.Context(), userID, modelType, configID)
	merged := mergeConfig(existing, payload, []string{"api_key"})
	if _, ok := merged["provider"]; !ok {
		if provider := inferProviderFromConfigID(configID); provider != "" {
			merged["provider"] = provider
		}
	}

	enabled, _ := merged["enabled"].(bool)
	if enabled {
		_ = h.disableOtherConfigs(r.Context(), userID, modelType, configID)
	}

	raw, _ := json.Marshal(merged)
	_, err = h.db.Exec(r.Context(), `
		INSERT INTO model_configs (user_id, id, type, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (user_id, id, type)
		DO UPDATE SET config=EXCLUDED.config, updated_at=NOW()
	`, userID, configID, modelType, raw)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save failed"})
		return
	}

	maskSecret(merged, "api_key")
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"config_id": configID,
			"config":    merged,
		},
		"message":   "配置已更新",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ModelConfigHandler) activateConfig(w http.ResponseWriter, r *http.Request, modelType string) {
	userID := UserIDFromContext(r.Context())
	configID := chi.URLParam(r, "configID")
	if configID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "config id required"})
		return
	}
	cfg, ok := h.loadConfig(r.Context(), userID, modelType, configID)
	if ok {
		cfg["enabled"] = true
		raw, _ := json.Marshal(cfg)
		_, _ = h.db.Exec(r.Context(), `
			UPDATE model_configs SET config=$4, updated_at=NOW()
			WHERE user_id=$1 AND type=$2 AND id=$3
		`, userID, modelType, configID, raw)
	}
	_ = h.disableOtherConfigs(r.Context(), userID, modelType, configID)
	writeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"message":   "配置已激活",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *ModelConfigHandler) testConfig(w http.ResponseWriter, r *http.Request, modelType string) {
	userID := UserIDFromContext(r.Context())
	configID := chi.URLParam(r, "configID")
	if configID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "config id required"})
		return
	}

	cfg, ok := h.loadConfig(r.Context(), userID, modelType, configID)
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"data": map[string]any{
				"error": "未找到配置",
			},
			"message": "连接测试失败",
		})
		return
	}

	if _, ok := cfg["provider"]; !ok {
		if provider := inferProviderFromConfigID(configID); provider != "" {
			cfg["provider"] = provider
		}
	}
	modelCfg, err := modelConfigFromMap(cfg, inferProviderFromConfigID(configID))
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"data": map[string]any{
				"error": err.Error(),
			},
			"message": "连接测试失败",
		})
		return
	}

	testMessages := []ai.ChatMessage{
		{Role: "system", Content: "You are a helpful assistant."},
	}
	if modelType == modelTypeContent {
		testMessages = append(testMessages, ai.ChatMessage{
			Role:    "user",
			Content: "Reply with a JSON object with keys ok (boolean) and note (string) to confirm the model connection.",
		})
	} else {
		testMessages = append(testMessages, ai.ChatMessage{
			Role:    "user",
			Content: "Reply with a short acknowledgement to confirm the model connection.",
		})
	}

	req := buildChatRequest(modelCfg, cfg, testMessages, modelType == modelTypeContent)
	raw, err := ai.GenerateChat(r.Context(), modelCfg, req)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"data": map[string]any{
				"error": err.Error(),
			},
			"message": "连接测试失败",
		})
		return
	}

	response := map[string]any{
		"response_preview": raw,
	}
	if modelType == modelTypeContent {
		if parsed, _, err := parseJSONFromText(raw); err == nil {
			response["structured_output"] = parsed
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    response,
		"message": "连接测试成功",
	})
}

func buildChatRequest(modelCfg ai.ModelConfig, cfg map[string]any, messages []ai.ChatMessage, enforceJSON bool) ai.ChatRequest {
	req := ai.ChatRequest{
		Model:    modelCfg.ModelName,
		Messages: messages,
		Extra:    modelCfg.ExtraParams,
	}
	if temp, ok := floatFromAny(cfg["temperature"]); ok {
		req.Temperature = &temp
	}
	if temp, ok := floatFromAny(modelCfg.ExtraParams["temperature"]); ok {
		req.Temperature = &temp
	}
	if topP, ok := floatFromAny(cfg["top_p"]); ok {
		req.TopP = &topP
	}
	if topP, ok := floatFromAny(modelCfg.ExtraParams["top_p"]); ok {
		req.TopP = &topP
	}
	if maxTokens, ok := intFromAny(cfg["max_tokens"]); ok {
		req.MaxTokens = &maxTokens
	}
	if maxTokens, ok := intFromAny(modelCfg.ExtraParams["max_tokens"]); ok {
		req.MaxTokens = &maxTokens
	}
	if respFormat, ok := cfg["response_format"].(map[string]any); ok {
		req.ResponseFormat = respFormat
	} else if respFormat, ok := modelCfg.ExtraParams["response_format"].(map[string]any); ok {
		req.ResponseFormat = respFormat
	} else if enforceJSON {
		req.ResponseFormat = map[string]any{"type": "json_object"}
	}
	return req
}

func floatFromAny(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		if err == nil {
			return f, true
		}
	}
	return 0, false
}

func intFromAny(val any) (int, bool) {
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return int(i), true
		}
	}
	return 0, false
}

func (h *ModelConfigHandler) loadConfig(ctx context.Context, userID, modelType, configID string) (map[string]any, bool) {
	var raw []byte
	err := h.db.QueryRow(ctx, `
		SELECT config
		FROM model_configs
		WHERE user_id=$1 AND type=$2 AND id=$3
	`, userID, modelType, configID).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return map[string]any{}, false
	}
	cfg := map[string]any{}
	_ = json.Unmarshal(raw, &cfg)
	return cfg, true
}

func (h *ModelConfigHandler) disableOtherConfigs(ctx context.Context, userID, modelType, activeID string) error {
	rows, err := h.db.Query(ctx, `
		SELECT id, config
		FROM model_configs
		WHERE user_id=$1 AND type=$2 AND id<>$3
	`, userID, modelType, activeID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			continue
		}
		cfg := map[string]any{}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &cfg)
		}
		if enabled, ok := cfg["enabled"].(bool); ok && !enabled {
			continue
		}
		cfg["enabled"] = false
		nextRaw, _ := json.Marshal(cfg)
		_, _ = h.db.Exec(ctx, `
			UPDATE model_configs SET config=$4, updated_at=NOW()
			WHERE user_id=$1 AND type=$2 AND id=$3
		`, userID, modelType, id, nextRaw)
	}
	return nil
}

func decodeConfigPayload(r *http.Request) (map[string]any, error) {
	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, err
	}
	if cfg, ok := payload["config"]; ok {
		if m, ok := cfg.(map[string]any); ok {
			return m, nil
		}
	}
	return payload, nil
}

func mergeConfig(existing, patch map[string]any, secretKeys []string) map[string]any {
	if existing == nil {
		existing = map[string]any{}
	}
	secretSet := map[string]struct{}{}
	for _, key := range secretKeys {
		secretSet[key] = struct{}{}
	}
	for key, val := range patch {
		if _, isSecret := secretSet[key]; isSecret {
			if str, ok := val.(string); ok && str == "***" {
				continue
			}
		}
		if key == "extra_params" {
			if patchMap, ok := val.(map[string]any); ok {
				base, _ := existing[key].(map[string]any)
				if base == nil {
					base = map[string]any{}
				}
				for k, v := range patchMap {
					base[k] = v
				}
				existing[key] = base
				continue
			}
		}
		existing[key] = val
	}
	return existing
}

func maskSecret(cfg map[string]any, key string) {
	if cfg == nil {
		return
	}
	if val, ok := cfg[key].(string); ok && strings.TrimSpace(val) != "" {
		cfg[key] = "***"
	}
}

func inferProviderFromConfigID(configID string) string {
	parts := strings.Split(configID, "_")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
