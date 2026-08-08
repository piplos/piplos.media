package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofiber/fiber/v3"
)

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

func TestFilesListHidesSizedWebPVariants(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"cover.webp", "cover-480.webp", "cover-960.webp", "icon.svg"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	h := NewFilesHandler(dir, "https://api.test")
	app := fiber.New()
	app.Get("/files", h.List)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/files", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var body struct {
		Files []struct {
			Name string `json:"name"`
		} `json:"files"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, f := range body.Files {
		names[f.Name] = true
	}
	if !names["cover.webp"] || !names["icon.svg"] {
		t.Fatalf("expected masters listed, got %#v", names)
	}
	if names["cover-480.webp"] || names["cover-960.webp"] {
		t.Fatalf("sized variants must be hidden, got %#v", names)
	}
}
