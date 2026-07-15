package models

import "testing"

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
