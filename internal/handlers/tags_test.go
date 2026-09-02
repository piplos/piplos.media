package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/piplos/piplos.media/internal/models"
)

func catalogForTags() []models.StackItem {
	return []models.StackItem{
		{Slug: "flutter", Label: "Flutter", Published: true},
		{Slug: "golang", Label: "Go", Published: true},
		{Slug: "figma", Label: "Figma", Published: true},
		{Slug: "php", Label: "PHP", Published: false},
	}
}

func TestCanonicalizeTagsMapsToCatalogLabels(t *testing.T) {
	got, err := canonicalizeTags([]string{"flutter", " GO ", "Flutter", "figma"}, catalogForTags())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"Flutter", "Go", "Figma"}
	if len(got) != len(want) {
		t.Fatalf("tags = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tags = %v, want %v", got, want)
		}
	}
}

func TestCanonicalizeTagsAcceptsDraftCatalogItems(t *testing.T) {
	// Validation covers the whole catalog, so a draft stack item (used by an
	// admin-created article) never breaks updates of that article.
	got, err := canonicalizeTags([]string{"php"}, catalogForTags())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "PHP" {
		t.Fatalf("tags = %v, want [PHP]", got)
	}
}

func TestCanonicalizeTagsErrors(t *testing.T) {
	cases := []struct {
		name string
		tags []string
		want string
	}{
		{"missing", nil, "missing: tags"},
		{"blank after trim", []string{"  ", ""}, "missing: tags"},
		{"unknown", []string{"Flutter", "Nocat"}, "unknown stack tags: Nocat"},
	}
	for _, tc := range cases {
		_, err := canonicalizeTags(tc.tags, catalogForTags())
		if err == nil {
			t.Fatalf("%s: expected error", tc.name)
		}
		if err.Error() != tc.want {
			t.Fatalf("%s: err = %q, want %q", tc.name, err.Error(), tc.want)
		}
	}
}

func TestCanonicalizeTagsLimit(t *testing.T) {
	// 21 distinct catalog-valid tags: the limit applies after
	// canonicalization, so duplicates cannot smuggle a bigger stack in.
	items := make([]models.StackItem, 0, agentStackTagLimit+5)
	tags := make([]string, 0, agentStackTagLimit+1)
	for i := 0; i <= agentStackTagLimit; i++ {
		label := fmt.Sprintf("Tech %d", i)
		items = append(items, models.StackItem{Slug: fmt.Sprintf("t%d", i), Label: label, Published: true})
		tags = append(tags, strings.ToLower(label))
	}
	_, err := canonicalizeTags(tags, items)
	if err == nil || err.Error() != "tags: at most 20 items allowed" {
		t.Fatalf("err = %v, want limit error", err)
	}
}
