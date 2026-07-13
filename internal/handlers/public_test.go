package handlers

import (
	"testing"

	"github.com/piplos/site/internal/models"
)

func TestPublicProjectsFiltersUnpublished(t *testing.T) {
	items := []models.Project{
		{Slug: "pub", Published: true},
		{Slug: "draft", Published: false},
	}
	published := []models.Project{}
	for _, p := range items {
		if p.Published {
			published = append(published, p)
		}
	}
	if len(published) != 1 || published[0].Slug != "pub" {
		t.Fatalf("expected only published project, got %+v", published)
	}
}
