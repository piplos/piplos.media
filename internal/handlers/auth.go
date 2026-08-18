// Package handlers contains Fiber HTTP handlers for the admin API.
package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos/piplos.media/internal/errors"
	authperms "github.com/piplos/piplos.media/internal/auth"
	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
	authsvc "github.com/piplos/piplos.media/internal/services/auth"
)

// AuthUserStore loads users for login/refresh/me.
type AuthUserStore interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}

// AuthSessionStore manages server-side refresh sessions.
type AuthSessionStore interface {
	CreateRefreshSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time, rotatedFrom *string) (string, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error
}

// AuthHandler serves login/refresh/me/logout endpoints.
type AuthHandler struct {
	auth     *authsvc.Service
	users    AuthUserStore
	sessions AuthSessionStore
	legacy   bool
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(auth *authsvc.Service, users AuthUserStore, sessions AuthSessionStore, legacyRefresh bool) *AuthHandler {
	return &AuthHandler{auth: auth, users: users, sessions: sessions, legacy: legacyRefresh}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) issueTokens(ctx context.Context, user *models.User, rotatedFrom *string) (access, refresh, sessionID string, err error) {
	refresh, err = h.auth.NewRefreshToken()
	if err != nil {
		return "", "", "", err
	}
	hash := h.auth.HashRefreshToken(refresh)
	expiresAt := time.Now().Add(h.auth.RefreshExpiration())
	sessionID, err = h.sessions.CreateRefreshSession(ctx, user.ID, hash, expiresAt, rotatedFrom)
	if err != nil {
		return "", "", "", err
	}
	access, err = h.auth.GenerateAccessToken(user, sessionID)
	if err != nil {
		return "", "", "", err
	}
	return access, refresh, sessionID, nil
}

func (h *AuthHandler) tokenResponse(c fiber.Ctx, user *models.User, rotatedFrom *string) error {
	access, refresh, _, err := h.issueTokens(c.Context(), user, rotatedFrom)
	if err != nil {
		return apperrors.ErrInternal("token generation failed")
	}
	return c.JSON(fiber.Map{
		"access_token":  access,
		"refresh_token": refresh,
		"user":          user,
	})
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

	user, err := h.users.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return apperrors.ErrInternal("login failed")
	}
	if user == nil || !h.auth.CheckPassword(user.PasswordHash, req.Password) {
		return apperrors.ErrUnauthorized("invalid email or password")
	}
	if !user.IsActive {
		return apperrors.ErrAccountDisabled("account is disabled")
	}

	return h.tokenResponse(c, user, nil)
}

// Refresh exchanges a refresh token for a new token pair (with rotation).
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req refreshRequest
	if err := c.Bind().Body(&req); err != nil || req.RefreshToken == "" {
		return apperrors.ErrInvalidRequest("refresh_token is required")
	}

	session, user, err := h.resolveRefreshSession(c, req.RefreshToken)
	if err != nil {
		return err
	}
	if user == nil || !user.IsActive {
		return apperrors.ErrUnauthorized("user not found or disabled")
	}

	if session.ID != "" {
		if err := h.sessions.RevokeSession(c.Context(), session.ID); err != nil {
			return apperrors.ErrInternal("refresh failed")
		}
		rotatedFrom := session.ID
		return h.tokenResponse(c, user, &rotatedFrom)
	}

	// Legacy stateless refresh: issue new session without revoking.
	return h.tokenResponse(c, user, nil)
}

func (h *AuthHandler) resolveRefreshSession(c fiber.Ctx, refreshToken string) (*models.RefreshSession, *models.User, error) {
	hash := h.auth.HashRefreshToken(refreshToken)
	session, err := h.sessions.GetSessionByTokenHash(c.Context(), hash)
	if err != nil {
		return nil, nil, apperrors.ErrInternal("refresh failed")
	}
	if session != nil {
		if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
			return nil, nil, apperrors.ErrUnauthorized("invalid refresh token")
		}
		user, err := h.users.GetUserByID(c.Context(), session.UserID)
		if err != nil {
			return nil, nil, apperrors.ErrInternal("refresh failed")
		}
		return session, user, nil
	}

	if !h.legacy {
		return nil, nil, apperrors.ErrUnauthorized("invalid refresh token")
	}

	claims, err := h.auth.ValidateLegacyRefreshToken(refreshToken)
	if err != nil {
		return nil, nil, apperrors.ErrUnauthorized("invalid refresh token")
	}
	user, err := h.users.GetUserByID(c.Context(), claims.UserID)
	if err != nil {
		return nil, nil, apperrors.ErrInternal("refresh failed")
	}
	// Legacy path: no DB session to rotate; issue new session pair.
	return &models.RefreshSession{ID: ""}, user, nil
}

// Logout revokes the refresh session (idempotent).
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req logoutRequest
	_ = c.Bind().Body(&req)

	if req.RefreshToken != "" {
		hash := h.auth.HashRefreshToken(req.RefreshToken)
		if err := h.sessions.RevokeSessionByTokenHash(c.Context(), hash); err != nil {
			return apperrors.ErrInternal("logout failed")
		}
	} else if sid, ok := c.Locals("session_id").(string); ok && sid != "" {
		if err := h.sessions.RevokeSession(c.Context(), sid); err != nil {
			return apperrors.ErrInternal("logout failed")
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Me returns the authenticated user (fresh from DB for notify_leads etc.).
func (h *AuthHandler) Me(c fiber.Ctx) error {
	current := middleware.CurrentUser(c)
	if current == nil {
		return apperrors.ErrUnauthorized("unauthorized")
	}
	user, err := h.users.GetUserByID(c.Context(), current.ID)
	if err != nil {
		return apperrors.ErrInternal("failed to load user")
	}
	if user == nil {
		return apperrors.ErrUnauthorized("user not found")
	}
	return c.JSON(fiber.Map{"user": user})
}

// Permissions returns the read-only API permission matrix.
func (h *AuthHandler) Permissions(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"groups": fiber.Map{
			authperms.GroupStaff:         authperms.StaffRoles,
			authperms.GroupAdmin:         authperms.AdminRoles,
			authperms.GroupAuthenticated: authperms.AuthenticatedRoles,
		},
		"routes": authperms.APIRoutes,
	})
}
