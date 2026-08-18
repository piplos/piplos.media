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

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/handlers"
	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
	authsvc "github.com/piplos/piplos.media/internal/services/auth"
)

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type memUserStore struct {
	byEmail map[string]*models.User
	byID    map[string]*models.User
}

func (m *memUserStore) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	return m.byEmail[strings.ToLower(email)], nil
}

func (m *memUserStore) GetUserByID(_ context.Context, id string) (*models.User, error) {
	return m.byID[id], nil
}

type memSessionStore struct {
	byHash map[string]*models.RefreshSession
	byID   map[string]*models.RefreshSession
	nextID int
}

func (m *memSessionStore) IsSessionValid(_ context.Context, sessionID string) (bool, error) {
	s, ok := m.byID[sessionID]
	if !ok || s.RevokedAt != nil || time.Now().After(s.ExpiresAt) {
		return false, nil
	}
	return true, nil
}

func (m *memSessionStore) CreateRefreshSession(_ context.Context, userID, tokenHash string, expiresAt time.Time, rotatedFrom *string) (string, error) {
	m.nextID++
	id := fmt.Sprintf("sess-%d", m.nextID)
	s := &models.RefreshSession{ID: id, UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt, RotatedFrom: rotatedFrom}
	m.byHash[tokenHash] = s
	m.byID[id] = s
	return id, nil
}

func (m *memSessionStore) GetSessionByTokenHash(_ context.Context, tokenHash string) (*models.RefreshSession, error) {
	return m.byHash[tokenHash], nil
}

func (m *memSessionStore) RevokeSession(_ context.Context, sessionID string) error {
	if s, ok := m.byID[sessionID]; ok && s.RevokedAt == nil {
		now := time.Now()
		s.RevokedAt = &now
	}
	return nil
}

func (m *memSessionStore) RevokeSessionByTokenHash(_ context.Context, tokenHash string) error {
	if s, ok := m.byHash[tokenHash]; ok && s.RevokedAt == nil {
		now := time.Now()
		s.RevokedAt = &now
	}
	return nil
}

func newAuthTestHandler(t *testing.T) (*handlers.AuthHandler, *authsvc.Service, *memSessionStore) {
	t.Helper()
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
	users := &memUserStore{
		byEmail: map[string]*models.User{"admin@test.com": user},
		byID:    map[string]*models.User{"user-1": user},
	}
	sessions := &memSessionStore{byHash: map[string]*models.RefreshSession{}, byID: map[string]*models.RefreshSession{}}
	return handlers.NewAuthHandler(authSvc, users, sessions, false), authSvc, sessions
}

func TestLoginReturnsTokensAndCreatesSession(t *testing.T) {
	h, authSvc, sessions := newAuthTestHandler(t)
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Post("/login", h.Login)

	body := `{"email":"admin@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}

	var out tokenResponse
	decodeBody(t, resp.Body, &out)
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatal("missing tokens")
	}
	if len(sessions.byHash) != 1 {
		t.Fatalf("sessions created: %d, want 1", len(sessions.byHash))
	}

	claims, err := authSvc.ValidateToken(out.AccessToken)
	if err != nil || claims.SessionID == "" {
		t.Fatalf("access token invalid: %v", err)
	}
}

func TestLoginBadPassword(t *testing.T) {
	h, _, _ := newAuthTestHandler(t)
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Post("/login", h.Login)

	body := `{"email":"admin@test.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: %d, want 401", resp.StatusCode)
	}
}

func TestRefreshRotatesToken(t *testing.T) {
	h, authSvc, sessions := newAuthTestHandler(t)
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Post("/login", h.Login)
	app.Post("/refresh", h.Refresh)

	loginReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"admin@test.com","password":"password123"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	if err != nil {
		t.Fatal(err)
	}
	var loginOut tokenResponse
	decodeBody(t, loginResp.Body, &loginOut)

	refreshReq := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"`+loginOut.RefreshToken+`"}`))
	refreshReq.Header.Set("Content-Type", "application/json")
	refreshResp, err := app.Test(refreshReq)
	if err != nil {
		t.Fatal(err)
	}
	if refreshResp.StatusCode != http.StatusOK {
		t.Fatalf("refresh status: %d", refreshResp.StatusCode)
	}
	var refreshOut tokenResponse
	decodeBody(t, refreshResp.Body, &refreshOut)
	if refreshOut.RefreshToken == loginOut.RefreshToken {
		t.Fatal("refresh token was not rotated")
	}

	hash := authSvc.HashRefreshToken(loginOut.RefreshToken)
	if s := sessions.byHash[hash]; s == nil || s.RevokedAt == nil {
		t.Fatal("old session should be revoked")
	}

	reuseReq := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"`+loginOut.RefreshToken+`"}`))
	reuseReq.Header.Set("Content-Type", "application/json")
	reuseResp, err := app.Test(reuseReq)
	if err != nil {
		t.Fatal(err)
	}
	if reuseResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("reused refresh: got %d, want 401", reuseResp.StatusCode)
	}
}

func TestLoginInactiveUser(t *testing.T) {
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
	inactive := &models.User{
		ID: "user-2", Email: "inactive@test.com", PasswordHash: hash,
		FullName: "Inactive", Role: models.RoleAdmin, IsActive: false,
	}
	inactiveHandler := handlers.NewAuthHandler(authSvc, &memUserStore{
		byEmail: map[string]*models.User{"inactive@test.com": inactive},
		byID:    map[string]*models.User{"user-2": inactive},
	}, &memSessionStore{byHash: map[string]*models.RefreshSession{}, byID: map[string]*models.RefreshSession{}}, false)
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Post("/login", inactiveHandler.Login)

	body := `{"email":"inactive@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("inactive login: got %d, want 403", resp.StatusCode)
	}
}

func TestLoginMissingFields(t *testing.T) {
	h, _, _ := newAuthTestHandler(t)
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Post("/login", h.Login)

	for _, body := range []string{`{}`, `{"email":"a@test.com"}`, `{"password":"x"}`} {
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("body %s: got %d, want 400", body, resp.StatusCode)
		}
	}
}

func TestRefreshInactiveUser(t *testing.T) {
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
	inactive := &models.User{
		ID: "user-2", Email: "inactive@test.com", PasswordHash: hash,
		FullName: "Inactive", Role: models.RoleAdmin, IsActive: false,
	}
	sessions := &memSessionStore{byHash: map[string]*models.RefreshSession{}, byID: map[string]*models.RefreshSession{}}
	h := handlers.NewAuthHandler(authSvc, &memUserStore{
		byEmail: map[string]*models.User{"inactive@test.com": inactive},
		byID:    map[string]*models.User{"user-2": inactive},
	}, sessions, false)

	refreshToken, err := authSvc.NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	tokenHash := authSvc.HashRefreshToken(refreshToken)
	sid, err := sessions.CreateRefreshSession(context.Background(), inactive.ID, tokenHash, time.Now().Add(time.Hour), nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = sid

	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Post("/refresh", h.Refresh)

	req := httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(`{"refresh_token":"`+refreshToken+`"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("inactive refresh: got %d, want 401", resp.StatusCode)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	h, authSvc, sessions := newAuthTestHandler(t)
	authMw := middleware.NewAuth(authSvc, sessions)
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Post("/login", h.Login)
	app.Post("/logout", authMw.RequireAuth(), h.Logout)

	loginReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"admin@test.com","password":"password123"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	if err != nil {
		t.Fatal(err)
	}
	var loginOut tokenResponse
	decodeBody(t, loginResp.Body, &loginOut)

	logoutBody := `{"refresh_token":"` + loginOut.RefreshToken + `"}`
	logoutReq := httptest.NewRequest(http.MethodPost, "/logout", strings.NewReader(logoutBody))
	logoutReq.Header.Set("Content-Type", "application/json")
	logoutReq.Header.Set("Authorization", "Bearer "+loginOut.AccessToken)
	logoutResp, err := app.Test(logoutReq)
	if err != nil {
		t.Fatal(err)
	}
	if logoutResp.StatusCode != http.StatusNoContent {
		t.Fatalf("logout status: %d", logoutResp.StatusCode)
	}

	hash := authSvc.HashRefreshToken(loginOut.RefreshToken)
	if s := sessions.byHash[hash]; s == nil || s.RevokedAt == nil {
		t.Fatal("session not revoked after logout")
	}
}

func decodeBody(t *testing.T, r io.Reader, dest any) {
	t.Helper()
	raw, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
}
