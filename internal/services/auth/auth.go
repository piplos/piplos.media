// Package auth implements JWT generation/validation and password hashing.
package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/models"
)

// Service handles authentication operations.
type Service struct {
	cfg *config.Config
}

// Claims are JWT access token claims.
type Claims struct {
	UserID    string          `json:"user_id"`
	Email     string          `json:"email"`
	Role      models.UserRole `json:"role"`
	Type      string          `json:"type"`
	SessionID string          `json:"sid"`
	jwt.RegisteredClaims
}

// New creates an auth Service.
func New(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

// RefreshExpiration returns refresh token TTL.
func (s *Service) RefreshExpiration() time.Duration {
	return s.cfg.JWTRefreshExpiration()
}

// GenerateAccessToken returns a signed access JWT linked to sessionID.
func (s *Service) GenerateAccessToken(user *models.User, sessionID string) (string, error) {
	claims := Claims{
		UserID: user.ID, Email: user.Email, Role: user.Role,
		Type: "access", SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWTExpiration())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return token, nil
}

// NewRefreshToken returns a cryptographically random opaque refresh token.
func (s *Service) NewRefreshToken() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return id.String(), nil
}

// HashRefreshToken returns SHA-256 hex digest of an opaque refresh token.
func (s *Service) HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// ValidateToken parses and validates an access JWT.
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

// UserFromClaims builds a User from JWT claims (middleware hot path).
func UserFromClaims(claims *Claims) *models.User {
	return &models.User{
		ID:       claims.UserID,
		Email:    claims.Email,
		Role:     claims.Role,
		IsActive: true,
	}
}

// LegacyRefreshClaims are JWT refresh token claims (deprecated stateless refresh).
type LegacyRefreshClaims struct {
	UserID string          `json:"user_id"`
	Email  string          `json:"email"`
	Role   models.UserRole `json:"role"`
	Type   string          `json:"type"`
	jwt.RegisteredClaims
}

// ValidateLegacyRefreshToken parses a stateless refresh JWT (AUTH_LEGACY_REFRESH grace period).
func (s *Service) ValidateLegacyRefreshToken(tokenString string) (*LegacyRefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &LegacyRefreshClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse legacy refresh token: %w", err)
	}
	claims, ok := token.Claims.(*LegacyRefreshClaims)
	if !ok || !token.Valid || claims.Type != "refresh" {
		return nil, fmt.Errorf("invalid legacy refresh token")
	}
	return claims, nil
}

// HashPassword hashes a password using bcrypt.
func (s *Service) HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hashed), nil
}

// CheckPassword compares a bcrypt hash with a plain-text password.
func (s *Service) CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
