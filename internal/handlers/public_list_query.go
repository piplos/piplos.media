package handlers

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/piplos/piplos.media/internal/media"
	"github.com/piplos/piplos.media/internal/models"
)

// maxPublicListLimit caps ?limit= for public catalog endpoints.
const maxPublicListLimit = 100

// projectListQuery is GET /public/projects query options.
type projectListQuery struct {
	lang         string
	featuredOnly bool
	category     string
	tags         []string
	slugs        []string
	limit        int // 0 = no limit
	mode         string
}

func parseCSVQuery(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	for part := range strings.SplitSeq(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseLimitQuery(raw string) int {
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0
	}
	return min(n, maxPublicListLimit)
}

// parseSummaryMode returns "summary" or "" (full payload).
func parseSummaryMode(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "summary") {
		return "summary"
	}
	return ""
}

func parseProjectListQuery(c fiber.Ctx) projectListQuery {
	return projectListQuery{
		lang:         c.Query("lang"),
		featuredOnly: c.Query("featured") == "true",
		category:     strings.TrimSpace(c.Query("category")),
		tags:         parseCSVQuery(c.Query("tags")),
		slugs:        parseCSVQuery(c.Query("slugs")),
		limit:        parseLimitQuery(c.Query("limit")),
		mode:         parseSummaryMode(c.Query("mode")),
	}
}

func projectInCategory(p models.Project, category string) bool {
	if p.Category == category {
		return true
	}
	return slices.Contains(p.Categories, category)
}

// hasAnyTag reports whether have contains any of wanted (case-insensitive).
// Empty wanted matches everything (caller usually skips the check).
func hasAnyTag(have, wanted []string) bool {
	if len(wanted) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(have))
	for _, tag := range have {
		set[strings.ToLower(tag)] = struct{}{}
	}
	for _, tag := range wanted {
		if _, ok := set[strings.ToLower(tag)]; ok {
			return true
		}
	}
	return false
}

// omitTranslationFields returns a copy without the listed keys (smaller JSON).
func omitTranslationFields(t models.Translations, fields ...string) models.Translations {
	if t == nil {
		return nil
	}
	out := make(models.Translations, len(t))
	for lang, values := range t {
		copied := make(map[string]string, len(values))
		for key, value := range values {
			copied[key] = value
		}
		for _, field := range fields {
			delete(copied, field)
		}
		out[lang] = copied
	}
	return out
}

func prepareProjectTranslations(t models.Translations, lang, mode string) models.Translations {
	t = filteredTranslations(t, lang)
	if mode == "summary" {
		// Lists/embeds: skip Markdown render of heavy HTML fields.
		return omitTranslationFields(t, "solution", "result")
	}
	return renderMarkdownFields(t, "solution")
}

// publishedProjects returns only published projects (ListProjects order preserved).
func publishedProjects(items []models.Project) []models.Project {
	out := make([]models.Project, 0, len(items))
	for _, p := range items {
		if p.Published {
			out = append(out, p)
		}
	}
	return out
}

// filterPublishedProjects applies public list filters.
// slugs win over featured/category/tags (same as site embed selection).
// With category (and no slugs): sort by group sort_order, then year DESC.
// Otherwise: keep ListProjects order (global_sort_order).
func filterPublishedProjects(items []models.Project, q projectListQuery) []models.Project {
	published := publishedProjects(items)

	var out []models.Project
	if len(q.slugs) > 0 {
		bySlug := make(map[string]models.Project, len(published))
		for _, p := range published {
			bySlug[p.Slug] = p
		}
		for _, slug := range q.slugs {
			if p, ok := bySlug[slug]; ok {
				out = append(out, p)
			}
		}
	} else {
		out = make([]models.Project, 0, len(published))
		for _, p := range published {
			if q.featuredOnly && !p.Featured {
				continue
			}
			if q.category != "" && !projectInCategory(p, q.category) {
				continue
			}
			if len(q.tags) > 0 && !hasAnyTag(p.Tags, q.tags) {
				continue
			}
			out = append(out, p)
		}
		if q.category != "" {
			slices.SortStableFunc(out, func(a, b models.Project) int {
				if c := cmp.Compare(a.SortOrder, b.SortOrder); c != 0 {
					return c
				}
				return cmp.Compare(b.Year, a.Year)
			})
		}
	}

	if q.limit > 0 && len(out) > q.limit {
		out = out[:q.limit]
	}
	return out
}

// serviceListQuery is GET /public/services query options.
type serviceListQuery struct {
	lang  string
	tags  []string
	slugs []string
	limit int
	mode  string
}

func parseServiceListQuery(c fiber.Ctx) serviceListQuery {
	return serviceListQuery{
		lang:  c.Query("lang"),
		tags:  parseCSVQuery(c.Query("tags")),
		slugs: parseCSVQuery(c.Query("slugs")),
		limit: parseLimitQuery(c.Query("limit")),
		mode:  parseSummaryMode(c.Query("mode")),
	}
}

func prepareServiceTranslations(t models.Translations, lang, mode string) models.Translations {
	t = filteredTranslations(t, lang)
	if mode == "summary" {
		return omitTranslationFields(t, "body")
	}
	return renderMarkdownFields(t, "body")
}

// preferWebPInTranslations rewrites /uploads/*.png|jpg in translation HTML/text to .webp.
func preferWebPInTranslations(t models.Translations) models.Translations {
	if t == nil {
		return nil
	}
	out := make(models.Translations, len(t))
	for lang, values := range t {
		copied := make(map[string]string, len(values))
		for key, value := range values {
			copied[key] = media.PreferWebPInText(value)
		}
		out[lang] = copied
	}
	return out
}

func publishedServices(items []models.Service) []models.Service {
	out := make([]models.Service, 0, len(items))
	for _, s := range items {
		if s.Published {
			out = append(out, s)
		}
	}
	return out
}

func filterPublishedServices(items []models.Service, q serviceListQuery) []models.Service {
	published := publishedServices(items)

	var out []models.Service
	if len(q.slugs) > 0 {
		bySlug := make(map[string]models.Service, len(published))
		for _, s := range published {
			bySlug[s.Slug] = s
		}
		for _, slug := range q.slugs {
			if s, ok := bySlug[slug]; ok {
				out = append(out, s)
			}
		}
	} else {
		out = make([]models.Service, 0, len(published))
		for _, s := range published {
			if len(q.tags) > 0 && !hasAnyTag(s.Tags, q.tags) {
				continue
			}
			out = append(out, s)
		}
		slices.SortStableFunc(out, func(a, b models.Service) int {
			return cmp.Compare(a.SortOrder, b.SortOrder)
		})
	}

	if q.limit > 0 && len(out) > q.limit {
		out = out[:q.limit]
	}
	return out
}
