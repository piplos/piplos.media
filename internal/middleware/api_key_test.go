package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
	"github.com/piplos/piplos.media/internal/utils"
)

type fakeAPIKeyChecker struct {
	byHash  map[string]*models.APIKey
	touched []string
}

func newFakeAPIKeyChecker(keys ...*models.APIKey) *fakeAPIKeyChecker {
	f := &fakeAPIKeyChecker{byHash: map[string]*models.APIKey{}}
	for _, k := range keys {
		// The middleware hashes whatever bearer token is presented; the fake
		// registry indexes keys by the hash of "pk_live_<id>".
		f.byHash[utils.HashAPIKey("pk_live_"+k.ID)] = k
	}
	return f
}

func (f *fakeAPIKeyChecker) GetAPIKeyByHash(_ context.Context, keyHash string) (*models.APIKey, error) {
	return f.byHash[keyHash], nil
}

func (f *fakeAPIKeyChecker) TouchAPIKeyLastUsed(_ context.Context, id string) error {
	f.touched = append(f.touched, id)
	return nil
}

func apiKeyTestApp(t *testing.T, checker *fakeAPIKeyChecker) *fiber.App {
	t.Helper()
	// JWT parts are nil: the agent middleware never touches them.
	authMw := middleware.NewAuth(nil, nil, checker)
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Get("/agent", authMw.RequireAPIKey(), func(c fiber.Ctx) error {
		key := middleware.CurrentAPIKey(c)
		raw, _ := json.Marshal(map[string]string{"id": key.ID})
		c.Set("Content-Type", "application/json")
		return c.Send(raw)
	})
	return app
}

func requestAgentStatus(t *testing.T, app *fiber.App, header string) (*http.Response, int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/agent", nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp, resp.StatusCode
}

func TestRequireAPIKeyAcceptsValidKey(t *testing.T) {
	key := &models.APIKey{ID: "key-1", Name: "manus"}
	checker := newFakeAPIKeyChecker(key)
	app := apiKeyTestApp(t, checker)

	resp, status := requestAgentStatus(t, app, "Bearer pk_live_key-1")
	if status != http.StatusOK {
		t.Fatalf("got %d, want 200", status)
	}
	var out struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out.ID != "key-1" {
		t.Fatalf("locals api_key id = %q, want key-1", out.ID)
	}
	if len(checker.touched) != 1 || checker.touched[0] != "key-1" {
		t.Fatalf("TouchAPIKeyLastUsed calls = %v, want [key-1]", checker.touched)
	}
}

func TestRequireAPIKeyRejections(t *testing.T) {
	revokedAt := time.Now()
	key := &models.APIKey{ID: "key-1", Name: "manus"}
	revokedKey := &models.APIKey{ID: "key-2", Name: "old", RevokedAt: &revokedAt}
	checker := newFakeAPIKeyChecker(key, revokedKey)
	app := apiKeyTestApp(t, checker)

	cases := []struct {
		name   string
		header string
	}{
		{"missing header", ""},
		{"scheme only", "Bearer"},
		{"wrong scheme", "Basic pk_live_key-1"},
		{"unknown key", "Bearer pk_live_nope"},
		{"revoked key", "Bearer pk_live_key-2"},
	}
	for _, tc := range cases {
		if _, status := requestAgentStatus(t, app, tc.header); status != http.StatusUnauthorized {
			t.Errorf("%s: got %d, want 401", tc.name, status)
		}
	}
	if len(checker.touched) != 0 {
		t.Fatalf("rejected requests must not touch last_used_at, got %v", checker.touched)
	}
}
