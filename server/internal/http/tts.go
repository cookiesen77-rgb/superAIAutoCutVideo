package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/polly"
	pollytypes "github.com/aws/aws-sdk-go-v2/service/polly/types"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/config"
	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/storage"
)

type TTSHandler struct {
	cfg config.Config
	db  *pgxpool.Pool
}

func NewTTSHandler(cfg config.Config, db *pgxpool.Pool) *TTSHandler {
	return &TTSHandler{cfg: cfg, db: db}
}

func (h *TTSHandler) GetEngines(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": []map[string]any{
			{
				"provider":        "edge_tts",
				"display_name":    "Edge TTS",
				"description":     "微软 Edge TTS（云端）",
				"required_fields": []string{},
				"optional_fields": []string{},
			},
			{
				"provider":        "tencent_tts",
				"display_name":    "腾讯云 TTS",
				"description":     "腾讯云语音合成（云端）",
				"required_fields": []string{"secret_id", "secret_key"},
				"optional_fields": []string{"region"},
			},
			{
				"provider":        "index_tts",
				"display_name":    "IndexTTS2",
				"description":     "IndexTTS2 云端引擎",
				"required_fields": []string{},
				"optional_fields": []string{},
			},
			{
				"provider":        "aws_polly",
				"display_name":    "AWS Polly",
				"description":     "AWS Polly 云端语音合成",
				"required_fields": []string{"secret_id", "secret_key"},
				"optional_fields": []string{"region"},
			},
		},
		"message": "获取引擎成功",
	})
}

func (h *TTSHandler) GetVoices(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	provider := strings.TrimSpace(r.URL.Query().Get("provider"))
	if provider == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "provider required"})
		return
	}
	switch provider {
	case "aws_polly":
		configID := ttsConfigIDForProvider(provider)
		cfg, ok := h.loadTTSConfig(r, userID, configID)
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先配置 AWS Polly"})
			return
		}
		if !hasTtsCredentials(cfg) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 SecretId 或 SecretKey"})
			return
		}
		voices, err := h.listPollyVoices(r.Context(), cfg)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "获取音色失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    voices,
			"message": "获取音色成功",
		})
	default:
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    []map[string]any{},
			"message": "暂无音色",
		})
	}
}

func (h *TTSHandler) GetConfigs(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	rows, err := h.db.Query(r.Context(), `
		SELECT id, provider, config, active
		FROM tts_configs
		WHERE user_id=$1
		ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}
	defer rows.Close()

	configs := make(map[string]any)
	activeID := ""
	for rows.Next() {
		var id string
		var provider string
		var raw []byte
		var active bool
		if err := rows.Scan(&id, &provider, &raw, &active); err != nil {
			continue
		}
		cfg := map[string]any{}
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &cfg)
		}
		if _, ok := cfg["provider"]; !ok {
			cfg["provider"] = provider
		}
		maskSecret(cfg, "secret_id")
		maskSecret(cfg, "secret_key")
		if active && activeID == "" {
			activeID = id
		}
		configs[id] = cfg
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"configs":          configs,
			"active_config_id": activeID,
		},
		"message":   "获取配置成功",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *TTSHandler) PatchConfig(w http.ResponseWriter, r *http.Request) {
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

	existing, existingActive := h.loadTTSConfig(r, userID, configID)
	merged := mergeConfig(existing, payload, []string{"secret_id", "secret_key"})

	if _, ok := merged["provider"]; !ok {
		if provider := inferProviderFromTtsConfigID(configID); provider != "" {
			merged["provider"] = provider
		}
	}
	if _, ok := merged["speed_ratio"]; !ok {
		merged["speed_ratio"] = 1.0
	}
	provider, _ := merged["provider"].(string)
	raw, _ := json.Marshal(merged)

	_, err = h.db.Exec(r.Context(), `
		INSERT INTO tts_configs (user_id, id, provider, config, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (user_id, id)
		DO UPDATE SET provider=EXCLUDED.provider, config=EXCLUDED.config, updated_at=NOW()
	`, userID, configID, provider, raw, existingActive)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "save failed"})
		return
	}

	maskSecret(merged, "secret_id")
	maskSecret(merged, "secret_key")
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

func (h *TTSHandler) ActivateConfig(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	configID := chi.URLParam(r, "configID")
	if configID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "config id required"})
		return
	}
	_, err := h.db.Exec(r.Context(), `
		UPDATE tts_configs
		SET active = CASE WHEN id=$2 THEN TRUE ELSE FALSE END,
		    updated_at = NOW()
		WHERE user_id=$1
	`, userID, configID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "activate failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"success":   true,
		"message":   "配置已激活",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *TTSHandler) TestConfig(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	configID := chi.URLParam(r, "configID")
	if configID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "config id required"})
		return
	}
	cfg, ok := h.loadTTSConfig(r, userID, configID)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "请先配置 TTS 引擎"})
		return
	}
	if _, ok := cfg["provider"]; !ok {
		cfg["provider"] = inferProviderFromTtsConfigID(configID)
	}
	provider := strings.TrimSpace(asString(cfg["provider"]))
	if provider == "" {
		provider = inferProviderFromTtsConfigID(configID)
	}
	if provider == "edge_tts" {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"config_id": configID,
				"provider":  provider,
			},
			"message": "Edge TTS 服务可用",
		})
		return
	}
	secretID := strings.TrimSpace(asString(cfg["secret_id"]))
	secretKey := strings.TrimSpace(asString(cfg["secret_key"]))
	if secretID == "" || secretKey == "" || secretID == "***" || secretKey == "***" {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"data": map[string]any{
				"error": "缺少 SecretId 或 SecretKey",
			},
			"message": "连接测试失败",
		})
		return
	}
	switch provider {
	case "aws_polly":
		if err := h.testPolly(r.Context(), cfg); err != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"success": false,
				"data": map[string]any{
					"error": err.Error(),
				},
				"message": "连接测试失败",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"config_id": configID,
				"provider":  provider,
			},
			"message": "连接测试成功",
		})
	default:
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"data": map[string]any{
				"error": "云端 TTS 暂未接入自动测试",
			},
			"message": "连接测试失败",
		})
	}
}

func (h *TTSHandler) PreviewVoice(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	voiceID := chi.URLParam(r, "voiceID")
	var payload struct {
		ConfigID string `json:"config_id"`
		Provider string `json:"provider"`
		Text     string `json:"text"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)
	provider := strings.TrimSpace(payload.Provider)
	configID := strings.TrimSpace(payload.ConfigID)
	if provider == "" && configID != "" {
		provider = inferProviderFromTtsConfigID(configID)
	}
	if provider == "" {
		provider = "edge_tts"
	}
	if configID == "" {
		configID = ttsConfigIDForProvider(provider)
	}
	cfg, _ := h.loadTTSConfig(r, userID, configID)
	if _, ok := cfg["provider"]; !ok {
		cfg["provider"] = provider
	}
	text := strings.TrimSpace(payload.Text)
	if text == "" {
		text = "您好，欢迎使用智能配音。"
	}
	switch provider {
	case "aws_polly":
		if !hasTtsCredentials(cfg) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "缺少 SecretId 或 SecretKey"})
			return
		}
		units := runeCount(text)
		if err := consumeUsage(r.Context(), h.db, userID, "tts_chars", units, map[string]any{
			"provider":  provider,
			"voice_id":  voiceID,
			"config_id": configID,
		}); err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "配额不足"})
			return
		}
		url, err := h.synthesizePolly(r.Context(), cfg, voiceID, text)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "试听失败"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"audio_url":      url,
				"sample_wav_url": url,
				"voice_id":       voiceID,
			},
			"message": "试听成功",
		})
	default:
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"data": map[string]any{
				"error": "云端 TTS 暂未配置试听接口",
			},
			"message": "试听失败",
		})
	}
}

func (h *TTSHandler) GetEmotions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": []map[string]any{
			{"id": "happy", "name": "开心", "icon": "😄"},
			{"id": "sad", "name": "悲伤", "icon": "😢"},
			{"id": "angry", "name": "愤怒", "icon": "😠"},
			{"id": "afraid", "name": "恐惧", "icon": "😨"},
			{"id": "calm", "name": "平静", "icon": "😌"},
			{"id": "surprised", "name": "惊讶", "icon": "😲"},
			{"id": "melancholic", "name": "忧郁", "icon": "🥲"},
			{"id": "disgusted", "name": "厌恶", "icon": "😣"},
		},
		"message": "获取情感类型成功",
	})
}

func (h *TTSHandler) GetIndexTtsStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"loaded":    false,
			"loading":   false,
			"available": false,
			"error":     "云端 IndexTTS 未启用",
		},
		"message": "状态获取成功",
	})
}

func (h *TTSHandler) PreloadIndexTts(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": false,
		"error":   "云端 IndexTTS 未启用",
		"message": "模型加载失败",
	})
}

func (h *TTSHandler) TestIndexTts(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": false,
		"error":   "云端 IndexTTS 未启用",
		"message": "测试失败",
	})
}

func (h *TTSHandler) loadTTSConfig(r *http.Request, userID, configID string) (map[string]any, bool) {
	var raw []byte
	var active bool
	err := h.db.QueryRow(r.Context(), `
		SELECT config, active
		FROM tts_configs
		WHERE user_id=$1 AND id=$2
	`, userID, configID).Scan(&raw, &active)
	if err != nil || len(raw) == 0 {
		return map[string]any{}, false
	}
	cfg := map[string]any{}
	_ = json.Unmarshal(raw, &cfg)
	return cfg, active
}

func inferProviderFromTtsConfigID(configID string) string {
	return strings.TrimSuffix(configID, "_default")
}

func ttsConfigIDForProvider(provider string) string {
	switch provider {
	case "tencent_tts":
		return "tencent_tts_default"
	case "index_tts":
		return "index_tts_default"
	default:
		return provider + "_default"
	}
}

func runeCount(text string) int {
	return len([]rune(text))
}

func hasTtsCredentials(cfg map[string]any) bool {
	secretID := strings.TrimSpace(asString(cfg["secret_id"]))
	secretKey := strings.TrimSpace(asString(cfg["secret_key"]))
	if secretID == "" || secretKey == "" {
		return false
	}
	if secretID == "***" || secretKey == "***" {
		return false
	}
	return true
}

func (h *TTSHandler) listPollyVoices(ctx context.Context, cfg map[string]any) ([]map[string]any, error) {
	client, err := h.pollyClient(ctx, cfg)
	if err != nil {
		return nil, err
	}
	out, err := client.DescribeVoices(ctx, &polly.DescribeVoicesInput{})
	if err != nil {
		return nil, err
	}
	voices := make([]map[string]any, 0, len(out.Voices))
	for _, v := range out.Voices {
		engines := []string{}
		for _, e := range v.SupportedEngines {
			engines = append(engines, string(e))
		}
		voices = append(voices, map[string]any{
			"id":             string(v.Id),
			"name":           aws.ToString(v.Name),
			"language":       string(v.LanguageCode),
			"gender":         string(v.Gender),
			"voice_quality":  strings.Join(engines, "/"),
			"voice_type_tag": aws.ToString(v.LanguageName),
		})
	}
	return voices, nil
}

func (h *TTSHandler) testPolly(ctx context.Context, cfg map[string]any) error {
	client, err := h.pollyClient(ctx, cfg)
	if err != nil {
		return err
	}
	_, err = client.DescribeVoices(ctx, &polly.DescribeVoicesInput{})
	return err
}

func (h *TTSHandler) synthesizePolly(ctx context.Context, cfg map[string]any, voiceID, text string) (string, error) {
	client, err := h.pollyClient(ctx, cfg)
	if err != nil {
		return "", err
	}
	req := &polly.SynthesizeSpeechInput{
		OutputFormat: pollytypes.OutputFormatMp3,
		Text:         aws.String(text),
		VoiceId:      pollytypes.VoiceId(voiceID),
	}
	out, err := client.SynthesizeSpeech(ctx, req)
	if err != nil {
		return "", err
	}
	defer out.AudioStream.Close()

	uploader, err := storage.NewS3Uploader(ctx, h.cfg.AWSRegion, h.cfg.S3Bucket)
	if err != nil {
		return "", err
	}
	key := strings.Join([]string{"tts", "previews", uuid.NewString() + ".mp3"}, "/")
	url, err := uploader.Upload(ctx, key, "audio/mpeg", out.AudioStream)
	if err != nil {
		return "", err
	}
	return url, nil
}

func (h *TTSHandler) pollyClient(ctx context.Context, cfg map[string]any) (*polly.Client, error) {
	region := strings.TrimSpace(asString(cfg["region"]))
	if region == "" {
		region = h.cfg.AWSRegion
	}
	secretID := strings.TrimSpace(asString(cfg["secret_id"]))
	secretKey := strings.TrimSpace(asString(cfg["secret_key"]))
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
	}
	if secretID != "" && secretKey != "" && secretID != "***" && secretKey != "***" {
		creds := credentials.NewStaticCredentialsProvider(secretID, secretKey, "")
		opts = append(opts, awsconfig.WithCredentialsProvider(creds))
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return polly.NewFromConfig(awsCfg), nil
}
