package handlers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	apperrors "github.com/piplos/piplos.media/internal/errors"
	"github.com/piplos/piplos.media/internal/models"
	"github.com/piplos/piplos.media/internal/repository"
	"github.com/piplos/piplos.media/internal/utils"
)

// AgentStore abstracts persistence for the agent articles API
// (implemented by *repository.Repository).
type AgentStore interface {
	ListLanguages(ctx context.Context) ([]models.Language, error)
	ListStackItems(ctx context.Context) ([]models.StackItem, error)
	ListPages(ctx context.Context) ([]models.Page, error)
	GetPage(ctx context.Context, id string) (*models.Page, error)
	GetPageBySlug(ctx context.Context, slug string) (*models.Page, error)
	CreatePage(ctx context.Context, p *models.Page) (*models.Page, error)
	UpdatePage(ctx context.Context, p *models.Page) (*models.Page, error)
	DeletePage(ctx context.Context, id string) error
	GetSEOPageByPath(ctx context.Context, path string) (*models.SEOPage, error)
	CreateSEOPage(ctx context.Context, p *models.SEOPage) (*models.SEOPage, error)
	UpdateSEOPage(ctx context.Context, p *models.SEOPage) (*models.SEOPage, error)
	DeleteSEOPage(ctx context.Context, id string) error
}

// AgentHandler serves external automation agents (API-key auth).
// Articles here are the same entities as the admin "pages" CRUD, but with
// strict per-language validation and computed publication status.
type AgentHandler struct {
	repo    AgentStore
	uploads *UploadsHandler
}

// NewAgentHandler creates an AgentHandler.
func NewAgentHandler(repo AgentStore, uploads *UploadsHandler) *AgentHandler {
	return &AgentHandler{repo: repo, uploads: uploads}
}

// agentArticleRequest is the create/update payload. translations must cover
// every enabled language with non-empty title/description/body.
type agentArticleRequest struct {
	Slug         string              `json:"slug"`
	Published    bool                `json:"published"`
	PublishAt    *time.Time          `json:"publish_at"`
	Image        string              `json:"image"`
	Tags         []string            `json:"tags"`
	Translations models.Translations `json:"translations"`
	SEO          *agentSEORequest    `json:"seo,omitempty"`
}

type agentSEORequest struct {
	Translations models.Translations `json:"translations"`
}

// agentArticle wraps a page with the computed publication status and its
// SEO entry (if one exists for /articles/<slug>).
type agentArticle struct {
	models.Page
	Status string          `json:"status"`
	SEO    *models.SEOPage `json:"seo,omitempty"`
}

var (
	// articleRequiredFields must be present in every enabled language.
	articleRequiredFields = []string{"title", "description", "body"}
	// seoRequiredFields must be present in every enabled language when the
	// seo block is provided.
	seoRequiredFields = []string{"title", "description"}

	agentSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

	// agentStackTagLimit caps the tags array: articles pick a stack, not a
	// replacement for the body.
	agentStackTagLimit = 20
)

// articleSEOPath is the SEO entry path convention for articles.
func articleSEOPath(slug string) string { return "/articles/" + slug }

// validateTranslations enforces full coverage: every enabled language must
// provide every required field with non-blank values, and no unexpected
// language keys may appear. Languages are iterated in repo order so the
// error message is deterministic.
func validateTranslations(t models.Translations, langs []models.Language, fields []string) *apperrors.AppError {
	enabled := models.EnabledLanguageCodes(langs)
	enabledSet := make(map[string]struct{}, len(enabled))
	for _, code := range enabled {
		enabledSet[code] = struct{}{}
	}

	var missing []string
	for _, code := range enabled {
		for _, field := range fields {
			if v, ok := t[code][field]; !ok || strings.TrimSpace(v) == "" {
				missing = append(missing, code+"."+field)
			}
		}
	}
	if len(missing) > 0 {
		return apperrors.ErrValidation("missing: " + strings.Join(missing, ", "))
	}

	var unsupported []string
	for code := range t {
		if _, ok := enabledSet[code]; !ok {
			unsupported = append(unsupported, code)
		}
	}
	if len(unsupported) > 0 {
		sort.Strings(unsupported)
		return apperrors.ErrValidation("unsupported languages: " + strings.Join(unsupported, ", "))
	}
	return nil
}

// languages loads enabled languages or fails the request.
func (h *AgentHandler) languages(c fiber.Ctx) ([]models.Language, error) {
	langs, err := h.repo.ListLanguages(c.Context())
	if err != nil {
		return nil, internalErr("failed to load languages", err)
	}
	if len(models.EnabledLanguageCodes(langs)) == 0 {
		return nil, apperrors.ErrInternal("no enabled languages configured")
	}
	return langs, nil
}

// resolveSlug returns the slug for a new article: the explicit one (charset
// validated) or one generated from the default-language title with -N
// de-duplication probing.
func (h *AgentHandler) resolveSlug(c fiber.Ctx, req *agentArticleRequest, langs []models.Language) (string, error) {
	req.Slug = strings.TrimSpace(req.Slug)
	if req.Slug != "" {
		if !agentSlugPattern.MatchString(req.Slug) {
			return "", apperrors.ErrInvalidRequest("slug must be lowercase letters, digits and dashes")
		}
		if existing, err := h.repo.GetPageBySlug(c.Context(), req.Slug); err != nil {
			return "", internalErr("failed to check slug", err)
		} else if existing != nil {
			return "", apperrors.ErrConflict("slug already exists")
		}
		return req.Slug, nil
	}

	// Generate from the default language title.
	var defaultCode string
	for _, l := range langs {
		if l.IsDefault {
			defaultCode = l.Code
			break
		}
	}
	if defaultCode == "" {
		if enabled := models.EnabledLanguageCodes(langs); len(enabled) > 0 {
			defaultCode = enabled[0]
		}
	}
	title := strings.TrimSpace(req.Translations[defaultCode]["title"])
	base := utils.Slugify(title)
	if base == "" {
		return "", apperrors.ErrInvalidRequest("slug or default-language title is required")
	}

	candidate := base
	for i := 2; i <= 50; i++ {
		existing, err := h.repo.GetPageBySlug(c.Context(), candidate)
		if err != nil {
			return "", internalErr("failed to check slug", err)
		}
		if existing == nil {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, i)
	}
	return "", apperrors.ErrConflict("could not generate unique slug")
}

// upsertSEO creates or updates the SEO entry at path for the given slug.
func (h *AgentHandler) upsertSEO(c fiber.Ctx, path string, tr models.Translations) (*models.SEOPage, error) {
	existing, err := h.repo.GetSEOPageByPath(c.Context(), path)
	if err != nil {
		return nil, internalErr("failed to load seo page", err)
	}
	if existing == nil {
		created, err := h.repo.CreateSEOPage(c.Context(), &models.SEOPage{Path: path, Translations: tr})
		if err != nil {
			return nil, internalErr("failed to create seo page", err)
		}
		return created, nil
	}
	existing.Translations = tr
	updated, err := h.repo.UpdateSEOPage(c.Context(), existing)
	if err != nil {
		return nil, internalErr("failed to update seo page", err)
	}
	return updated, nil
}

// canonicalizeTags validates article tags against the stack catalog: tags
// are required, trimmed, deduped (case-insensitive) and mapped onto the
// catalog labels, so stored values always match the catalog exactly.
// Returns error (not *AppError): callers assign the result into an existing
// error variable, and a typed nil would make that check misfire.
func canonicalizeTags(tags []string, items []models.StackItem) ([]string, error) {
	byLabel := make(map[string]string, len(items))
	for _, it := range items {
		byLabel[strings.ToLower(strings.TrimSpace(it.Label))] = it.Label
	}

	canonical := []string{}
	seen := make(map[string]struct{}, len(tags))
	var unknown []string
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		label, ok := byLabel[strings.ToLower(t)]
		if !ok {
			unknown = append(unknown, t)
			continue
		}
		if _, dup := seen[strings.ToLower(label)]; dup {
			continue
		}
		seen[strings.ToLower(label)] = struct{}{}
		canonical = append(canonical, label)
	}
	if len(unknown) > 0 {
		return nil, apperrors.ErrValidation("unknown stack tags: " + strings.Join(unknown, ", "))
	}
	if len(canonical) == 0 {
		return nil, apperrors.ErrValidation("missing: tags")
	}
	if len(canonical) > agentStackTagLimit {
		return nil, apperrors.ErrValidation(fmt.Sprintf("tags: at most %d items allowed", agentStackTagLimit))
	}
	return canonical, nil
}

// bindArticleRequest parses the body and validates translations for the
// request's languages (strict, article fields + optional SEO block).
// Shared by create and update (DRY).
func (h *AgentHandler) bindArticleRequest(c fiber.Ctx) (*agentArticleRequest, []models.Language, error) {
	var req agentArticleRequest
	if err := c.Bind().Body(&req); err != nil {
		return nil, nil, apperrors.ErrInvalidRequest("invalid request body")
	}
	langs, err := h.languages(c)
	if err != nil {
		return nil, nil, err
	}
	if err := validateTranslations(req.Translations, langs, articleRequiredFields); err != nil {
		return nil, nil, err
	}
	if req.SEO != nil {
		if err := validateTranslations(req.SEO.Translations, langs, seoRequiredFields); err != nil {
			return nil, nil, err
		}
	}

	items, err := h.repo.ListStackItems(c.Context())
	if err != nil {
		return nil, nil, internalErr("failed to load stack catalog", err)
	}
	req.Tags, err = canonicalizeTags(req.Tags, items)
	if err != nil {
		return nil, nil, err
	}
	return &req, langs, nil
}

// ListStack returns the stack catalog for tag selection: the label of each
// entry is a valid value for the article's tags field. Published items only —
// the same set the public site shows; validation accepts any catalog label
// so drafts never break updates of existing articles.
func (h *AgentHandler) ListStack(c fiber.Ctx) error {
	items, err := h.repo.ListStackItems(c.Context())
	if err != nil {
		return internalErr("failed to load stack catalog", err)
	}
	published := []models.StackItem{}
	for _, it := range items {
		if it.Published {
			published = append(published, it)
		}
	}
	return c.JSON(fiber.Map{"stack": published})
}

// CreateArticle adds an article on behalf of an external agent.
func (h *AgentHandler) CreateArticle(c fiber.Ctx) error {
	req, langs, err := h.bindArticleRequest(c)
	if err != nil {
		return err
	}

	slug, err := h.resolveSlug(c, req, langs)
	if err != nil {
		return err
	}

	page := &models.Page{
		Slug:         slug,
		Published:    req.Published,
		PublishAt:    req.PublishAt,
		Image:        strings.TrimSpace(req.Image),
		Tags:         req.Tags,
		Translations: req.Translations,
	}
	created, err := h.repo.CreatePage(c.Context(), page)
	if err != nil {
		// DB UNIQUE constraint is the backstop for slug probe races.
		if errors.Is(err, repository.ErrDuplicateSlug) {
			return apperrors.ErrConflict("slug already exists")
		}
		return internalErr("failed to create article", err)
	}

	var seo *models.SEOPage
	if req.SEO != nil {
		seo, err = h.upsertSEO(c, articleSEOPath(slug), req.SEO.Translations)
		if err != nil {
			// Статья уже создана: повторный POST с авто-slug создаст дубль.
			// Явно сообщаем агенту, что контент сохранён и нужно догнать SEO
			// через PUT этого же id.
			return apperrors.ErrInternal("article created but seo update failed; retry via PUT /articles/" + created.ID)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"article": agentArticle{Page: *created, Status: created.ArticleStatus(time.Now()), SEO: seo},
	})
}

// ListArticles returns all articles (drafts and scheduled included).
func (h *AgentHandler) ListArticles(c fiber.Ctx) error {
	items, err := h.repo.ListPages(c.Context())
	if err != nil {
		return internalErr("failed to list articles", err)
	}
	now := time.Now()
	articles := make([]agentArticle, 0, len(items))
	for _, p := range items {
		articles = append(articles, agentArticle{Page: p, Status: p.ArticleStatus(now)})
	}
	return c.JSON(fiber.Map{"articles": articles})
}

// articleID validates the :id path param as UUID (pages.id is UUID; an
// unvalidated non-UUID would surface as a 500 from the driver).
func articleID(c fiber.Ctx) (string, error) {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		return "", apperrors.ErrInvalidRequest("invalid article id")
	}
	return id, nil
}

// GetArticle returns one article with its SEO entry.
func (h *AgentHandler) GetArticle(c fiber.Ctx) error {
	id, err := articleID(c)
	if err != nil {
		return err
	}
	p, err := h.repo.GetPage(c.Context(), id)
	if err != nil {
		return internalErr("failed to get article", err)
	}
	if p == nil {
		return apperrors.ErrNotFound("article not found")
	}
	seo, err := h.repo.GetSEOPageByPath(c.Context(), articleSEOPath(p.Slug))
	if err != nil {
		return internalErr("failed to load seo page", err)
	}
	return c.JSON(fiber.Map{
		"article": agentArticle{Page: *p, Status: p.ArticleStatus(time.Now()), SEO: seo},
	})
}

// UpdateArticle replaces an article. The payload must again cover every
// enabled language. An empty slug keeps the current one; a slug change
// moves the SEO entry to the new path (409 when the target path is taken).
func (h *AgentHandler) UpdateArticle(c fiber.Ctx) error {
	id, err := articleID(c)
	if err != nil {
		return err
	}
	existing, err := h.repo.GetPage(c.Context(), id)
	if err != nil {
		return internalErr("failed to get article", err)
	}
	if existing == nil {
		return apperrors.ErrNotFound("article not found")
	}

	req, _, err := h.bindArticleRequest(c)
	if err != nil {
		return err
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = existing.Slug
	} else {
		if !agentSlugPattern.MatchString(slug) {
			return apperrors.ErrInvalidRequest("slug must be lowercase letters, digits and dashes")
		}
		if slug != existing.Slug {
			if occupant, err := h.repo.GetPageBySlug(c.Context(), slug); err != nil {
				return internalErr("failed to check slug", err)
			} else if occupant != nil && occupant.ID != existing.ID {
				return apperrors.ErrConflict("slug already exists")
			}
		}
	}

	page := &models.Page{
		ID:           existing.ID,
		Slug:         slug,
		Published:    req.Published,
		PublishAt:    req.PublishAt,
		Image:        strings.TrimSpace(req.Image),
		Tags:         req.Tags,
		Translations: req.Translations,
	}
	updated, err := h.repo.UpdatePage(c.Context(), page)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateSlug) {
			return apperrors.ErrConflict("slug already exists")
		}
		return internalErr("failed to update article", err)
	}
	if updated == nil {
		return apperrors.ErrNotFound("article not found")
	}

	seo, err := h.rebindSEO(c, existing.Slug, slug, req.SEO)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"article": agentArticle{Page: *updated, Status: updated.ArticleStatus(time.Now()), SEO: seo},
	})
}

// rebindSEO reconciles the SEO entry after an update. Slug unchanged: the
// provided block (if any) upserts at the current path; without a block the
// existing entry is left untouched. Slug changed: the entry moves to the new
// path unless that path is already taken (409).
func (h *AgentHandler) rebindSEO(c fiber.Ctx, oldSlug, newSlug string, reqSEO *agentSEORequest) (*models.SEOPage, error) {
	newPath := articleSEOPath(newSlug)
	if oldSlug == newSlug {
		if reqSEO == nil {
			return nil, nil // leave existing SEO untouched; response omits it
		}
		return h.upsertSEO(c, newPath, reqSEO.Translations)
	}

	oldSEO, err := h.repo.GetSEOPageByPath(c.Context(), articleSEOPath(oldSlug))
	if err != nil {
		return nil, internalErr("failed to load seo page", err)
	}
	occupant, err := h.repo.GetSEOPageByPath(c.Context(), newPath)
	if err != nil {
		return nil, internalErr("failed to load seo page", err)
	}

	switch {
	case oldSEO != nil && occupant != nil && occupant.ID != oldSEO.ID:
		return nil, apperrors.ErrConflict("seo path " + newPath + " is already taken")
	case oldSEO != nil:
		oldSEO.Path = newPath
		if reqSEO != nil {
			oldSEO.Translations = reqSEO.Translations
		}
		moved, err := h.repo.UpdateSEOPage(c.Context(), oldSEO)
		if err != nil {
			return nil, internalErr("failed to update seo page", err)
		}
		return moved, nil
	case reqSEO != nil:
		return h.upsertSEO(c, newPath, reqSEO.Translations)
	default:
		return nil, nil
	}
}

// DeleteArticle removes an article and its SEO entry.
func (h *AgentHandler) DeleteArticle(c fiber.Ctx) error {
	id, err := articleID(c)
	if err != nil {
		return err
	}
	p, err := h.repo.GetPage(c.Context(), id)
	if err != nil {
		return internalErr("failed to get article", err)
	}
	if p == nil {
		return apperrors.ErrNotFound("article not found")
	}
	if err := h.repo.DeletePage(c.Context(), p.ID); err != nil {
		return internalErr("failed to delete article", err)
	}
	seo, err := h.repo.GetSEOPageByPath(c.Context(), articleSEOPath(p.Slug))
	if err != nil {
		return internalErr("failed to load seo page", err)
	}
	if seo != nil {
		if err := h.repo.DeleteSEOPage(c.Context(), seo.ID); err != nil {
			return internalErr("failed to delete seo page", err)
		}
	}
	return c.JSON(fiber.Map{"ok": true})
}

// Uploads delegates to the shared uploads handler (multipart; it reads no
// identity from locals, so it works under API-key auth).
func (h *AgentHandler) Uploads(c fiber.Ctx) error {
	if h.uploads == nil {
		return apperrors.ErrInternal("uploads handler is not configured")
	}
	return h.uploads.Upload(c)
}
