package backup

import (
	"testing"
	"time"
)

func TestParseSettingsDefaults(t *testing.T) {
	s, err := ParseSettings("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != DefaultSettings() {
		t.Fatalf("expected defaults, got %+v", s)
	}
}

func TestParseSettingsNormalization(t *testing.T) {
	raw := `{"enabled":true,"type":"weird","interval_hours":0,"keep":-5,"storage":"ftp"}`
	s, err := ParseSettings(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Type != TypeFull {
		t.Errorf("type: got %q, want %q", s.Type, TypeFull)
	}
	if s.Storage != StorageLocal {
		t.Errorf("storage: got %q, want %q", s.Storage, StorageLocal)
	}
	if s.IntervalHours != 24 {
		t.Errorf("interval: got %d, want 24", s.IntervalHours)
	}
	if s.Keep != 0 {
		t.Errorf("keep: got %d, want 0", s.Keep)
	}
	if !s.Enabled {
		t.Error("enabled flag lost")
	}
}

func TestParseSettingsClampsUpperBounds(t *testing.T) {
	s, err := ParseSettings(`{"interval_hours":9000,"keep":500}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.IntervalHours != 720 {
		t.Errorf("interval: got %d, want 720", s.IntervalHours)
	}
	if s.Keep != 100 {
		t.Errorf("keep: got %d, want 100", s.Keep)
	}
}

func TestParseSettingsInvalidJSON(t *testing.T) {
	if _, err := ParseSettings("{not json"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestArchiveNames(t *testing.T) {
	ts := time.Date(2026, 7, 19, 10, 30, 0, 0, time.UTC)
	name := NewArchiveName(TypeFull, ts)
	if name != "backup-full-20260719-103000.tar.gz" {
		t.Fatalf("unexpected archive name: %s", name)
	}
	if !ValidArchiveName(name) {
		t.Fatalf("generated name must be valid: %s", name)
	}
	if got := TypeFromName(name); got != TypeFull {
		t.Fatalf("type from name: got %q", got)
	}

	invalid := []string{
		"", "backup-full-20260719-103000.tar", "../../etc/passwd",
		"backup-x-20260719-103000.tar.gz", "backup-db-2026-103000.tar.gz",
		"prefix-backup-db-20260719-103000.tar.gz",
	}
	for _, n := range invalid {
		if ValidArchiveName(n) {
			t.Errorf("name %q must be invalid", n)
		}
	}
}
