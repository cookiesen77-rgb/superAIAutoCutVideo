package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/cookiesen77-rgb/superAIAutoCutVideo/server/internal/ai"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	scriptPromptFeatureKey = "short_drama_narration:script_generation"
	maxSubtitleBytes       = 512 * 1024
	maxSubtitleChars       = 8000
)

func resolveScriptPrompt(ctx context.Context, db *pgxpool.Pool, userID, projectID string) (promptDetail, error) {
	selection, _ := loadPromptSelection(ctx, db, userID, projectID, scriptPromptFeatureKey)
	if selection != nil {
		mode := strings.ToLower(strings.TrimSpace(asString(selection["type"])))
		key := strings.TrimSpace(asString(selection["key_or_id"]))
		switch mode {
		case "user":
			if key != "" {
				if parts := strings.Split(key, ":"); len(parts) > 1 {
					key = parts[len(parts)-1]
				}
				item, err := loadPromptDetail(ctx, db, userID, key)
				if err == nil {
					return item, nil
				}
			}
		case "official":
			if key != "" {
				if item, ok := getOfficialPromptDetail(key); ok {
					return item, nil
				}
			}
		}
	}
	if item, ok := getOfficialPromptDetail(scriptPromptFeatureKey); ok {
		return item, nil
	}
	return promptDetail{}, errors.New("prompt not found")
}

func loadPromptSelection(ctx context.Context, db *pgxpool.Pool, userID, projectID, featureKey string) (map[string]any, error) {
	var raw []byte
	err := db.QueryRow(ctx, `
		SELECT selection
		FROM prompt_selections
		WHERE user_id=$1 AND project_id=$2 AND feature_key=$3
	`, userID, projectID, featureKey).Scan(&raw)
	if err != nil || len(raw) == 0 {
		return nil, err
	}
	selection := map[string]any{}
	if err := json.Unmarshal(raw, &selection); err != nil {
		return nil, err
	}
	return selection, nil
}

func loadPromptDetail(ctx context.Context, db *pgxpool.Pool, userID, promptID string) (promptDetail, error) {
	var item promptDetail
	err := db.QueryRow(ctx, `
		SELECT id, name, category, origin, template, variables, system_prompt, tags, version
		FROM prompts
		WHERE user_id=$1 AND id=$2
	`, userID, promptID).Scan(
		&item.IDOrKey,
		&item.Name,
		&item.Category,
		&item.Origin,
		&item.Template,
		&item.Variables,
		&item.SystemPrompt,
		&item.Tags,
		&item.Version,
	)
	return item, err
}

func fetchSubtitleContent(ctx context.Context, subtitleURL string) (string, error) {
	if strings.TrimSpace(subtitleURL) == "" {
		return "", nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, subtitleURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("subtitle fetch failed")
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSubtitleBytes))
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(string(body))
	text = strings.ToValidUTF8(text, "")
	if maxSubtitleChars > 0 && utf8.RuneCountInString(text) > maxSubtitleChars {
		text = truncateRunes(text, maxSubtitleChars)
	}
	return text, nil
}

func truncateRunes(s string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	out := make([]rune, 0, limit)
	for _, r := range s {
		out = append(out, r)
		if len(out) >= limit {
			break
		}
	}
	return string(out)
}

func buildScriptMessages(prompt promptDetail, vars map[string]any) []ai.ChatMessage {
	content := renderTemplate(prompt.Template, vars)
	messages := make([]ai.ChatMessage, 0, 2)
	if prompt.SystemPrompt != nil && strings.TrimSpace(*prompt.SystemPrompt) != "" {
		messages = append(messages, ai.ChatMessage{
			Role:    "system",
			Content: strings.TrimSpace(*prompt.SystemPrompt),
		})
	}
	messages = append(messages, ai.ChatMessage{
		Role:    "user",
		Content: content,
	})
	return messages
}

func extractScriptItems(data map[string]any) ([]map[string]any, error) {
	raw, ok := data["items"]
	if !ok {
		return nil, errors.New("items missing")
	}
	list, ok := raw.([]any)
	if !ok {
		return nil, errors.New("items invalid")
	}
	items := make([]map[string]any, 0, len(list))
	for _, entry := range list {
		if m, ok := entry.(map[string]any); ok {
			items = append(items, m)
		}
	}
	if len(items) == 0 {
		return nil, errors.New("items empty")
	}
	return items, nil
}

func buildVideoScriptFromItems(items []map[string]any) map[string]any {
	segments := make([]map[string]any, 0, len(items))
	var maxEnd float64
	for idx, item := range items {
		ts := strings.TrimSpace(asString(item["timestamp"]))
		start, end, ok := parseTimestampRange(ts)
		if !ok {
			continue
		}
		if end > maxEnd {
			maxEnd = end
		}
		id := strings.TrimSpace(asString(item["_id"]))
		if id == "" {
			id = strconv.Itoa(idx + 1)
		}
		if num, ok := item["_id"].(float64); ok && id == "" {
			id = strconv.Itoa(int(num))
		}
		seg := map[string]any{
			"id":         id,
			"start_time": start,
			"end_time":   end,
			"text":       strings.TrimSpace(asString(item["narration"])),
		}
		if pic := strings.TrimSpace(asString(item["picture"])); pic != "" {
			seg["subtitle"] = pic
		}
		segments = append(segments, seg)
	}
	return map[string]any{
		"version":        "v2.0",
		"total_duration": maxEnd,
		"segments":       segments,
		"metadata": map[string]any{
			"created_at": time.Now().UTC().Format(time.RFC3339),
		},
	}
}

func parseTimestampRange(raw string) (float64, float64, bool) {
	if raw == "" {
		return 0, 0, false
	}
	sep := ""
	if strings.Contains(raw, "-->") {
		sep = "-->"
	} else if strings.Contains(raw, "-") {
		sep = "-"
	} else if strings.Contains(raw, "~") {
		sep = "~"
	}
	if sep == "" {
		return 0, 0, false
	}
	parts := strings.Split(raw, sep)
	if len(parts) < 2 {
		return 0, 0, false
	}
	start, ok1 := parseTimecode(strings.TrimSpace(parts[0]))
	end, ok2 := parseTimecode(strings.TrimSpace(parts[1]))
	if !ok1 || !ok2 {
		return 0, 0, false
	}
	return start, end, true
}

func parseTimecode(raw string) (float64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}
	raw = strings.ReplaceAll(raw, ",", ".")
	parts := strings.Split(raw, ":")
	if len(parts) == 1 {
		sec, err := strconv.ParseFloat(parts[0], 64)
		return sec, err == nil
	}

	secPart := parts[len(parts)-1]
	minPart := parts[len(parts)-2]
	hourPart := "0"
	if len(parts) >= 3 {
		hourPart = parts[len(parts)-3]
	}
	sec, err := strconv.ParseFloat(secPart, 64)
	if err != nil {
		return 0, false
	}
	minutes, err := strconv.Atoi(minPart)
	if err != nil {
		return 0, false
	}
	hours, err := strconv.Atoi(hourPart)
	if err != nil {
		return 0, false
	}
	total := float64(hours*3600+minutes*60) + sec
	return total, true
}

func asString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		return ""
	}
}
