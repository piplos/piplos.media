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

	"github.com/piplos-media/site/internal/config"
	"github.com/piplos-media/site/internal/middleware"
	"github.com/piplos-media/site/internal/models"
	authsvc "github.com/piplos-media/site/internal/services/auth"
)

type routeCase struct {
	method string
	path   string
}

var (
	publicRoutes = []routeCase{
		{http.MethodGet, "/api/v1/public/projects"},
		{http.MethodGet, "/api/v1/public/services"},
		{http.MethodGet, "/api/v1/public/stack"},
		{http.MethodGet, "/api/v1/public/seo"},
		{http.MethodGet, "/api/v1/public/legal"},
		{http.MethodGet, "/api/v1/public/languages"},
		{http.MethodPost, "/api/v1/leads"},
		{http.MethodPost, "/api/v1/auth/login"},
		{http.MethodPost, "/api/v1/auth/refresh"},
	}

	staffRoutes = []routeCase{
		{http.MethodGet, "/api/v1/projects"},
		{http.MethodPost, "/api/v1/projects"},
		{http.MethodPost, "/api/v1/projects/reorder"},
		{http.MethodGet, "/api/v1/services"},
		{http.MethodPost, "/api/v1/services"},
		{http.MethodGet, "/api/v1/stack"},
		{http.MethodPost, "/api/v1/stack"},
		{http.MethodGet, "/api/v1/seo"},
		{http.MethodPost, "/api/v1/seo"},
		{http.MethodGet, "/api/v1/legal"},
		{http.MethodGet, "/api/v1/leads"},
		{http.MethodPost, "/api/v1/uploads"},
		{http.MethodPost, "/api/v1/translate"},
		{http.MethodGet, "/api/v1/languages"},
		{http.MethodGet, "/api/v1/auth/me"},
	}

	adminRoutes = []routeCase{
		{http.MethodGet, "/api/v1/users"},
		{http.MethodPost, "/api/v1/users"},
		{http.MethodGet, "/api/v1/settings"},
		{http.MethodPut, "/api/v1/settings/SMTP"},
		{http.MethodPost, "/api/v1/settings/test"},
		{http.MethodPost, "/api/v1/languages"},
		{http.MethodGet, "/api/v1/ai-models"},
		{http.MethodPost, "/api/v1/ai-models"},
	}
)

func routeSuffix(path string) string {
	return strings.TrimPrefix(path, "/api/v1")
}

type fakeUserLookup struct {
	users map[string]*models.User
}

func (f *fakeUserLookup) GetUserByID(_ context.Context, id string) (*models.User, error) {
	return f.users[id], nil
}

func newAuthTestApp(t *testing.T) (*fiber.App, *authsvc.Service, *fakeUserLookup) {
	t.Helper()

	cfg := &config.Config{
		JWTSecret:               "test-jwt-secret-with-enough-length!!",
		JWTExpirationMinutes:    15,
		JWTRefreshExpirationHrs: 168,
	}
	authService := authsvc.New(cfg)
	users := &fakeUserLookup{users: map[string]*models.User{
		"admin-id": {
			ID: "admin-id", Email: "admin@test.com", Role: models.RoleAdmin, IsActive: true,
		},
		"manager-id": {
			ID: "manager-id", Email: "manager@test.com", Role: models.RoleManager, IsActive: true,
		},
	}}
	authMw := middleware.NewAuth(authService, users)

	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	ok := func(c fiber.Ctx) error { return c.SendStatus(http.StatusNoContent) }

	api := app.Group("/api/v1")
	api.Post("/leads", ok)
	pub := api.Group("/public")
	pub.Get("/projects", ok)
	pub.Get("/services", ok)
	pub.Get("/stack", ok)
	pub.Get("/seo", ok)
	pub.Get("/legal", ok)
	pub.Get("/languages", ok)

	api.Post("/auth/login", ok)
	api.Post("/auth/refresh", ok)
	api.Get("/auth/me", authMw.RequireAuth(), ok)

	staff := api.Group("", authMw.RequireAuth(), authMw.RequireRole(models.RoleAdmin, models.RoleManager))
	for _, r := range staffRoutes {
		registerProbe(staff, routeSuffix(r.path), r.method, ok)
	}

	adm := api.Group("", authMw.RequireAuth(), authMw.RequireRole(models.RoleAdmin))
	for _, r := range adminRoutes {
		registerProbe(adm, routeSuffix(r.path), r.method, ok)
	}

	return app, authService, users
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

func tokenFor(t *testing.T, auth *authsvc.Service, user *models.User) string {
	t.Helper()
	token, _, err := auth.GenerateTokens(user)
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
	app, authService, users := newAuthTestApp(t)
	manager := users.users["manager-id"]
	token := tokenFor(t, authService, manager)

	for _, r := range adminRoutes {
		resp := doRequest(t, app, r.method, r.path, token, "")
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("%s %s as manager: got %d, want 403", r.method, r.path, resp.StatusCode)
		}
	}
}

func TestStaffRoutesAllowManager(t *testing.T) {
	app, authService, users := newAuthTestApp(t)
	manager := users.users["manager-id"]
	token := tokenFor(t, authService, manager)

	for _, r := range staffRoutes {
		resp := doRequest(t, app, r.method, r.path, token, "")
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("%s %s as manager: got %d, want 204", r.method, r.path, resp.StatusCode)
		}
	}
}

func TestAdminRoutesAllowAdmin(t *testing.T) {
	app, authService, users := newAuthTestApp(t)
	admin := users.users["admin-id"]
	token := tokenFor(t, authService, admin)

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
		if r.path == "/api/v1/leads" {
			body = `{"types":["web"],"first_name":"T","email":"t@test.com","lang":"en"}`
		}
		resp := doRequest(t, app, r.method, r.path, "", body)
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			t.Fatalf("%s %s should be public, got %d", r.method, r.path, resp.StatusCode)
		}
	}
}

func TestInvalidTokenRejected(t *testing.T) {
	app, authService, users := newAuthTestApp(t)
	admin := users.users["admin-id"]
	_, refreshToken, err := authService.GenerateTokens(admin)
	if err != nil {
		t.Fatal(err)
	}

	resp := doRequest(t, app, http.MethodGet, "/api/v1/users", "not-a-jwt", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("invalid token: got %d, want 401", resp.StatusCode)
	}

	resp = doRequest(t, app, http.MethodGet, "/api/v1/users", refreshToken, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("refresh token as access: got %d, want 401", resp.StatusCode)
	}
}
