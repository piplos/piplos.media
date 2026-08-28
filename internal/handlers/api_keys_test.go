package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"github.com/piplos/piplos.media/internal/handlers"
	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
)

type fakeAPIKeyStore struct {
	byID   map[string]*models.APIKey
	byHash map[string]*models.APIKey
	nextID int
}

func newFakeAPIKeyStore(keys ...*models.APIKey) *fakeAPIKeyStore {
	f := &fakeAPIKeyStore{byID: map[string]*models.APIKey{}, byHash: map[string]*models.APIKey{}}
	for _, k := range keys {
		f.byID[k.ID] = k
		f.byHash[k.KeyHash] = k
	}
	return f
}

func (f *fakeAPIKeyStore) CreateAPIKey(_ context.Context, name, keyHash, keyPrefix string, createdBy *string) (*models.APIKey, error) {
	f.nextID++
	k := &models.APIKey{
		ID: fmt.Sprintf("k-%d", f.nextID), Name: name,
		KeyHash: keyHash, KeyPrefix: keyPrefix, CreatedBy: createdBy,
	}
	f.byID[k.ID] = k
	f.byHash[keyHash] = k
	return k, nil
}

func (f *fakeAPIKeyStore) ListAPIKeys(_ context.Context) ([]models.APIKey, error) {
	out := []models.APIKey{}
	for _, k := range f.byID {
		out = append(out, *k)
	}
	return out, nil
}

func (f *fakeAPIKeyStore) RevokeAPIKey(_ context.Context, id string) (*models.APIKey, error) {
	k := f.byID[id]
	if k == nil || k.RevokedAt != nil {
		return nil, nil
	}
	now := time.Now()
	k.RevokedAt = &now
	return k, nil
}

func (f *fakeAPIKeyStore) DeleteAPIKey(_ context.Context, id string) error {
	delete(f.byID, id)
	return nil
}

func newAPIKeysTestApp(t *testing.T, store *fakeAPIKeyStore) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	h := handlers.NewAPIKeysHandler(store)
	app.Get("/v1/api-keys", h.List)
	app.Post("/v1/api-keys", h.Create)
	app.Post("/v1/api-keys/:id/revoke", h.Revoke)
	app.Delete("/v1/api-keys/:id", h.Delete)
	return app
}

func doAPIKeyRequest(t *testing.T, app *fiber.App, method, path, body string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestAPIKeysCreateReturnsRawKeyOnce(t *testing.T) {
	store := newFakeAPIKeyStore()
	app := newAPIKeysTestApp(t, store)

	resp := doAPIKeyRequest(t, app, http.MethodPost, "/v1/api-keys", `{"name":"manus"}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: got %d, want 201", resp.StatusCode)
	}
	var out struct {
		APIKey *models.APIKey `json:"api_key"`
		Key    string         `json:"key"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	if out.Key == "" || !strings.HasPrefix(out.Key, "pk_live_") {
		t.Fatalf("raw key = %q, want pk_live_...", out.Key)
	}
	if out.APIKey == nil || out.APIKey.KeyHash != "" {
		t.Fatal("api_key must not expose the hash")
	}

	listResp := doAPIKeyRequest(t, app, http.MethodGet, "/v1/api-keys", "")
	rawList, _ := io.ReadAll(listResp.Body)
	if strings.Contains(string(rawList), out.Key) {
		t.Fatal("list response leaks the raw key")
	}
	if !strings.Contains(string(rawList), out.Key[:16]) {
		t.Fatal("list response should include the key prefix")
	}
}

func TestAPIKeysCreateValidation(t *testing.T) {
	store := newFakeAPIKeyStore()
	app := newAPIKeysTestApp(t, store)

	cases := []struct{ name, body string }{
		{"empty name", `{"name":"  "}`},
		{"missing name", `{}`},
		{"name too long", `{"name":"` + strings.Repeat("x", 101) + `"}`},
	}
	for _, tc := range cases {
		resp := doAPIKeyRequest(t, app, http.MethodPost, "/v1/api-keys", tc.body)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", tc.name, resp.StatusCode)
		}
	}
}

func TestAPIKeysRevokeAndDelete(t *testing.T) {
	key := &models.APIKey{ID: "k-1", Name: "manus", KeyHash: "hash-1"}
	store := newFakeAPIKeyStore(key)
	app := newAPIKeysTestApp(t, store)

	resp := doAPIKeyRequest(t, app, http.MethodPost, "/v1/api-keys/k-1/revoke", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("revoke: got %d, want 200", resp.StatusCode)
	}
	if store.byID["k-1"].RevokedAt == nil {
		t.Fatal("revoked_at must be set in the store")
	}
	resp = doAPIKeyRequest(t, app, http.MethodPost, "/v1/api-keys/k-1/revoke", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("revoke twice: got %d, want 404", resp.StatusCode)
	}
	resp = doAPIKeyRequest(t, app, http.MethodPost, "/v1/api-keys/unknown/revoke", "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("revoke unknown: got %d, want 404", resp.StatusCode)
	}

	resp = doAPIKeyRequest(t, app, http.MethodDelete, "/v1/api-keys/k-1", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete: got %d, want 200", resp.StatusCode)
	}
	if len(store.byID) != 0 {
		t.Fatal("key row must be gone after delete")
	}
}
