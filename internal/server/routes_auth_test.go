package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	authperms "github.com/piplos/piplos.media/internal/auth"
	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
	authsvc "github.com/piplos/piplos.media/internal/services/auth"
)

type routeCase struct {
	method string
	path   string
}

var (
	publicRoutes = []routeCase{
		{http.MethodGet, "/v1/public/projects"},
		{http.MethodGet, "/v1/public/services"},
		{http.MethodGet, "/v1/public/services/web"},
		{http.MethodGet, "/v1/public/stack"},
		{http.MethodGet, "/v1/public/seo"},
		{http.MethodGet, "/v1/public/pages"},
		{http.MethodGet, "/v1/public/legal"},
		{http.MethodGet, "/v1/public/languages"},
		{http.MethodPost, "/v1/leads"},
		{http.MethodPost, "/v1/auth/login"},
		{http.MethodPost, "/v1/auth/refresh"},
	}

	staffRoutes = []routeCase{
		{http.MethodGet, "/v1/projects"},
		{http.MethodPost, "/v1/projects"},
		{http.MethodPost, "/v1/projects/reorder"},
		{http.MethodPost, "/v1/projects/reorder-global"},
		{http.MethodGet, "/v1/services"},
		{http.MethodPost, "/v1/services"},
		{http.MethodGet, "/v1/stack"},
		{http.MethodPost, "/v1/stack"},
		{http.MethodGet, "/v1/seo"},
		{http.MethodPost, "/v1/seo"},
		{http.MethodGet, "/v1/pages"},
		{http.MethodPost, "/v1/pages"},
		{http.MethodGet, "/v1/legal"},
		{http.MethodGet, "/v1/leads"},
		{http.MethodPost, "/v1/uploads"},
		{http.MethodGet, "/v1/files"},
		{http.MethodPost, "/v1/files/folders"},
		{http.MethodPost, "/v1/files/rename"},
		{http.MethodPost, "/v1/files/move"},
		{http.MethodPost, "/v1/files/delete"},
		{http.MethodPost, "/v1/translate"},
		{http.MethodGet, "/v1/languages"},
		{http.MethodGet, "/v1/auth/me"},
	}

	adminRoutes = []routeCase{
		{http.MethodGet, "/v1/users"},
		{http.MethodPost, "/v1/users"},
		{http.MethodGet, "/v1/settings"},
		{http.MethodPut, "/v1/settings/SMTP"},
		{http.MethodPost, "/v1/settings/test"},
		{http.MethodPost, "/v1/languages"},
		{http.MethodGet, "/v1/ai-models"},
		{http.MethodPost, "/v1/ai-models"},
	}
)

func routeSuffix(path string) string {
	return strings.TrimPrefix(path, "/v1")
}

type fakeSessionChecker struct {
	validSessions map[string]bool
}

func (f *fakeSessionChecker) IsSessionValid(_ context.Context, sessionID string) (bool, error) {
	return f.validSessions[sessionID], nil
}

func newAuthTestApp(t *testing.T) (*fiber.App, *authsvc.Service, *fakeSessionChecker) {
	t.Helper()

	cfg := &config.Config{
		JWTSecret:               "test-jwt-secret-with-enough-length!!",
		JWTExpirationMinutes:    15,
		JWTRefreshExpirationHrs: 168,
	}
	authService := authsvc.New(cfg)
	sessions := &fakeSessionChecker{validSessions: map[string]bool{
		"admin-session":   true,
		"manager-session": true,
	}}
	authMw := middleware.NewAuth(authService, sessions)

	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	ok := func(c fiber.Ctx) error { return c.SendStatus(http.StatusNoContent) }

	api := app.Group("/v1")
	api.Post("/leads", ok)
	pub := api.Group("/public")
	pub.Get("/projects", ok)
	pub.Get("/projects/:slug", ok)
	pub.Get("/services", ok)
	pub.Get("/services/:slug", ok)
	pub.Get("/stack", ok)
	pub.Get("/seo", ok)
	pub.Get("/pages", ok)
	pub.Get("/pages/:slug", ok)
	pub.Get("/legal", ok)
	pub.Get("/legal/:slug", ok)
	pub.Get("/languages", ok)

	api.Post("/auth/login", ok)
	api.Post("/auth/refresh", ok)
	authn := api.Group("", authMw.RequireAuth())
	authn.Get("/auth/me", ok)

	staff := api.Group("", authMw.RequireAuth(), authMw.RequireRole(authperms.StaffRoles...))
	for _, r := range staffRoutes {
		registerProbe(staff, routeSuffix(r.path), r.method, ok)
	}

	adm := api.Group("", authMw.RequireAuth(), authMw.RequireRole(authperms.AdminRoles...))
	for _, r := range adminRoutes {
		registerProbe(adm, routeSuffix(r.path), r.method, ok)
	}

	return app, authService, sessions
}

func registerProbe(router fiber.Router, path, method string, h fiber.Handler) {
	switch method {
	case http.MethodGet:
		router.Get(path, h)
	case http.MethodPost:
		router.Post(path, h)
	case http.MethodPut:
		router.Put(path, h)
	case http.MethodPatch:
		router.Patch(path, h)
	case http.MethodDelete:
		router.Delete(path, h)
	}
}

func doRequest(t *testing.T, app *fiber.App, method, path, token string, body string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func tokenFor(t *testing.T, auth *authsvc.Service, user *models.User, sessionID string) string {
	t.Helper()
	token, err := auth.GenerateAccessToken(user, sessionID)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func TestStaffAndAdminRoutesRequireAuth(t *testing.T) {
	app, _, _ := newAuthTestApp(t)

	for _, routes := range [][]routeCase{staffRoutes, adminRoutes} {
		for _, r := range routes {
			resp := doRequest(t, app, r.method, r.path, "", "")
			if resp.StatusCode != http.StatusUnauthorized {
				t.Fatalf("%s %s without auth: got %d, want 401", r.method, r.path, resp.StatusCode)
			}
		}
	}
}

func TestAdminRoutesRejectManager(t *testing.T) {
	app, authService, _ := newAuthTestApp(t)
	manager := &models.User{ID: "manager-id", Email: "manager@test.com", Role: models.RoleManager, IsActive: true}
	token := tokenFor(t, authService, manager, "manager-session")

	for _, r := range adminRoutes {
		resp := doRequest(t, app, r.method, r.path, token, "")
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("%s %s as manager: got %d, want 403", r.method, r.path, resp.StatusCode)
		}
	}
}

func TestStaffRoutesAllowManager(t *testing.T) {
	app, authService, _ := newAuthTestApp(t)
	manager := &models.User{ID: "manager-id", Email: "manager@test.com", Role: models.RoleManager, IsActive: true}
	token := tokenFor(t, authService, manager, "manager-session")

	for _, r := range staffRoutes {
		resp := doRequest(t, app, r.method, r.path, token, "")
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("%s %s as manager: got %d, want 204", r.method, r.path, resp.StatusCode)
		}
	}
}

func TestAdminRoutesAllowAdmin(t *testing.T) {
	app, authService, _ := newAuthTestApp(t)
	admin := &models.User{ID: "admin-id", Email: "admin@test.com", Role: models.RoleAdmin, IsActive: true}
	token := tokenFor(t, authService, admin, "admin-session")

	for _, r := range adminRoutes {
		resp := doRequest(t, app, r.method, r.path, token, "")
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("%s %s as admin: got %d, want 204", r.method, r.path, resp.StatusCode)
		}
	}
}

func TestPublicRoutesDoNotRequireAuth(t *testing.T) {
	app, _, _ := newAuthTestApp(t)

	for _, r := range publicRoutes {
		body := ""
		if r.path == "/v1/leads" {
			body = `{"types":["web"],"first_name":"T","email":"t@test.com","lang":"en"}`
		}
		resp := doRequest(t, app, r.method, r.path, "", body)
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			t.Fatalf("%s %s should be public, got %d", r.method, r.path, resp.StatusCode)
		}
	}
}

func TestInvalidTokenRejected(t *testing.T) {
	app, authService, sessions := newAuthTestApp(t)
	admin := &models.User{ID: "admin-id", Email: "admin@test.com", Role: models.RoleAdmin, IsActive: true}
	token := tokenFor(t, authService, admin, "admin-session")

	resp := doRequest(t, app, http.MethodGet, "/v1/users", "not-a-jwt", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("invalid token: got %d, want 401", resp.StatusCode)
	}

	// Revoked session should reject valid JWT.
	sessions.validSessions["admin-session"] = false
	resp = doRequest(t, app, http.MethodGet, "/v1/users", token, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("revoked session: got %d, want 401", resp.StatusCode)
	}
}

func TestTokenWithoutSessionIDRejected(t *testing.T) {
	app, authService, _ := newAuthTestApp(t)
	admin := &models.User{ID: "admin-id", Email: "admin@test.com", Role: models.RoleAdmin, IsActive: true}
	// Empty sid — middleware must reject.
	token, err := authService.GenerateAccessToken(admin, "")
	if err != nil {
		t.Fatal(err)
	}
	resp := doRequest(t, app, http.MethodGet, "/v1/users", token, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("empty sid: got %d, want 401", resp.StatusCode)
	}
}

func TestUnknownSessionRejected(t *testing.T) {
	app, authService, _ := newAuthTestApp(t)
	admin := &models.User{ID: "admin-id", Email: "admin@test.com", Role: models.RoleAdmin, IsActive: true}
	token := tokenFor(t, authService, admin, "unknown-session")
	resp := doRequest(t, app, http.MethodGet, "/v1/users", token, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unknown session: got %d, want 401", resp.StatusCode)
	}
}
