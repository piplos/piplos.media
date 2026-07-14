package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/piplos/piplos.media/internal/models"
)

const serviceColumns = "id, slug, icon, tags, published, sort_order, translations, created_at, updated_at"

func scanService(row pgx.Row) (*models.Service, error) {
	var s models.Service
	var raw []byte
	err := row.Scan(&s.ID, &s.Slug, &s.Icon, &s.Tags, &s.Published, &s.SortOrder, &raw, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan service: %w", err)
	}
	s.Translations = translationsFromJSON(raw)
	return &s, nil
}

// ListServices returns all services.
func (r *Repository) ListServices(ctx context.Context) ([]models.Service, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+serviceColumns+" FROM services ORDER BY sort_order, created_at")
	if err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()

	items := []models.Service{}
	for rows.Next() {
		s, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	return items, rows.Err()
}

// GetService returns a service by UUID or slug, or nil.
func (r *Repository) GetService(ctx context.Context, ref string) (*models.Service, error) {
	row := r.pool.QueryRow(ctx,
		"SELECT "+serviceColumns+" FROM services WHERE id::text = $1 OR slug = $1", ref)
	return scanService(row)
}

// CreateService inserts a service.
func (r *Repository) CreateService(ctx context.Context, s *models.Service) (*models.Service, error) {
	tr, err := s.Translations.JSON()
	if err != nil {
		return nil, fmt.Errorf("marshal translations: %w", err)
	}
	row := r.pool.QueryRow(ctx,
		`INSERT INTO services (slug, icon, tags, published, sort_order, translations)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+serviceColumns,
		s.Slug, s.Icon, s.Tags, s.Published, s.SortOrder, tr)
	return scanService(row)
}

// UpdateService updates a service.
func (r *Repository) UpdateService(ctx context.Context, s *models.Service) (*models.Service, error) {
	tr, err := s.Translations.JSON()
	if err != nil {
		return nil, fmt.Errorf("marshal translations: %w", err)
	}
	row := r.pool.QueryRow(ctx,
		`UPDATE services SET slug = $2, icon = $3, tags = $4, published = $5, sort_order = $6,
			translations = $7, updated_at = now()
		 WHERE id = $1 RETURNING `+serviceColumns,
		s.ID, s.Slug, s.Icon, s.Tags, s.Published, s.SortOrder, tr)
	return scanService(row)
}

// DeleteService removes a service.
func (r *Repository) DeleteService(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM services WHERE id = $1", id)
	return err
}

// ReorderServices sets sort_order from the position of each id in the slice.
func (r *Repository) ReorderServices(ctx context.Context, ids []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin reorder services tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for i, id := range ids {
		if _, err := tx.Exec(ctx,
			"UPDATE services SET sort_order = $2, updated_at = now() WHERE id = $1",
			id, i,
		); err != nil {
			return fmt.Errorf("reorder service %s: %w", id, err)
		}
	}

	return tx.Commit(ctx)
}
