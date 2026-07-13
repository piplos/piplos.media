// Package auth implements JWT generation/validation and password hashing.
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/piplos-media/site/internal/config"
	"github.com/piplos-media/site/internal/models"
)

// Service handles authentication operations.
type Service struct {
	cfg *config.Config
}

// Claims are JWT token claims.
type Claims struct {
	UserID string          `json:"user_id"`
	Email  string          `json:"email"`
	Role   models.UserRole `json:"role"`
	Type   string          `json:"type"`
	jwt.RegisteredClaims
}

// New creates an auth Service.
func New(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) sign(user *models.User, tokenType string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: user.ID, Email: user.Email, Role: user.Role, Type: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("sign %s token: %w", tokenType, err)
	}
	return token, nil
}

// GenerateTokens returns access and refresh JWT tokens for a user.
func (s *Service) GenerateTokens(user *models.User) (accessToken, refreshToken string, err error) {
	if accessToken, err = s.sign(user, "access", s.cfg.JWTExpiration()); err != nil {
		return "", "", err
	}
	if refreshToken, err = s.sign(user, "refresh", s.cfg.JWTRefreshExpiration()); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

// ValidateToken parses and validates a JWT.
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
