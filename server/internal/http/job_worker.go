package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/ai"
	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type jobRecord struct {
	ID       string
	UserID   string
	Type     string
	InputURL string
	Payload  map[string]any
}

func StartJobWorker(ctx context.Context, cfg config.Config, db *pgxpool.Pool) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := processNextJob(ctx, cfg, db); err != nil {
				continue
			}
		}
	}
}

func processNextJob(ctx context.Context, cfg config.Config, db *pgxpool.Pool) error {
	job, err := claimQueuedJob(ctx, db)
	if err != nil {
		return err
	}
	if job == nil {
		return nil
	}

	switch strings.ToLower(job.Type) {
	case "script":
		result, err := handleScriptJob(ctx, db, job)
		if err != nil {
			_ = failJob(ctx, db, job.ID, err.Error())
			return err
		}
		_ = completeJob(ctx, db, job.ID, "", result)
	case "tts":
		result, outputURL, err := handleTtsJob(ctx, cfg, db, job)
		if err != nil {
			_ = failJob(ctx, db, job.ID, err.Error())
			return err
		}
		_ = completeJob(ctx, db, job.ID, outputURL, result)
	default:
		_ = failJob(ctx, db, job.ID, "unsupported job type")
	}
	return nil
}

func claimQueuedJob(ctx context.Context, db *pgxpool.Pool) (*jobRecord, error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var id, userID, jobType, inputURL string
	var payloadRaw []byte
	err = tx.QueryRow(ctx, `
		SELECT id, user_id, type, input_url, payload
		FROM jobs
		WHERE status='queued'
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`).Scan(&id, &userID, &jobType, &inputURL, &payloadRaw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	_, err = tx.Exec(ctx, `
		UPDATE jobs SET status='processing', progress=10, updated_at=NOW()
		WHERE id=$1
	`, id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	payload := map[string]any{}
	if len(payloadRaw) > 0 {
		_ = json.Unmarshal(payloadRaw, &payload)
	}
	return &jobRecord{
		ID:       id,
		UserID:   userID,
		Type:     jobType,
		InputURL: inputURL,
		Payload:  payload,
	}, nil
}

func handleScriptJob(ctx context.Context, db *pgxpool.Pool, job *jobRecord) (map[string]any, error) {
	projectID := strings.TrimSpace(asString(job.Payload["project_id"]))
	subtitleURL := strings.TrimSpace(asString(job.Payload["subtitle_url"]))
	dramaName := strings.TrimSpace(asString(job.Payload["drama_name"]))
	plotAnalysis := strings.TrimSpace(asString(job.Payload["plot_analysis"]))

	if projectID != "" {
		meta, err := loadProjectMeta(ctx, db, job.UserID, projectID)
		if err == nil {
			if dramaName == "" {
				dramaName = meta.Name
			}
			if plotAnalysis == "" && meta.Description != nil {
				plotAnalysis = strings.TrimSpace(*meta.Description)
			}
			if subtitleURL == "" && meta.SubtitlePath != nil {
				subtitleURL = *meta.SubtitlePath
			}
		}
	}
	if subtitleURL == "" && strings.HasPrefix(job.InputURL, "http") {
		subtitleURL = job.InputURL
	}
	if dramaName == "" {
		dramaName = "未命名短剧"
	}

	modelID, cfgMap, ok := loadActiveModelConfig(ctx, db, job.UserID, modelTypeContent)
	if !ok {
		return nil, errors.New("missing content model config")
	}
	if _, ok := cfgMap["provider"]; !ok {
		if provider := inferProviderFromConfigID(modelID); provider != "" {
			cfgMap["provider"] = provider
		}
	}
	modelCfg, err := modelConfigFromMap(cfgMap, inferProviderFromConfigID(modelID))
	if err != nil {
		return nil, err
	}

	prompt, err := resolveScriptPrompt(ctx, db, job.UserID, projectID)
	if err != nil {
		return nil, err
	}

	subtitleText, err := fetchSubtitleContent(ctx, subtitleURL)
	if err != nil {
		return nil, err
	}
	vars := map[string]any{
		"drama_name":       dramaName,
		"plot_analysis":    plotAnalysis,
		"subtitle_content": subtitleText,
	}
	messages := buildScriptMessages(prompt, vars)
	req := buildChatRequest(modelCfg, cfgMap, messages, true)
	units := 0
	for _, msg := range messages {
		units += runeCount(msg.Content)
	}
	if err := consumeUsage(ctx, db, job.UserID, "llm_chars", units, map[string]any{
		"provider": modelCfg.Provider,
		"model":    modelCfg.ModelName,
	}); err != nil {
		return nil, errors.New("quota exceeded")
	}

	raw, err := ai.GenerateChat(ctx, modelCfg, req)
	if err != nil {
		return nil, err
	}
	data, _, err := parseJSONFromText(raw)
	if err != nil {
		return nil, err
	}
	items, err := extractScriptItems(data)
	if err != nil {
		return nil, err
	}
	script := buildVideoScriptFromItems(items)

	if projectID != "" {
		scriptJSON, _ := json.Marshal(script)
		_, _ = db.Exec(ctx, `
			UPDATE projects SET script=$3, updated_at=NOW()
			WHERE id=$1 AND user_id=$2
		`, projectID, job.UserID, scriptJSON)
	}

	return map[string]any{
		"script":        script,
		"plot_analysis": plotAnalysis,
	}, nil
}

func handleTtsJob(ctx context.Context, cfg config.Config, db *pgxpool.Pool, job *jobRecord) (map[string]any, string, error) {
	provider := strings.TrimSpace(asString(job.Payload["provider"]))
	configID := strings.TrimSpace(asString(job.Payload["config_id"]))
	voiceID := strings.TrimSpace(asString(job.Payload["voice_id"]))
	text := strings.TrimSpace(asString(job.Payload["text"]))
	if provider == "" {
		provider = "aws_polly"
	}
	if configID == "" {
		configID = ttsConfigIDForProvider(provider)
	}
	if text == "" {
		text = "您好，欢迎使用智能配音。"
	}
	if voiceID == "" {
		return nil, "", errors.New("voice_id required")
	}

	switch provider {
	case "aws_polly":
		cfgMap, err := loadTtsConfig(ctx, db, job.UserID, configID)
		if err != nil {
			return nil, "", err
		}
		if !hasTtsCredentials(cfgMap) {
			return nil, "", errors.New("missing credentials")
		}
		units := runeCount(text)
		if err := consumeUsage(ctx, db, job.UserID, "tts_chars", units, map[string]any{
			"provider":  provider,
			"voice_id":  voiceID,
			"config_id": configID,
		}); err != nil {
			return nil, "", errors.New("quota exceeded")
		}
		handler := &TTSHandler{cfg: cfg, db: db}
		url, err := handler.synthesizePolly(ctx, cfgMap, voiceID, text)
		if err != nil {
			return nil, "", err
		}
		return map[string]any{
			"audio_url":      url,
			"sample_wav_url": url,
			"voice_id":       voiceID,
		}, url, nil
	default:
		return nil, "", errors.New("unsupported tts provider")
	}
}

type projectMeta struct {
	ID           string
	Name         string
	Description  *string
	SubtitlePath *string
}

func loadProjectMeta(ctx context.Context, db *pgxpool.Pool, userID, projectID string) (projectMeta, error) {
	var meta projectMeta
	err := db.QueryRow(ctx, `
		SELECT id, name, description, subtitle_path
		FROM projects
		WHERE id=$1 AND user_id=$2
	`, projectID, userID).Scan(&meta.ID, &meta.Name, &meta.Description, &meta.SubtitlePath)
	return meta, err
}

func loadTtsConfig(ctx context.Context, db *pgxpool.Pool, userID, configID string) (map[string]any, error) {
	var raw []byte
	err := db.QueryRow(ctx, `
		SELECT config
		FROM tts_configs
		WHERE user_id=$1 AND id=$2
	`, userID, configID).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return nil, errors.New("tts config not found")
	}
	cfg := map[string]any{}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func completeJob(ctx context.Context, db *pgxpool.Pool, jobID, outputURL string, result map[string]any) error {
	raw, _ := json.Marshal(result)
	_, err := db.Exec(ctx, `
		UPDATE jobs
		SET status='completed', progress=100, output_url=$2, result=$3, updated_at=NOW()
		WHERE id=$1
	`, jobID, nullIfEmpty(outputURL), raw)
	return err
}

func failJob(ctx context.Context, db *pgxpool.Pool, jobID, message string) error {
	_, err := db.Exec(ctx, `
		UPDATE jobs
		SET status='failed', error=$2, progress=100, updated_at=NOW()
		WHERE id=$1
	`, jobID, message)
	return err
}

func nullIfEmpty(val string) *string {
	if strings.TrimSpace(val) == "" {
		return nil
	}
	out := val
	return &out
}
