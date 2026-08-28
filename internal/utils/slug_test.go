package utils

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Привет, мир!", "privet-mir"},
		{"Ёлка и Ёж", "elka-i-ezh"},
		{"Тестовая статья про Flutter", "testovaya-statya-pro-flutter"},
		{"Café & Bar", "cafe-bar"},
		{"Already ASCII", "already-ascii"},
		{"  --trimmed--  ", "trimmed"},
		{"", ""},
		{"!!!", ""},
		{"ТЪЯЖЁЛЫЙ ЩУК", "tyazhelyy-shchuk"},
	}

	for _, tt := range tests {
		if got := Slugify(tt.in); got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
