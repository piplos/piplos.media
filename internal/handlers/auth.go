// Package handlers contains Fiber HTTP handlers for the admin API.
package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos/site/internal/errors"
	"github.com/piplos/site/internal/middleware"
	"github.com/piplos/site/internal/repository"
	authsvc "github.com/piplos/site/internal/services/auth"
)

// AuthHandler serves login/refresh/me endpoints.
type AuthHandler struct {
	auth *authsvc.Service
	repo *repository.Repository
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(auth *authsvc.Service, repo *repository.Repository) *AuthHandler {
	return &AuthHandler{auth: auth, repo: repo}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Login authenticates by email/password and returns JWT tokens.
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req loginRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		return apperrors.ErrInvalidRequest("email and password are required")
	}

	user, err := h.repo.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return apperrors.ErrInternal("login failed")
	}
	if user == nil || !h.auth.CheckPassword(user.PasswordHash, req.Password) {
		return apperrors.ErrUnauthorized("invalid email or password")
	}
	if !user.IsActive {
		return apperrors.ErrAccountDisabled("account is disabled")
	}

	access, refresh, err := h.auth.GenerateTokens(user)
	if err != nil {
		return apperrors.ErrInternal("token generation failed")
	}
	return c.JSON(fiber.Map{
		"access_token":  access,
		"refresh_token": refresh,
		"user":          user,
	})
}

// Refresh exchanges a refresh token for a new token pair.
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req refreshRequest
	if err := c.Bind().Body(&req); err != nil || req.RefreshToken == "" {
		return apperrors.ErrInvalidRequest("refresh_token is required")
	}

	claims, err := h.auth.ValidateToken(req.RefreshToken)
	if err != nil || claims.Type != "refresh" {
		return apperrors.ErrUnauthorized("invalid refresh token")
	}

	user, err := h.repo.GetUserByID(c.Context(), claims.UserID)
	if err != nil {
		return apperrors.ErrInternal("refresh failed")
	}
	if user == nil || !user.IsActive {
		return apperrors.ErrUnauthorized("user not found or disabled")
	}

	access, refresh, err := h.auth.GenerateTokens(user)
	if err != nil {
		return apperrors.ErrInternal("token generation failed")
	}
	return c.JSON(fiber.Map{
		"access_token":  access,
		"refresh_token": refresh,
		"user":          user,
	})
}

// Me returns the authenticated user.
func (h *AuthHandler) Me(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"user": middleware.CurrentUser(c)})
}
