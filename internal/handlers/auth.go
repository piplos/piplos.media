// Package handlers contains Fiber HTTP handlers for the admin API.
package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	authperms "github.com/piplos/piplos.media/internal/auth"
	apperrors "github.com/piplos/piplos.media/internal/errors"
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

// AtomicRefreshSessionStore is the hardened AuthSessionStore contract used by
// refresh rotation: ClaimRefreshSession atomically revokes the session if it
// is still active and reports whether this call won the claim (RowsAffected>0),
// closing the TOCTOU window of check-then-revoke rotation; RevokeSessionChain
// revokes every active session rotated from one — the reuse-detection defense
// for leaked refresh tokens. The production repository implements it; stores
// that predate it transparently fall back to legacy check-then-revoke (see
// AuthHandler.refresh).
type AtomicRefreshSessionStore interface {
	AuthSessionStore
	ClaimRefreshSession(ctx context.Context, sessionID string) (bool, error)
	RevokeSessionChain(ctx context.Context, sessionID string) error
}

// AuthHandler serves login/refresh/me/logout endpoints.
type AuthHandler struct {
	auth     *authsvc.Service
	users    AuthUserStore
	sessions AuthSessionStore
	atomic   AtomicRefreshSessionStore // nil unless sessions supports atomic claims
	legacy   bool
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(auth *authsvc.Service, users AuthUserStore, sessions AuthSessionStore, legacyRefresh bool) *AuthHandler {
	atomic, _ := sessions.(AtomicRefreshSessionStore)
	return &AuthHandler{auth: auth, users: users, sessions: sessions, atomic: atomic, legacy: legacyRefresh}
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

// dummyBcryptHash hashes an unguessable string; compared against when the
// email is unknown so login takes the same time as the known-email path
// (otherwise response timing reveals which addresses exist).
const dummyBcryptHash = "$2a$10$5rdJOcuN.BkI3.Gn30J20Ok283B75V10dCwNO9ITYaudC4/Y6QMYy"

// replayGraceWindow separates token theft from a benign concurrent
// double-submit of the same refresh token: only reuse of a token revoked
// earlier than this window triggers chain revocation (see Refresh).
const replayGraceWindow = 5 * time.Second

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
		return internalErr("token generation failed", err)
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
		return internalErr("login failed", err)
	}
	if user == nil {
		h.auth.CheckPassword(dummyBcryptHash, req.Password)
		return apperrors.ErrUnauthorized("invalid email or password")
	}
	if !h.auth.CheckPassword(user.PasswordHash, req.Password) {
		return apperrors.ErrUnauthorized("invalid email or password")
	}
	if !user.IsActive {
		return apperrors.ErrAccountDisabled("account is disabled")
	}

	return h.tokenResponse(c, user, nil)
}

// Refresh exchanges a refresh token for a new token pair (with rotation).
//
// Rotation is guarded by an atomic claim (ClaimRefreshSession): the UPDATE
// flips revoked_at only while the row is still active, so of several
// concurrent refreshes with the same token exactly one wins and every loser
// is a replay rejected with 401. Presenting a token that was already rotated
// earlier additionally triggers reuse detection below. Session stores without
// atomic-claim support (see AtomicRefreshSessionStore) fall back to the
// legacy check-then-revoke rotation.
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req refreshRequest
	if err := c.Bind().Body(&req); err != nil || req.RefreshToken == "" {
		return apperrors.ErrInvalidRequest("refresh_token is required")
	}

	hash := h.auth.HashRefreshToken(req.RefreshToken)
	session, err := h.sessions.GetSessionByTokenHash(c.Context(), hash)
	if err != nil {
		return internalErr("refresh failed", err)
	}
	if session == nil {
		return h.refreshLegacy(c, req.RefreshToken)
	}

	if session.RevokedAt != nil {
		// Reuse detection: this token was already rotated/revoked before the
		// request — the classic refresh-theft signal. Kill every active
		// session derived from it so a thief cannot keep riding the chain.
		// A revocation younger than replayGraceWindow is more likely a benign
		// concurrent duplicate (HTTP retry racing the winning request), so it
		// gets no chain revocation. Best-effort either way: the 401 below is
		// enforced regardless of cleanup outcome.
		if h.atomic != nil && time.Since(*session.RevokedAt) > replayGraceWindow {
			_ = h.atomic.RevokeSessionChain(c.Context(), session.ID)
		}
		return apperrors.ErrUnauthorized("invalid refresh token")
	}
	if time.Now().After(session.ExpiresAt) {
		// Plain expiry without revocation is benign (client was offline):
		// reject, but do not treat it as theft.
		return apperrors.ErrUnauthorized("invalid refresh token")
	}

	user, err := h.users.GetUserByID(c.Context(), session.UserID)
	if err != nil {
		return internalErr("refresh failed", err)
	}
	if user == nil || !user.IsActive {
		return apperrors.ErrUnauthorized("user not found or disabled")
	}

	if h.atomic != nil {
		claimed, err := h.atomic.ClaimRefreshSession(c.Context(), session.ID)
		if err != nil {
			return internalErr("refresh failed", err)
		}
		if !claimed {
			// Lost the claim to a concurrent request with the same token.
			// Unlike the revoked-at-read path above this is not theft
			// evidence — a client may legitimately double-submit — so no
			// chain revocation.
			return apperrors.ErrUnauthorized("invalid refresh token")
		}
	} else if err := h.sessions.RevokeSession(c.Context(), session.ID); err != nil {
		// Legacy store: best-effort revoke before rotating (racy by design —
		// superseded by AtomicRefreshSessionStore).
		return internalErr("refresh failed", err)
	}

	rotatedFrom := session.ID
	return h.tokenResponse(c, user, &rotatedFrom)
}

// refreshLegacy serves the pre-sessions stateless refresh tokens when enabled.
func (h *AuthHandler) refreshLegacy(c fiber.Ctx, refreshToken string) error {
	if !h.legacy {
		return apperrors.ErrUnauthorized("invalid refresh token")
	}
	claims, err := h.auth.ValidateLegacyRefreshToken(refreshToken)
	if err != nil {
		return apperrors.ErrUnauthorized("invalid refresh token")
	}
	user, err := h.users.GetUserByID(c.Context(), claims.UserID)
	if err != nil {
		return internalErr("refresh failed", err)
	}
	// Legacy path: no DB session to rotate; issue new session pair.
	return h.tokenResponse(c, user, nil)
}

// Logout revokes the refresh session (idempotent).
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	var req logoutRequest
	_ = c.Bind().Body(&req)

	if req.RefreshToken != "" {
		hash := h.auth.HashRefreshToken(req.RefreshToken)
		if err := h.sessions.RevokeSessionByTokenHash(c.Context(), hash); err != nil {
			return internalErr("logout failed", err)
		}
	} else if sid, ok := c.Locals("session_id").(string); ok && sid != "" {
		if err := h.sessions.RevokeSession(c.Context(), sid); err != nil {
			return internalErr("logout failed", err)
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
		return internalErr("failed to load user", err)
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
