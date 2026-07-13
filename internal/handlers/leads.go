package handlers

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos/site/internal/errors"
	"github.com/piplos/site/internal/models"
	"github.com/piplos/site/internal/repository"
)

type leadNotifier interface {
	NotifyNewLead(ctx context.Context, lead *models.Lead)
}

// LeadsHandler serves the public order form endpoint and admin lead management.
type LeadsHandler struct {
	repo   *repository.Repository
	notify leadNotifier
}

// NewLeadsHandler creates a LeadsHandler.
func NewLeadsHandler(repo *repository.Repository, notify leadNotifier) *LeadsHandler {
	return &LeadsHandler{repo: repo, notify: notify}
}

type leadRequest struct {
	Types       []string `json:"types"`
	ProjectName string   `json:"project_name"`
	Description string   `json:"description"`
	Stack       string   `json:"stack"`
	References  string   `json:"references"`
	Budget      int      `json:"budget"`
	Currency    string   `json:"currency"`
	Timeline    string   `json:"timeline"`
	Stage       string   `json:"stage"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Email       string   `json:"email"`
	Company     string   `json:"company"`
	Phone       string   `json:"phone"`
	HowFound    string   `json:"how_found"`
	Notes       string   `json:"notes"`
	Lang        string   `json:"lang"`
}

// Create is the public endpoint the site order form posts to.
func (h *LeadsHandler) Create(c fiber.Ctx) error {
	var req leadRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		return apperrors.ErrInvalidRequest("valid email is required")
	}
	if req.FirstName == "" {
		return apperrors.ErrInvalidRequest("first_name is required")
	}
	if len(req.Types) == 0 {
		return apperrors.ErrInvalidRequest("at least one project type is required")
	}
	if req.Lang == "" {
		req.Lang = "en"
	}
	if req.Currency == "" {
		req.Currency = "USD"
	}

	lead := &models.Lead{
		Types: req.Types, ProjectName: req.ProjectName, Description: req.Description,
		Stack: req.Stack, ReferenceURLs: req.References, Budget: req.Budget,
		Currency: req.Currency, Timeline: req.Timeline, Stage: req.Stage,
		FirstName: req.FirstName, LastName: req.LastName, Email: req.Email,
		Company: req.Company, Phone: req.Phone, HowFound: req.HowFound,
		Notes: req.Notes, Lang: req.Lang,
	}
	created, err := h.repo.CreateLead(c.Context(), lead)
	if err != nil {
		return apperrors.ErrInternal("failed to save request")
	}
	if h.notify != nil {
		notifyLead := *created
		go h.notify.NotifyNewLead(context.Background(), &notifyLead)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": created.ID})
}

// List returns leads with optional ?status= filter and pagination.
func (h *LeadsHandler) List(c fiber.Ctx) error {
	status := c.Query("status")
	if status != "" {
		switch models.LeadStatus(status) {
		case models.LeadNew, models.LeadInProgress, models.LeadDone, models.LeadSpam:
		default:
			return apperrors.ErrInvalidRequest("invalid status filter")
		}
	}
	limit := fiber.Query(c, "limit", 50)
	if limit < 1 || limit > 200 {
		limit = 50
	}
	offset := fiber.Query(c, "offset", 0)
	if offset < 0 {
		offset = 0
	}

	leads, total, err := h.repo.ListLeads(c.Context(), status, limit, offset)
	if err != nil {
		return apperrors.ErrInternal("failed to list leads")
	}
	return c.JSON(fiber.Map{"leads": leads, "total": total})
}

// Get returns one lead.
func (h *LeadsHandler) Get(c fiber.Ctx) error {
	lead, err := h.repo.GetLead(c.Context(), c.Params("id"))
	if err != nil {
		return apperrors.ErrInternal("failed to get lead")
	}
	if lead == nil {
		return apperrors.ErrNotFound("lead not found")
	}
	return c.JSON(fiber.Map{"lead": lead})
}

type leadStatusRequest struct {
	Status string `json:"status"`
}

// UpdateStatus changes lead processing status.
func (h *LeadsHandler) UpdateStatus(c fiber.Ctx) error {
	var req leadStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	status := models.LeadStatus(req.Status)
	switch status {
	case models.LeadNew, models.LeadInProgress, models.LeadDone, models.LeadSpam:
	default:
		return apperrors.ErrInvalidRequest("invalid status")
	}

	lead, err := h.repo.UpdateLeadStatus(c.Context(), c.Params("id"), status)
	if err != nil {
		return apperrors.ErrInternal("failed to update lead")
	}
	if lead == nil {
		return apperrors.ErrNotFound("lead not found")
	}
	return c.JSON(fiber.Map{"lead": lead})
}

// Delete removes a lead.
func (h *LeadsHandler) Delete(c fiber.Ctx) error {
	if err := h.repo.DeleteLead(c.Context(), c.Params("id")); err != nil {
		return apperrors.ErrInternal("failed to delete lead")
	}
	return c.JSON(fiber.Map{"ok": true})
}
