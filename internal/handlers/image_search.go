package handlers

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos/piplos.media/internal/errors"
	"github.com/piplos/piplos.media/internal/media"
	"github.com/piplos/piplos.media/internal/models"
	"github.com/piplos/piplos.media/internal/repository"
)

// ImageSearchHandler finds /uploads/*.png|jpg references inserted into content texts
// and rewrites them to the existing WebP siblings in the database.
type ImageSearchHandler struct {
	repo *repository.Repository
	dir  string

	// Замены сериализуются: каждая сохраняет строку целиком из снятого
	// снапшота, поэтому параллельные Replace/ReplaceAll откатывали бы
	// изменения друг друга.
	mu sync.Mutex
}

// NewImageSearchHandler creates a ImageSearchHandler; dir is the upload root.
func NewImageSearchHandler(repo *repository.Repository, dir string) *ImageSearchHandler {
	return &ImageSearchHandler{repo: repo, dir: dir}
}

// refUsage is one place (entity + field) where a raster file is referenced.
type refUsage struct {
	Entity string `json:"entity"`
	ID     string `json:"id"`
	Label  string `json:"label"`
	Field  string `json:"field"`
	Href   string `json:"href,omitempty"` // admin edit URL, empty when the entity has no own page
}

// imageFileRef aggregates all usages of one uploaded raster file.
type imageFileRef struct {
	Path       string     `json:"path"`      // /uploads/<rel>.png|jpg (path case as stored, ext lowercase)
	WebPPath   string     `json:"webp_path"` // /uploads/<rel>.webp
	WebPExists bool       `json:"webp_exists"`
	Usages     []refUsage `json:"usages"`
}

// Entity kinds reported in usages.
const (
	entityProject = "project"
	entityService = "service"
	entityStack   = "stack"
	entitySEO     = "seo"
	entityLegal   = "legal"
	entityPage    = "page"
)

// uploadsRasterRefRE finds /uploads/*.png|jpg|jpeg references (optional
// scheme+host, optional ?query/#fragment suffix) — the same URL shapes that
// media.PreferWebPInText rewrites for rendering. The regex alone is broader
// than a real reference: it also matches inside foreign URLs
// ("…/wp-content/uploads/x.jpg") and filename continuations
// ("logo.png.bak", "logo.pngx"); rasterRefMatches filters those out.
var uploadsRasterRefRE = regexp.MustCompile(
	`(?i)((?:https?://[^"'>\s]+)?(/uploads/[^"'>\s?#]+))\.(png|jpe?g)([?#][^"'>\s]*)?`,
)

// scanField is one scannable text field of an entity.
type scanField struct {
	name  string
	value string
	set   func(string)
}

// scanEntity adapts a content entity: text fields to scan plus a save
// closure persisting the modified model through its repository update.
type scanEntity struct {
	entity string
	id     string
	label  string
	href   string // admin edit URL surfaced in the references response
	fields []scanField
	save   func(ctx context.Context) error
}

// References GET /v1/media/image-references — scan content tables and report
// every inserted PNG/JPG with a flag whether its WebP sibling exists on disk.
func (h *ImageSearchHandler) References(c fiber.Ctx) error {
	files, err := h.scan(c.Context())
	if err != nil {
		return internalErr("failed to scan image references", err)
	}
	return c.JSON(fiber.Map{"files": files})
}

type imageReplaceRequest struct {
	Path string `json:"path"`
}

// Replace POST /v1/media/image-references/replace {"path": "/uploads/x/y.png"} —
// rewrite one raster reference to .webp everywhere it is used.
func (h *ImageSearchHandler) Replace(c fiber.Ctx) error {
	var req imageReplaceRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	target, err := normalizeRasterRefPath(req.Path)
	if err != nil {
		return err
	}
	if !h.webPExists(rasterWebPPath(target)) {
		return apperrors.ErrInvalidRequest("webp version not found")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	entities, err := collectScanEntities(c.Context(), h.repo)
	if err != nil {
		return internalErr("failed to scan image references", err)
	}
	updated, err := applyRasterReplacements(c.Context(), entities, map[string]bool{target: true})
	if err != nil {
		return partialReplaceErr(updated, err)
	}
	return c.JSON(fiber.Map{"updated": updated})
}

// ReplaceAll POST /v1/media/image-references/replace-all — rewrite every raster
// reference whose WebP sibling exists.
func (h *ImageSearchHandler) ReplaceAll(c fiber.Ctx) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Одна загрузка сущностей даёт и цели (пути + наличие WebP), и поля для замены.
	entities, err := collectScanEntities(c.Context(), h.repo)
	if err != nil {
		return internalErr("failed to scan image references", err)
	}
	usages := scanRasterRefs(entities)
	targets := make(map[string]bool, len(usages))
	for refPath := range usages {
		if h.webPExists(rasterWebPPath(refPath)) {
			targets[refPath] = true
		}
	}
	if len(targets) == 0 {
		return apperrors.ErrInvalidRequest("no replaceable image references")
	}

	updated, err := applyRasterReplacements(c.Context(), entities, targets)
	if err != nil {
		return partialReplaceErr(updated, err)
	}
	return c.JSON(fiber.Map{"files": len(targets), "updated": updated})
}

// partialReplaceErr reports a failed multi-entity replace together with the
// number of records already rewritten: saves continue past individual failures,
// so the operator must see the real state instead of a bare failure.
func partialReplaceErr(updated int, cause error) error {
	return internalErr(
		fmt.Sprintf("replace interrupted after %d records were already updated", updated),
		cause,
	)
}

func (h *ImageSearchHandler) scan(ctx context.Context) ([]imageFileRef, error) {
	entities, err := collectScanEntities(ctx, h.repo)
	if err != nil {
		return nil, err
	}
	usages := scanRasterRefs(entities)

	out := make([]imageFileRef, 0, len(usages))
	for refPath, uses := range usages {
		out = append(out, imageFileRef{
			Path:       refPath,
			WebPPath:   rasterWebPPath(refPath),
			WebPExists: h.webPExists(rasterWebPPath(refPath)),
			Usages:     uses,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

// webPExists reports whether the WebP sibling file exists inside the upload dir.
func (h *ImageSearchHandler) webPExists(webpPath string) bool {
	_, abs, ok := resolveUploadPath(h.dir, strings.TrimPrefix(webpPath, "/uploads/"))
	if !ok {
		return false
	}
	_, err := os.Stat(abs)
	return err == nil
}

// replaceRefs reloads all content entities, rewrites the target references
// and saves changed entities through their repository updates.
func (h *ImageSearchHandler) replaceRefs(ctx context.Context, targets map[string]bool) (int, error) {
	entities, err := collectScanEntities(ctx, h.repo)
	if err != nil {
		return 0, err
	}
	return applyRasterReplacements(ctx, entities, targets)
}

// applyRasterReplacements rewrites the target references in the given entities
// and saves changed ones. Saves continue past individual failures (best
// effort — a mid-loop abort would strand already-rewritten rows silently);
// returns the number of saved entities and the first save error.
func applyRasterReplacements(ctx context.Context, entities []scanEntity, targets map[string]bool) (int, error) {
	updated := 0
	var firstErr error
	for _, e := range entities {
		changed := false
		for _, f := range e.fields {
			next := replaceRasterTargets(f.value, targets)
			if next != f.value {
				f.set(next)
				changed = true
			}
		}
		if changed {
			if err := e.save(ctx); err != nil {
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			updated++
		}
	}
	return updated, firstErr
}

// ---------- чистая логика скана и замены (покрывается тестами) ----------

// collectScanEntities loads all content entities and exposes their text fields.
func collectScanEntities(ctx context.Context, repo *repository.Repository) ([]scanEntity, error) {
	var out []scanEntity

	// Услуги первыми: их слаги нужны для ссылки на проект («Без группы» →
	// сегмент unassigned, как в списке проектов админки).
	services, err := repo.ListServices(ctx)
	if err != nil {
		return nil, err
	}
	serviceSlugs := make(map[string]bool, len(services))
	for i := range services {
		s := &services[i]
		serviceSlugs[s.Slug] = true
		e := scanEntity{
			entity: entityService,
			id:     s.ID,
			label:  translationsTitle(s.Translations, s.Slug),
			href:   "/services/" + s.Slug,
			fields: []scanField{{name: "icon", value: s.Icon, set: func(v string) { s.Icon = v }}},
			save: func(ctx context.Context) error {
				_, err := repo.UpdateService(ctx, s)
				return err
			},
		}
		e.fields = append(e.fields, translationFields(s.Translations)...)
		out = append(out, e)
	}

	projects, err := repo.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	for i := range projects {
		p := &projects[i]
		group := p.Category
		if !serviceSlugs[group] {
			group = "unassigned"
		}
		e := scanEntity{
			entity: entityProject,
			id:     p.ID,
			label:  translationsTitle(p.Translations, p.Slug),
			href:   "/projects/" + group + "/" + p.Slug,
			fields: []scanField{{name: "image", value: p.Image, set: func(v string) { p.Image = v }}},
			save: func(ctx context.Context) error {
				_, err := repo.UpdateProject(ctx, p)
				return err
			},
		}
		e.fields = append(e.fields, translationFields(p.Translations)...)
		out = append(out, e)
	}

	stackItems, err := repo.ListStackItems(ctx)
	if err != nil {
		return nil, err
	}
	for i := range stackItems {
		s := &stackItems[i]
		out = append(out, scanEntity{
			entity: entityStack,
			id:     s.ID,
			label:  s.Label,
			href:   "/stack",
			fields: []scanField{
				{name: "icon", value: s.Icon, set: func(v string) { s.Icon = v }},
				{name: "icon_alt", value: s.IconAlt, set: func(v string) { s.IconAlt = v }},
			},
			save: func(ctx context.Context) error {
				_, err := repo.UpdateStackItem(ctx, s)
				return err
			},
		})
	}

	seoPages, err := repo.ListSEOPages(ctx)
	if err != nil {
		return nil, err
	}
	for i := range seoPages {
		p := &seoPages[i]
		// SEO редактируется внутри формы своей сущности — отдельной страницы нет.
		out = append(out, scanEntity{
			entity: entitySEO,
			id:     p.ID,
			label:  p.Path,
			fields: translationFields(p.Translations),
			save: func(ctx context.Context) error {
				_, err := repo.UpdateSEOPage(ctx, p)
				return err
			},
		})
	}

	legalPages, err := repo.ListLegalPages(ctx)
	if err != nil {
		return nil, err
	}
	for i := range legalPages {
		p := &legalPages[i]
		out = append(out, scanEntity{
			entity: entityLegal,
			id:     p.ID,
			label:  p.Slug,
			href:   "/pages/legal/" + p.ID,
			fields: legalTranslationFields(p.Translations),
			save: func(ctx context.Context) error {
				_, err := repo.UpdateLegalPage(ctx, p)
				return err
			},
		})
	}

	pages, err := repo.ListPages(ctx)
	if err != nil {
		return nil, err
	}
	for i := range pages {
		p := &pages[i]
		e := scanEntity{
			entity: entityPage,
			id:     p.ID,
			label:  translationsTitle(p.Translations, p.Slug),
			href:   "/pages/" + p.ID,
			fields: []scanField{{name: "image", value: p.Image, set: func(v string) { p.Image = v }}},
			save: func(ctx context.Context) error {
				_, err := repo.UpdatePage(ctx, p)
				return err
			},
		}
		e.fields = append(e.fields, translationFields(p.Translations)...)
		out = append(out, e)
	}

	return out, nil
}

// translationFields exposes every non-empty string value of Translations.
func translationFields(t models.Translations) []scanField {
	var fields []scanField
	for lang, values := range t {
		keys := make([]string, 0, len(values))
		for key := range values {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := values[key]
			if value == "" {
				continue
			}
			lang, key := lang, key
			fields = append(fields, scanField{
				name:  "translations." + lang + "." + key,
				value: value,
				set:   func(v string) { t[lang][key] = v },
			})
		}
	}
	return fields
}

// legalTranslationFields exposes rendered text of legal documents: locale
// label/title and every section's title/body — all of it goes through
// renderLegalMarkdown to the public page.
func legalTranslationFields(t models.LegalTranslations) []scanField {
	var fields []scanField
	for lang, locale := range t {
		lang := lang
		if label := locale.Label; label != "" {
			fields = append(fields, scanField{
				name:  "translations." + lang + ".label",
				value: label,
				set: func(v string) {
					updated := t[lang]
					updated.Label = v
					t[lang] = updated
				},
			})
		}
		if title := locale.Title; title != "" {
			fields = append(fields, scanField{
				name:  "translations." + lang + ".title",
				value: title,
				set: func(v string) {
					updated := t[lang]
					updated.Title = v
					t[lang] = updated
				},
			})
		}
		for i := range locale.Sections {
			lang, i := lang, i
			if body := locale.Sections[i].Body; body != "" {
				fields = append(fields, scanField{
					name:  "translations." + lang + ".sections[" + strconv.Itoa(i) + "].body",
					value: body,
					set:   func(v string) { t[lang].Sections[i].Body = v },
				})
			}
			if title := locale.Sections[i].Title; title != "" {
				fields = append(fields, scanField{
					name:  "translations." + lang + ".sections[" + strconv.Itoa(i) + "].title",
					value: title,
					set:   func(v string) { t[lang].Sections[i].Title = v },
				})
			}
		}
	}
	return fields
}

// scanRasterRefs groups all raster occurrences by file path
// (/uploads/<rel>.png|jpg|jpeg, extension lowercased). Usages are deduplicated
// per entity field.
func scanRasterRefs(entities []scanEntity) map[string][]refUsage {
	found := make(map[string][]refUsage)
	seen := make(map[string]bool)
	for _, e := range entities {
		for _, f := range e.fields {
			rasterRefMatches(f.value, func(m []string) {
				refPath := rasterRefKey(m)
				key := e.entity + "\x00" + e.id + "\x00" + f.name + "\x00" + refPath
				if seen[key] {
					return
				}
				seen[key] = true
				found[refPath] = append(found[refPath], refUsage{
					Entity: e.entity,
					ID:     e.id,
					Label:  e.label,
					Field:  f.name,
					Href:   e.href,
				})
			})
		}
	}
	for _, uses := range found {
		sort.Slice(uses, func(i, j int) bool {
			if uses[i].Entity != uses[j].Entity {
				return uses[i].Entity < uses[j].Entity
			}
			if uses[i].Label != uses[j].Label {
				return uses[i].Label < uses[j].Label
			}
			return uses[i].Field < uses[j].Field
		})
	}
	return found
}

// rasterRefKey normalizes a match to the library path used as scan key:
// /uploads/<rel>.<lowercase ext>.
func rasterRefKey(m []string) string {
	return m[2] + "." + strings.ToLower(m[3])
}

// rasterRefMatches calls fn for every standalone local raster reference in s —
// skipping matches glued into longer paths (foreign "…/wp-content/uploads/…"
// URLs) and filename continuations ("logo.pngx", "logo.png.bak").
func rasterRefMatches(s string, fn func(m []string)) {
	for _, loc := range uploadsRasterRefRE.FindAllStringSubmatchIndex(s, -1) {
		m := refSubmatches(s, loc)
		if !isLocalUploadRef(m[1], m[2]) || !refEndsToken(s, loc[0], loc[1]) {
			continue
		}
		fn(m)
	}
}

// replaceRasterRefs rebuilds s in one pass, substituting every standalone local
// reference for which replace returns a replacement.
func replaceRasterRefs(s string, replace func(m []string) (string, bool)) string {
	var b strings.Builder
	last := 0
	for _, loc := range uploadsRasterRefRE.FindAllStringSubmatchIndex(s, -1) {
		m := refSubmatches(s, loc)
		if !isLocalUploadRef(m[1], m[2]) || !refEndsToken(s, loc[0], loc[1]) {
			continue
		}
		out, ok := replace(m)
		if !ok {
			continue
		}
		b.WriteString(s[last:loc[0]])
		b.WriteString(out)
		last = loc[1]
	}
	if last == 0 {
		return s
	}
	b.WriteString(s[last:])
	return b.String()
}

// replaceRasterRefInValue rewrites every occurrence of targetPath (any host
// prefix, any ?/# suffix) to its .webp sibling, leaving other URLs intact.
func replaceRasterRefInValue(s, targetPath string) string {
	return replaceRasterRefs(s, func(m []string) (string, bool) {
		if rasterRefKey(m) != targetPath {
			return "", false
		}
		return m[1] + ".webp" + m[4], true
	})
}

// replaceRasterTargets rewrites every reference whose normalized path is in
// targets to its .webp sibling — one regex pass regardless of |targets|.
func replaceRasterTargets(value string, targets map[string]bool) string {
	if !strings.Contains(strings.ToLower(value), "/uploads/") {
		return value
	}
	return replaceRasterRefs(value, func(m []string) (string, bool) {
		if !targets[rasterRefKey(m)] {
			return "", false
		}
		return m[1] + ".webp" + m[4], true
	})
}

func refSubmatches(s string, loc []int) []string {
	m := make([]string, len(loc)/2)
	for i := range m {
		if loc[2*i] >= 0 {
			m[i] = s[loc[2*i]:loc[2*i+1]]
		}
	}
	return m
}

// isLocalUploadRef rejects URLs that keep walking a path before /uploads/
// ("https://cdn.example.org/wp-content/uploads/…"): only bare /uploads/…
// paths and scheme+host-root URLs point at our media library.
func isLocalUploadRef(fullPath, uploadPath string) bool {
	prefix := strings.TrimSuffix(fullPath, uploadPath)
	if prefix == "" {
		return true
	}
	lowered := strings.ToLower(prefix)
	host, ok := strings.CutPrefix(lowered, "https://")
	if !ok {
		host, ok = strings.CutPrefix(lowered, "http://")
		if !ok {
			return false // префикс не схема+хост: склейка с окружающим путём
		}
	}
	return !strings.Contains(host, "/")
}

// refEndsToken reports whether the match stands alone: nothing path-like
// directly before it and no name continuation after the extension
// ("logo.pngx") or an extra suffix segment ("logo.png.bak"). A sentence dot
// after the extension ("see /uploads/a/b.png.") is fine.
func refEndsToken(s string, start, end int) bool {
	if start > 0 {
		switch s[start-1] {
		case '"', '\'', '(', '=', '>', '<', ' ', '\t', '\n', '\r':
		default:
			return false
		}
	}
	if end < len(s) {
		c := s[end]
		if isNameChar(c) {
			return false // logo.pngx
		}
		if c == '.' && end+1 < len(s) && isNameChar(s[end+1]) {
			return false // logo.png.bak
		}
	}
	return true
}

func isNameChar(c byte) bool {
	return c == '_' || c == '-' ||
		(c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z')
}

// rasterWebPPath converts /uploads/<stem>.<png|jpg|jpeg> to /uploads/<stem>.webp
// (any extension case). Other extensions are returned unchanged. Stem/ext logic
// is media's own (FullWebPName), so scan keys can't drift from the variants.
func rasterWebPPath(refPath string) string {
	if !media.IsDisposableOriginal(path.Ext(refPath)) {
		return refPath
	}
	rel := strings.TrimPrefix(refPath, "/uploads/")
	return "/uploads/" + media.FullWebPName(rel)
}

// normalizeRasterRefPath validates a client-supplied reference path and returns
// its canonical form: absolute /uploads/… path, lowercase raster extension,
// cleaned segments ("/uploads/a/b/../../c.PNG" → "/uploads/c.png").
func normalizeRasterRefPath(raw string) (string, error) {
	p := strings.TrimSpace(raw)
	if !strings.HasPrefix(p, "/uploads/") || len(p) > 512 {
		return "", apperrors.ErrInvalidRequest("path must be an /uploads/ png or jpg path")
	}
	ext := strings.ToLower(path.Ext(p))
	switch ext {
	case ".png", ".jpg", ".jpeg":
	default:
		return "", apperrors.ErrInvalidRequest("path must end with .png, .jpg or .jpeg")
	}
	stem := strings.TrimSuffix(strings.TrimPrefix(p, "/uploads/"), path.Ext(p))
	// Clean от корня нейтрализует «..» и лишние слэши: они не могут выйти
	// выше /uploads/, поэтому после Clean остаётся безопасный относительный путь.
	rel := strings.TrimPrefix(path.Clean("/"+stem), "/")
	if rel == "" || rel == "." || rel == "/" || !validEntryName(path.Base(rel)) {
		return "", apperrors.ErrInvalidRequest("invalid path")
	}
	if dir := path.Dir(rel); dir != "." && !validFolderPath(dir) {
		return "", apperrors.ErrInvalidRequest("invalid path")
	}
	return "/uploads/" + rel + ext, nil
}

// translationsTitle picks the first non-empty title across languages
// (en first, then alphabetical), falling back when absent.
func translationsTitle(t models.Translations, fallback string) string {
	codes := make([]string, 0, len(t))
	for code := range t {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	if _, ok := t["en"]; ok {
		codes = append([]string{"en"}, codes...)
	}
	seen := make(map[string]bool, len(codes))
	for _, code := range codes {
		if seen[code] {
			continue
		}
		seen[code] = true
		if title := strings.TrimSpace(t[code]["title"]); title != "" {
			return title
		}
	}
	return fallback
}
