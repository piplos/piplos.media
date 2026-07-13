package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	apperrors "github.com/piplos/site/internal/errors"
	"github.com/piplos/site/internal/models"
	"github.com/piplos/site/internal/repository"
)

// AIModelsHandler manages the ai_provider_models catalog.
type AIModelsHandler struct {
	repo *repository.Repository
}

// NewAIModelsHandler creates an AIModelsHandler.
func NewAIModelsHandler(repo *repository.Repository) *AIModelsHandler {
	return &AIModelsHandler{repo: repo}
}

type createAIModelRequest struct {
	Provider    string `json:"provider"`
	ModelID     string `json:"model_id"`
	DisplayName string `json:"display_name"`
}

// ListAIModels returns all catalog models.
func (h *AIModelsHandler) ListAIModels(c fiber.Ctx) error {
	list, err := h.repo.ListAIProviderModels(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to list AI models")
	}
	if list == nil {
		list = []models.AIProviderModel{}
	}
	return c.JSON(fiber.Map{"models": list})
}

// CreateAIModel adds a model to the catalog.
func (h *AIModelsHandler) CreateAIModel(c fiber.Ctx) error {
	var req createAIModelRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	req.Provider = strings.TrimSpace(strings.ToLower(req.Provider))
	req.ModelID = strings.TrimSpace(req.ModelID)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.Provider == "" || req.ModelID == "" || req.DisplayName == "" {
		return apperrors.ErrInvalidRequest("provider, model_id and display_name are required")
	}
	allowed := map[string]bool{"gemini": true, "grok": true, "openai": true, "openrouter": true}
	if !allowed[req.Provider] {
		return apperrors.ErrInvalidRequest("unknown provider")
	}
	model, err := h.repo.CreateAIProviderModel(c.Context(), req.Provider, req.ModelID, req.DisplayName)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return apperrors.ErrConflict("model already exists")
		}
		return apperrors.ErrInternal("failed to create AI model")
	}
	return c.Status(fiber.StatusCreated).JSON(model)
}

type updateAIModelRequest struct {
	DisplayName string `json:"display_name"`
	Enabled     *bool  `json:"enabled"`
}

// UpdateAIModel updates display name and/or enabled flag.
func (h *AIModelsHandler) UpdateAIModel(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.ErrInvalidRequest("invalid model id")
	}
	var req updateAIModelRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	if req.DisplayName == "" && req.Enabled == nil {
		return apperrors.ErrInvalidRequest("nothing to update")
	}
	found, err := h.repo.GetAIProviderModelByID(c.Context(), id)
	if err != nil {
		return apperrors.ErrInternal("failed to load AI model")
	}
	if found == nil {
		return apperrors.ErrNotFound("model not found")
	}
	displayName := req.DisplayName
	if displayName == "" {
		displayName = found.DisplayName
	}
	enabled := found.Enabled
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	model, err := h.repo.UpdateAIProviderModel(c.Context(), id, displayName, enabled)
	if err != nil {
		return apperrors.ErrInternal("failed to update AI model")
	}
	if model == nil {
		return apperrors.ErrNotFound("model not found")
	}
	return c.JSON(model)
}

// DeleteAIModel removes a catalog model.
func (h *AIModelsHandler) DeleteAIModel(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return apperrors.ErrInvalidRequest("invalid model id")
	}
	deleted, err := h.repo.DeleteAIProviderModel(c.Context(), id)
	if err != nil {
		return apperrors.ErrInternal("failed to delete AI model")
	}
	if !deleted {
		return apperrors.ErrNotFound("model not found")
	}
	return c.JSON(fiber.Map{"ok": true})
}
