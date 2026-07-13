package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos-media/site/internal/errors"
	"github.com/piplos-media/site/internal/models"
	"github.com/piplos-media/site/internal/repository"
)

// ContentHandler manages projects, services, stack items and SEO pages.
type ContentHandler struct {
	repo *repository.Repository
}

// NewContentHandler creates a ContentHandler.
func NewContentHandler(repo *repository.Repository) *ContentHandler {
	return &ContentHandler{repo: repo}
}

// ---------- Projects ----------

type projectRequest struct {
	Slug         string              `json:"slug"`
	Category     string              `json:"category"`
	Categories   []string            `json:"categories"`
	Tags         []string            `json:"tags"`
	Year         int                 `json:"year"`
	Featured     bool                `json:"featured"`
	Published    bool                `json:"published"`
	SortOrder    int                 `json:"sort_order"`
	Translations models.Translations `json:"translations"`
}

func (req *projectRequest) toModel(id string) (*models.Project, error) {
	req.Slug = strings.TrimSpace(req.Slug)
	if req.Slug == "" {
		return nil, apperrors.ErrInvalidRequest("slug is required")
	}
	if req.Categories == nil {
		req.Categories = []string{}
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}
	if req.Translations == nil {
		req.Translations = models.Translations{}
	}
	return &models.Project{
		ID: id, Slug: req.Slug, Category: req.Category, Categories: req.Categories,
		Tags: req.Tags, Year: req.Year, Featured: req.Featured, Published: req.Published,
		SortOrder: req.SortOrder, Translations: req.Translations,
	}, nil
}

// ListProjects returns all projects.
func (h *ContentHandler) ListProjects(c fiber.Ctx) error {
	items, err := h.repo.ListProjects(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to list projects")
	}
	return c.JSON(fiber.Map{"projects": items})
}

// GetProject returns one project.
func (h *ContentHandler) GetProject(c fiber.Ctx) error {
	p, err := h.repo.GetProject(c.Context(), c.Params("slug"))
	if err != nil {
		return apperrors.ErrInternal("failed to get project")
	}
	if p == nil {
		return apperrors.ErrNotFound("project not found")
	}
	return c.JSON(fiber.Map{"project": p})
}

// CreateProject adds a project.
func (h *ContentHandler) CreateProject(c fiber.Ctx) error {
	var req projectRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	p, err := req.toModel("")
	if err != nil {
		return err
	}
	created, err := h.repo.CreateProject(c.Context(), p)
	if err != nil {
		return apperrors.ErrInternal("failed to create project")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"project": created})
}

// UpdateProject modifies a project.
func (h *ContentHandler) UpdateProject(c fiber.Ctx) error {
	existing, err := h.repo.GetProject(c.Context(), c.Params("slug"))
	if err != nil {
		return apperrors.ErrInternal("failed to get project")
	}
	if existing == nil {
		return apperrors.ErrNotFound("project not found")
	}

	var req projectRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	p, err := req.toModel(existing.ID)
	if err != nil {
		return err
	}
	updated, err := h.repo.UpdateProject(c.Context(), p)
	if err != nil {
		return apperrors.ErrInternal("failed to update project")
	}
	if updated == nil {
		return apperrors.ErrNotFound("project not found")
	}
	return c.JSON(fiber.Map{"project": updated})
}

// DeleteProject removes a project.
func (h *ContentHandler) DeleteProject(c fiber.Ctx) error {
	if err := h.repo.DeleteProject(c.Context(), c.Params("slug")); err != nil {
		return apperrors.ErrInternal("failed to delete project")
	}
	return c.JSON(fiber.Map{"ok": true})
}

type reorderProjectsGroupRequest struct {
	GroupID string   `json:"group_id"`
	IDs     []string `json:"ids"`
}

type reorderProjectsRequest struct {
	Groups []reorderProjectsGroupRequest `json:"groups"`
}

// ReorderProjects updates category membership and order within each service group.
func (h *ContentHandler) ReorderProjects(c fiber.Ctx) error {
	var req reorderProjectsRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	if len(req.Groups) == 0 {
		return apperrors.ErrInvalidRequest("groups is required")
	}

	groups := make([]repository.ProjectGroupOrder, 0, len(req.Groups))
	for _, g := range req.Groups {
		groupID := strings.TrimSpace(g.GroupID)
		if groupID == "" {
			return apperrors.ErrInvalidRequest("group_id is required")
		}
		if len(g.IDs) == 0 {
			continue
		}
		groups = append(groups, repository.ProjectGroupOrder{GroupID: groupID, IDs: g.IDs})
	}
	if len(groups) == 0 {
		return apperrors.ErrInvalidRequest("groups is required")
	}

	if err := h.repo.ReorderProjects(c.Context(), groups); err != nil {
		return apperrors.ErrInternal("failed to reorder projects")
	}
	return c.JSON(fiber.Map{"ok": true})
}

// ---------- Services ----------

type serviceRequest struct {
	Slug         string              `json:"slug"`
	Icon         string              `json:"icon"`
	Tags         []string            `json:"tags"`
	Published    bool                `json:"published"`
	SortOrder    int                 `json:"sort_order"`
	Translations models.Translations `json:"translations"`
}

func (req *serviceRequest) toModel(id string) (*models.Service, error) {
	req.Slug = strings.TrimSpace(req.Slug)
	if req.Slug == "" {
		return nil, apperrors.ErrInvalidRequest("slug is required")
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}
	if req.Translations == nil {
		req.Translations = models.Translations{}
	}
	return &models.Service{
		ID: id, Slug: req.Slug, Icon: req.Icon, Tags: req.Tags,
		Published: req.Published, SortOrder: req.SortOrder, Translations: req.Translations,
	}, nil
}

// ListServices returns all services.
func (h *ContentHandler) ListServices(c fiber.Ctx) error {
	items, err := h.repo.ListServices(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to list services")
	}
	return c.JSON(fiber.Map{"services": items})
}

// CreateService adds a service.
func (h *ContentHandler) CreateService(c fiber.Ctx) error {
	var req serviceRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	s, err := req.toModel("")
	if err != nil {
		return err
	}
	created, err := h.repo.CreateService(c.Context(), s)
	if err != nil {
		return apperrors.ErrInternal("failed to create service")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"service": created})
}

// UpdateService modifies a service.
func (h *ContentHandler) UpdateService(c fiber.Ctx) error {
	var req serviceRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	s, err := req.toModel(c.Params("id"))
	if err != nil {
		return err
	}
	updated, err := h.repo.UpdateService(c.Context(), s)
	if err != nil {
		return apperrors.ErrInternal("failed to update service")
	}
	if updated == nil {
		return apperrors.ErrNotFound("service not found")
	}
	return c.JSON(fiber.Map{"service": updated})
}

// DeleteService removes a service.
func (h *ContentHandler) DeleteService(c fiber.Ctx) error {
	if err := h.repo.DeleteService(c.Context(), c.Params("id")); err != nil {
		return apperrors.ErrInternal("failed to delete service")
	}
	return c.JSON(fiber.Map{"ok": true})
}

type reorderServicesRequest struct {
	IDs []string `json:"ids"`
}

// ReorderServices updates display order from an ordered list of service ids.
func (h *ContentHandler) ReorderServices(c fiber.Ctx) error {
	var req reorderServicesRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	if len(req.IDs) == 0 {
		return apperrors.ErrInvalidRequest("ids is required")
	}
	if err := h.repo.ReorderServices(c.Context(), req.IDs); err != nil {
		return apperrors.ErrInternal("failed to reorder services")
	}
	return c.JSON(fiber.Map{"ok": true})
}

// ---------- Stack ----------

type stackRequest struct {
	Slug      string `json:"slug"`
	Label     string `json:"label"`
	GroupID   string `json:"group_id"`
	Published bool   `json:"published"`
	SortOrder int    `json:"sort_order"`
}

func (req *stackRequest) toModel(id string) (*models.StackItem, error) {
	req.Slug = strings.TrimSpace(req.Slug)
	req.Label = strings.TrimSpace(req.Label)
	if req.Slug == "" || req.Label == "" {
		return nil, apperrors.ErrInvalidRequest("slug and label are required")
	}
	return &models.StackItem{
		ID: id, Slug: req.Slug, Label: req.Label, GroupID: req.GroupID,
		Published: req.Published, SortOrder: req.SortOrder,
	}, nil
}

// ListStack returns all stack items.
func (h *ContentHandler) ListStack(c fiber.Ctx) error {
	items, err := h.repo.ListStackItems(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to list stack")
	}
	return c.JSON(fiber.Map{"stack": items})
}

// CreateStackItem adds a stack item.
func (h *ContentHandler) CreateStackItem(c fiber.Ctx) error {
	var req stackRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	s, err := req.toModel("")
	if err != nil {
		return err
	}
	created, err := h.repo.CreateStackItem(c.Context(), s)
	if err != nil {
		return apperrors.ErrInternal("failed to create stack item")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"item": created})
}

// UpdateStackItem modifies a stack item.
func (h *ContentHandler) UpdateStackItem(c fiber.Ctx) error {
	var req stackRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	s, err := req.toModel(c.Params("id"))
	if err != nil {
		return err
	}
	updated, err := h.repo.UpdateStackItem(c.Context(), s)
	if err != nil {
		return apperrors.ErrInternal("failed to update stack item")
	}
	if updated == nil {
		return apperrors.ErrNotFound("stack item not found")
	}
	return c.JSON(fiber.Map{"item": updated})
}

// DeleteStackItem removes a stack item.
func (h *ContentHandler) DeleteStackItem(c fiber.Ctx) error {
	if err := h.repo.DeleteStackItem(c.Context(), c.Params("id")); err != nil {
		return apperrors.ErrInternal("failed to delete stack item")
	}
	return c.JSON(fiber.Map{"ok": true})
}

type reorderStackGroupRequest struct {
	GroupID string   `json:"group_id"`
	IDs     []string `json:"ids"`
}

type reorderStackRequest struct {
	Groups []reorderStackGroupRequest `json:"groups"`
}

// ReorderStack updates group membership and order within each service group.
func (h *ContentHandler) ReorderStack(c fiber.Ctx) error {
	var req reorderStackRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	if len(req.Groups) == 0 {
		return apperrors.ErrInvalidRequest("groups is required")
	}

	groups := make([]repository.StackGroupOrder, 0, len(req.Groups))
	for _, g := range req.Groups {
		groupID := strings.TrimSpace(g.GroupID)
		if groupID == "" {
			return apperrors.ErrInvalidRequest("group_id is required")
		}
		if len(g.IDs) == 0 {
			continue
		}
		groups = append(groups, repository.StackGroupOrder{GroupID: groupID, IDs: g.IDs})
	}
	if len(groups) == 0 {
		return apperrors.ErrInvalidRequest("groups is required")
	}

	if err := h.repo.ReorderStack(c.Context(), groups); err != nil {
		return apperrors.ErrInternal("failed to reorder stack")
	}
	return c.JSON(fiber.Map{"ok": true})
}

// ---------- SEO ----------

type seoRequest struct {
	Path         string              `json:"path"`
	Translations models.Translations `json:"translations"`
}

func (req *seoRequest) toModel(id string) (*models.SEOPage, error) {
	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" || !strings.HasPrefix(req.Path, "/") {
		return nil, apperrors.ErrInvalidRequest("path must start with /")
	}
	if models.IsLegalPath(req.Path) {
		return nil, apperrors.ErrInvalidRequest("legal pages are managed in the Legal section")
	}
	if req.Translations == nil {
		req.Translations = models.Translations{}
	}
	return &models.SEOPage{ID: id, Path: req.Path, Translations: req.Translations}, nil
}

// ListSEO returns all SEO entries.
func (h *ContentHandler) ListSEO(c fiber.Ctx) error {
	items, err := h.repo.ListSEOPages(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to list seo pages")
	}
	return c.JSON(fiber.Map{"pages": items})
}

// CreateSEO adds a SEO entry.
func (h *ContentHandler) CreateSEO(c fiber.Ctx) error {
	var req seoRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	p, err := req.toModel("")
	if err != nil {
		return err
	}
	created, err := h.repo.CreateSEOPage(c.Context(), p)
	if err != nil {
		return apperrors.ErrInternal("failed to create seo page")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"page": created})
}

// UpdateSEO modifies a SEO entry.
func (h *ContentHandler) UpdateSEO(c fiber.Ctx) error {
	var req seoRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	p, err := req.toModel(c.Params("id"))
	if err != nil {
		return err
	}
	updated, err := h.repo.UpdateSEOPage(c.Context(), p)
	if err != nil {
		return apperrors.ErrInternal("failed to update seo page")
	}
	if updated == nil {
		return apperrors.ErrNotFound("seo page not found")
	}
	return c.JSON(fiber.Map{"page": updated})
}

// DeleteSEO removes a SEO entry.
func (h *ContentHandler) DeleteSEO(c fiber.Ctx) error {
	if err := h.repo.DeleteSEOPage(c.Context(), c.Params("id")); err != nil {
		return apperrors.ErrInternal("failed to delete seo page")
	}
	return c.JSON(fiber.Map{"ok": true})
}

// ---------- Legal ----------

type legalRequest struct {
	Translations models.LegalTranslations `json:"translations"`
}

// ListLegal returns all legal documents.
func (h *ContentHandler) ListLegal(c fiber.Ctx) error {
	items, err := h.repo.ListLegalPages(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to list legal pages")
	}
	return c.JSON(fiber.Map{"pages": items})
}

// GetLegal returns one legal document.
func (h *ContentHandler) GetLegal(c fiber.Ctx) error {
	p, err := h.repo.GetLegalPage(c.Context(), c.Params("id"))
	if err != nil {
		return apperrors.ErrInternal("failed to load legal page")
	}
	if p == nil {
		return apperrors.ErrNotFound("legal page not found")
	}
	return c.JSON(fiber.Map{"page": p})
}

// UpdateLegal updates legal document content (path is fixed).
func (h *ContentHandler) UpdateLegal(c fiber.Ctx) error {
	var req legalRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	if req.Translations == nil {
		req.Translations = models.LegalTranslations{}
	}
	existing, err := h.repo.GetLegalPage(c.Context(), c.Params("id"))
	if err != nil {
		return apperrors.ErrInternal("failed to load legal page")
	}
	if existing == nil {
		return apperrors.ErrNotFound("legal page not found")
	}
	existing.Translations = req.Translations
	updated, err := h.repo.UpdateLegalPage(c.Context(), existing)
	if err != nil {
		return apperrors.ErrInternal("failed to update legal page")
	}
	return c.JSON(fiber.Map{"page": updated})
}
