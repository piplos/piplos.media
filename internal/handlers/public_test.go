package handlers

import (
	"testing"

	"github.com/piplos/piplos.media/internal/models"
)

func TestPublishedProjectsFiltersUnpublished(t *testing.T) {
	items := []models.Project{
		{Slug: "pub", Published: true},
		{Slug: "draft", Published: false},
	}
	published := publishedProjects(items)
	if len(published) != 1 || published[0].Slug != "pub" {
		t.Fatalf("expected only published project, got %+v", published)
	}
}

func TestFilterPublishedProjectsFeaturedOnly(t *testing.T) {
	items := []models.Project{
		{Slug: "featured", Published: true, Featured: true},
		{Slug: "regular", Published: true, Featured: false},
		{Slug: "draft-featured", Published: false, Featured: true},
	}
	got := filterPublishedProjects(items, projectListQuery{featuredOnly: true})
	if len(got) != 1 || got[0].Slug != "featured" {
		t.Fatalf("expected only published featured project, got %+v", got)
	}
}

func TestFilteredTranslations(t *testing.T) {
	tr := models.Translations{
		"en": {"title": "Hello"},
		"ru": {"title": "Привет"},
	}

	got := filteredTranslations(tr, "ru")
	if len(got) != 1 || got["ru"]["title"] != "Привет" {
		t.Fatalf("expected only ru translation, got %+v", got)
	}

	// Пустой lang — без фильтрации.
	if got := filteredTranslations(tr, ""); len(got) != 2 {
		t.Fatalf("expected all translations for empty lang, got %+v", got)
	}

	// Неизвестный язык — полный набор для клиентского fallback.
	if got := filteredTranslations(tr, "de"); len(got) != 2 {
		t.Fatalf("expected all translations for unknown lang, got %+v", got)
	}
}

func TestFilteredLegalTranslations(t *testing.T) {
	tr := models.LegalTranslations{
		"en": {Title: "Privacy"},
		"ru": {Title: "Конфиденциальность"},
	}

	got := filteredLegalTranslations(tr, "en")
	if len(got) != 1 || got["en"].Title != "Privacy" {
		t.Fatalf("expected only en translation, got %+v", got)
	}

	if got := filteredLegalTranslations(tr, "de"); len(got) != 2 {
		t.Fatalf("expected all translations for unknown lang, got %+v", got)
	}
}

func TestFilterPublishedProjectsFeaturedLimitPreservesGlobalOrder(t *testing.T) {
	// ListProjects order = global_sort_order; featured+limit must keep that relative order.
	items := []models.Project{
		{Slug: "g0", Published: true, Featured: false},
		{Slug: "f1", Published: true, Featured: true},
		{Slug: "g1", Published: true, Featured: false},
		{Slug: "f2", Published: true, Featured: true},
		{Slug: "f3", Published: true, Featured: true},
		{Slug: "f4", Published: true, Featured: true},
	}
	got := filterPublishedProjects(items, projectListQuery{featuredOnly: true, limit: 3})
	if len(got) != 3 || got[0].Slug != "f1" || got[1].Slug != "f2" || got[2].Slug != "f3" {
		t.Fatalf("expected f1,f2,f3 in ListProjects order, got %+v", got)
	}
}

func TestFilterPublishedProjectsByCategoryLimit(t *testing.T) {
	items := []models.Project{
		{Slug: "a", Published: true, Category: "web", SortOrder: 2, Year: 2024},
		{Slug: "b", Published: true, Category: "web", SortOrder: 1, Year: 2023},
		{Slug: "c", Published: true, Category: "mobile", SortOrder: 1, Year: 2025},
		{Slug: "draft", Published: false, Category: "web", SortOrder: 0, Year: 2025},
	}
	got := filterPublishedProjects(items, projectListQuery{category: "web", limit: 1})
	if len(got) != 1 || got[0].Slug != "b" {
		t.Fatalf("expected first by sort_order=b, got %+v", got)
	}
}

func TestFilterPublishedProjectsByCategoriesSlice(t *testing.T) {
	items := []models.Project{
		{Slug: "primary", Published: true, Category: "web"},
		{Slug: "extra-only", Published: true, Category: "other", Categories: []string{"web", "mobile"}},
		{Slug: "no", Published: true, Category: "mobile", Categories: []string{"mobile"}},
	}
	got := filterPublishedProjects(items, projectListQuery{category: "web"})
	if len(got) != 2 || got[0].Slug != "primary" || got[1].Slug != "extra-only" {
		t.Fatalf("expected primary+extra-only via Category/Categories[], got %+v", got)
	}
}

func TestFilterPublishedProjectsBySlugs(t *testing.T) {
	items := []models.Project{
		{Slug: "a", Published: true},
		{Slug: "b", Published: true},
		{Slug: "c", Published: false},
	}
	got := filterPublishedProjects(items, projectListQuery{slugs: []string{"b", "missing", "a", "c"}})
	if len(got) != 2 || got[0].Slug != "b" || got[1].Slug != "a" {
		t.Fatalf("expected slug order b,a (skip draft/missing), got %+v", got)
	}
}

func TestFilterPublishedProjectsSlugsOverrideFilters(t *testing.T) {
	items := []models.Project{
		{Slug: "a", Published: true, Featured: false, Category: "web", Tags: []string{"Go"}},
		{Slug: "b", Published: true, Featured: true, Category: "mobile", Tags: []string{"TS"}},
	}
	got := filterPublishedProjects(items, projectListQuery{
		slugs:        []string{"a", "b"},
		featuredOnly: true,
		category:     "mobile",
		tags:         []string{"TS"},
	})
	if len(got) != 2 || got[0].Slug != "a" || got[1].Slug != "b" {
		t.Fatalf("slugs must ignore featured/category/tags, got %+v", got)
	}
}

func TestFilterPublishedProjectsByTags(t *testing.T) {
	items := []models.Project{
		{Slug: "go", Published: true, Tags: []string{"Go", "Postgres"}},
		{Slug: "js", Published: true, Tags: []string{"TypeScript"}},
	}
	got := filterPublishedProjects(items, projectListQuery{tags: []string{"go"}, limit: 5})
	if len(got) != 1 || got[0].Slug != "go" {
		t.Fatalf("expected go project, got %+v", got)
	}
}

func TestPrepareProjectTranslationsSummary(t *testing.T) {
	tr := models.Translations{
		"en": {"title": "T", "solution": "# Big", "result": "Done"},
	}
	got := prepareProjectTranslations(tr, "en", "summary")
	if got["en"]["title"] != "T" {
		t.Fatalf("summary should keep title, got %+v", got)
	}
	if _, ok := got["en"]["solution"]; ok {
		t.Fatalf("summary should omit solution key, got %+v", got)
	}
	if _, ok := got["en"]["result"]; ok {
		t.Fatalf("summary should omit result key, got %+v", got)
	}
}

func TestPrepareServiceTranslationsSummary(t *testing.T) {
	tr := models.Translations{
		"en": {"title": "Web", "body": "# Long"},
	}
	got := prepareServiceTranslations(tr, "en", "summary")
	if got["en"]["title"] != "Web" {
		t.Fatalf("summary should keep title, got %+v", got)
	}
	if _, ok := got["en"]["body"]; ok {
		t.Fatalf("summary should omit body key, got %+v", got)
	}
}

func TestFilterPublishedServicesBySlugsAndLimit(t *testing.T) {
	items := []models.Service{
		{Slug: "web", Published: true, SortOrder: 2},
		{Slug: "mobile", Published: true, SortOrder: 1},
		{Slug: "draft", Published: false, SortOrder: 0},
	}
	got := filterPublishedServices(items, serviceListQuery{slugs: []string{"mobile", "web"}, limit: 1})
	if len(got) != 1 || got[0].Slug != "mobile" {
		t.Fatalf("expected mobile only (limit 1 in slug order), got %+v", got)
	}
}

func TestFilterPublishedServicesByTagsCaseInsensitive(t *testing.T) {
	items := []models.Service{
		{Slug: "web", Published: true, Tags: []string{"Frontend"}, SortOrder: 1},
		{Slug: "api", Published: true, Tags: []string{"Backend"}, SortOrder: 2},
	}
	got := filterPublishedServices(items, serviceListQuery{tags: []string{"frontend"}})
	if len(got) != 1 || got[0].Slug != "web" {
		t.Fatalf("expected web via case-insensitive tag, got %+v", got)
	}
}

func TestParseCSVQueryEmptyAndWhitespace(t *testing.T) {
	if got := parseCSVQuery(""); got != nil {
		t.Fatalf("empty CSV → nil, got %+v", got)
	}
	if got := parseCSVQuery(" , , "); got != nil && len(got) != 0 {
		t.Fatalf("whitespace-only CSV → empty, got %+v", got)
	}
	got := parseCSVQuery(" a, ,b ,")
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("expected [a b], got %+v", got)
	}
}

func TestParseLimitQueryInvalidAndCap(t *testing.T) {
	cases := []struct {
		raw  string
		want int
	}{
		{"", 0},
		{"0", 0},
		{"-3", 0},
		{"abc", 0},
		{"5", 5},
		{"1000", maxPublicListLimit},
	}
	for _, tc := range cases {
		if got := parseLimitQuery(tc.raw); got != tc.want {
			t.Fatalf("parseLimitQuery(%q)=%d, want %d", tc.raw, got, tc.want)
		}
	}
}

func TestParseSummaryMode(t *testing.T) {
	if got := parseSummaryMode("SUMMARY"); got != "summary" {
		t.Fatalf("expected summary, got %q", got)
	}
	if got := parseSummaryMode("full"); got != "" {
		t.Fatalf("expected empty for non-summary, got %q", got)
	}
}

func TestHasAnyTag(t *testing.T) {
	if !hasAnyTag([]string{"Go"}, []string{"go"}) {
		t.Fatal("expected case-insensitive match")
	}
	if hasAnyTag([]string{"Go"}, []string{"rust"}) {
		t.Fatal("expected no match")
	}
	if !hasAnyTag([]string{"Go"}, nil) {
		t.Fatal("empty wanted should match")
	}
}
