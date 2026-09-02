package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/piplos/piplos.media/internal/handlers"
	"github.com/piplos/piplos.media/internal/middleware"
	"github.com/piplos/piplos.media/internal/models"
	"github.com/piplos/piplos.media/internal/repository"
)

// fakeAgentStore implements handlers.AgentStore in memory, with the pages.slug
// unique constraint emulated via repository.ErrDuplicateSlug.
type fakeAgentStore struct {
	langs []models.Language
	stack []models.StackItem
	pages map[string]*models.Page
	seo   map[string]*models.SEOPage
}

func newFakeAgentStore() *fakeAgentStore {
	return &fakeAgentStore{
		langs: []models.Language{
			{Code: "en", Name: "English", IsDefault: true, Enabled: true},
			{Code: "ru", Name: "Русский", Enabled: true, SortOrder: 1},
		},
		stack: []models.StackItem{
			{ID: "s-flutter", Slug: "flutter", Label: "Flutter", Published: true},
			{ID: "s-go", Slug: "golang", Label: "Go", Published: true},
			{ID: "s-figma", Slug: "figma", Label: "Figma", Published: true},
			{ID: "s-php", Slug: "php", Label: "PHP", Published: false},
		},
		pages: map[string]*models.Page{},
		seo:   map[string]*models.SEOPage{},
	}
}

func (f *fakeAgentStore) ListLanguages(_ context.Context) ([]models.Language, error) {
	return f.langs, nil
}

func (f *fakeAgentStore) ListStackItems(_ context.Context) ([]models.StackItem, error) {
	return f.stack, nil
}

func (f *fakeAgentStore) slugTaken(slug, exceptID string) bool {
	for _, p := range f.pages {
		if p.Slug == slug && p.ID != exceptID {
			return true
		}
	}
	return false
}

func (f *fakeAgentStore) ListPages(_ context.Context) ([]models.Page, error) {
	out := []models.Page{}
	for _, p := range f.pages {
		out = append(out, *p)
	}
	return out, nil
}

func (f *fakeAgentStore) GetPage(_ context.Context, id string) (*models.Page, error) {
	if p := f.pages[id]; p != nil {
		c := *p
		return &c, nil
	}
	return nil, nil
}

func (f *fakeAgentStore) GetPageBySlug(_ context.Context, slug string) (*models.Page, error) {
	for _, p := range f.pages {
		if p.Slug == slug {
			c := *p
			return &c, nil
		}
	}
	return nil, nil
}

func (f *fakeAgentStore) CreatePage(_ context.Context, p *models.Page) (*models.Page, error) {
	if f.slugTaken(p.Slug, "") {
		return nil, fmt.Errorf("create page: %w", repository.ErrDuplicateSlug)
	}
	c := *p
	// Реальные UUID: обработчики валидируют :id как UUID.
	c.ID = uuid.Must(uuid.NewRandom()).String()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	f.pages[c.ID] = &c
	return &c, nil
}

func (f *fakeAgentStore) UpdatePage(_ context.Context, p *models.Page) (*models.Page, error) {
	stored := f.pages[p.ID]
	if stored == nil {
		return nil, nil
	}
	if f.slugTaken(p.Slug, p.ID) {
		return nil, fmt.Errorf("update page: %w", repository.ErrDuplicateSlug)
	}
	stored.Slug = p.Slug
	stored.Published = p.Published
	stored.PublishAt = p.PublishAt
	stored.Image = p.Image
	stored.Tags = p.Tags
	stored.Translations = p.Translations
	stored.UpdatedAt = time.Now()
	c := *stored
	return &c, nil
}

func (f *fakeAgentStore) DeletePage(_ context.Context, id string) error {
	delete(f.pages, id)
	return nil
}

func (f *fakeAgentStore) GetSEOPageByPath(_ context.Context, path string) (*models.SEOPage, error) {
	for _, s := range f.seo {
		if s.Path == path {
			c := *s
			return &c, nil
		}
	}
	return nil, nil
}

func (f *fakeAgentStore) CreateSEOPage(_ context.Context, p *models.SEOPage) (*models.SEOPage, error) {
	c := *p
	c.ID = uuid.Must(uuid.NewRandom()).String()
	f.seo[c.ID] = &c
	return &c, nil
}

func (f *fakeAgentStore) UpdateSEOPage(_ context.Context, p *models.SEOPage) (*models.SEOPage, error) {
	stored := f.seo[p.ID]
	if stored == nil {
		return nil, nil
	}
	stored.Path = p.Path
	stored.Translations = p.Translations
	c := *stored
	return &c, nil
}

func (f *fakeAgentStore) DeleteSEOPage(_ context.Context, id string) error {
	delete(f.seo, id)
	return nil
}

// --- test helpers ---

type agentArticleOut struct {
	ID           string              `json:"id"`
	Slug         string              `json:"slug"`
	Published    bool                `json:"published"`
	PublishAt    *time.Time          `json:"publish_at"`
	Tags         []string            `json:"tags"`
	Translations models.Translations `json:"translations"`
	Status       string              `json:"status"`
	SEO          *models.SEOPage     `json:"seo"`
}

type agentErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func newAgentTestApp(t *testing.T) (*fiber.App, *fakeAgentStore) {
	t.Helper()
	store := newFakeAgentStore()
	app := fiber.New()
	app.Use(middleware.ErrorHandler(zerolog.Nop()))
	h := handlers.NewAgentHandler(store, nil)
	app.Post("/v1/agent/articles", h.CreateArticle)
	app.Get("/v1/agent/stack", h.ListStack)
	app.Get("/v1/agent/articles", h.ListArticles)
	app.Get("/v1/agent/articles/:id", h.GetArticle)
	app.Put("/v1/agent/articles/:id", h.UpdateArticle)
	app.Delete("/v1/agent/articles/:id", h.DeleteArticle)
	return app, store
}

func doAgentRequest(t *testing.T, app *fiber.App, method, path, body string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func decodeAgentJSON(t *testing.T, r io.Reader, dest any) {
	t.Helper()
	raw, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
}

// fullTranslations returns complete en+ru translations.
func fullTranslations() models.Translations {
	return models.Translations{
		"en": {"title": "Test Article", "description": "A test", "body": "# Body"},
		"ru": {"title": "Тестовая статья", "description": "Тест", "body": "# Текст"},
	}
}

func TestAgentCreateRejectsMissingTranslations(t *testing.T) {
	app, _ := newAgentTestApp(t)

	tr := fullTranslations()
	delete(tr["ru"], "body")
	tr["en"]["title"] = "   "
	body := fmt.Sprintf(`{"tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, tr))

	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("got %d, want 422", resp.StatusCode)
	}
	var out agentErrorResponse
	decodeAgentJSON(t, resp.Body, &out)
	if out.Error != "validation_failed" {
		t.Fatalf("error code = %q, want validation_failed", out.Error)
	}
	// Deterministic order: en first (sort_order 0), then ru; fields in order.
	want := "missing: en.title, ru.body"
	if out.Message != want {
		t.Fatalf("message = %q, want %q", out.Message, want)
	}
}

func TestAgentCreateRejectsUnsupportedLanguage(t *testing.T) {
	app, _ := newAgentTestApp(t)

	tr := fullTranslations()
	tr["de"] = map[string]string{"title": "x", "description": "y", "body": "z"}
	body := fmt.Sprintf(`{"tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, tr))

	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("got %d, want 422", resp.StatusCode)
	}
	var out agentErrorResponse
	decodeAgentJSON(t, resp.Body, &out)
	if out.Message != "unsupported languages: de" {
		t.Fatalf("message = %q, want unsupported languages: de", out.Message)
	}
}

func TestAgentCreateAutoSlugFromCyrillicTitle(t *testing.T) {
	app, store := newAgentTestApp(t)

	tr := fullTranslations()
	body := fmt.Sprintf(`{"tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, tr))

	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("got %d, want 201", resp.StatusCode)
	}
	var out struct {
		Article agentArticleOut `json:"article"`
	}
	decodeAgentJSON(t, resp.Body, &out)
	// Default language (en) title "Test Article" -> slug.
	if out.Article.Slug != "test-article" {
		t.Fatalf("auto slug = %q, want test-article", out.Article.Slug)
	}
	if out.Article.Status != "published" {
		t.Fatalf("status = %q, want published", out.Article.Status)
	}

	// Second create with the same titles gets a -2 suffix.
	resp = doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("second create: got %d, want 201", resp.StatusCode)
	}
	decodeAgentJSON(t, resp.Body, &out)
	if out.Article.Slug != "test-article-2" {
		t.Fatalf("second slug = %q, want test-article-2", out.Article.Slug)
	}
	_ = store
}

func TestAgentCreateSlugValidationAndConflict(t *testing.T) {
	app, store := newAgentTestApp(t)

	// Invalid charset.
	body := fmt.Sprintf(`{"slug":"My Slug","tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, fullTranslations()))
	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid slug: got %d, want 400", resp.StatusCode)
	}

	// Explicit duplicate.
	store.pages["p-1"] = &models.Page{ID: "p-1", Slug: "taken"}
	body = fmt.Sprintf(`{"slug":"taken","tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, fullTranslations()))
	resp = doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate slug: got %d, want 409", resp.StatusCode)
	}
	var out agentErrorResponse
	decodeAgentJSON(t, resp.Body, &out)
	if out.Error != "conflict" {
		t.Fatalf("error code = %q, want conflict", out.Error)
	}
}

func TestAgentCreateStatuses(t *testing.T) {
	app, _ := newAgentTestApp(t)
	future := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)

	cases := []struct {
		name string
		body string
		want string
	}{
		{"draft", fmt.Sprintf(`{"tags":["Flutter"],"published":false,"translations":%s}`, mustJSON(t, fullTranslations())), "draft"},
		{"scheduled", fmt.Sprintf(`{"tags":["Flutter"],"published":true,"publish_at":%q,"translations":%s}`, future, mustJSON(t, fullTranslations())), "scheduled"},
		{"published past", fmt.Sprintf(`{"tags":["Flutter"],"published":true,"publish_at":%q,"translations":%s}`, past, mustJSON(t, fullTranslations())), "published"},
	}
	for _, tc := range cases {
		resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", tc.body)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("%s: got %d, want 201", tc.name, resp.StatusCode)
		}
		var out struct {
			Article agentArticleOut `json:"article"`
		}
		decodeAgentJSON(t, resp.Body, &out)
		if out.Article.Status != tc.want {
			t.Errorf("%s: status = %q, want %q", tc.name, out.Article.Status, tc.want)
		}
	}
}

func TestAgentCreateWithSEO(t *testing.T) {
	app, store := newAgentTestApp(t)

	seo := models.Translations{
		"en": {"title": "SEO En", "description": "Desc"},
		"ru": {"title": "SEO Ru", "description": "Описание"},
	}
	body := fmt.Sprintf(`{"slug":"with-seo","tags":["Flutter"],"published":true,"translations":%s,"seo":{"translations":%s}}`,
		mustJSON(t, fullTranslations()), mustJSON(t, seo))

	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("got %d, want 201", resp.StatusCode)
	}
	var out struct {
		Article agentArticleOut `json:"article"`
	}
	decodeAgentJSON(t, resp.Body, &out)
	if out.Article.SEO == nil || out.Article.SEO.Path != "/articles/with-seo" {
		t.Fatalf("seo = %+v, want path /articles/with-seo", out.Article.SEO)
	}
	if _, ok := store.seo[out.Article.SEO.ID]; !ok {
		t.Fatal("seo row not persisted")
	}

	// GET by id returns the SEO too.
	resp = doAgentRequest(t, app, http.MethodGet, "/v1/agent/articles/"+out.Article.ID, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get: got %d, want 200", resp.StatusCode)
	}
	decodeAgentJSON(t, resp.Body, &out)
	if out.Article.SEO == nil || out.Article.SEO.Path != "/articles/with-seo" {
		t.Fatalf("get seo = %+v, want /articles/with-seo", out.Article.SEO)
	}
}

func TestAgentUpdateMovesSEOOnSlugChange(t *testing.T) {
	app, _ := newAgentTestApp(t)

	seo := models.Translations{"en": {"title": "SEO", "description": "D"}, "ru": {"title": "СЕО", "description": "О"}}
	createBody := fmt.Sprintf(`{"slug":"old-slug","tags":["Flutter"],"published":true,"translations":%s,"seo":{"translations":%s}}`,
		mustJSON(t, fullTranslations()), mustJSON(t, seo))
	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", createBody)
	var created struct {
		Article agentArticleOut `json:"article"`
	}
	decodeAgentJSON(t, resp.Body, &created)

	// Update with a new slug and no seo block: entry must move, translations kept.
	updateBody := fmt.Sprintf(`{"slug":"new-slug","tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, fullTranslations()))
	resp = doAgentRequest(t, app, http.MethodPut, "/v1/agent/articles/"+created.Article.ID, updateBody)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update: got %d, want 200", resp.StatusCode)
	}
	var updated struct {
		Article agentArticleOut `json:"article"`
	}
	decodeAgentJSON(t, resp.Body, &updated)
	if updated.Article.SEO == nil || updated.Article.SEO.Path != "/articles/new-slug" {
		t.Fatalf("seo path = %+v, want /articles/new-slug", updated.Article.SEO)
	}
	if updated.Article.SEO.Translations["en"]["title"] != "SEO" {
		t.Fatal("moved SEO must keep its translations")
	}
}

func TestAgentUpdateSEOOccupiedPath(t *testing.T) {
	app, store := newAgentTestApp(t)

	// Article A with SEO at /articles/a; article B occupying /articles/b.
	seoA := models.Translations{"en": {"title": "A", "description": "D"}, "ru": {"title": "А", "description": "Д"}}
	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles",
		fmt.Sprintf(`{"slug":"a","tags":["Flutter"],"published":true,"translations":%s,"seo":{"translations":%s}}`,
			mustJSON(t, fullTranslations()), mustJSON(t, seoA)))
	var a struct {
		Article agentArticleOut `json:"article"`
	}
	decodeAgentJSON(t, resp.Body, &a)

	store.seo[uuid.Must(uuid.NewRandom()).String()] = &models.SEOPage{ID: "seo-b", Path: "/articles/b"}

	// Rename A onto B's path -> 409.
	updateBody := fmt.Sprintf(`{"slug":"b","tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, fullTranslations()))
	resp = doAgentRequest(t, app, http.MethodPut, "/v1/agent/articles/"+a.Article.ID, updateBody)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("occupied seo path: got %d, want 409", resp.StatusCode)
	}
}

func TestAgentUpdateValidation(t *testing.T) {
	app, _ := newAgentTestApp(t)

	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles",
		fmt.Sprintf(`{"slug":"upd","tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, fullTranslations())))
	var created struct {
		Article agentArticleOut `json:"article"`
	}
	decodeAgentJSON(t, resp.Body, &created)

	// Partial translations on update must fail with 422 too.
	tr := fullTranslations()
	delete(tr["en"], "body")
	resp = doAgentRequest(t, app, http.MethodPut, "/v1/agent/articles/"+created.Article.ID,
		fmt.Sprintf(`{"tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, tr)))
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("partial update: got %d, want 422", resp.StatusCode)
	}

	// Unknown UUID -> 404 (valid UUID that does not exist).
	missingID := uuid.Must(uuid.NewRandom()).String()
	resp = doAgentRequest(t, app, http.MethodPut, "/v1/agent/articles/"+missingID,
		fmt.Sprintf(`{"tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, fullTranslations())))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown id: got %d, want 404", resp.StatusCode)
	}

	// Non-UUID id -> 400 (not a 500 from the driver).
	resp = doAgentRequest(t, app, http.MethodPut, "/v1/agent/articles/not-a-uuid",
		fmt.Sprintf(`{"tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, fullTranslations())))
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("non-uuid id: got %d, want 400", resp.StatusCode)
	}
}

func TestAgentDeleteRemovesSEO(t *testing.T) {
	app, store := newAgentTestApp(t)

	seo := models.Translations{"en": {"title": "SEO", "description": "D"}, "ru": {"title": "СЕО", "description": "О"}}
	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles",
		fmt.Sprintf(`{"slug":"del","tags":["Flutter"],"published":true,"translations":%s,"seo":{"translations":%s}}`,
			mustJSON(t, fullTranslations()), mustJSON(t, seo)))
	var created struct {
		Article agentArticleOut `json:"article"`
	}
	decodeAgentJSON(t, resp.Body, &created)

	resp = doAgentRequest(t, app, http.MethodDelete, "/v1/agent/articles/"+created.Article.ID, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("delete: got %d, want 200", resp.StatusCode)
	}
	if len(store.pages) != 0 || len(store.seo) != 0 {
		t.Fatalf("pages/seo must be empty, got %d/%d", len(store.pages), len(store.seo))
	}

	resp = doAgentRequest(t, app, http.MethodDelete, "/v1/agent/articles/"+created.Article.ID, "")
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("delete twice: got %d, want 404", resp.StatusCode)
	}
}

func TestAgentListIncludesDrafts(t *testing.T) {
	app, _ := newAgentTestApp(t)

	doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles",
		fmt.Sprintf(`{"slug":"one","tags":["Flutter"],"published":true,"translations":%s}`, mustJSON(t, fullTranslations())))
	doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles",
		fmt.Sprintf(`{"slug":"two","tags":["Flutter"],"published":false,"translations":%s}`, mustJSON(t, fullTranslations())))

	resp := doAgentRequest(t, app, http.MethodGet, "/v1/agent/articles", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list: got %d, want 200", resp.StatusCode)
	}
	var out struct {
		Articles []agentArticleOut `json:"articles"`
	}
	decodeAgentJSON(t, resp.Body, &out)
	if len(out.Articles) != 2 {
		t.Fatalf("list len = %d, want 2", len(out.Articles))
	}
}

func TestAgentCreateRequiresTags(t *testing.T) {
	app, _ := newAgentTestApp(t)

	var out agentErrorResponse

	// No tags at all.
	body := fmt.Sprintf(`{"published":true,"translations":%s}`, mustJSON(t, fullTranslations()))
	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("no tags: got %d, want 422", resp.StatusCode)
	}
	decodeAgentJSON(t, resp.Body, &out)
	if out.Error != "validation_failed" || out.Message != "missing: tags" {
		t.Fatalf("no tags: got %s / %q, want validation_failed / missing: tags", out.Error, out.Message)
	}

	// Blank strings after trim count as missing too.
	body = fmt.Sprintf(`{"tags":["  ",""],"published":true,"translations":%s}`, mustJSON(t, fullTranslations()))
	resp = doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("blank tags: got %d, want 422", resp.StatusCode)
	}
	decodeAgentJSON(t, resp.Body, &out)
	if out.Message != "missing: tags" {
		t.Fatalf("blank tags: message = %q, want missing: tags", out.Message)
	}
}

func TestAgentCreateRejectsUnknownTags(t *testing.T) {
	app, _ := newAgentTestApp(t)

	body := fmt.Sprintf(`{"tags":["Flutter","Nocat"],"published":true,"translations":%s}`,
		mustJSON(t, fullTranslations()))
	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("got %d, want 422", resp.StatusCode)
	}
	var out agentErrorResponse
	decodeAgentJSON(t, resp.Body, &out)
	if out.Error != "validation_failed" {
		t.Fatalf("error code = %q, want validation_failed", out.Error)
	}
	if out.Message != "unknown stack tags: Nocat" {
		t.Fatalf("message = %q, want unknown stack tags: Nocat", out.Message)
	}
}

func TestAgentCreateCanonicalizesTags(t *testing.T) {
	app, _ := newAgentTestApp(t)

	// Case-insensitive catalog match, trim and dedupe: stored values are the
	// canonical catalog labels in request order.
	body := fmt.Sprintf(`{"tags":["flutter"," GO ","Flutter","figma"],"published":true,"translations":%s}`,
		mustJSON(t, fullTranslations()))
	resp := doAgentRequest(t, app, http.MethodPost, "/v1/agent/articles", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("got %d, want 201", resp.StatusCode)
	}
	var out struct {
		Article agentArticleOut `json:"article"`
	}
	decodeAgentJSON(t, resp.Body, &out)
	want := []string{"Flutter", "Go", "Figma"}
	if !slices.Equal(out.Article.Tags, want) {
		t.Fatalf("tags = %v, want %v", out.Article.Tags, want)
	}

	// Update accepts the canonical labels as-is.
	resp = doAgentRequest(t, app, http.MethodPut, "/v1/agent/articles/"+out.Article.ID,
		fmt.Sprintf(`{"tags":["Figma"],"published":true,"translations":%s}`, mustJSON(t, fullTranslations())))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update: got %d, want 200", resp.StatusCode)
	}
	decodeAgentJSON(t, resp.Body, &out)
	if !slices.Equal(out.Article.Tags, []string{"Figma"}) {
		t.Fatalf("updated tags = %v, want [Figma]", out.Article.Tags)
	}
}

func TestAgentListStack(t *testing.T) {
	app, _ := newAgentTestApp(t)

	resp := doAgentRequest(t, app, http.MethodGet, "/v1/agent/stack", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d, want 200", resp.StatusCode)
	}
	var out struct {
		Stack []models.StackItem `json:"stack"`
	}
	decodeAgentJSON(t, resp.Body, &out)
	// Published items only — the agent picks tags from what the site shows.
	if len(out.Stack) != 3 {
		t.Fatalf("stack len = %d, want 3", len(out.Stack))
	}
	for _, it := range out.Stack {
		if it.Slug == "php" {
			t.Fatal("unpublished item must not be listed")
		}
	}
}

// mustJSON marshals v or fails the test.
func mustJSON(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}
