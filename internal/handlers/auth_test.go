package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
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

// memSessionStore mirrors the repository's session semantics, including the
// atomic claim: exactly one concurrent ClaimRefreshSession for a session wins.
type memSessionStore struct {
	mu     sync.Mutex
	byHash map[string]*models.RefreshSession
	byID   map[string]*models.RefreshSession
	nextID int
}

func (m *memSessionStore) activeLocked(s *models.RefreshSession) bool {
	return s != nil && s.RevokedAt == nil && time.Now().Before(s.ExpiresAt)
}

func (m *memSessionStore) IsSessionValid(_ context.Context, sessionID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.activeLocked(m.byID[sessionID]), nil
}

func (m *memSessionStore) CreateRefreshSession(_ context.Context, userID, tokenHash string, expiresAt time.Time, rotatedFrom *string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	id := fmt.Sprintf("sess-%d", m.nextID)
	s := &models.RefreshSession{ID: id, UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt, RotatedFrom: rotatedFrom}
	m.byHash[tokenHash] = s
	m.byID[id] = s
	return id, nil
}

func (m *memSessionStore) GetSessionByTokenHash(_ context.Context, tokenHash string) (*models.RefreshSession, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s := m.byHash[tokenHash]; s != nil {
		cp := *s // snapshot: callers must not observe later mutations
		return &cp, nil
	}
	return nil, nil
}

func (m *memSessionStore) revokeLocked(s *models.RefreshSession) {
	now := time.Now()
	s.RevokedAt = &now
}

func (m *memSessionStore) RevokeSession(_ context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s := m.byID[sessionID]; m.activeLocked(s) {
		m.revokeLocked(s)
	}
	return nil
}

func (m *memSessionStore) RevokeSessionByTokenHash(_ context.Context, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s := m.byHash[tokenHash]; m.activeLocked(s) {
		m.revokeLocked(s)
	}
	return nil
}

// ClaimRefreshSession atomically revokes the session if still active and
// reports whether this call won — the in-memory analogue of the repository's
// single-row UPDATE ... WHERE revoked_at IS NULL AND expires_at > now().
func (m *memSessionStore) ClaimRefreshSession(_ context.Context, sessionID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.byID[sessionID]
	if !m.activeLocked(s) {
		return false, nil
	}
	m.revokeLocked(s)
	return true, nil
}

// RevokeSessionChain revokes the session and every active descendant that was
// rotated from it, directly or transitively.
func (m *memSessionStore) RevokeSessionChain(_ context.Context, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	children := make(map[string][]*models.RefreshSession, len(m.byID))
	for _, s := range m.byID {
		if s.RotatedFrom != nil {
			children[*s.RotatedFrom] = append(children[*s.RotatedFrom], s)
		}
	}
	queue := []*models.RefreshSession{m.byID[sessionID]}
	for len(queue) > 0 {
		s := queue[0]
		queue = queue[1:]
		if s == nil {
			continue
		}
		if m.activeLocked(s) {
			m.revokeLocked(s)
		}
		queue = append(queue, children[s.ID]...)
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

func TestLoginUnknownEmailReturns401(t *testing.T) {
	h, _, _ := newAuthTestHandler(t)
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Post("/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"ghost@test.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unknown email: got %d, want 401", resp.StatusCode)
	}
}

// newAuthTestApp wires a fresh app with login+refresh routes over the shared
// test handler.
func newAuthTestApp(t *testing.T) (*fiber.App, *memSessionStore) {
	t.Helper()
	h, _, sessions := newAuthTestHandler(t)
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	app.Post("/login", h.Login)
	app.Post("/refresh", h.Refresh)
	return app, sessions
}

// loginAndDecode performs a successful login and returns its token pair.
func loginAndDecode(t *testing.T, app *fiber.App) tokenResponse {
	t.Helper()
	resp := postJSON(t, app, "/login", `{"email":"admin@test.com","password":"password123"}`)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status: %d", resp.StatusCode)
	}
	var out tokenResponse
	decodeBody(t, resp.Body, &out)
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatal("missing tokens in login response")
	}
	return out
}

func postJSON(t *testing.T, app *fiber.App, path, body string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestRefreshReusedRotatedTokenRejectedAndKillsChain(t *testing.T) {
	app, sessions := newAuthTestApp(t)

	t1 := loginAndDecode(t, app)

	rotate := func(token string) tokenResponse {
		t.Helper()
		resp := postJSON(t, app, "/refresh", `{"refresh_token":"`+token+`"}`)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("refresh status: %d", resp.StatusCode)
		}
		var out tokenResponse
		decodeBody(t, resp.Body, &out)
		if out.RefreshToken == token {
			t.Fatal("refresh token was not rotated")
		}
		return out
	}

	t2 := rotate(t1.RefreshToken) // sess-2 (rotated from sess-1)
	t3 := rotate(t2.RefreshToken) // sess-3 (rotated from sess-2)

	// Benign-looking immediate duplicate of the first rotation: rejected, but
	// inside replayGraceWindow it must NOT burn the still-valid chain.
	dupResp := postJSON(t, app, "/refresh", `{"refresh_token":"`+t1.RefreshToken+`"}`)
	if dupResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("duplicate refresh: got %d, want 401", dupResp.StatusCode)
	}
	// sess-2 is legitimately revoked (rotated into sess-3); the current live
	// token sess-3 must have survived the grace-window duplicate.
	if valid, _ := sessions.IsSessionValid(context.Background(), "sess-3"); !valid {
		t.Fatal("grace-window duplicate must not revoke the live chain")
	}

	// Simulate theft: the reused token was revoked long ago (outside the
	// grace window). Reuse must be rejected and every active session derived
	// from it must die.
	stale := time.Now().Add(-10 * time.Second)
	sessions.mu.Lock()
	sessions.byID["sess-1"].RevokedAt = &stale
	sessions.mu.Unlock()

	reuseResp := postJSON(t, app, "/refresh", `{"refresh_token":"`+t1.RefreshToken+`"}`)
	if reuseResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("reused refresh: got %d, want 401", reuseResp.StatusCode)
	}
	if valid, _ := sessions.IsSessionValid(context.Background(), "sess-3"); valid {
		t.Fatal("session sess-3 must be revoked by reuse detection")
	}

	// The winner tokens of the burned chain no longer refresh either.
	deadResp := postJSON(t, app, "/refresh", `{"refresh_token":"`+t3.RefreshToken+`"}`)
	if deadResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("chain descendant after reuse: got %d, want 401", deadResp.StatusCode)
	}
}

func TestRefreshConcurrentOnlyOneWins(t *testing.T) {
	app, sessions := newAuthTestApp(t)

	t1 := loginAndDecode(t, app)

	const racers = 8
	start := make(chan struct{})
	statuses := make([]int, racers)
	var wg sync.WaitGroup
	for i := 0; i < racers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			statuses[i] = postJSON(t, app, "/refresh",
				`{"refresh_token":"`+t1.RefreshToken+`"}`).StatusCode
		}(i)
	}
	close(start)
	wg.Wait()

	ok, unauthorized := 0, 0
	for _, st := range statuses {
		switch st {
		case http.StatusOK:
			ok++
		case http.StatusUnauthorized:
			unauthorized++
		default:
			t.Fatalf("unexpected status %d in race", st)
		}
	}
	if ok != 1 || unauthorized != racers-1 {
		t.Fatalf("race outcome: ok=%d unauthorized=%d, want 1/%d", ok, unauthorized, racers-1)
	}
	// Exactly one rotation happened: login session + one successor.
	sessions.mu.Lock()
	count := len(sessions.byID)
	var liveSuccessors int
	for _, s := range sessions.byID {
		if s.RotatedFrom != nil && sessions.activeLocked(s) {
			liveSuccessors++
		}
	}
	sessions.mu.Unlock()
	if count != 2 {
		t.Fatalf("sessions created = %d, want 2 (login + single winner)", count)
	}
	if liveSuccessors != 1 {
		t.Fatalf("live successor sessions = %d, want 1", liveSuccessors)
	}
}

func TestRefreshExpiredTokenRejectedKeepsChain(t *testing.T) {
	app, sessions := newAuthTestApp(t)
	_, authSvc, _ := newAuthTestHandler(t)

	// Expired parent with an active child: presenting the expired token is a
	// benign timeout, so the child must survive.
	expiredToken, err := authSvc.NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	parentID, err := sessions.CreateRefreshSession(context.Background(), "user-1",
		authSvc.HashRefreshToken(expiredToken), time.Now().Add(-time.Hour), nil)
	if err != nil {
		t.Fatal(err)
	}
	childToken, err := authSvc.NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	childHash := authSvc.HashRefreshToken(childToken)
	if _, err := sessions.CreateRefreshSession(context.Background(), "user-1", childHash,
		time.Now().Add(time.Hour), &parentID); err != nil {
		t.Fatal(err)
	}

	resp := postJSON(t, app, "/refresh", `{"refresh_token":"`+expiredToken+`"}`)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expired refresh: got %d, want 401", resp.StatusCode)
	}
	if valid, _ := sessions.IsSessionValid(context.Background(), parentID); valid {
		t.Fatal("expired parent must be invalid")
	}

	// The child session must remain usable: its own refresh succeeds.
	childResp := postJSON(t, app, "/refresh", `{"refresh_token":"`+childToken+`"}`)
	if childResp.StatusCode != http.StatusOK {
		t.Fatalf("child refresh after parent expiry: got %d, want 200", childResp.StatusCode)
	}
}
