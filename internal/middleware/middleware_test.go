package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
	authsvc "github.com/piplos/piplos.media/internal/services/auth"
)

type fakeSessionChecker struct {
	valid map[string]bool
	calls int
}

func (f *fakeSessionChecker) IsSessionValid(_ context.Context, sessionID string) (bool, error) {
	f.calls++
	return f.valid[sessionID], nil
}

func TestRequireAuthUsesSessionCheckerNotUserLookup(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-jwt-secret-with-enough-length!!", JWTExpirationMinutes: 15}
	authSvc := authsvc.New(cfg)
	sessions := &fakeSessionChecker{valid: map[string]bool{"sid-1": true}}
	authMw := middleware.NewAuth(authSvc, sessions)

	user := &models.User{ID: "u1", Email: "a@test.com", Role: models.RoleAdmin, IsActive: true}
	token, err := authSvc.GenerateAccessToken(user, "sid-1")
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Get("/protected", authMw.RequireAuth(), func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("got %d, want 204", resp.StatusCode)
	}
	if sessions.calls != 1 {
		t.Fatalf("IsSessionValid calls: got %d, want 1", sessions.calls)
	}
}

func TestRequireAuthRejectsRevokedSession(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-jwt-secret-with-enough-length!!", JWTExpirationMinutes: 15}
	authSvc := authsvc.New(cfg)
	sessions := &fakeSessionChecker{valid: map[string]bool{"sid-1": false}}
	authMw := middleware.NewAuth(authSvc, sessions)

	user := &models.User{ID: "u1", Email: "a@test.com", Role: models.RoleAdmin, IsActive: true}
	token, err := authSvc.GenerateAccessToken(user, "sid-1")
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Get("/protected", authMw.RequireAuth(), func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", resp.StatusCode)
	}
}

func TestRequireAuthRejectsMissingSessionID(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-jwt-secret-with-enough-length!!", JWTExpirationMinutes: 15}
	authSvc := authsvc.New(cfg)
	sessions := &fakeSessionChecker{valid: map[string]bool{}}
	authMw := middleware.NewAuth(authSvc, sessions)

	user := &models.User{ID: "u1", Email: "a@test.com", Role: models.RoleAdmin, IsActive: true}
	token, err := authSvc.GenerateAccessToken(user, "")
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Get("/protected", authMw.RequireAuth(), func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", resp.StatusCode)
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-jwt-secret-with-enough-length!!", JWTExpirationMinutes: 15}
	authSvc := authsvc.New(cfg)
	sessions := &fakeSessionChecker{valid: map[string]bool{"sid-1": true}}
	authMw := middleware.NewAuth(authSvc, sessions)

	user := &models.User{ID: "u1", Email: "m@test.com", Role: models.RoleManager, IsActive: true}
	token, err := authSvc.GenerateAccessToken(user, "sid-1")
	if err != nil {
		t.Fatal(err)
	}

	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Get("/admin", authMw.RequireAuth(), authMw.RequireRole(models.RoleAdmin), func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("got %d, want 403", resp.StatusCode)
	}
}
