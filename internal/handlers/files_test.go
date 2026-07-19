package handlers

import "testing"

func TestResolveUploadPath(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		raw     string
		wantRel string
		wantOK  bool
	}{
		{"", "", true},
		{"/", "", true},
		{"folder", "folder", true},
		{"folder/sub", "folder/sub", true},
		{"/folder//sub/", "folder/sub", true},
		{"../escape", "escape", true}, // Clean("/../escape") = "/escape" — stays inside root
		{"folder/../../etc", "etc", true},
		{"a/./b", "a/b", true},
	}
	for _, tc := range cases {
		rel, _, ok := resolveUploadPath(dir, tc.raw)
		if ok != tc.wantOK || rel != tc.wantRel {
			t.Errorf("resolveUploadPath(%q) = (%q, %v), want (%q, %v)", tc.raw, rel, ok, tc.wantRel, tc.wantOK)
		}
	}
}

func TestValidEntryName(t *testing.T) {
	valid := []string{"image.png", "My Folder", "проект-1", "a"}
	for _, name := range valid {
		if !validEntryName(name) {
			t.Errorf("validEntryName(%q) = false, want true", name)
		}
	}
	invalid := []string{"", ".hidden", "a/b", `a\b`, "a:b", "a*b", "a?b", `a"b`, "a<b", "a|b"}
	for _, name := range invalid {
		if validEntryName(name) {
			t.Errorf("validEntryName(%q) = true, want false", name)
		}
	}
}

func TestValidFolderPath(t *testing.T) {
	valid := []string{"", "folder", "folder/sub", "проекты/site-dev"}
	for _, p := range valid {
		if !validFolderPath(p) {
			t.Errorf("validFolderPath(%q) = false, want true", p)
		}
	}
	invalid := []string{".hidden", "folder/.hidden", "a:b/c", "ok/a*b"}
	for _, p := range invalid {
		if validFolderPath(p) {
			t.Errorf("validFolderPath(%q) = true, want false", p)
		}
	}
}

func TestUploadsFileURL(t *testing.T) {
	if got := uploadsFileURL("", "folder/img.png"); got != "/uploads/folder/img.png" {
		t.Errorf("uploadsFileURL empty base = %q", got)
	}
	if got := uploadsFileURL("https://api.test", "img.png"); got != "https://api.test/uploads/img.png" {
		t.Errorf("uploadsFileURL with base = %q", got)
	}
	if got := uploadsFileURL("", "папка/с пробелом.png"); got != "/uploads/%D0%BF%D0%B0%D0%BF%D0%BA%D0%B0/%D1%81%20%D0%BF%D1%80%D0%BE%D0%B1%D0%B5%D0%BB%D0%BE%D0%BC.png" {
		t.Errorf("uploadsFileURL escaping = %q", got)
	}
}
