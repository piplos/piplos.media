package models

import (
	"testing"
	"time"
)

func TestPageIsLive(t *testing.T) {
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name string
		page Page
		want bool
	}{
		{"draft", Page{Published: false}, false},
		{"published no schedule", Page{Published: true}, true},
		{"published in past", Page{Published: true, PublishAt: &past}, true},
		{"scheduled for future", Page{Published: true, PublishAt: &future}, false},
		{"draft with past schedule", Page{Published: false, PublishAt: &past}, false},
	}

	for _, tt := range tests {
		if got := tt.page.IsLive(now); got != tt.want {
			t.Errorf("%s: IsLive() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestEnabledLanguageCodes(t *testing.T) {
	got := EnabledLanguageCodes([]Language{
		{Code: "en", Enabled: true},
		{Code: "ru", Enabled: false},
		{Code: "de", Enabled: true},
	})
	if len(got) != 2 || got[0] != "en" || got[1] != "de" {
		t.Fatalf("EnabledLanguageCodes() = %v, want [en de]", got)
	}
}

func TestIsLegalPath(t *testing.T) {
	langs := []string{"en", "ru"}

	tests := []struct {
		path string
		want bool
	}{
		{"/legal/privacy", true},
		{"/en/legal/privacy", true},
		{"/ru/legal/terms", true},
		{"/de/legal/privacy", false},
		{"/en/portfolio", false},
		{"/legal/unknown", false},
	}

	for _, tt := range tests {
		if got := IsLegalPath(tt.path, langs); got != tt.want {
			t.Errorf("IsLegalPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
