// Package storage provides pluggable blob storage backends (local disk and
// S3-compatible services such as Cloudflare R2). It is used for backups and
// is designed to be reused by other admin features.
package storage

import (
	"context"
	"io"
	"time"
)

// ObjectInfo describes a stored object.
type ObjectInfo struct {
	Key     string
	Size    int64
	ModTime time.Time
}

// Storage is a minimal blob store abstraction.
type Storage interface {
	// Put stores r under key. size is the exact content length in bytes.
	Put(ctx context.Context, key string, r io.Reader, size int64) error
	// Get opens an object for reading; the caller must close the reader.
	// Returns fs.ErrNotExist (wrapped) when the object is missing.
	Get(ctx context.Context, key string) (io.ReadCloser, int64, error)
	// List returns objects whose keys start with prefix.
	List(ctx context.Context, prefix string) ([]ObjectInfo, error)
	// Delete removes an object. Deleting a missing object is not an error.
	Delete(ctx context.Context, key string) error
}
