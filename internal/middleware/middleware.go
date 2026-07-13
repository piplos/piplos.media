// Package middleware provides auth, CORS and error handling for Fiber.
package middleware

import (
	"errors"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/rs/zerolog"

	apperrors "github.com/piplos/site/internal/errors"
	"github.com/piplos/site/internal/models"
	authsvc "github.com/piplos/site/internal/services/auth"
)

// CORS returns configured CORS middleware.
func CORS(origins []string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: origins,
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	})
}

// ErrorHandler converts AppError (and unknown errors) into JSON responses.
func ErrorHandler(log zerolog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}
		var appErr *apperrors.AppError
		if !errors.As(err, &appErr) {
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				return c.Status(fiberErr.Code).JSON(fiber.Map{"error": "http_error", "message": fiberErr.Message})
			}
			appErr = apperrors.From(err)
		}
		if appErr.Status >= 500 {
			log.Error().Err(err).Str("path", c.Path()).Msg("request failed")
		}
		return c.Status(appErr.Status).JSON(fiber.Map{"error": appErr.Code, "message": appErr.Message})
	}
}

// Auth handles JWT authentication and role checks.
type Auth struct {
	authService *authsvc.Service
	repo        UserLookup
}

// NewAuth creates auth middleware.
func NewAuth(authService *authsvc.Service, repo UserLookup) *Auth {
	return &Auth{authService: authService, repo: repo}
}

// RequireAuth validates the Bearer token and stores the user in locals.
func (m *Auth) RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return apperrors.ErrUnauthorized("missing Authorization header")
		}
		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return apperrors.ErrUnauthorized("invalid Authorization format")
		}

		claims, err := m.authService.ValidateToken(parts[1])
		if err != nil || claims.Type != "access" {
			return apperrors.ErrUnauthorized("invalid token")
		}

		user, err := m.repo.GetUserByID(c.Context(), claims.UserID)
		if err != nil {
			return apperrors.ErrInternal("authentication failed")
		}
		if user == nil {
			return apperrors.ErrUnauthorized("user not found")
		}
		if !user.IsActive {
			return apperrors.ErrAccountDisabled("account is disabled")
		}

		c.Locals("user", user)
		return c.Next()
	}
}

// RequireRole allows only users with one of the given roles.
func (m *Auth) RequireRole(allowed ...models.UserRole) fiber.Handler {
	return func(c fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User)
		if !ok {
			return apperrors.ErrUnauthorized("unauthorized")
		}
		if !slices.Contains(allowed, user.Role) {
			return apperrors.ErrForbidden("insufficient permissions")
		}
		return c.Next()
	}
}

// CurrentUser returns the authenticated user from locals (nil if absent).
func CurrentUser(c fiber.Ctx) *models.User {
	user, _ := c.Locals("user").(*models.User)
	return user
}
