package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/piplos/piplos.media/internal/models"
)

const apiKeyColumns = "id, name, key_hash, key_prefix, created_by, last_used_at, revoked_at, created_at, updated_at"

func scanAPIKey(row pgx.Row) (*models.APIKey, error) {
	var k models.APIKey
	err := row.Scan(&k.ID, &k.Name, &k.KeyHash, &k.KeyPrefix, &k.CreatedBy,
		&k.LastUsedAt, &k.RevokedAt, &k.CreatedAt, &k.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan api key: %w", err)
	}
	return &k, nil
}

// CreateAPIKey inserts a key (hash only — the raw key never touches the DB).
func (r *Repository) CreateAPIKey(ctx context.Context, name, keyHash, keyPrefix string, createdBy *string) (*models.APIKey, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO api_keys (name, key_hash, key_prefix, created_by)
		 VALUES ($1, $2, $3, $4) RETURNING `+apiKeyColumns,
		name, keyHash, keyPrefix, createdBy)
	return scanAPIKey(row)
}

// ListAPIKeys returns all keys, newest first.
func (r *Repository) ListAPIKeys(ctx context.Context) ([]models.APIKey, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT "+apiKeyColumns+" FROM api_keys ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	items := []models.APIKey{}
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *k)
	}
	return items, rows.Err()
}

// GetAPIKeyByHash returns a key by SHA-256 hex digest or nil.
func (r *Repository) GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+apiKeyColumns+" FROM api_keys WHERE key_hash = $1", keyHash)
	return scanAPIKey(row)
}

// RevokeAPIKey sets revoked_at = now() if the key is still active and returns
// it; nil means the id does not exist or the key was already revoked.
func (r *Repository) RevokeAPIKey(ctx context.Context, id string) (*models.APIKey, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE api_keys SET revoked_at = now(), updated_at = now()
		 WHERE id = $1 AND revoked_at IS NULL RETURNING `+apiKeyColumns, id)
	return scanAPIKey(row)
}

// DeleteAPIKey removes a key row by id.
func (r *Repository) DeleteAPIKey(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM api_keys WHERE id = $1", id)
	return err
}

// TouchAPIKeyLastUsed records key usage, throttled to at most one write per
// minute per key by the WHERE clause. Callers ignore the error: a failed
// touch must never fail a request.
func (r *Repository) TouchAPIKeyLastUsed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE api_keys SET last_used_at = now()
		 WHERE id = $1 AND (last_used_at IS NULL OR last_used_at < now() - INTERVAL '1 minute')`, id)
	return err
}
