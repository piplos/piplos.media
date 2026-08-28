package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos/piplos.media/internal/errors"
	"github.com/piplos/piplos.media/internal/models"
	"github.com/piplos/piplos.media/internal/utils"
)

// RequireAPIKey validates the agent API key (Authorization: Bearer pk_live_...)
// and stores *models.APIKey in locals under "api_key".
//
// Deliberately JWT-independent: agent routes never touch RequireAuth/RequireRole,
// so external automations do not need user accounts or refresh sessions.
// Keys have no expiry — they stay valid until revoked.
func (m *Auth) RequireAPIKey() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Misconfiguration check first: an unconfigured checker must 500
		// regardless of header shape.
		if m.apiKeys == nil {
			return apperrors.ErrInternal("api key checker is not configured")
		}
		header := c.Get("Authorization")
		if header == "" {
			return apperrors.ErrUnauthorized("missing Authorization header")
		}
		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return apperrors.ErrUnauthorized("invalid Authorization format")
		}

		key, err := m.apiKeys.GetAPIKeyByHash(c.Context(), utils.HashAPIKey(parts[1]))
		if err != nil {
			// Attach the cause so ErrorHandler logs the root failure.
			e := apperrors.ErrInternal("authentication failed")
			e.Cause = err
			return e
		}
		if key == nil {
			return apperrors.ErrUnauthorized("invalid api key")
		}
		if key.RevokedAt != nil {
			return apperrors.ErrUnauthorized("api key revoked")
		}

		c.Locals("api_key", key)
		_ = m.apiKeys.TouchAPIKeyLastUsed(c.Context(), key.ID) // throttled, best-effort
		return c.Next()
	}
}

// CurrentAPIKey returns the authenticated API key from locals (nil if absent).
func CurrentAPIKey(c fiber.Ctx) *models.APIKey {
	key, _ := c.Locals("api_key").(*models.APIKey)
	return key
}
