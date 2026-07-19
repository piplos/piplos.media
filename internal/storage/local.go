package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/piplos/piplos.media/internal/utils"
)

// Local stores objects as files under a root directory.
type Local struct {
	root string
}

// NewLocal creates a Local storage rooted at dir (created if missing).
func NewLocal(dir string) (*Local, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve storage dir: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}
	return &Local{root: abs}, nil
}

// resolve validates key and maps it to an absolute path inside root.
func (l *Local) resolve(key string) (string, error) {
	abs, err := utils.SafeJoin(l.root, key)
	if err != nil {
		return "", fmt.Errorf("invalid storage key %q", key)
	}
	return abs, nil
}

// Put writes r to a temp file and renames it into place (atomic on one fs).
func (l *Local) Put(_ context.Context, key string, r io.Reader, _ int64) error {
	abs, err := l.resolve(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(abs), ".tmp-put-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		return fmt.Errorf("write object: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close object: %w", err)
	}
	if err := os.Rename(tmp.Name(), abs); err != nil {
		return fmt.Errorf("finalize object: %w", err)
	}
	return nil
}

// Get opens an object for reading.
func (l *Local) Get(_ context.Context, key string) (io.ReadCloser, int64, error) {
	abs, err := l.resolve(key)
	if err != nil {
		return nil, 0, err
	}
	f, err := os.Open(abs)
	if err != nil {
		return nil, 0, err // preserves fs.ErrNotExist
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, err
	}
	return f, st.Size(), nil
}

// List walks the root and returns objects with slash-separated keys.
func (l *Local) List(_ context.Context, prefix string) ([]ObjectInfo, error) {
	out := []ObjectInfo{}
	err := filepath.WalkDir(l.root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(l.root, p)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(rel)
		if !strings.HasPrefix(key, prefix) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		out = append(out, ObjectInfo{Key: key, Size: info.Size(), ModTime: info.ModTime()})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}
	return out, nil
}

// Delete removes an object; missing objects are ignored.
func (l *Local) Delete(_ context.Context, key string) error {
	abs, err := l.resolve(key)
	if err != nil {
		return err
	}
	if err := os.Remove(abs); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}
