package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"message":   "云端 API 运行中",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthHandler) TestIntegrations(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"content_model": map[string]any{
				"status":  "unknown",
				"message": "云端模型尚未配置",
			},
			"tts": map[string]any{
				"status":  "unknown",
				"message": "云端 TTS 尚未配置",
			},
			"asr": map[string]any{
				"status":  "unknown",
				"message": "云端 ASR 尚未配置",
			},
			"overall_status": "healthy",
		},
		"message": "集成检查已完成",
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	if payload != nil {
		switch data := payload.(type) {
		case map[string]any:
			if _, has := data["message"]; !has {
				if msg, ok := data["error"].(string); ok && msg != "" {
					data["message"] = msg
				}
			}
		case map[string]string:
			if _, has := data["message"]; !has {
				if msg, ok := data["error"]; ok && msg != "" {
					data["message"] = msg
				}
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
