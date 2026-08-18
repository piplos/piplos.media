package server

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

	authperms "github.com/piplos/piplos.media/internal/auth"
	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/handlers"
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

// --- Auth handler integration (login / refresh / inactive) ---

type authRouteUserStore struct {
	byEmail map[string]*models.User
	byID    map[string]*models.User
}

func (m *authRouteUserStore) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	return m.byEmail[strings.ToLower(email)], nil
}

func (m *authRouteUserStore) GetUserByID(_ context.Context, id string) (*models.User, error) {
	return m.byID[id], nil
}

type authRouteSessionStore struct {
	byHash map[string]*models.RefreshSession
	byID   map[string]*models.RefreshSession
	nextID int
}

func (m *authRouteSessionStore) IsSessionValid(_ context.Context, sessionID string) (bool, error) {
	s, ok := m.byID[sessionID]
	if !ok || s.RevokedAt != nil || time.Now().After(s.ExpiresAt) {
		return false, nil
	}
	return true, nil
}

func (m *authRouteSessionStore) CreateRefreshSession(_ context.Context, userID, tokenHash string, expiresAt time.Time, rotatedFrom *string) (string, error) {
	m.nextID++
	id := fmt.Sprintf("sess-%d", m.nextID)
	s := &models.RefreshSession{ID: id, UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt, RotatedFrom: rotatedFrom}
	m.byHash[tokenHash] = s
	m.byID[id] = s
	return id, nil
}

func (m *authRouteSessionStore) GetSessionByTokenHash(_ context.Context, tokenHash string) (*models.RefreshSession, error) {
	return m.byHash[tokenHash], nil
}

func (m *authRouteSessionStore) RevokeSession(_ context.Context, sessionID string) error {
	if s, ok := m.byID[sessionID]; ok && s.RevokedAt == nil {
		now := time.Now()
		s.RevokedAt = &now
	}
	return nil
}

func (m *authRouteSessionStore) RevokeSessionByTokenHash(_ context.Context, tokenHash string) error {
	if s, ok := m.byHash[tokenHash]; ok && s.RevokedAt == nil {
		now := time.Now()
		s.RevokedAt = &now
	}
	return nil
}

type authTokenResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *models.User `json:"user"`
}

func newAuthHandlerRouteApp(t *testing.T, users *authRouteUserStore) (*fiber.App, *authsvc.Service, *authRouteSessionStore) {
	t.Helper()
	cfg := &config.Config{
		JWTSecret:               "test-jwt-secret-with-enough-length!!",
		JWTExpirationMinutes:    15,
		JWTRefreshExpirationHrs: 168,
	}
	authService := authsvc.New(cfg)
	sessions := &authRouteSessionStore{byHash: map[string]*models.RefreshSession{}, byID: map[string]*models.RefreshSession{}}
	authHandler := handlers.NewAuthHandler(authService, users, sessions, false)
	authMw := middleware.NewAuth(authService, sessions)

	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	ok := func(c fiber.Ctx) error { return c.SendStatus(http.StatusNoContent) }

	api := app.Group("/v1")
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/refresh", authHandler.Refresh)
	authn := api.Group("", authMw.RequireAuth())
	authn.Get("/auth/me", authHandler.Me)
	authn.Post("/auth/logout", authHandler.Logout)
	authn.Get("/users", ok)

	return app, authService, sessions
}

func decodeAuthJSON(t *testing.T, r io.Reader, dest any) {
	t.Helper()
	raw, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
}

func TestAuthRouteLoginSuccess(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:               "test-jwt-secret-with-enough-length!!",
		JWTExpirationMinutes:    15,
		JWTRefreshExpirationHrs: 168,
	}
	authSvc := authsvc.New(cfg)
	hash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	user := &models.User{
		ID: "user-1", Email: "admin@test.com", PasswordHash: hash,
		FullName: "Admin", Role: models.RoleAdmin, IsActive: true,
	}
	users := &authRouteUserStore{
		byEmail: map[string]*models.User{"admin@test.com": user},
		byID:    map[string]*models.User{"user-1": user},
	}
	app, authService, sessions := newAuthHandlerRouteApp(t, users)

	resp := doRequest(t, app, http.MethodPost, "/v1/auth/login", "", `{"email":"admin@test.com","password":"password123"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login: got %d, want 200", resp.StatusCode)
	}
	var out authTokenResponse
	decodeAuthJSON(t, resp.Body, &out)
	if out.AccessToken == "" || out.RefreshToken == "" || out.User == nil {
		t.Fatal("missing tokens or user")
	}
	if len(sessions.byHash) != 1 {
		t.Fatalf("sessions: got %d, want 1", len(sessions.byHash))
	}
	claims, err := authService.ValidateToken(out.AccessToken)
	if err != nil || claims.SessionID == "" {
		t.Fatalf("invalid access token: %v", err)
	}
}

func TestAuthRouteLoginInactiveUser(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:               "test-jwt-secret-with-enough-length!!",
		JWTExpirationMinutes:    15,
		JWTRefreshExpirationHrs: 168,
	}
	authSvc := authsvc.New(cfg)
	hash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	user := &models.User{
		ID: "user-2", Email: "inactive@test.com", PasswordHash: hash,
		FullName: "Inactive", Role: models.RoleAdmin, IsActive: false,
	}
	users := &authRouteUserStore{
		byEmail: map[string]*models.User{"inactive@test.com": user},
		byID:    map[string]*models.User{"user-2": user},
	}
	app, _, _ := newAuthHandlerRouteApp(t, users)

	resp := doRequest(t, app, http.MethodPost, "/v1/auth/login", "", `{"email":"inactive@test.com","password":"password123"}`)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("inactive login: got %d, want 403", resp.StatusCode)
	}
}

func TestAuthRouteRefreshRotatesToken(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:               "test-jwt-secret-with-enough-length!!",
		JWTExpirationMinutes:    15,
		JWTRefreshExpirationHrs: 168,
	}
	authSvc := authsvc.New(cfg)
	hash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	user := &models.User{
		ID: "user-1", Email: "admin@test.com", PasswordHash: hash,
		FullName: "Admin", Role: models.RoleAdmin, IsActive: true,
	}
	users := &authRouteUserStore{
		byEmail: map[string]*models.User{"admin@test.com": user},
		byID:    map[string]*models.User{"user-1": user},
	}
	app, authService, sessions := newAuthHandlerRouteApp(t, users)

	loginResp := doRequest(t, app, http.MethodPost, "/v1/auth/login", "", `{"email":"admin@test.com","password":"password123"}`)
	var loginOut authTokenResponse
	decodeAuthJSON(t, loginResp.Body, &loginOut)

	refreshBody := `{"refresh_token":"` + loginOut.RefreshToken + `"}`
	refreshResp := doRequest(t, app, http.MethodPost, "/v1/auth/refresh", "", refreshBody)
	if refreshResp.StatusCode != http.StatusOK {
		t.Fatalf("refresh: got %d, want 200", refreshResp.StatusCode)
	}
	var refreshOut authTokenResponse
	decodeAuthJSON(t, refreshResp.Body, &refreshOut)
	if refreshOut.RefreshToken == loginOut.RefreshToken {
		t.Fatal("refresh token not rotated")
	}

	oldHash := authService.HashRefreshToken(loginOut.RefreshToken)
	if s := sessions.byHash[oldHash]; s == nil || s.RevokedAt == nil {
		t.Fatal("old session should be revoked")
	}

	reuseResp := doRequest(t, app, http.MethodPost, "/v1/auth/refresh", "", refreshBody)
	if reuseResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("reused refresh: got %d, want 401", reuseResp.StatusCode)
	}
}

func TestAuthRouteFullFlowLoginMeRefreshLogout(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:               "test-jwt-secret-with-enough-length!!",
		JWTExpirationMinutes:    15,
		JWTRefreshExpirationHrs: 168,
	}
	authSvc := authsvc.New(cfg)
	hash, err := authSvc.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	user := &models.User{
		ID: "user-1", Email: "admin@test.com", PasswordHash: hash,
		FullName: "Admin", Role: models.RoleAdmin, IsActive: true,
	}
	users := &authRouteUserStore{
		byEmail: map[string]*models.User{"admin@test.com": user},
		byID:    map[string]*models.User{"user-1": user},
	}
	app, authService, sessions := newAuthHandlerRouteApp(t, users)

	loginResp := doRequest(t, app, http.MethodPost, "/v1/auth/login", "", `{"email":"admin@test.com","password":"password123"}`)
	var loginOut authTokenResponse
	decodeAuthJSON(t, loginResp.Body, &loginOut)

	meResp := doRequest(t, app, http.MethodGet, "/v1/auth/me", loginOut.AccessToken, "")
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("me: got %d, want 200", meResp.StatusCode)
	}

	refreshResp := doRequest(t, app, http.MethodPost, "/v1/auth/refresh", "", `{"refresh_token":"`+loginOut.RefreshToken+`"}`)
	if refreshResp.StatusCode != http.StatusOK {
		t.Fatalf("refresh: got %d, want 200", refreshResp.StatusCode)
	}
	var refreshOut authTokenResponse
	decodeAuthJSON(t, refreshResp.Body, &refreshOut)

	logoutResp := doRequest(t, app, http.MethodPost, "/v1/auth/logout", refreshOut.AccessToken, `{"refresh_token":"`+refreshOut.RefreshToken+`"}`)
	if logoutResp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: got %d, want 204", logoutResp.StatusCode)
	}

	tokenHash := authService.HashRefreshToken(refreshOut.RefreshToken)
	if s := sessions.byHash[tokenHash]; s == nil || s.RevokedAt == nil {
		t.Fatal("session not revoked after logout")
	}

	postLogout := doRequest(t, app, http.MethodPost, "/v1/auth/refresh", "", `{"refresh_token":"`+refreshOut.RefreshToken+`"}`)
	if postLogout.StatusCode != http.StatusUnauthorized {
		t.Fatalf("refresh after logout: got %d, want 401", postLogout.StatusCode)
	}
}
