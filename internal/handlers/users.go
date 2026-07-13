package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos-media/site/internal/errors"
	"github.com/piplos-media/site/internal/middleware"
	"github.com/piplos-media/site/internal/models"
	"github.com/piplos-media/site/internal/repository"
	authsvc "github.com/piplos-media/site/internal/services/auth"
)

// UsersHandler manages admin panel accounts (admin role only).
type UsersHandler struct {
	auth *authsvc.Service
	repo *repository.Repository
}

// NewUsersHandler creates a UsersHandler.
func NewUsersHandler(auth *authsvc.Service, repo *repository.Repository) *UsersHandler {
	return &UsersHandler{auth: auth, repo: repo}
}

type userRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	IsActive *bool  `json:"is_active"`
}

func parseRole(role string) (models.UserRole, error) {
	switch models.UserRole(role) {
	case models.RoleAdmin, models.RoleManager:
		return models.UserRole(role), nil
	default:
		return "", apperrors.ErrInvalidRequest("role must be admin or manager")
	}
}

// List returns all users.
func (h *UsersHandler) List(c fiber.Ctx) error {
	users, err := h.repo.ListUsers(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to list users")
	}
	return c.JSON(fiber.Map{"users": users})
}

// Create adds a new user.
func (h *UsersHandler) Create(c fiber.Ctx) error {
	var req userRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || len(req.Password) < 8 {
		return apperrors.ErrInvalidRequest("email and password (min 8 chars) are required")
	}
	role, err := parseRole(req.Role)
	if err != nil {
		return err
	}

	existing, err := h.repo.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return apperrors.ErrInternal("failed to create user")
	}
	if existing != nil {
		return apperrors.ErrConflict("user with this email already exists")
	}

	hash, err := h.auth.HashPassword(req.Password)
	if err != nil {
		return apperrors.ErrInternal("failed to hash password")
	}
	user, err := h.repo.CreateUser(c.Context(), req.Email, hash, req.FullName, role)
	if err != nil {
		return apperrors.ErrInternal("failed to create user")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user": user})
}

// Update modifies name/role/active/password.
func (h *UsersHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var req userRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	role, err := parseRole(req.Role)
	if err != nil {
		return err
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Защита от самоблокировки/самопонижения последнего админа.
	current := middleware.CurrentUser(c)
	if current != nil && current.ID == id && (role != models.RoleAdmin || !isActive) {
		return apperrors.ErrInvalidRequest("you cannot demote or deactivate your own account")
	}

	hash := ""
	if req.Password != "" {
		if len(req.Password) < 8 {
			return apperrors.ErrInvalidRequest("password must be at least 8 chars")
		}
		if hash, err = h.auth.HashPassword(req.Password); err != nil {
			return apperrors.ErrInternal("failed to hash password")
		}
	}

	user, err := h.repo.UpdateUser(c.Context(), id, req.FullName, role, isActive, hash)
	if err != nil {
		return apperrors.ErrInternal("failed to update user")
	}
	if user == nil {
		return apperrors.ErrNotFound("user not found")
	}
	return c.JSON(fiber.Map{"user": user})
}

// Delete removes a user.
func (h *UsersHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	current := middleware.CurrentUser(c)
	if current != nil && current.ID == id {
		return apperrors.ErrInvalidRequest("you cannot delete your own account")
	}
	if err := h.repo.DeleteUser(c.Context(), id); err != nil {
		return apperrors.ErrInternal("failed to delete user")
	}
	return c.JSON(fiber.Map{"ok": true})
}
