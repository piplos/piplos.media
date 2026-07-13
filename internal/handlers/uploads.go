package handlers

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	apperrors "github.com/piplos-media/site/internal/errors"
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

// Upload accepts a multipart image and returns its public URL.
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

	name := uuid.NewString() + ext
	destPath := filepath.Join(h.dir, name)

	if err := os.WriteFile(destPath, body, 0o644); err != nil {
		return apperrors.ErrInternal("failed to save upload")
	}

	rel := "/uploads/" + name
	url := h.publicURL + rel
	if h.publicURL == "" {
		url = rel
	}

	return c.JSON(fiber.Map{
		"url":      url,
		"path":     rel,
		"filename": name,
		"mime":     mime.TypeByExtension(ext),
	})
}
