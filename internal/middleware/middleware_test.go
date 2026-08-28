package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
	authsvc "github.com/piplos/piplos.media/internal/services/auth"
)

const testJWTSecret = "test-jwt-secret-with-enough-length!!"

type fakeSessionChecker struct {
	valid map[string]bool
	calls int
}

func (f *fakeSessionChecker) IsSessionValid(_ context.Context, sessionID string) (bool, error) {
	f.calls++
	return f.valid[sessionID], nil
}

func newTestAuth(t *testing.T, sessions *fakeSessionChecker) (*authsvc.Service, *middleware.Auth) {
	t.Helper()
	cfg := &config.Config{JWTSecret: testJWTSecret, JWTExpirationMinutes: 15}
	authSvc := authsvc.New(cfg)
	return authSvc, middleware.NewAuth(authSvc, sessions, nil)
}

func protectedApp(t *testing.T, authMw *middleware.Auth, afterAuth ...fiber.Handler) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	chain := []any{authMw.RequireAuth()}
	for _, h := range afterAuth {
		chain = append(chain, h)
	}
	chain = append(chain, func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})
	app.Get("/protected", chain[0], chain[1:]...)
	return app
}

func requestStatusAt(t *testing.T, app *fiber.App, path, header string) int {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if header != "" {
		req.Header.Set("Authorization", header)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func requestStatus(t *testing.T, app *fiber.App, header string) int {
	t.Helper()
	return requestStatusAt(t, app, "/protected", header)
}

func accessToken(t *testing.T, authSvc *authsvc.Service, user *models.User, sessionID string) string {
	t.Helper()
	token, err := authSvc.GenerateAccessToken(user, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

// craftToken signs custom access-shaped claims, for cases the service API cannot produce.
func craftToken(t *testing.T, secret string, mutate func(*authsvc.Claims)) string {
	t.Helper()
	claims := authsvc.Claims{
		UserID: "u1", Email: "a@test.com", Role: models.RoleAdmin,
		Type: "access", SessionID: "sid-1",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	if mutate != nil {
		mutate(&claims)
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestRequireAuthUsesSessionCheckerNotUserLookup(t *testing.T) {
	sessions := &fakeSessionChecker{valid: map[string]bool{"sid-1": true}}
	authSvc, authMw := newTestAuth(t, sessions)

	user := &models.User{ID: "u1", Email: "a@test.com", Role: models.RoleAdmin, IsActive: true}
	token := accessToken(t, authSvc, user, "sid-1")

	status := requestStatus(t, protectedApp(t, authMw), "Bearer "+token)
	if status != http.StatusNoContent {
		t.Fatalf("got %d, want 204", status)
	}
	if sessions.calls != 1 {
		t.Fatalf("IsSessionValid calls: got %d, want 1", sessions.calls)
	}
}

func TestRequireAuthRejectsRevokedSession(t *testing.T) {
	sessions := &fakeSessionChecker{valid: map[string]bool{"sid-1": false}}
	authSvc, authMw := newTestAuth(t, sessions)

	user := &models.User{ID: "u1", Email: "a@test.com", Role: models.RoleAdmin, IsActive: true}
	token := accessToken(t, authSvc, user, "sid-1")

	status := requestStatus(t, protectedApp(t, authMw), "Bearer "+token)
	if status != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", status)
	}
}

func TestRequireAuthRejectsMissingSessionID(t *testing.T) {
	sessions := &fakeSessionChecker{valid: map[string]bool{}}
	authSvc, authMw := newTestAuth(t, sessions)

	user := &models.User{ID: "u1", Email: "a@test.com", Role: models.RoleAdmin, IsActive: true}
	token := accessToken(t, authSvc, user, "")

	status := requestStatus(t, protectedApp(t, authMw), "Bearer "+token)
	if status != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", status)
	}
}

func TestRequireAuthRejectsRefreshTypeToken(t *testing.T) {
	sessions := &fakeSessionChecker{valid: map[string]bool{"sid-1": true}}
	_, authMw := newTestAuth(t, sessions)

	// A refresh-typed token (legacy stateless refresh) must never authenticate,
	// even with an otherwise valid signature, expiry and session ID.
	token := craftToken(t, testJWTSecret, func(c *authsvc.Claims) {
		c.Type = "refresh"
	})

	status := requestStatus(t, protectedApp(t, authMw), "Bearer "+token)
	if status != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", status)
	}
	if sessions.calls != 0 {
		t.Fatalf("IsSessionValid calls: got %d, want 0 (rejected before session lookup)", sessions.calls)
	}
}

func TestRequireAuthRejectsExpiredAccessToken(t *testing.T) {
	sessions := &fakeSessionChecker{valid: map[string]bool{"sid-1": true}}
	_, authMw := newTestAuth(t, sessions)

	token := craftToken(t, testJWTSecret, func(c *authsvc.Claims) {
		c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
	})

	status := requestStatus(t, protectedApp(t, authMw), "Bearer "+token)
	if status != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", status)
	}
}

func TestRequireAuthRejectsForeignSignature(t *testing.T) {
	sessions := &fakeSessionChecker{valid: map[string]bool{"sid-1": true}}
	_, authMw := newTestAuth(t, sessions)

	token := craftToken(t, "another-secret-with-enough-length!", nil)

	status := requestStatus(t, protectedApp(t, authMw), "Bearer "+token)
	if status != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", status)
	}
}

func TestRequireAuthRejectsMalformedAuthorizationHeader(t *testing.T) {
	sessions := &fakeSessionChecker{valid: map[string]bool{"sid-1": true}}
	authSvc, authMw := newTestAuth(t, sessions)

	user := &models.User{ID: "u1", Email: "a@test.com", Role: models.RoleAdmin, IsActive: true}
	token := accessToken(t, authSvc, user, "sid-1")
	app := protectedApp(t, authMw)

	cases := []struct{ name, header string }{
		{"missing header", ""},
		{"scheme only", "Bearer"},
		{"wrong scheme", "Basic " + token},
		{"extra parts", "Bearer " + token + " extra"},
		{"double space", "Bearer  " + token},
		{"not a jwt", "Bearer garbage"},
	}
	for _, tc := range cases {
		if status := requestStatus(t, app, tc.header); status != http.StatusUnauthorized {
			t.Errorf("%s: got %d, want 401", tc.name, status)
		}
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	sessions := &fakeSessionChecker{valid: map[string]bool{"sid-1": true}}
	authSvc, authMw := newTestAuth(t, sessions)

	user := &models.User{ID: "u1", Email: "m@test.com", Role: models.RoleManager, IsActive: true}
	token := accessToken(t, authSvc, user, "sid-1")

	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Get("/admin", authMw.RequireAuth(), authMw.RequireRole(models.RoleAdmin), func(c fiber.Ctx) error {
		return c.SendStatus(http.StatusNoContent)
	})

	status := requestStatusAt(t, app, "/admin", "Bearer "+token)
	if status != http.StatusForbidden {
		t.Fatalf("got %d, want 403", status)
	}
}
