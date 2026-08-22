package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// jsonString кодирует строку как JSON-литерал (для сборки ответов с фенсами).
func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func TestStripJSONFences(t *testing.T) {
	cases := map[string]string{
		`{"a":1}`:                   `{"a":1}`,
		"```json\n{\"a\":1}\n```":   `{"a":1}`,
		"```JSON\n{\"a\":1}\n```":   `{"a":1}`, // регистр метки языка
		"```\n{\"a\":1}\n```":       `{"a":1}`,
		"```{\"a\":1}```":           `{"a":1}`, // без перевода строки
		"  ```json\n {\"a\":1} ```": `{"a":1}`,
		"text without fences":       "text without fences",
		"```json\nnot closed":       "not closed",
		"```js\n{\"a\":1}\n```":     "js\n{\"a\":1}", // незнакомая метка не срезается
	}
	for in, want := range cases {
		if got := stripJSONFences(in); got != want {
			t.Errorf("stripJSONFences(%q)=%q want %q", in, got, want)
		}
	}
}

func TestTruncateKeepsValidUTF8(t *testing.T) {
	s := strings.Repeat("ж", 300) // многобайтовые руны
	got := truncate(s, 10)
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected ellipsis suffix: %q", got)
	}
	body := strings.TrimSuffix(got, "...")
	if len(body) > 10 {
		t.Fatalf("body too long: %d bytes", len(body))
	}
	if strings.ContainsRune(body, '�') && len(body) == 0 {
		t.Fatal("empty body")
	}
}

// TestGeminiChatUsesHeaderNotURLQuery guards against API-key leakage through
// *url.Error messages (they embed the full request URL).
func TestGeminiChatUsesHeaderNotURLQuery(t *testing.T) {
	var gotKey, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("x-goog-api-key")
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"ok\":true}"}]}}]}`))
	}))
	defer srv.Close()

	c := &geminiClient{apiKey: "secret-key", model: "gemini-x", provider: "gemini", base: srv.URL, http: newHTTPClient(0)}
	content, err := c.ChatJSON(context.Background(), "sys", "user", 0.3)
	if err != nil {
		t.Fatalf("ChatJSON: %v", err)
	}
	if content != `{"ok":true}` {
		t.Fatalf("content: %q", content)
	}
	if gotKey != "secret-key" {
		t.Fatalf("x-goog-api-key header: %q", gotKey)
	}
	if gotQuery != "" {
		t.Fatalf("API key must not travel in the URL query, got %q", gotQuery)
	}
}

func TestGeminiTestAPIKeyUsesHeader(t *testing.T) {
	var gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("x-goog-api-key")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &geminiClient{apiKey: "secret-key", model: "gemini-x", provider: "gemini", base: srv.URL, http: newHTTPClient(0)}
	if err := c.TestAPIKey(context.Background()); err != nil {
		t.Fatalf("TestAPIKey: %v", err)
	}
	if gotKey != "secret-key" {
		t.Fatalf("x-goog-api-key header: %q", gotKey)
	}
}

func TestOpenAICompatChatJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer k" {
			t.Errorf("authorization header: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fenced := "```json\n{\"t\":\"перевод\"}\n```"
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":` + jsonString(fenced) + `}}]}`))
	}))
	defer srv.Close()

	c := newOpenAICompat("k", srv.URL, "m", "openai", 0)
	content, err := c.ChatJSON(context.Background(), "sys", "user", 0.3)
	if err != nil {
		t.Fatalf("ChatJSON: %v", err)
	}
	if content != `{"t":"перевод"}` {
		t.Fatalf("fences must be stripped: %q", content)
	}
}

func TestOpenAICompatErrorIncludesStatusAndTruncatedBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(strings.Repeat("ошибка", 200)))
	}))
	defer srv.Close()

	c := newOpenAICompat("k", srv.URL, "m", "openai", 0)
	_, err := c.ChatJSON(context.Background(), "sys", "user", 0.3)
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("expected status in error, got %v", err)
	}
	if len(err.Error()) > 600 {
		t.Fatalf("error body must be truncated, len=%d", len(err.Error()))
	}
}
