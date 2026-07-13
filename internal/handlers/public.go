package handlers

import (
	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos/site/internal/errors"
	"github.com/piplos/site/internal/models"
	"github.com/piplos/site/internal/repository"
)

// PublicHandler exposes published content to the site (no auth).
type PublicHandler struct {
	repo *repository.Repository
}

// NewPublicHandler creates a PublicHandler.
func NewPublicHandler(repo *repository.Repository) *PublicHandler {
	return &PublicHandler{repo: repo}
}

// Projects returns published projects.
func (h *PublicHandler) Projects(c fiber.Ctx) error {
	items, err := h.repo.ListProjects(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load projects")
	}
	published := []models.Project{}
	for _, p := range items {
		if p.Published {
			published = append(published, p)
		}
	}
	return c.JSON(fiber.Map{"projects": published})
}

// Services returns published services.
func (h *PublicHandler) Services(c fiber.Ctx) error {
	items, err := h.repo.ListServices(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load services")
	}
	published := []models.Service{}
	for _, s := range items {
		if s.Published {
			published = append(published, s)
		}
	}
	return c.JSON(fiber.Map{"services": published})
}

// Stack returns published stack items.
func (h *PublicHandler) Stack(c fiber.Ctx) error {
	items, err := h.repo.ListStackItems(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load stack")
	}
	published := []models.StackItem{}
	for _, s := range items {
		if s.Published {
			published = append(published, s)
		}
	}
	return c.JSON(fiber.Map{"stack": published})
}

// SEO returns all SEO page entries.
func (h *PublicHandler) SEO(c fiber.Ctx) error {
	items, err := h.repo.ListSEOPages(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load seo")
	}
	return c.JSON(fiber.Map{"pages": items})
}

// Legal returns all legal documents for the public site.
func (h *PublicHandler) Legal(c fiber.Ctx) error {
	items, err := h.repo.ListLegalPages(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load legal pages")
	}
	return c.JSON(fiber.Map{"pages": items})
}

// Languages returns enabled content languages.
func (h *PublicHandler) Languages(c fiber.Ctx) error {
	langs, err := h.repo.ListLanguages(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load languages")
	}
	enabled := []models.Language{}
	for _, l := range langs {
		if l.Enabled {
			enabled = append(enabled, l)
		}
	}
	return c.JSON(fiber.Map{"languages": enabled})
}
