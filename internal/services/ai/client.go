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
	"unicode/utf8"
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
	provider string
	http     *http.Client
}

func newOpenAICompat(apiKey, baseURL, model, provider string, timeout time.Duration) *openAICompat {
	return &openAICompat{apiKey: apiKey, baseURL: baseURL, model: model, provider: provider, http: newHTTPClient(timeout)}
}

// newHTTPClient builds a client with the given overall request timeout
// (defaulting to 120s when unset) so connections are reused across requests.
func newHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &http.Client{Timeout: timeout}
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
	resp, err := c.http.Do(req)
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
	resp, err := c.http.Do(req)
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
	provider string
	http     *http.Client
	base     string // переопределяет базовый URL (для тестов); пусто — боевой endpoint
}

func newGemini(apiKey, model string, timeout time.Duration) *geminiClient {
	return &geminiClient{apiKey: apiKey, model: model, provider: "gemini", http: newHTTPClient(timeout)}
}

func (c *geminiClient) baseURL() string {
	if c.base != "" {
		return c.base
	}
	return fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", c.model)
}

// setAuth passes the API key via header instead of the ?key= query parameter:
// transport errors embed the full request URL and would leak the key into logs.
func (c *geminiClient) setAuth(req *http.Request) {
	req.Header.Set("x-goog-api-key", c.apiKey)
}

func (c *geminiClient) TestAPIKey(ctx context.Context) error {
	if c.apiKey == "" {
		return fmt.Errorf("API key is empty")
	}
	body := []byte(`{"contents":[{"parts":[{"text":"Hi"}]}]}`)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)
	resp, err := c.http.Do(req)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL(), bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)
	resp, err := c.http.Do(req)
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
	if !strings.HasPrefix(content, "```") {
		return content
	}
	after := content[3:]
	// Необязательная метка языка (```json, ```JSON, ...) до перевода строки.
	if len(after) >= 4 && strings.EqualFold(after[:4], "json") {
		after = after[4:]
	}
	after = strings.TrimPrefix(after, "\n")
	after = strings.TrimPrefix(after, "```")
	after = strings.TrimSuffix(after, "```")
	return strings.TrimSpace(after)
}

// truncate shortens s to at most n bytes without splitting a UTF-8 rune.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n] + "..."
}
