package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/storage"
)

// s3Prefix namespaces backup objects inside a bucket shared with other features.
const s3Prefix = "backups/"

// ErrBusy is returned when a backup or restore is already running.
var ErrBusy = errors.New("another backup operation is already running")

// settingsReader provides decrypted composite settings (implemented by repository).
type settingsReader interface {
	GetDecryptedValue(ctx context.Context, key string) (string, error)
}

// Archive is a stored backup archive.
type Archive struct {
	Name    string    `json:"name"`
	Type    string    `json:"type"`
	Storage string    `json:"storage"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

// Result describes a finished backup/restore operation.
type Result struct {
	Op         string    `json:"op"` // backup | restore
	Type       string    `json:"type,omitempty"`
	Archive    string    `json:"archive,omitempty"`
	Storage    string    `json:"storage,omitempty"`
	OK         bool      `json:"ok"`
	Error      string    `json:"error,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	SizeBytes  int64     `json:"size_bytes,omitempty"`
}

// Status is the current state of the backup subsystem.
type Status struct {
	Running   bool      `json:"running"`
	Op        string    `json:"op,omitempty"`
	StartedAt time.Time `json:"started_at,omitzero"`
	Last      *Result   `json:"last,omitempty"`
}

// Service runs backups and restores. One operation at a time.
type Service struct {
	pool      *pgxpool.Pool
	settings  settingsReader
	uploadDir string
	localDir  string
	log       zerolog.Logger

	mu      sync.Mutex
	running bool
	op      string
	started time.Time
	last    *Result
}

// New creates a backup Service. localDir is the local archive directory.
func New(pool *pgxpool.Pool, settings settingsReader, uploadDir, localDir string, log zerolog.Logger) *Service {
	return &Service{
		pool: pool, settings: settings,
		uploadDir: uploadDir, localDir: localDir,
		log: log.With().Str("component", "backup").Logger(),
	}
}

// Settings loads current backup settings (defaults when unset).
func (s *Service) Settings(ctx context.Context) (Settings, error) {
	raw, err := s.settings.GetDecryptedValue(ctx, config.KeyBackup)
	if err != nil {
		return DefaultSettings(), fmt.Errorf("load backup settings: %w", err)
	}
	return ParseSettings(raw)
}

// s3Config loads the shared S3 setting ("" fields when unset).
func (s *Service) s3Config(ctx context.Context) (storage.S3Config, error) {
	var cfg storage.S3Config
	raw, err := s.settings.GetDecryptedValue(ctx, config.KeyS3)
	if err != nil {
		return cfg, fmt.Errorf("load S3 settings: %w", err)
	}
	if strings.TrimSpace(raw) == "" {
		return cfg, nil
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return cfg, fmt.Errorf("parse S3 settings: %w", err)
	}
	return cfg, nil
}

// storageFor returns the storage backend and key prefix for a storage kind.
func (s *Service) storageFor(ctx context.Context, kind string) (storage.Storage, string, error) {
	switch kind {
	case StorageLocal:
		st, err := storage.NewLocal(s.localDir)
		return st, "", err
	case StorageS3:
		cfg, err := s.s3Config(ctx)
		if err != nil {
			return nil, "", err
		}
		if err := cfg.Validate(); err != nil {
			return nil, "", fmt.Errorf("S3 storage is not configured: %w", err)
		}
		st, err := storage.NewS3(cfg)
		return st, s3Prefix, err
	default:
		return nil, "", fmt.Errorf("unknown storage kind %q", kind)
	}
}

// Status returns the current operation state.
func (s *Service) Status() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := Status{Running: s.running, Op: s.op, StartedAt: s.started}
	if s.last != nil {
		last := *s.last
		st.Last = &last
	}
	return st
}

// begin marks an operation as running (fails when busy).
func (s *Service) begin(op string) (time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return time.Time{}, ErrBusy
	}
	s.running = true
	s.op = op
	s.started = time.Now()
	return s.started, nil
}

// finish records the operation result.
func (s *Service) finish(res Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.running = false
	s.op = ""
	s.last = &res
	if res.OK {
		s.log.Info().Str("op", res.Op).Str("archive", res.Archive).
			Int64("size", res.SizeBytes).Dur("took", res.FinishedAt.Sub(res.StartedAt)).
			Msg("operation finished")
	} else {
		s.log.Error().Str("op", res.Op).Str("archive", res.Archive).
			Str("error", res.Error).Msg("operation failed")
	}
}

// complete finalizes res with err and records it as the last result.
func (s *Service) complete(res Result, err error) Result {
	res.FinishedAt = time.Now()
	res.OK = err == nil
	if err != nil {
		res.Error = err.Error()
	}
	s.finish(res)
	return res
}

// opTimeout limits a single background backup/restore operation.
const opTimeout = 30 * time.Minute

// validateParams checks backup type and storage kind values.
func validateParams(backupType, storageKind string) error {
	if backupType != TypeFull && backupType != TypeDB && backupType != TypeFiles {
		return fmt.Errorf("unknown backup type %q", backupType)
	}
	if storageKind != StorageLocal && storageKind != StorageS3 {
		return fmt.Errorf("unknown storage kind %q", storageKind)
	}
	return nil
}

// Run performs a backup of backupType to storageKind synchronously.
// Returns ErrBusy when another operation is in progress.
func (s *Service) Run(ctx context.Context, backupType, storageKind string) (Result, error) {
	if err := validateParams(backupType, storageKind); err != nil {
		return Result{}, err
	}
	started, err := s.begin("backup")
	if err != nil {
		return Result{}, err
	}
	return s.executeBackup(ctx, backupType, storageKind, started)
}

// StartBackup launches a backup in the background. The operation slot is
// acquired synchronously, so ErrBusy is reported to the caller immediately.
func (s *Service) StartBackup(backupType, storageKind string) error {
	if err := validateParams(backupType, storageKind); err != nil {
		return err
	}
	started, err := s.begin("backup")
	if err != nil {
		return err
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
		defer cancel()
		s.executeBackup(ctx, backupType, storageKind, started)
	}()
	return nil
}

// executeBackup runs one backup operation and records its result.
// The caller must have acquired the operation slot via begin.
func (s *Service) executeBackup(ctx context.Context, backupType, storageKind string, started time.Time) (Result, error) {
	res := Result{Op: "backup", Type: backupType, Storage: storageKind, StartedAt: started}
	err := s.runBackup(ctx, &res)
	return s.complete(res, err), err
}

func (s *Service) runBackup(ctx context.Context, res *Result) error {
	st, prefix, err := s.storageFor(ctx, res.Storage)
	if err != nil {
		return err
	}

	name := NewArchiveName(res.Type, res.StartedAt)
	res.Archive = name

	// Архив собирается во временный файл: для S3 и tar нужен известный размер.
	tmp, err := os.CreateTemp("", "piplos-archive-*.tar.gz")
	if err != nil {
		return fmt.Errorf("create temp archive: %w", err)
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	if _, err := writeArchive(ctx, tmp, s.pool, s.uploadDir, res.Type, res.StartedAt); err != nil {
		return err
	}
	size, err := tmp.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("size archive: %w", err)
	}
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind archive: %w", err)
	}
	res.SizeBytes = size

	if err := st.Put(ctx, prefix+name, tmp, size); err != nil {
		return err
	}

	if err := s.applyRetention(ctx, st, prefix); err != nil {
		// Бекап уже сохранён; ошибка ретеншна не должна ронять операцию.
		s.log.Warn().Err(err).Msg("retention cleanup failed")
	}
	return nil
}

// applyRetention keeps only the newest N archives in st (per settings).
func (s *Service) applyRetention(ctx context.Context, st storage.Storage, prefix string) error {
	cfg, err := s.Settings(ctx)
	if err != nil {
		return err
	}
	if cfg.Keep <= 0 {
		return nil
	}
	archives, err := listArchives(ctx, st, prefix)
	if err != nil {
		return err
	}
	if len(archives) <= cfg.Keep {
		return nil
	}
	// listArchives сортирует новые -> старые; удаляем хвост.
	for _, old := range archives[cfg.Keep:] {
		if err := st.Delete(ctx, prefix+old.Name); err != nil {
			return err
		}
		s.log.Info().Str("archive", old.Name).Msg("old backup removed by retention")
	}
	return nil
}

// listArchives lists valid backup archives, newest first.
func listArchives(ctx context.Context, st storage.Storage, prefix string) ([]Archive, error) {
	objects, err := st.List(ctx, prefix)
	if err != nil {
		return nil, err
	}
	out := []Archive{}
	for _, obj := range objects {
		name := strings.TrimPrefix(obj.Key, prefix)
		if !ValidArchiveName(name) {
			continue
		}
		out = append(out, Archive{
			Name: name, Type: TypeFromName(name),
			Size: obj.Size, ModTime: obj.ModTime,
		})
	}
	// Имя содержит UTC-метку, поэтому сортировка по имени = по времени создания.
	sort.Slice(out, func(i, j int) bool { return out[i].Name > out[j].Name })
	return out, nil
}

// List returns archives from local storage and, when configured, from S3.
func (s *Service) List(ctx context.Context) ([]Archive, error) {
	local, prefix, err := s.storageFor(ctx, StorageLocal)
	if err != nil {
		return nil, err
	}
	archives, err := listArchives(ctx, local, prefix)
	if err != nil {
		return nil, err
	}
	for i := range archives {
		archives[i].Storage = StorageLocal
	}

	// S3 не настроен или недоступен — показываем только локальные архивы.
	if s3st, s3prefix, err := s.storageFor(ctx, StorageS3); err == nil {
		s3archives, err := listArchives(ctx, s3st, s3prefix)
		if err != nil {
			s.log.Warn().Err(err).Msg("list S3 backups failed")
		} else {
			for i := range s3archives {
				s3archives[i].Storage = StorageS3
			}
			archives = append(archives, s3archives...)
		}
	}

	sort.Slice(archives, func(i, j int) bool {
		if archives[i].Name != archives[j].Name {
			return archives[i].Name > archives[j].Name
		}
		return archives[i].Storage < archives[j].Storage
	})
	return archives, nil
}

// Open returns a reader with the archive content (caller closes it).
func (s *Service) Open(ctx context.Context, storageKind, name string) (io.ReadCloser, int64, error) {
	if !ValidArchiveName(name) {
		return nil, 0, fmt.Errorf("invalid archive name")
	}
	st, prefix, err := s.storageFor(ctx, storageKind)
	if err != nil {
		return nil, 0, err
	}
	return st.Get(ctx, prefix+name)
}

// Delete removes an archive.
func (s *Service) Delete(ctx context.Context, storageKind, name string) error {
	if !ValidArchiveName(name) {
		return fmt.Errorf("invalid archive name")
	}
	st, prefix, err := s.storageFor(ctx, storageKind)
	if err != nil {
		return err
	}
	return st.Delete(ctx, prefix+name)
}

// StartRestore launches a restore in the background. The operation slot is
// acquired synchronously, so ErrBusy is reported to the caller immediately.
func (s *Service) StartRestore(storageKind, name string) error {
	if !ValidArchiveName(name) {
		return fmt.Errorf("invalid archive name")
	}
	started, err := s.begin("restore")
	if err != nil {
		return err
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
		defer cancel()
		res := Result{Op: "restore", Archive: name, Storage: storageKind, StartedAt: started}
		err := s.runRestore(ctx, &res)
		s.complete(res, err)
	}()
	return nil
}

func (s *Service) runRestore(ctx context.Context, res *Result) error {
	rc, size, err := s.Open(ctx, res.Storage, res.Archive)
	if err != nil {
		return err
	}
	defer rc.Close()
	res.SizeBytes = size

	manifest, err := restoreArchive(ctx, rc, s.pool, s.uploadDir)
	if err != nil {
		return err
	}
	res.Type = manifest.Type
	return nil
}

// StartScheduler launches automatic backups in a background goroutine.
// The interval is measured from the newest archive in the configured storage,
// so the schedule survives restarts without extra state.
func (s *Service) StartScheduler(ctx context.Context) {
	const tick = 5 * time.Minute
	go func() {
		timer := time.NewTicker(tick)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				s.schedulerTick(ctx)
			}
		}
	}()
}

func (s *Service) schedulerTick(ctx context.Context) {
	cfg, err := s.Settings(ctx)
	if err != nil {
		s.log.Warn().Err(err).Msg("scheduler: load settings failed")
		return
	}
	if !cfg.Enabled {
		return
	}
	st, prefix, err := s.storageFor(ctx, cfg.Storage)
	if err != nil {
		s.log.Warn().Err(err).Msg("scheduler: storage unavailable")
		return
	}
	archives, err := listArchives(ctx, st, prefix)
	if err != nil {
		s.log.Warn().Err(err).Msg("scheduler: list archives failed")
		return
	}
	if len(archives) > 0 && time.Since(archives[0].ModTime) < cfg.Interval() {
		return
	}
	// Как и у ручных операций, время выполнения ограничено opTimeout —
	// зависший бекап не должен навсегда занять слот running.
	runCtx, cancel := context.WithTimeout(ctx, opTimeout)
	defer cancel()
	if _, err := s.Run(runCtx, cfg.Type, cfg.Storage); err != nil && !errors.Is(err, ErrBusy) {
		s.log.Error().Err(err).Msg("scheduled backup failed")
	}
}
