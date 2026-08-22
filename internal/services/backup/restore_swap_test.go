package backup

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Сбой подмены uploadDir не должен уничтожать распакованные файлы: скрытый
// .restore-* каталог остаётся для ручного разбора (следующий успешный restore
// приберёт его сам).
func TestRestoreArchiveKeepsRestoredFilesOnFailedSwap(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission-based failure simulation requires non-root")
	}
	ctx := context.Background()
	dst := t.TempDir()

	// Неудаляемая запись ломает очистку uploadDir -> swap падает уже после
	// успешной распаковки архива.
	sticky := filepath.Join(dst, "sticky")
	if err := os.MkdirAll(sticky, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sticky, "x"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(sticky, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(sticky, 0o755) })

	archive := gzipTar(t, [][2]string{
		{"manifest.json", `{"version":1,"type":"files","created_at":"2026-07-19T10:00:00Z","file_count":1}`},
		{"files/a.txt", "alpha"},
	})
	if _, err := restoreArchive(ctx, bytes.NewReader(archive), nil, dst); err == nil {
		t.Fatal("expected restore to fail on undeletable entry")
	}

	var kept string
	entries, err := os.ReadDir(dst)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".restore-") {
			kept = e.Name()
		}
	}
	if kept == "" {
		t.Fatal(".restore-* dir with restored files must be kept on failed swap")
	}
	data, err := os.ReadFile(filepath.Join(dst, kept, "a.txt"))
	if err != nil || string(data) != "alpha" {
		t.Fatalf("restored file must survive the failed swap: %v", err)
	}
}
