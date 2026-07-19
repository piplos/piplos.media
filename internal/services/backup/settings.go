// Package backup implements site backups: PostgreSQL dump via pgx (no external
// tools), uploaded files archive, retention, scheduling and restore. Archives
// are stored locally or in an S3-compatible bucket (Cloudflare R2).
package backup

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Backup types.
const (
	TypeFull  = "full"
	TypeDB    = "db"
	TypeFiles = "files"
)

// Storage kinds.
const (
	StorageLocal = "local"
	StorageS3    = "s3"
)

// Settings is the BACKUP composite setting (stored in the settings table).
type Settings struct {
	// Enabled turns automatic scheduled backups on.
	Enabled bool `json:"enabled"`
	// Type of automatic backups: full | db | files.
	Type string `json:"type"`
	// IntervalHours between automatic backups (1..720).
	IntervalHours int `json:"interval_hours"`
	// Keep is how many newest archives to retain (0 = unlimited).
	Keep int `json:"keep"`
	// Storage backend: local | s3.
	Storage string `json:"storage"`
}

// DefaultSettings returns the default backup configuration.
func DefaultSettings() Settings {
	return Settings{Enabled: false, Type: TypeFull, IntervalHours: 24, Keep: 7, Storage: StorageLocal}
}

// ParseSettings decodes raw JSON ("" -> defaults) and normalizes values.
func ParseSettings(raw string) (Settings, error) {
	s := DefaultSettings()
	if strings.TrimSpace(raw) == "" {
		return s, nil
	}
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		return DefaultSettings(), fmt.Errorf("parse backup settings: %w", err)
	}
	if s.Type != TypeFull && s.Type != TypeDB && s.Type != TypeFiles {
		s.Type = TypeFull
	}
	if s.Storage != StorageLocal && s.Storage != StorageS3 {
		s.Storage = StorageLocal
	}
	if s.IntervalHours < 1 {
		s.IntervalHours = 24
	}
	if s.IntervalHours > 720 {
		s.IntervalHours = 720
	}
	if s.Keep < 0 {
		s.Keep = 0
	}
	if s.Keep > 100 {
		s.Keep = 100
	}
	return s, nil
}

// Interval returns the schedule interval as a duration.
func (s Settings) Interval() time.Duration {
	return time.Duration(s.IntervalHours) * time.Hour
}

// archiveNameRe validates backup archive names (also guards against path tricks).
var archiveNameRe = regexp.MustCompile(`^backup-(full|db|files)-\d{8}-\d{6}\.tar\.gz$`)

// ValidArchiveName reports whether name is a well-formed backup archive name.
func ValidArchiveName(name string) bool {
	return archiveNameRe.MatchString(name)
}

// TypeFromName extracts the backup type from an archive name ("" if invalid).
func TypeFromName(name string) string {
	m := archiveNameRe.FindStringSubmatch(name)
	if m == nil {
		return ""
	}
	return m[1]
}

// NewArchiveName builds an archive name for a backup started at t (UTC).
func NewArchiveName(backupType string, t time.Time) string {
	return fmt.Sprintf("backup-%s-%s.tar.gz", backupType, t.UTC().Format("20060102-150405"))
}
