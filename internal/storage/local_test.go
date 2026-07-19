package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalPutGetListDelete(t *testing.T) {
	ctx := context.Background()
	st, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}

	content := []byte("hello backup")
	if err := st.Put(ctx, "a/b.tar.gz", bytes.NewReader(content), int64(len(content))); err != nil {
		t.Fatalf("Put: %v", err)
	}

	rc, size, err := st.Get(ctx, "a/b.tar.gz")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	got, _ := io.ReadAll(rc)
	rc.Close()
	if !bytes.Equal(got, content) || size != int64(len(content)) {
		t.Fatalf("Get returned wrong content/size: %q %d", got, size)
	}

	objects, err := st.List(ctx, "a/")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(objects) != 1 || objects[0].Key != "a/b.tar.gz" {
		t.Fatalf("List: unexpected %+v", objects)
	}

	if err := st.Delete(ctx, "a/b.tar.gz"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := st.Delete(ctx, "a/b.tar.gz"); err != nil {
		t.Fatalf("Delete of missing object must succeed: %v", err)
	}
	if _, _, err := st.Get(ctx, "a/b.tar.gz"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Get of missing object: expected fs.ErrNotExist, got %v", err)
	}
}

func TestLocalKeysCannotEscapeRoot(t *testing.T) {
	ctx := context.Background()
	parent := t.TempDir()
	root := filepath.Join(parent, "store")
	st, err := NewLocal(root)
	if err != nil {
		t.Fatalf("NewLocal: %v", err)
	}

	// Пустые ключи отклоняются.
	for _, key := range []string{"", "."} {
		if err := st.Put(ctx, key, bytes.NewReader([]byte("x")), 1); err == nil {
			t.Errorf("key %q must be rejected", key)
		}
	}

	// Ключи с ".." нейтрализуются внутри корня (как в resolveUploadPath).
	for _, key := range []string{"../evil", "a/../../evil"} {
		if err := st.Put(ctx, key, bytes.NewReader([]byte("x")), 1); err != nil {
			t.Fatalf("Put %q: %v", key, err)
		}
	}
	if _, err := os.Stat(filepath.Join(parent, "evil")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("object escaped the storage root")
	}
	if _, err := os.Stat(filepath.Join(root, "evil")); err != nil {
		t.Fatalf("neutralized object must stay inside root: %v", err)
	}
}
