package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	apperrors "github.com/piplos/piplos.media/internal/errors"
	"github.com/piplos/piplos.media/internal/media"
)

const maxUploadBytes = 5 << 20 // 5 MiB

var allowedImageMIMEs = map[string]string{
	"image/jpeg":    ".jpg",
	"image/png":     ".png",
	"image/webp":    ".webp",
	"image/gif":     ".gif",
	"image/svg+xml": ".svg",
}

// UploadsHandler stores and serves uploaded media files.
type UploadsHandler struct {
	dir       string
	publicURL string

	rebuildMu sync.Mutex
	rebuild   rebuildJobStatus
}

// rebuildJobStatus is the async WebP rebuild progress (staff).
type rebuildJobStatus struct {
	Running    bool      `json:"running"`
	OK         int       `json:"ok"`
	Failed     int       `json:"failed"`
	Error      string    `json:"error,omitempty"`
	StartedAt  time.Time `json:"started_at,omitempty"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}

// NewUploadsHandler creates an UploadsHandler.
func NewUploadsHandler(dir, publicURL string) *UploadsHandler {
	return &UploadsHandler{dir: dir, publicURL: strings.TrimRight(publicURL, "/")}
}

// uniqueName returns name, or name with a numeric suffix if it (or its WebP family) is taken.
func uniqueName(dir, name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	candidate := name
	for i := 1; ; i++ {
		if !media.NameConflicts(dir, candidate) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d%s", base, i, ext)
	}
}

// isSVGContent reports whether body looks like an SVG image.
// http.DetectContentType often classifies SVG as text/plain or text/xml.
func isSVGContent(body []byte) bool {
	trimmed := bytes.TrimSpace(body)
	trimmed = bytes.TrimPrefix(trimmed, []byte{0xEF, 0xBB, 0xBF})
	trimmed = bytes.TrimSpace(trimmed)
	if len(trimmed) > 512 {
		trimmed = trimmed[:512]
	}
	if bytes.HasPrefix(trimmed, []byte("<svg")) {
		return true
	}
	return bytes.HasPrefix(trimmed, []byte("<?xml")) && bytes.Contains(trimmed, []byte("<svg"))
}

func detectUploadType(body []byte) (ext, mimeType string, ok bool) {
	detected := http.DetectContentType(body)
	if ext, ok = allowedImageMIMEs[detected]; ok {
		return ext, detected, true
	}
	if isSVGContent(body) {
		return ".svg", "image/svg+xml", true
	}
	return "", "", false
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

	ext, mimeType, ok := detectUploadType(body)
	if !ok {
		return apperrors.ErrInvalidRequest("only JPEG, PNG, WebP, GIF and SVG images are allowed")
	}

	folderRel, folderAbs, ok := resolveUploadPath(h.dir, c.FormValue("path"))
	if !ok || !validFolderPath(folderRel) {
		return apperrors.ErrInvalidRequest("invalid path")
	}
	// Target folder is created on demand: forms upload straight into an
	// entity folder (projects/<slug> etc.) that may not exist yet.
	if err := os.MkdirAll(folderAbs, 0o755); err != nil {
		return apperrors.ErrInternal("failed to create target folder")
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
	resp := fiber.Map{
		"url":      uploadsFileURL(h.publicURL, fileRel),
		"path":     "/uploads/" + fileRel,
		"filename": name,
		"mime":     mimeType,
	}

	// Raster uploads get WebP siblings (full + 480/960) for the public site.
	if media.IsRasterExt(ext) {
		variants, err := media.GenerateVariants(destPath)
		if err != nil {
			// Original kept on failure; partial WebP outputs are rolled back.
			resp["variants_error"] = err.Error()
		} else if variants.Full != "" {
			fullRel := path.Join(folderRel, variants.Full)
			webpURL := uploadsFileURL(h.publicURL, fullRel)
			resp["webp"] = webpURL
			resp["webp_path"] = "/uploads/" + fullRel
			sized := fiber.Map{}
			for _, w := range media.VariantWidths {
				sizedName := media.SizedWebPName(variants.Full, w)
				sizedRel := path.Join(folderRel, sizedName)
				sized[fmt.Sprintf("%d", w)] = uploadsFileURL(h.publicURL, sizedRel)
			}
			resp["webp_sizes"] = sized
			resp["url"] = webpURL
			resp["path"] = resp["webp_path"]
			resp["filename"] = variants.Full
			resp["mime"] = "image/webp"
		}
	}

	return c.JSON(resp)
}

// RebuildVariants starts a background walk that regenerates WebP sidecars
// for every raster original (staff; used after deploy / for legacy PNG).
// Returns 202 immediately so reverse proxies do not 504 on long jobs.
func (h *UploadsHandler) RebuildVariants(c fiber.Ctx) error {
	h.rebuildMu.Lock()
	if h.rebuild.Running {
		h.rebuildMu.Unlock()
		return apperrors.ErrConflict("rebuild already running")
	}
	h.rebuild = rebuildJobStatus{Running: true, StartedAt: time.Now().UTC()}
	st := h.rebuild
	h.rebuildMu.Unlock()

	go h.runRebuild()

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": st})
}

// RebuildVariantsStatus returns the last/current rebuild job state.
func (h *UploadsHandler) RebuildVariantsStatus(c fiber.Ctx) error {
	h.rebuildMu.Lock()
	st := h.rebuild
	h.rebuildMu.Unlock()
	return c.JSON(fiber.Map{"status": st})
}

func (h *UploadsHandler) runRebuild() {
	ok, failed, err := media.RebuildDir(h.dir)
	h.rebuildMu.Lock()
	defer h.rebuildMu.Unlock()
	h.rebuild.Running = false
	h.rebuild.OK = ok
	h.rebuild.Failed = failed
	h.rebuild.FinishedAt = time.Now().UTC()
	if err != nil {
		h.rebuild.Error = err.Error()
	} else {
		h.rebuild.Error = ""
	}
}
