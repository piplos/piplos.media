package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeUploads fills dir with a small fixture tree.
func writeUploads(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		abs := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
}

// TestFilesBackupRestoreRoundtrip covers writeArchive -> restoreArchive for a
// files-only backup: manifest contents, extraction and upload dir swap.
func TestFilesBackupRestoreRoundtrip(t *testing.T) {
	ctx := context.Background()
	src := t.TempDir()
	writeUploads(t, src, map[string]string{
		"a.txt":     "alpha",
		"img/b.png": "beta",
	})

	var buf bytes.Buffer
	manifest, err := writeArchive(ctx, &buf, nil, src, TypeFiles, time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("writeArchive: %v", err)
	}
	if manifest.Type != TypeFiles || manifest.FileCount != 2 || len(manifest.Tables) != 0 {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}

	dst := t.TempDir()
	writeUploads(t, dst, map[string]string{"stale.txt": "must disappear"})

	restored, err := restoreArchive(ctx, bytes.NewReader(buf.Bytes()), nil, dst)
	if err != nil {
		t.Fatalf("restoreArchive: %v", err)
	}
	if restored.Type != TypeFiles {
		t.Fatalf("restored manifest type: %q", restored.Type)
	}

	for rel, want := range map[string]string{"a.txt": "alpha", "img/b.png": "beta"} {
		got, err := os.ReadFile(filepath.Join(dst, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read restored %s: %v", rel, err)
		}
		if string(got) != want {
			t.Errorf("restored %s = %q, want %q", rel, got, want)
		}
	}
	if _, err := os.Stat(filepath.Join(dst, "stale.txt")); !os.IsNotExist(err) {
		t.Error("stale file must be removed by restore")
	}
	// Временный каталог восстановления не должен оставаться в uploads.
	entries, err := os.ReadDir(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".restore-") {
			t.Errorf("restore temp dir left behind: %s", e.Name())
		}
	}
}

// gzipTar builds a tar.gz archive from ordered name -> content pairs.
func gzipTar(t *testing.T, entries [][2]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		hdr := &tar.Header{Name: e[0], Mode: 0o644, Size: int64(len(e[1]))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("tar header: %v", err)
		}
		if _, err := tw.Write([]byte(e[1])); err != nil {
			t.Fatalf("tar write: %v", err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return buf.Bytes()
}

func TestRestoreRejectsMalformedArchives(t *testing.T) {
	ctx := context.Background()

	cases := map[string][][2]string{
		"manifest not first":  {{"data.txt", "x"}},
		"unsupported version": {{"manifest.json", `{"version":99,"type":"files"}`}},
		"invalid json":        {{"manifest.json", `{broken`}},
	}
	for name, entries := range cases {
		archive := gzipTar(t, entries)
		if _, err := restoreArchive(ctx, bytes.NewReader(archive), nil, t.TempDir()); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}

	if _, err := restoreArchive(ctx, bytes.NewReader([]byte("not gzip")), nil, t.TempDir()); err == nil {
		t.Error("plain bytes: expected gzip error")
	}
}

// TestRestoreNeutralizesEscapingPaths ensures "files/../x" entries cannot
// write outside the upload dir.
func TestRestoreNeutralizesEscapingPaths(t *testing.T) {
	ctx := context.Background()
	archive := gzipTar(t, [][2]string{
		{"manifest.json", `{"version":1,"type":"files","created_at":"2026-07-19T10:00:00Z","file_count":1}`},
		{"files/../../evil.txt", "x"},
	})

	parent := t.TempDir()
	dst := filepath.Join(parent, "uploads")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, err := restoreArchive(ctx, bytes.NewReader(archive), nil, dst); err != nil {
		t.Fatalf("restoreArchive: %v", err)
	}
	if _, err := os.Stat(filepath.Join(parent, "evil.txt")); !os.IsNotExist(err) {
		t.Fatal("archive entry escaped the upload dir")
	}
	if _, err := os.Stat(filepath.Join(dst, "evil.txt")); err != nil {
		t.Fatalf("neutralized entry must stay inside upload dir: %v", err)
	}
}
