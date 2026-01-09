package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GeminiClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewGeminiClient(apiKey, baseURL string) *GeminiClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	return &GeminiClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *GeminiClient) GenerateChat(ctx context.Context, req ChatRequest) (string, error) {
	contents := make([]map[string]any, 0, len(req.Messages))
	systemText := ""
	for _, msg := range req.Messages {
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		if role == "system" {
			if systemText != "" {
				systemText += "\n"
			}
			systemText += msg.Content
			continue
		}
		geminiRole := "user"
		if role == "assistant" || role == "model" {
			geminiRole = "model"
		}
		contents = append(contents, map[string]any{
			"role": geminiRole,
			"parts": []map[string]string{
				{"text": msg.Content},
			},
		})
	}

	generationConfig := map[string]any{}
	if req.Temperature != nil {
		generationConfig["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		generationConfig["topP"] = *req.TopP
	}
	if req.MaxTokens != nil {
		generationConfig["maxOutputTokens"] = *req.MaxTokens
	}
	if req.ResponseFormat != nil {
		if formatType, ok := req.ResponseFormat["type"].(string); ok && strings.Contains(strings.ToLower(formatType), "json") {
			generationConfig["responseMimeType"] = "application/json"
		}
	}
	for k, v := range req.Extra {
		switch k {
		case "max_output_tokens":
			generationConfig["maxOutputTokens"] = v
		case "top_k":
			generationConfig["topK"] = v
		case "response_mime_type":
			generationConfig["responseMimeType"] = v
		case "response_format":
			continue
		default:
			if _, exists := generationConfig[k]; !exists {
				generationConfig[k] = v
			}
		}
	}

	payload := map[string]any{
		"contents": contents,
	}
	if systemText != "" {
		payload["systemInstruction"] = map[string]any{
			"parts": []map[string]string{
				{"text": systemText},
			},
		}
	}
	if len(generationConfig) > 0 {
		payload["generationConfig"] = generationConfig
	}
	if v, ok := req.Extra["safety_settings"]; ok {
		payload["safetySettings"] = v
	} else if v, ok := req.Extra["safetySettings"]; ok {
		payload["safetySettings"] = v
	}

	body, _ := json.Marshal(payload)
	endpoint := buildGeminiURL(c.baseURL, req.Model)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(c.apiKey) != "" {
		httpReq.Header.Set("x-goog-api-key", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errPayload map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errPayload)
		return "", fmt.Errorf("gemini error: %v", errPayload)
	}

	var data struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data.Candidates) == 0 || len(data.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response")
	}

	parts := make([]string, 0, len(data.Candidates[0].Content.Parts))
	for _, part := range data.Candidates[0].Content.Parts {
		if strings.TrimSpace(part.Text) != "" {
			parts = append(parts, part.Text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n")), nil
}

func buildGeminiURL(baseURL, model string) string {
	base := strings.TrimRight(baseURL, "/")
	model = strings.TrimSpace(model)
	if strings.Contains(base, ":generateContent") {
		return base
	}
	if strings.HasSuffix(base, "/models") {
		return base + "/" + url.PathEscape(model) + ":generateContent"
	}
	return base + "/models/" + url.PathEscape(model) + ":generateContent"
}
