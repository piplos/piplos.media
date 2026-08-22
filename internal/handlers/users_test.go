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

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/handlers"
	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
	"github.com/piplos/piplos.media/internal/repository"
	authsvc "github.com/piplos/piplos.media/internal/services/auth"
)

// fakeUsersStore implements handlers.UserStore in memory.
type fakeUsersStore struct {
	byID    map[string]*models.User
	byEmail map[string]*models.User

	createErr error
	updateErr error
	deleteErr error

	nextID      int
	revocations int
}

func newFakeUsersStore(users ...*models.User) *fakeUsersStore {
	f := &fakeUsersStore{byID: map[string]*models.User{}, byEmail: map[string]*models.User{}}
	for _, u := range users {
		f.byID[u.ID] = u
		f.byEmail[strings.ToLower(u.Email)] = u
	}
	return f
}

func (f *fakeUsersStore) ListUsers(_ context.Context) ([]models.User, error) {
	out := []models.User{}
	for _, u := range f.byID {
		out = append(out, *u)
	}
	return out, nil
}

func (f *fakeUsersStore) GetUserByEmail(_ context.Context, email string) (*models.User, error) {
	if u := f.byEmail[strings.ToLower(email)]; u != nil {
		c := *u
		return &c, nil
	}
	return nil, nil
}

func (f *fakeUsersStore) GetUserByID(_ context.Context, id string) (*models.User, error) {
	if u := f.byID[id]; u != nil {
		c := *u
		return &c, nil
	}
	return nil, nil
}

func (f *fakeUsersStore) CreateUser(_ context.Context, email, passwordHash, fullName string, role models.UserRole, notifyLeads bool) (*models.User, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.nextID++
	u := &models.User{
		ID: fmt.Sprintf("u-%d", f.nextID), Email: email, PasswordHash: passwordHash,
		FullName: fullName, Role: role, IsActive: true, NotifyLeads: notifyLeads,
	}
	f.byID[u.ID] = u
	f.byEmail[strings.ToLower(email)] = u
	return u, nil
}

func (f *fakeUsersStore) UpdateUser(_ context.Context, id, fullName string, role models.UserRole, isActive bool, notifyLeads bool, passwordHash string) (*models.User, error) {
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	u := f.byID[id]
	if u == nil {
		return nil, nil
	}
	u.FullName = fullName
	u.Role = role
	u.IsActive = isActive
	u.NotifyLeads = notifyLeads
	if passwordHash != "" {
		u.PasswordHash = passwordHash
	}
	c := *u
	return &c, nil
}

func (f *fakeUsersStore) DeleteUser(_ context.Context, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	delete(f.byID, id)
	return nil
}

func (f *fakeUsersStore) RevokeAllUserSessions(_ context.Context, _ string) error {
	f.revocations++
	return nil
}

func newUsersTestApp(t *testing.T, store *fakeUsersStore) *fiber.App {
	t.Helper()
	cfg := &config.Config{JWTSecret: "test-jwt-secret-with-enough-length!!", JWTExpirationMinutes: 15}
	h := handlers.NewUsersHandler(authsvc.New(cfg), store)

	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	asAdmin := func(c fiber.Ctx) error {
		c.Locals("user", &models.User{ID: "admin-1", Email: "admin@test.com", Role: models.RoleAdmin, IsActive: true})
		return c.Next()
	}
	app.Post("/v1/users", asAdmin, h.Create)
	app.Put("/v1/users/:id", asAdmin, h.Update)
	app.Delete("/v1/users/:id", asAdmin, h.Delete)
	return app
}

func doUsersRequest(t *testing.T, app *fiber.App, method, path, body string) (*http.Response, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatalf("decode %s: %v", raw, err)
		}
	}
	return resp, out
}

func assertUsersStatus(t *testing.T, resp *http.Response, wantStatus int) {
	t.Helper()
	if resp.StatusCode != wantStatus {
		t.Fatalf("status: got %d, want %d", resp.StatusCode, wantStatus)
	}
}

func TestCreateUserMapsDuplicateEmailToConflict(t *testing.T) {
	store := newFakeUsersStore()
	store.createErr = fmt.Errorf("create user: %w", repository.ErrDuplicateEmail)
	app := newUsersTestApp(t, store)

	// Гонка: предварительная проверка email прошла, INSERT нарушил unique-индекс.
	resp, out := doUsersRequest(t, app, http.MethodPost, "/v1/users",
		`{"email":"dup@test.com","password":"12345678","role":"manager"}`)
	assertUsersStatus(t, resp, http.StatusConflict)
	if out["error"] != "conflict" {
		t.Fatalf("error code: got %v, want conflict", out["error"])
	}
}

func TestCreateUserPreCheckDuplicateStillConflict(t *testing.T) {
	existing := &models.User{ID: "u-1", Email: "dup@test.com", Role: models.RoleManager, IsActive: true}
	app := newUsersTestApp(t, newFakeUsersStore(existing))

	resp, out := doUsersRequest(t, app, http.MethodPost, "/v1/users",
		`{"email":"dup@test.com","password":"12345678","role":"manager"}`)
	assertUsersStatus(t, resp, http.StatusConflict)
	if out["error"] != "conflict" {
		t.Fatalf("error code: got %v, want conflict", out["error"])
	}
}

func TestCreateUserSuccess(t *testing.T) {
	app := newUsersTestApp(t, newFakeUsersStore())

	resp, _ := doUsersRequest(t, app, http.MethodPost, "/v1/users",
		`{"email":"new@test.com","password":"12345678","role":"manager","full_name":"New Manager"}`)
	assertUsersStatus(t, resp, http.StatusCreated)
}

func TestUpdateUserMapsLastActiveAdminToConflict(t *testing.T) {
	victim := &models.User{ID: "u-9", Email: "victim@test.com", Role: models.RoleAdmin, IsActive: true}
	store := newFakeUsersStore(victim)
	store.updateErr = fmt.Errorf("update user: %w", repository.ErrLastActiveAdmin)
	app := newUsersTestApp(t, store)

	resp, out := doUsersRequest(t, app, http.MethodPut, "/v1/users/u-9",
		`{"role":"manager","is_active":true}`)
	assertUsersStatus(t, resp, http.StatusConflict)
	if out["error"] != "conflict" {
		t.Fatalf("error code: got %v, want conflict", out["error"])
	}
}

func TestDeleteUserMapsLastActiveAdminToConflict(t *testing.T) {
	store := newFakeUsersStore()
	store.deleteErr = repository.ErrLastActiveAdmin
	app := newUsersTestApp(t, store)

	resp, out := doUsersRequest(t, app, http.MethodDelete, "/v1/users/u-9", "")
	assertUsersStatus(t, resp, http.StatusConflict)
	if out["error"] != "conflict" {
		t.Fatalf("error code: got %v, want conflict", out["error"])
	}
}

func TestUpdateUserRevokesSessionsOnRoleChange(t *testing.T) {
	manager := &models.User{ID: "mgr-1", Email: "m@test.com", Role: models.RoleManager, IsActive: true}
	store := newFakeUsersStore(manager)
	app := newUsersTestApp(t, store)

	resp, _ := doUsersRequest(t, app, http.MethodPut, "/v1/users/mgr-1",
		`{"role":"admin","full_name":"Promoted"}`)
	assertUsersStatus(t, resp, http.StatusOK)
	if store.revocations != 1 {
		t.Fatalf("revocations: got %d, want 1", store.revocations)
	}
}

func TestDeleteMissingUserStillOK(t *testing.T) {
	app := newUsersTestApp(t, newFakeUsersStore())

	resp, out := doUsersRequest(t, app, http.MethodDelete, "/v1/users/nope", "")
	assertUsersStatus(t, resp, http.StatusOK)
	if out["ok"] != true {
		t.Fatalf("ok: got %v, want true", out["ok"])
	}
}

func TestSelfDemoteStillRejectedBeforeRepo(t *testing.T) {
	app := newUsersTestApp(t, newFakeUsersStore())

	// admin-1 понижает сам себя — UX-защита хендлера остаётся.
	req := httptest.NewRequest(http.MethodPut, "/v1/users/admin-1",
		strings.NewReader(`{"role":"manager","is_active":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", resp.StatusCode)
	}
}
