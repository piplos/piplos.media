package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ProviderSettings is the composite JSON for GEMINI/GROK/OPENAI/OPENROUTER keys.
type ProviderSettings struct {
	Enable         bool   `json:"enable"`
	APIKey         string `json:"apiKey"`
	RateLimit      int    `json:"rateLimit"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
}

// TaskSettings is the composite JSON for AI_TRANSLATION.
type TaskSettings struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Prompt   string `json:"prompt"`
}

// Client performs chat completion requests for translation.
type Client interface {
	TestAPIKey(ctx context.Context) error
	ChatJSON(ctx context.Context, systemPrompt, userPrompt string, temperature float64) (string, error)
}

type openAICompat struct {
	apiKey   string
	baseURL  string
	model    string
	timeout  time.Duration
	provider string
}

func newOpenAICompat(apiKey, baseURL, model, provider string, timeout time.Duration) *openAICompat {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &openAICompat{apiKey: apiKey, baseURL: baseURL, model: model, timeout: timeout, provider: provider}
}

func (c *openAICompat) TestAPIKey(ctx context.Context) error {
	if c.apiKey == "" {
		return fmt.Errorf("API key is empty")
	}
	body := []byte(fmt.Sprintf(`{"model":%q,"messages":[{"role":"user","content":"Hi"}],"max_tokens":1}`, c.model))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := (&http.Client{Timeout: c.timeout}).Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusOK, http.StatusTooManyRequests:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("invalid API key")
	default:
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
}

func (c *openAICompat) ChatJSON(ctx context.Context, systemPrompt, userPrompt string, temperature float64) (string, error) {
	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     temperature,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	resp, err := (&http.Client{Timeout: c.timeout}).Do(req)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API status %d: %s", resp.StatusCode, truncate(string(body), 500))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("no choices in AI response")
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	content = stripJSONFences(content)
	if content == "" {
		return "", fmt.Errorf("empty content in AI response")
	}
	return content, nil
}

type geminiClient struct {
	apiKey   string
	model    string
	timeout  time.Duration
	provider string
}

func newGemini(apiKey, model string, timeout time.Duration) *geminiClient {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &geminiClient{apiKey: apiKey, model: model, timeout: timeout, provider: "gemini"}
}

func (c *geminiClient) baseURL() string {
	return fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", c.model)
}

func (c *geminiClient) TestAPIKey(ctx context.Context) error {
	if c.apiKey == "" {
		return fmt.Errorf("API key is empty")
	}
	body := []byte(`{"contents":[{"parts":[{"text":"Hi"}]}]}`)
	url := fmt.Sprintf("%s?key=%s", c.baseURL(), c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: c.timeout}).Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("invalid API key")
	default:
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
}

func (c *geminiClient) ChatJSON(ctx context.Context, systemPrompt, userPrompt string, temperature float64) (string, error) {
	full := strings.TrimSpace(systemPrompt)
	if userPrompt != "" {
		full += "\n\n" + strings.TrimSpace(userPrompt)
	}
	payload := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]string{{"text": full}}},
		},
		"generationConfig": map[string]any{
			"temperature":      temperature,
			"responseMimeType": "application/json",
		},
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s?key=%s", c.baseURL(), c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := (&http.Client{Timeout: c.timeout}).Do(req)
	if err != nil {
		return "", fmt.Errorf("AI request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API status %d: %s", resp.StatusCode, truncate(string(body), 500))
	}
	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty content in AI response")
	}
	content := strings.TrimSpace(parsed.Candidates[0].Content.Parts[0].Text)
	content = stripJSONFences(content)
	if content == "" {
		return "", fmt.Errorf("empty content in AI response")
	}
	return content, nil
}

// NewClient builds a provider client for the given slug and model.
func NewClient(provider, apiKey, model string, timeout time.Duration) Client {
	switch provider {
	case "gemini":
		return newGemini(apiKey, model, timeout)
	case "grok":
		return newOpenAICompat(apiKey, "https://api.x.ai/v1/chat/completions", model, provider, timeout)
	case "openrouter":
		return newOpenAICompat(apiKey, "https://openrouter.ai/api/v1/chat/completions", model, provider, timeout)
	default:
		return newOpenAICompat(apiKey, "https://api.openai.com/v1/chat/completions", model, "openai", timeout)
	}
}

// ProviderSettingKey returns the composite settings key for a provider slug.
func ProviderSettingKey(provider string) string {
	switch provider {
	case "gemini":
		return "GEMINI"
	case "grok":
		return "GROK"
	case "openrouter":
		return "OPENROUTER"
	default:
		return "OPENAI"
	}
}

func stripJSONFences(content string) string {
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}
	return strings.TrimSpace(content)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
