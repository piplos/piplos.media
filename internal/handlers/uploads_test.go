package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

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
