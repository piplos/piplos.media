package handlers

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
)

func TestIsSVGContent(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"bare svg", `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path/></svg>`, true},
		{"xml svg", `<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"><path/></svg>`, true},
		{"bom svg", "\ufeff<svg xmlns=\"http://www.w3.org/2000/svg\"><path/></svg>", true},
		{"png", "\x89PNG\r\n\x1a\n", false},
		{"html", "<html><body>hi</body></html>", false},
		{"xml not svg", `<?xml version="1.0"?><root/>`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isSVGContent([]byte(tc.body)); got != tc.want {
				t.Fatalf("isSVGContent() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDetectUploadTypeSVG(t *testing.T) {
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle r="1"/></svg>`)
	ext, mimeType, ok := detectUploadType(svg)
	if !ok || ext != ".svg" || mimeType != "image/svg+xml" {
		t.Fatalf("detectUploadType(svg) = (%q, %q, %v), want (.svg, image/svg+xml, true)", ext, mimeType, ok)
	}
	if sniffed := http.DetectContentType(svg); sniffed == "image/svg+xml" {
		t.Logf("note: DetectContentType returned %q", sniffed)
	}
}

func TestUploadsHandlerAcceptsSVG(t *testing.T) {
	dir := t.TempDir()
	stackDir := filepath.Join(dir, "stack")
	if err := os.MkdirAll(stackDir, 0o755); err != nil {
		t.Fatal(err)
	}

	h := NewUploadsHandler(dir, "")
	app := fiber.New()
	app.Post("/upload", h.Upload)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "postgresql.svg")
	if err != nil {
		t.Fatal(err)
	}
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M0 0"/></svg>`)
	if _, err := part.Write(svg); err != nil {
		t.Fatal(err)
	}
	_ = writer.WriteField("path", "stack")
	_ = writer.WriteField("name", "postgresql.svg")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", resp.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(stackDir, "postgresql.svg")); err != nil {
		t.Fatalf("saved file missing: %v", err)
	}
}

func TestUploadsHandlerGeneratesWebPVariants(t *testing.T) {
	dir := t.TempDir()
	h := NewUploadsHandler(dir, "https://api.test")
	app := fiber.New()
	app.Post("/upload", h.Upload)

	var pngBuf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 32, 24))
	if err := png.Encode(&pngBuf, img); err != nil {
		t.Fatal(err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "cover.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(pngBuf.Bytes()); err != nil {
		t.Fatal(err)
	}
	_ = writer.WriteField("path", "projects/demo")
	_ = writer.WriteField("name", "cover.png")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", resp.StatusCode)
	}

	dest := filepath.Join(dir, "projects", "demo")
	for _, name := range []string{"cover.webp", "cover-480.webp", "cover-960.webp"} {
		if _, err := os.Stat(filepath.Join(dest, name)); err != nil {
			t.Fatalf("expected %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dest, "cover.png")); !os.IsNotExist(err) {
		t.Fatalf("expected original PNG removed, got err=%v", err)
	}
}

func TestRebuildVariantsReturnsAccepted(t *testing.T) {
	dir := t.TempDir()
	writePNG := func(name string) {
		t.Helper()
		var buf bytes.Buffer
		if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 16, 12))); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writePNG("a.png")

	h := NewUploadsHandler(dir, "")
	app := fiber.New()
	app.Post("/rebuild", h.RebuildVariants)
	app.Get("/rebuild", h.RebuildVariantsStatus)

	resp, err := app.Test(httptest.NewRequest(http.MethodPost, "/rebuild", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusAccepted {
		t.Fatalf("status = %d, want 202", resp.StatusCode)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		h.rebuildMu.Lock()
		running := h.rebuild.Running
		ok := h.rebuild.OK
		h.rebuildMu.Unlock()
		if !running {
			if ok < 1 {
				t.Fatalf("expected ok>=1, got %d", ok)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("rebuild did not finish")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestUploadDoesNotOverwriteExistingFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "icon.svg"), []byte("<svg old/>"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := NewUploadsHandler(dir, "")
	app := fiber.New()
	app.Post("/upload", h.Upload)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "icon.svg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg"><path/></svg>`)); err != nil {
		t.Fatal(err)
	}
	_ = writer.WriteField("name", "icon.svg")
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload status = %d, want 200", resp.StatusCode)
	}

	var out struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Filename != "icon-1.svg" {
		t.Fatalf("filename = %q, want icon-1.svg (existing must not be overwritten)", out.Filename)
	}
	raw, err := os.ReadFile(filepath.Join(dir, "icon.svg"))
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "<svg old/>" {
		t.Fatalf("existing file was overwritten: %q", raw)
	}
}
