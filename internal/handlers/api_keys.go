package handlers

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos/piplos.media/internal/errors"
	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
	"github.com/piplos/piplos.media/internal/utils"
)

// APIKeyStore abstracts API key persistence (implemented by *repository.Repository).
type APIKeyStore interface {
	CreateAPIKey(ctx context.Context, name, keyHash, keyPrefix string, createdBy *string) (*models.APIKey, error)
	ListAPIKeys(ctx context.Context) ([]models.APIKey, error)
	RevokeAPIKey(ctx context.Context, id string) (*models.APIKey, error)
	DeleteAPIKey(ctx context.Context, id string) error
}

// APIKeysHandler manages agent API keys (admin only).
type APIKeysHandler struct {
	repo APIKeyStore
}

// NewAPIKeysHandler creates an APIKeysHandler.
func NewAPIKeysHandler(repo APIKeyStore) *APIKeysHandler {
	return &APIKeysHandler{repo: repo}
}

// List returns all API keys (hashes are excluded from the JSON shape).
func (h *APIKeysHandler) List(c fiber.Ctx) error {
	keys, err := h.repo.ListAPIKeys(c.Context())
	if err != nil {
		return internalErr("failed to list api keys", err)
	}
	return c.JSON(fiber.Map{"api_keys": keys})
}

// Create generates a new key. The raw key is returned exactly once — only its
// SHA-256 hash is persisted.
func (h *APIKeysHandler) Create(c fiber.Ctx) error {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return apperrors.ErrInvalidRequest("name is required")
	}
	if len(req.Name) > 100 {
		return apperrors.ErrInvalidRequest("name must be at most 100 characters")
	}

	raw, hash, prefix, err := utils.GenerateAPIKey()
	if err != nil {
		return internalErr("failed to generate api key", err)
	}
	var createdBy *string
	if user := middleware.CurrentUser(c); user != nil {
		createdBy = &user.ID
	}
	created, err := h.repo.CreateAPIKey(c.Context(), req.Name, hash, prefix, createdBy)
	if err != nil {
		return internalErr("failed to create api key", err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"api_key": created, "key": raw})
}

// Revoke soft-revokes a key (sets revoked_at); the row stays for audit.
func (h *APIKeysHandler) Revoke(c fiber.Ctx) error {
	revoked, err := h.repo.RevokeAPIKey(c.Context(), c.Params("id"))
	if err != nil {
		return internalErr("failed to revoke api key", err)
	}
	if revoked == nil {
		return apperrors.ErrNotFound("api key not found or already revoked")
	}
	return c.JSON(fiber.Map{"api_key": revoked})
}

// Delete hard-deletes a key row.
func (h *APIKeysHandler) Delete(c fiber.Ctx) error {
	if err := h.repo.DeleteAPIKey(c.Context(), c.Params("id")); err != nil {
		return internalErr("failed to delete api key", err)
	}
	return c.JSON(fiber.Map{"ok": true})
}
