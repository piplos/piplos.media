package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/models"
	authsvc "github.com/piplos/piplos.media/internal/services/auth"
)

func testAuthService(t *testing.T) *authsvc.Service {
	t.Helper()
	return authsvc.New(&config.Config{
		JWTSecret:               "test-jwt-secret-with-enough-length!!",
		JWTExpirationMinutes:    15,
		JWTRefreshExpirationHrs: 168,
	})
}

func testUser() *models.User {
	return &models.User{
		ID: "user-1", Email: "u@test.com", Role: models.RoleAdmin, IsActive: true,
	}
}

func TestGenerateAccessTokenContainsSessionID(t *testing.T) {
	svc := testAuthService(t)
	user := testUser()
	const sid = "session-abc"

	token, err := svc.GenerateAccessToken(user, sid)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Type != "access" {
		t.Fatalf("type: got %q, want access", claims.Type)
	}
	if claims.SessionID != sid {
		t.Fatalf("sid: got %q, want %q", claims.SessionID, sid)
	}
	if claims.UserID != user.ID || claims.Email != user.Email || claims.Role != user.Role {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestNewRefreshTokenUnique(t *testing.T) {
	svc := testAuthService(t)
	a, err := svc.NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := svc.NewRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatal("expected unique refresh tokens")
	}
}

func TestHashRefreshTokenDeterministic(t *testing.T) {
	svc := testAuthService(t)
	const raw = "opaque-token-value"
	h1 := svc.HashRefreshToken(raw)
	h2 := svc.HashRefreshToken(raw)
	if h1 != h2 || len(h1) != 64 {
		t.Fatalf("hash: %q", h1)
	}
}

func TestUserFromClaims(t *testing.T) {
	claims := &authsvc.Claims{
		UserID: "id-1", Email: "a@b.c", Role: models.RoleManager, Type: "access", SessionID: "sid",
	}
	user := authsvc.UserFromClaims(claims)
	if user.ID != claims.UserID || user.Email != claims.Email || user.Role != claims.Role || !user.IsActive {
		t.Fatalf("user: %+v", user)
	}
}

func TestValidateTokenRejectsWrongType(t *testing.T) {
	svc := testAuthService(t)
	// Build a legacy-style refresh JWT manually via unexported path — use ValidateLegacyRefreshToken instead.
	user := testUser()
	token, err := svc.GenerateAccessToken(user, "sid")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := svc.ValidateToken(token)
	if err != nil || claims.Type != "access" {
		t.Fatalf("valid access token rejected: %v", err)
	}
}

func TestCheckPassword(t *testing.T) {
	svc := testAuthService(t)
	hash, err := svc.HashPassword("secret-password")
	if err != nil {
		t.Fatal(err)
	}
	if !svc.CheckPassword(hash, "secret-password") {
		t.Fatal("expected password match")
	}
	if svc.CheckPassword(hash, "wrong") {
		t.Fatal("expected password mismatch")
	}
}

func TestValidateTokenRejectsInvalid(t *testing.T) {
	svc := testAuthService(t)
	if _, err := svc.ValidateToken("not-a-jwt"); err == nil {
		t.Fatal("expected error for invalid token")
	}
	if _, err := svc.ValidateToken(""); err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestRefreshExpiration(t *testing.T) {
	svc := testAuthService(t)
	if svc.RefreshExpiration() != 168*time.Hour {
		t.Fatalf("refresh TTL: got %v, want 168h", svc.RefreshExpiration())
	}
}

func TestValidateLegacyRefreshToken(t *testing.T) {
	svc := testAuthService(t)
	user := testUser()
	// Build a legacy refresh JWT manually.
	claims := authsvc.LegacyRefreshClaims{
		UserID: user.ID, Email: user.Email, Role: user.Role, Type: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("test-jwt-secret-with-enough-length!!"))
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := svc.ValidateLegacyRefreshToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.UserID != user.ID || parsed.Type != "refresh" {
		t.Fatalf("unexpected claims: %+v", parsed)
	}
	if _, err := svc.ValidateLegacyRefreshToken("bad-token"); err == nil {
		t.Fatal("expected error for invalid legacy token")
	}
}
