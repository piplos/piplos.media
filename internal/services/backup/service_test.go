package backup

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/storage"
)

// fakeSettings implements settingsReader for tests.
type fakeSettings map[string]string

func (f fakeSettings) GetDecryptedValue(_ context.Context, key string) (string, error) {
	return f[key], nil
}

func TestValidateParams(t *testing.T) {
	valid := [][2]string{
		{TypeFull, StorageLocal}, {TypeDB, StorageS3}, {TypeFiles, StorageLocal},
	}
	for _, c := range valid {
		if err := validateParams(c[0], c[1]); err != nil {
			t.Errorf("validateParams(%q, %q): unexpected error %v", c[0], c[1], err)
		}
	}
	invalid := [][2]string{
		{"", StorageLocal}, {"weird", StorageLocal}, {TypeFull, ""}, {TypeFull, "ftp"},
	}
	for _, c := range invalid {
		if err := validateParams(c[0], c[1]); err == nil {
			t.Errorf("validateParams(%q, %q): expected error", c[0], c[1])
		}
	}
}

// putArchive stores a valid archive with the given type and timestamp offset.
func putArchive(t *testing.T, st storage.Storage, backupType string, createdAt time.Time) string {
	t.Helper()
	name := NewArchiveName(backupType, createdAt)
	if err := st.Put(context.Background(), name, bytes.NewReader([]byte("x")), 1); err != nil {
		t.Fatalf("Put %s: %v", name, err)
	}
	return name
}

func TestListArchivesFiltersAndSorts(t *testing.T) {
	ctx := context.Background()
	st, err := storage.NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}

	base := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	old := putArchive(t, st, TypeDB, base)
	fresh := putArchive(t, st, TypeFull, base.Add(time.Hour))
	// Посторонние объекты не должны попадать в список.
	if err := st.Put(ctx, "notes.txt", bytes.NewReader([]byte("x")), 1); err != nil {
		t.Fatalf("Put: %v", err)
	}

	archives, err := listArchives(ctx, st, "")
	if err != nil {
		t.Fatalf("listArchives: %v", err)
	}
	if len(archives) != 2 {
		t.Fatalf("expected 2 archives, got %+v", archives)
	}
	if archives[0].Name != fresh || archives[1].Name != old {
		t.Errorf("expected newest first, got %s, %s", archives[0].Name, archives[1].Name)
	}
	if archives[0].Type != TypeFull || archives[1].Type != TypeDB {
		t.Errorf("wrong types: %+v", archives)
	}
}

func TestApplyRetention(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	st, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}

	settings := fakeSettings{config.KeyBackup: `{"enabled":false,"type":"full","interval_hours":24,"keep":2,"storage":"local"}`}
	svc := New(nil, settings, "", dir, zerolog.Nop())

	base := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		putArchive(t, st, TypeFull, base.Add(time.Duration(i)*time.Hour))
	}

	if err := svc.applyRetention(ctx, st, ""); err != nil {
		t.Fatalf("applyRetention: %v", err)
	}
	archives, err := listArchives(ctx, st, "")
	if err != nil {
		t.Fatalf("listArchives: %v", err)
	}
	if len(archives) != 2 {
		t.Fatalf("expected 2 archives after retention, got %d", len(archives))
	}
	want := []string{
		NewArchiveName(TypeFull, base.Add(4*time.Hour)),
		NewArchiveName(TypeFull, base.Add(3*time.Hour)),
	}
	for i, name := range want {
		if archives[i].Name != name {
			t.Errorf("archive %d: got %s, want %s", i, archives[i].Name, name)
		}
	}
}

func TestApplyRetentionKeepZeroIsUnlimited(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	st, err := storage.NewLocal(dir)
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}
	settings := fakeSettings{config.KeyBackup: `{"keep":0}`}
	svc := New(nil, settings, "", dir, zerolog.Nop())

	base := time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		putArchive(t, st, TypeFull, base.Add(time.Duration(i)*time.Hour))
	}
	if err := svc.applyRetention(ctx, st, ""); err != nil {
		t.Fatalf("applyRetention: %v", err)
	}
	archives, err := listArchives(ctx, st, "")
	if err != nil {
		t.Fatalf("listArchives: %v", err)
	}
	if len(archives) != 3 {
		t.Fatalf("keep=0 must retain all archives, got %d", len(archives))
	}
}

func TestStatusLifecycle(t *testing.T) {
	svc := New(nil, fakeSettings{}, "", t.TempDir(), zerolog.Nop())

	if st := svc.Status(); st.Running || st.Last != nil {
		t.Fatalf("fresh service must be idle: %+v", st)
	}

	started, err := svc.begin("backup")
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if _, err := svc.begin("backup"); err != ErrBusy {
		t.Fatalf("second begin must return ErrBusy, got %v", err)
	}
	if st := svc.Status(); !st.Running || st.Op != "backup" {
		t.Fatalf("expected running backup, got %+v", st)
	}

	res := svc.complete(Result{Op: "backup", StartedAt: started}, fmt.Errorf("boom"))
	if res.OK || res.Error != "boom" || res.FinishedAt.IsZero() {
		t.Fatalf("complete result: %+v", res)
	}
	st := svc.Status()
	if st.Running || st.Last == nil || st.Last.Error != "boom" {
		t.Fatalf("status after complete: %+v", st)
	}
}
