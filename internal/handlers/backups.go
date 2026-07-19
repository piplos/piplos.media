package handlers

import (
	"errors"
	"io/fs"

	"github.com/gofiber/fiber/v3"

	apperrors "github.com/piplos/piplos.media/internal/errors"
	"github.com/piplos/piplos.media/internal/services/backup"
)

// BackupsHandler manages backup archives (create, restore, download, delete).
type BackupsHandler struct {
	svc *backup.Service
}

// NewBackupsHandler creates a BackupsHandler.
func NewBackupsHandler(svc *backup.Service) *BackupsHandler {
	return &BackupsHandler{svc: svc}
}

// List returns stored archives (local + S3 when configured) and the current status.
func (h *BackupsHandler) List(c fiber.Ctx) error {
	archives, err := h.svc.List(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to list backups: " + err.Error())
	}
	return c.JSON(fiber.Map{"archives": archives, "status": h.svc.Status()})
}

// Status returns the current backup/restore operation state.
func (h *BackupsHandler) Status(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": h.svc.Status()})
}

type runBackupRequest struct {
	Type    string `json:"type"`
	Storage string `json:"storage"`
}

// Run starts a backup in the background and returns 202.
// Body: {type: full|db|files, storage: local|s3}.
func (h *BackupsHandler) Run(c fiber.Ctx) error {
	var req runBackupRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	if req.Type == "" || req.Storage == "" {
		// По умолчанию — как в настройках расписания.
		cfg, err := h.svc.Settings(c.Context())
		if err != nil {
			return apperrors.ErrInternal("failed to load backup settings")
		}
		if req.Type == "" {
			req.Type = cfg.Type
		}
		if req.Storage == "" {
			req.Storage = cfg.Storage
		}
	}

	if err := h.svc.StartBackup(req.Type, req.Storage); err != nil {
		if errors.Is(err, backup.ErrBusy) {
			return apperrors.ErrConflict("another backup operation is already running")
		}
		return apperrors.ErrInvalidRequest(err.Error())
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": h.svc.Status()})
}

type restoreRequest struct {
	Storage string `json:"storage"`
	Name    string `json:"name"`
}

// Restore starts a restore in the background and returns 202.
// Body: {storage: local|s3, name: backup-....tar.gz}.
func (h *BackupsHandler) Restore(c fiber.Ctx) error {
	var req restoreRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	if !backup.ValidArchiveName(req.Name) {
		return apperrors.ErrInvalidRequest("invalid archive name")
	}
	// Проверяем доступность архива до старта фоновой операции.
	rc, _, err := h.svc.Open(c.Context(), req.Storage, req.Name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return apperrors.ErrNotFound("archive not found")
		}
		return apperrors.ErrInvalidRequest(err.Error())
	}
	rc.Close()

	if err := h.svc.StartRestore(req.Storage, req.Name); err != nil {
		if errors.Is(err, backup.ErrBusy) {
			return apperrors.ErrConflict("another backup operation is already running")
		}
		return apperrors.ErrInvalidRequest(err.Error())
	}
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"status": h.svc.Status()})
}

// Download streams an archive to the client.
// Query: storage — local|s3, name — archive file name.
func (h *BackupsHandler) Download(c fiber.Ctx) error {
	name := c.Query("name")
	storageKind := c.Query("storage")
	if !backup.ValidArchiveName(name) {
		return apperrors.ErrInvalidRequest("invalid archive name")
	}
	rc, size, err := h.svc.Open(c.Context(), storageKind, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return apperrors.ErrNotFound("archive not found")
		}
		return apperrors.ErrInternal("failed to open archive: " + err.Error())
	}
	c.Set(fiber.HeaderContentType, "application/gzip")
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+name+`"`)
	// fasthttp закрывает io.ReadCloser после отправки тела.
	return c.SendStream(rc, int(size))
}

type deleteBackupRequest struct {
	Storage string `json:"storage"`
	Name    string `json:"name"`
}

// Delete removes an archive from storage.
func (h *BackupsHandler) Delete(c fiber.Ctx) error {
	var req deleteBackupRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	if !backup.ValidArchiveName(req.Name) {
		return apperrors.ErrInvalidRequest("invalid archive name")
	}
	if err := h.svc.Delete(c.Context(), req.Storage, req.Name); err != nil {
		return apperrors.ErrInternal("failed to delete archive: " + err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}
