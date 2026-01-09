package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAIClient(apiKey, baseURL string) *OpenAIClient {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openai.com"
	}
	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (c *OpenAIClient) GenerateChat(ctx context.Context, req ChatRequest) (string, error) {
	payload := map[string]any{
		"model":    req.Model,
		"messages": req.Messages,
	}
	if req.Temperature != nil {
		payload["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		payload["top_p"] = *req.TopP
	}
	if req.MaxTokens != nil {
		payload["max_tokens"] = *req.MaxTokens
	}
	if req.ResponseFormat != nil {
		payload["response_format"] = req.ResponseFormat
	}
	for k, v := range req.Extra {
		if _, exists := payload[k]; !exists {
			payload[k] = v
		}
	}

	body, _ := json.Marshal(payload)
	endpoint := buildOpenAIURL(c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(c.apiKey) != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errPayload map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errPayload)
		return "", fmt.Errorf("openai error: %v", errPayload)
	}

	var data struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if len(data.Choices) == 0 {
		return "", errors.New("empty response")
	}
	return strings.TrimSpace(data.Choices[0].Message.Content), nil
}

func buildOpenAIURL(baseURL string) string {
	if strings.Contains(baseURL, "/chat/completions") {
		return baseURL
	}
	if strings.HasSuffix(baseURL, "/v1") {
		return baseURL + "/chat/completions"
	}
	return baseURL + "/v1/chat/completions"
}
