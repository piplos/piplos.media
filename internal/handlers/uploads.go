package handlers

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	apperrors "github.com/piplos/piplos.media/internal/errors"
)

const maxUploadBytes = 5 << 20 // 5 MiB

var allowedImageMIMEs = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// UploadsHandler stores and serves uploaded media files.
type UploadsHandler struct {
	dir       string
	publicURL string
}

// NewUploadsHandler creates an UploadsHandler.
func NewUploadsHandler(dir, publicURL string) *UploadsHandler {
	return &UploadsHandler{dir: dir, publicURL: strings.TrimRight(publicURL, "/")}
}

// uniqueName returns name, or name with a numeric suffix if it is taken in dir.
func uniqueName(dir, name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	candidate := name
	for i := 1; ; i++ {
		if _, err := os.Stat(filepath.Join(dir, candidate)); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d%s", base, i, ext)
	}
}

// Upload accepts a multipart image and returns its public URL.
// Optional form fields: path — target folder inside the upload dir;
// name — desired filename (extension is normalized by detected MIME).
func (h *UploadsHandler) Upload(c fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil || file == nil {
		return apperrors.ErrInvalidRequest("file is required")
	}
	if file.Size > maxUploadBytes {
		return apperrors.ErrInvalidRequest("file exceeds 5 MiB limit")
	}

	src, err := file.Open()
	if err != nil {
		return apperrors.ErrInternal("failed to open upload")
	}
	defer src.Close()

	body, err := io.ReadAll(io.LimitReader(src, maxUploadBytes+1))
	if err != nil {
		return apperrors.ErrInternal("failed to read upload")
	}
	if len(body) > maxUploadBytes {
		return apperrors.ErrInvalidRequest("file exceeds 5 MiB limit")
	}
	if len(body) == 0 {
		return apperrors.ErrInvalidRequest("empty file")
	}

	detected := http.DetectContentType(body)
	ext, ok := allowedImageMIMEs[detected]
	if !ok {
		return apperrors.ErrInvalidRequest("only JPEG, PNG, WebP and GIF images are allowed")
	}

	folderRel, folderAbs, ok := resolveUploadPath(h.dir, c.FormValue("path"))
	if !ok {
		return apperrors.ErrInvalidRequest("invalid path")
	}
	if st, err := os.Stat(folderAbs); err != nil || !st.IsDir() {
		return apperrors.ErrInvalidRequest("target folder not found")
	}

	name := uuid.NewString() + ext
	if requested := strings.TrimSpace(c.FormValue("name")); requested != "" {
		base := strings.TrimSuffix(requested, filepath.Ext(requested))
		if !validEntryName(base + ext) {
			return apperrors.ErrInvalidRequest("invalid file name")
		}
		name = uniqueName(folderAbs, base+ext)
	}
	destPath := filepath.Join(folderAbs, name)

	if err := os.WriteFile(destPath, body, 0o644); err != nil {
		return apperrors.ErrInternal("failed to save upload")
	}

	fileRel := path.Join(folderRel, name)
	return c.JSON(fiber.Map{
		"url":      uploadsFileURL(h.publicURL, fileRel),
		"path":     "/uploads/" + fileRel,
		"filename": name,
		"mime":     mime.TypeByExtension(ext),
	})
}
