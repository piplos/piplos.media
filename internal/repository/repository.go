// Package repository contains all database queries.
package repository

import (
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/piplos-media/site/internal/models"
)

// Repository executes queries against PostgreSQL.
type Repository struct {
	pool          *pgxpool.Pool
	encryptionKey []byte
}

// New creates a Repository.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// SetEncryptionKey sets the key for encrypting/decrypting settings. Must be 32 bytes.
func (r *Repository) SetEncryptionKey(key []byte) {
	r.encryptionKey = key
}

// translationsFromJSON decodes a JSONB column value.
func translationsFromJSON(raw []byte) models.Translations {
	t := models.Translations{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &t)
	}
	return t
}
