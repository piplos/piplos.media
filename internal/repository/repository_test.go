package repository

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/piplos/piplos.media/internal/models"
)

func TestSyncProjectCategories(t *testing.T) {
	cases := []struct {
		name              string
		oldCategory       string
		newCategory       string
		categories        []string
		want              []string
		wantSameSlice     bool // newCategory == "" must return the input slice unchanged
	}{
		{
			name:          "empty new category returns input as-is",
			oldCategory:   "web",
			newCategory:   "",
			categories:    []string{"web", "design"},
			want:          []string{"web", "design"},
			wantSameSlice: true,
		},
		{
			name:        "moves project into new category",
			oldCategory: "web",
			newCategory: "design",
			categories:  []string{"web"},
			want:        []string{"design"},
		},
		{
			name:        "prepends new category when absent",
			oldCategory: "",
			newCategory: "design",
			categories:  []string{"web"},
			want:        []string{"design", "web"},
		},
		{
			name:        "drops empty entries",
			oldCategory: "",
			newCategory: "design",
			categories:  []string{"", "web"},
			want:        []string{"design", "web"},
		},
		{
			name:        "deduplicates repeated entries",
			oldCategory: "",
			newCategory: "design",
			categories:  []string{"web", "web"},
			want:        []string{"design", "web"},
		},
		{
			name:        "same old and new category keeps position of others",
			oldCategory: "web",
			newCategory: "web",
			categories:  []string{"web", "mobile"},
			want:        []string{"web", "mobile"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := syncProjectCategories(tc.oldCategory, tc.newCategory, tc.categories)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			if tc.wantSameSlice && !reflect.DeepEqual(&got[0], &tc.categories[0]) {
				t.Fatal("empty new category must return the input slice")
			}
		})
	}
}

func TestNormalizeMaskedSensitiveFields(t *testing.T) {
	t.Run("masked field keeps stored value", func(t *testing.T) {
		got := normalizeMaskedSensitiveFields(
			`{"api_key":"stored-secret","url":"https://x"}`,
			`{"api_key":"****","url":"https://y"}`,
			[]string{"api_key"},
		)
		var m map[string]any
		if err := json.Unmarshal([]byte(got), &m); err != nil {
			t.Fatal(err)
		}
		if m["api_key"] != "stored-secret" {
			t.Fatalf("api_key = %v, want stored-secret", m["api_key"])
		}
		if m["url"] != "https://y" {
			t.Fatalf("url = %v, want https://y (non-sensitive edit lost)", m["url"])
		}
	})

	t.Run("masked field with no stored value becomes empty string", func(t *testing.T) {
		got := normalizeMaskedSensitiveFields(`{}`, `{"api_key":"****"}`, []string{"api_key"})
		var m map[string]any
		if err := json.Unmarshal([]byte(got), &m); err != nil {
			t.Fatal(err)
		}
		if v, ok := m["api_key"].(string); !ok || v != "" {
			t.Fatalf("api_key = %v, want empty string", m["api_key"])
		}
	})

	t.Run("unmasked values are untouched", func(t *testing.T) {
		got := normalizeMaskedSensitiveFields(
			`{"api_key":"old"}`,
			`{"api_key":"new-value"}`,
			[]string{"api_key"},
		)
		var m map[string]any
		if err := json.Unmarshal([]byte(got), &m); err != nil {
			t.Fatal(err)
		}
		if m["api_key"] != "new-value" {
			t.Fatalf("api_key = %v, want new-value", m["api_key"])
		}
	})

	t.Run("no sensitive fields returns input unchanged", func(t *testing.T) {
		in := `{"a":1}`
		if got := normalizeMaskedSensitiveFields(`{}`, in, nil); got != in {
			t.Fatalf("got %s, want %s", got, in)
		}
	})

	t.Run("invalid new JSON returned as-is", func(t *testing.T) {
		in := `{not-json`
		if got := normalizeMaskedSensitiveFields(`{}`, in, []string{"k"}); got != in {
			t.Fatalf("got %s, want %s", got, in)
		}
	})

	t.Run("field missing from new JSON is not re-added", func(t *testing.T) {
		got := normalizeMaskedSensitiveFields(
			`{"api_key":"stored-secret"}`,
			`{"other":"value"}`,
			[]string{"api_key"},
		)
		var m map[string]any
		if err := json.Unmarshal([]byte(got), &m); err != nil {
			t.Fatal(err)
		}
		if _, ok := m["api_key"]; ok {
			t.Fatal("removed sensitive field must not be resurrected from old value")
		}
	})
}

func TestTranslationsFromJSON(t *testing.T) {
	if got := translationsFromJSON(nil); len(got) != 0 {
		t.Fatalf("nil raw: got %v, want empty", got)
	}
	if got := translationsFromJSON([]byte(`{not-json`)); len(got) != 0 {
		t.Fatalf("invalid raw: got %v, want empty", got)
	}
	got := translationsFromJSON([]byte(`{"en":{"title":"Hi"}}`))
	if got["en"]["title"] != "Hi" {
		t.Fatalf("got %v, want en.title=Hi", got)
	}
}

func TestProjectLinksJSONRoundTrip(t *testing.T) {
	raw, err := projectLinksJSON([]models.ProjectLink{
		{URL: "https://x.io", Label: "X", Kind: models.ProjectLinkWebsite},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := projectLinksFromJSON(raw)
	if len(got) != 1 || got[0].URL != "https://x.io" || got[0].Kind != models.ProjectLinkWebsite {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

func TestProjectLinksFromJSONNilSafety(t *testing.T) {
	if got := projectLinksFromJSON(nil); got == nil || len(got) != 0 {
		t.Fatalf("nil raw: got %#v, want non-nil empty slice", got)
	}
	if got := projectLinksFromJSON([]byte(`{not-json`)); got == nil || len(got) != 0 {
		t.Fatalf("invalid raw: got %#v, want non-nil empty slice", got)
	}
	if got := projectLinksFromJSON([]byte(`null`)); got == nil || len(got) != 0 {
		t.Fatalf("null raw: got %#v, want non-nil empty slice", got)
	}
}
