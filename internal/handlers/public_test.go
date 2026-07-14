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
	published := publishedProjects(items, false)
	if len(published) != 1 || published[0].Slug != "pub" {
		t.Fatalf("expected only published project, got %+v", published)
	}
}

func TestPublishedProjectsFeaturedOnly(t *testing.T) {
	items := []models.Project{
		{Slug: "featured", Published: true, Featured: true},
		{Slug: "regular", Published: true, Featured: false},
		{Slug: "draft-featured", Published: false, Featured: true},
	}
	featured := publishedProjects(items, true)
	if len(featured) != 1 || featured[0].Slug != "featured" {
		t.Fatalf("expected only published featured project, got %+v", featured)
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
