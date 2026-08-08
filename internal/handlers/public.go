package handlers

import (
	"time"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos/piplos.media/internal/errors"
	"github.com/piplos/piplos.media/internal/media"
	"github.com/piplos/piplos.media/internal/models"
	"github.com/piplos/piplos.media/internal/repository"
)

// PublicHandler exposes published content to the site (no auth).
type PublicHandler struct {
	repo *repository.Repository
}

// NewPublicHandler creates a PublicHandler.
func NewPublicHandler(repo *repository.Repository) *PublicHandler {
	return &PublicHandler{repo: repo}
}

// filteredTranslations keeps only the requested language. If the language is
// missing, the full set is returned so the client can apply its own fallback.
func filteredTranslations(t models.Translations, lang string) models.Translations {
	if lang == "" {
		return t
	}
	if tr, ok := t[lang]; ok {
		return models.Translations{lang: tr}
	}
	return t
}

// filteredLegalTranslations is filteredTranslations for legal documents.
func filteredLegalTranslations(t models.LegalTranslations, lang string) models.LegalTranslations {
	if lang == "" {
		return t
	}
	if tr, ok := t[lang]; ok {
		return models.LegalTranslations{lang: tr}
	}
	return t
}

// Projects returns published projects.
// Query:
//   - lang — return only this translation
//   - featured=true — featured only
//   - category — primary category or categories[]
//   - tags — comma-separated, any match
//   - slugs — comma-separated; order preserved; overrides featured/category/tags
//   - limit — max items (capped)
//   - mode=summary — omit solution/result HTML (lists/embeds)
func (h *PublicHandler) Projects(c fiber.Ctx) error {
	items, err := h.repo.ListProjects(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load projects")
	}
	q := parseProjectListQuery(c)
	published := filterPublishedProjects(items, q)
	for i := range published {
		published[i].Translations = prepareProjectTranslations(published[i].Translations, q.lang, q.mode)
		published[i].Translations = preferWebPInTranslations(published[i].Translations)
		published[i].Image = media.PreferWebPURL(published[i].Image)
	}
	return c.JSON(fiber.Map{"projects": published})
}

// Project returns a single published project by slug.
// Query: lang — return only this translation.
func (h *PublicHandler) Project(c fiber.Ctx) error {
	p, err := h.repo.GetProject(c.Context(), c.Params("slug"))
	if err != nil {
		return apperrors.ErrInternal("failed to load project")
	}
	if p == nil || !p.Published {
		return apperrors.ErrNotFound("project not found")
	}
	p.Translations = renderMarkdownFields(
		filteredTranslations(p.Translations, c.Query("lang")), "solution")
	p.Image = media.PreferWebPURL(p.Image)
	p.Translations = preferWebPInTranslations(p.Translations)
	return c.JSON(fiber.Map{"project": p})
}

// Services returns published services.
// Query:
//   - lang — return only this translation
//   - tags — comma-separated, any match
//   - slugs — comma-separated; order preserved; overrides tags
//   - limit — max items (capped)
//   - mode=summary — omit body HTML (embeds/cards)
func (h *PublicHandler) Services(c fiber.Ctx) error {
	items, err := h.repo.ListServices(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load services")
	}
	q := parseServiceListQuery(c)
	published := filterPublishedServices(items, q)
	for i := range published {
		published[i].Translations = prepareServiceTranslations(published[i].Translations, q.lang, q.mode)
		published[i].Translations = preferWebPInTranslations(published[i].Translations)
	}
	return c.JSON(fiber.Map{"services": published})
}

// Service returns a single published service by slug.
// Query: lang — return only this translation.
func (h *PublicHandler) Service(c fiber.Ctx) error {
	s, err := h.repo.GetService(c.Context(), c.Params("slug"))
	if err != nil {
		return apperrors.ErrInternal("failed to load service")
	}
	if s == nil || !s.Published {
		return apperrors.ErrNotFound("service not found")
	}
	s.Translations = renderMarkdownFields(
		filteredTranslations(s.Translations, c.Query("lang")), "body")
	s.Translations = preferWebPInTranslations(s.Translations)
	return c.JSON(fiber.Map{"service": s})
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
			s.Icon = media.PreferWebPURL(s.Icon)
			s.IconAlt = media.PreferWebPURL(s.IconAlt)
			published = append(published, s)
		}
	}
	return c.JSON(fiber.Map{"stack": published})
}

// SEO returns SEO page entries.
// Query: path — return only the entry for this page path.
func (h *PublicHandler) SEO(c fiber.Ctx) error {
	if path := c.Query("path"); path != "" {
		p, err := h.repo.GetSEOPageByPath(c.Context(), path)
		if err != nil {
			return apperrors.ErrInternal("failed to load seo")
		}
		return c.JSON(fiber.Map{"page": p})
	}
	items, err := h.repo.ListSEOPages(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load seo")
	}
	return c.JSON(fiber.Map{"pages": items})
}

// Pages returns live custom pages (site "Articles" section):
// published and not scheduled for a future date.
// Query: lang — return only this translation.
func (h *PublicHandler) Pages(c fiber.Ctx) error {
	items, err := h.repo.ListPages(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load pages")
	}
	lang := c.Query("lang")
	now := time.Now()
	live := []models.Page{}
	for _, p := range items {
		if !p.IsLive(now) {
			continue
		}
		p.Translations = renderMarkdownFields(filteredTranslations(p.Translations, lang), "body")
		p.Translations = preferWebPInTranslations(p.Translations)
		p.Image = media.PreferWebPURL(p.Image)
		live = append(live, p)
	}
	return c.JSON(fiber.Map{"pages": live})
}

// Page returns a single live custom page by slug.
// Query: lang — return only this translation.
func (h *PublicHandler) Page(c fiber.Ctx) error {
	p, err := h.repo.GetPageBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return apperrors.ErrInternal("failed to load page")
	}
	if p == nil || !p.IsLive(time.Now()) {
		return apperrors.ErrNotFound("page not found")
	}
	p.Translations = renderMarkdownFields(
		filteredTranslations(p.Translations, c.Query("lang")), "body")
	p.Translations = preferWebPInTranslations(p.Translations)
	p.Image = media.PreferWebPURL(p.Image)
	return c.JSON(fiber.Map{"page": p})
}

// Legal returns all legal documents for the public site.
// Query: lang — return only this translation.
func (h *PublicHandler) Legal(c fiber.Ctx) error {
	items, err := h.repo.ListLegalPages(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to load legal pages")
	}
	lang := c.Query("lang")
	for i := range items {
		items[i].Translations = renderLegalMarkdown(filteredLegalTranslations(items[i].Translations, lang))
	}
	return c.JSON(fiber.Map{"pages": items})
}

// LegalPage returns a single legal document by slug.
// Query: lang — return only this translation.
func (h *PublicHandler) LegalPage(c fiber.Ctx) error {
	p, err := h.repo.GetLegalPageBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		return apperrors.ErrInternal("failed to load legal page")
	}
	if p == nil {
		return apperrors.ErrNotFound("legal page not found")
	}
	p.Translations = renderLegalMarkdown(filteredLegalTranslations(p.Translations, c.Query("lang")))
	return c.JSON(fiber.Map{"page": p})
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
