package handlers

import (
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos/piplos.media/internal/errors"
)

// FilesHandler manages the media library (folders and files) inside the upload dir.
type FilesHandler struct {
	dir       string
	publicURL string
}

// NewFilesHandler creates a FilesHandler.
func NewFilesHandler(dir, publicURL string) *FilesHandler {
	return &FilesHandler{dir: dir, publicURL: strings.TrimRight(publicURL, "/")}
}

// resolveUploadPath validates a relative path and resolves it inside dir.
// The leading-slash Clean trick makes ".." unable to escape the root.
func resolveUploadPath(dir, raw string) (rel string, abs string, ok bool) {
	clean := path.Clean("/" + strings.ReplaceAll(raw, "\\", "/"))
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." {
		clean = ""
	}
	abs = filepath.Join(dir, filepath.FromSlash(clean))
	if abs != dir && !strings.HasPrefix(abs, dir+string(filepath.Separator)) {
		return "", "", false
	}
	return clean, abs, true
}

var invalidNameChars = regexp.MustCompile(`[/\\:*?"<>|]`)

// validEntryName reports whether name is acceptable for a folder or file.
func validEntryName(name string) bool {
	if name == "" || len(name) > 128 || strings.HasPrefix(name, ".") {
		return false
	}
	return !invalidNameChars.MatchString(name)
}

// validFolderPath reports whether every segment of a relative path is a valid
// entry name ("" = upload root, always valid).
func validFolderPath(rel string) bool {
	if rel == "" {
		return true
	}
	for _, seg := range strings.Split(rel, "/") {
		if !validEntryName(seg) {
			return false
		}
	}
	return true
}

// uploadsFileURL builds a public URL for a file relative to the upload dir.
func uploadsFileURL(publicURL, rel string) string {
	segs := strings.Split(rel, "/")
	for i, s := range segs {
		segs[i] = url.PathEscape(s)
	}
	p := "/uploads/" + strings.Join(segs, "/")
	if publicURL == "" {
		return p
	}
	return publicURL + p
}

type folderInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type fileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	URL     string    `json:"url"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// List returns folders and files of one directory.
// Query: path — relative folder path ("" = root).
func (h *FilesHandler) List(c fiber.Ctx) error {
	rel, abs, ok := resolveUploadPath(h.dir, c.Query("path"))
	if !ok {
		return apperrors.ErrInvalidRequest("invalid path")
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return apperrors.ErrNotFound("folder not found")
		}
		return apperrors.ErrInternal("failed to read folder")
	}

	folders := []folderInfo{}
	files := []fileInfo{}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		entryRel := path.Join(rel, name)
		if e.IsDir() {
			folders = append(folders, folderInfo{Name: name, Path: entryRel})
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{
			Name: name, Path: entryRel, URL: uploadsFileURL(h.publicURL, entryRel),
			Size: info.Size(), ModTime: info.ModTime(),
		})
	}
	sort.Slice(folders, func(i, j int) bool { return folders[i].Name < folders[j].Name })
	sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })

	return c.JSON(fiber.Map{"path": rel, "folders": folders, "files": files})
}

// CreateFolder creates a folder (with parents) inside the upload dir.
func (h *FilesHandler) CreateFolder(c fiber.Ctx) error {
	var req struct {
		Path string `json:"path"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	rel, abs, ok := resolveUploadPath(h.dir, req.Path)
	if !ok || rel == "" || !validEntryName(path.Base(rel)) {
		return apperrors.ErrInvalidRequest("invalid folder name")
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return apperrors.ErrInternal("failed to create folder")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"path": rel})
}

// Rename moves a single file or folder to a new relative path.
func (h *FilesHandler) Rename(c fiber.Ctx) error {
	var req struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	fromRel, fromAbs, okFrom := resolveUploadPath(h.dir, req.From)
	toRel, toAbs, okTo := resolveUploadPath(h.dir, req.To)
	if !okFrom || !okTo || fromRel == "" || toRel == "" || !validEntryName(path.Base(toRel)) {
		return apperrors.ErrInvalidRequest("invalid path")
	}
	if _, err := os.Stat(fromAbs); err != nil {
		return apperrors.ErrNotFound("source not found")
	}
	if _, err := os.Stat(toAbs); err == nil {
		return apperrors.ErrInvalidRequest("target already exists")
	}
	if err := os.MkdirAll(filepath.Dir(toAbs), 0o755); err != nil {
		return apperrors.ErrInternal("failed to prepare target folder")
	}
	if err := os.Rename(fromAbs, toAbs); err != nil {
		return apperrors.ErrInternal("failed to rename")
	}
	return c.JSON(fiber.Map{"path": toRel, "url": uploadsFileURL(h.publicURL, toRel)})
}

// Move relocates files/folders into a destination folder.
func (h *FilesHandler) Move(c fiber.Ctx) error {
	var req struct {
		Paths []string `json:"paths"`
		Dest  string   `json:"dest"`
	}
	if err := c.Bind().Body(&req); err != nil || len(req.Paths) == 0 {
		return apperrors.ErrInvalidRequest("paths are required")
	}
	destRel, destAbs, ok := resolveUploadPath(h.dir, req.Dest)
	if !ok {
		return apperrors.ErrInvalidRequest("invalid destination")
	}
	if st, err := os.Stat(destAbs); err != nil || !st.IsDir() {
		return apperrors.ErrInvalidRequest("destination folder not found")
	}

	moved := []string{}
	for _, p := range req.Paths {
		srcRel, srcAbs, ok := resolveUploadPath(h.dir, p)
		if !ok || srcRel == "" {
			return apperrors.ErrInvalidRequest("invalid path: " + p)
		}
		name := path.Base(srcRel)
		targetRel := path.Join(destRel, name)
		if targetRel == srcRel {
			continue
		}
		if strings.HasPrefix(targetRel+"/", srcRel+"/") {
			return apperrors.ErrInvalidRequest("cannot move a folder into itself: " + name)
		}
		targetAbs := filepath.Join(destAbs, name)
		if _, err := os.Stat(targetAbs); err == nil {
			return apperrors.ErrInvalidRequest("already exists in destination: " + name)
		}
		if err := os.Rename(srcAbs, targetAbs); err != nil {
			return apperrors.ErrInternal("failed to move " + name)
		}
		moved = append(moved, targetRel)
	}
	return c.JSON(fiber.Map{"moved": moved})
}

// Delete removes files/folders (folders — recursively).
func (h *FilesHandler) Delete(c fiber.Ctx) error {
	var req struct {
		Paths []string `json:"paths"`
	}
	if err := c.Bind().Body(&req); err != nil || len(req.Paths) == 0 {
		return apperrors.ErrInvalidRequest("paths are required")
	}
	for _, p := range req.Paths {
		rel, abs, ok := resolveUploadPath(h.dir, p)
		if !ok || rel == "" {
			return apperrors.ErrInvalidRequest("invalid path: " + p)
		}
		if err := os.RemoveAll(abs); err != nil {
			return apperrors.ErrInternal("failed to delete " + rel)
		}
	}
	return c.JSON(fiber.Map{"deleted": len(req.Paths)})
}
