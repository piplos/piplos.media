package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/piplos/piplos.media/internal/models"
)

const seoColumns = "id, path, translations, created_at, updated_at"

func scanSEOPage(row pgx.Row) (*models.SEOPage, error) {
	var p models.SEOPage
	var raw []byte
	err := row.Scan(&p.ID, &p.Path, &raw, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan seo page: %w", err)
	}
	p.Translations = translationsFromJSON(raw)
	return &p, nil
}

// ListSEOPages returns all SEO page entries.
func (r *Repository) ListSEOPages(ctx context.Context) ([]models.SEOPage, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+seoColumns+" FROM seo_pages ORDER BY path")
	if err != nil {
		return nil, fmt.Errorf("list seo pages: %w", err)
	}
	defer rows.Close()

	items := []models.SEOPage{}
	for rows.Next() {
		p, err := scanSEOPage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *p)
	}
	return items, rows.Err()
}

// GetSEOPage returns a SEO entry by id or nil.
func (r *Repository) GetSEOPage(ctx context.Context, id string) (*models.SEOPage, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+seoColumns+" FROM seo_pages WHERE id = $1", id)
	return scanSEOPage(row)
}

// CreateSEOPage inserts a SEO entry.
func (r *Repository) CreateSEOPage(ctx context.Context, p *models.SEOPage) (*models.SEOPage, error) {
	tr, err := p.Translations.JSON()
	if err != nil {
		return nil, fmt.Errorf("marshal translations: %w", err)
	}
	row := r.pool.QueryRow(ctx,
		"INSERT INTO seo_pages (path, translations) VALUES ($1, $2) RETURNING "+seoColumns,
		p.Path, tr)
	return scanSEOPage(row)
}

// UpdateSEOPage updates a SEO entry.
func (r *Repository) UpdateSEOPage(ctx context.Context, p *models.SEOPage) (*models.SEOPage, error) {
	tr, err := p.Translations.JSON()
	if err != nil {
		return nil, fmt.Errorf("marshal translations: %w", err)
	}
	row := r.pool.QueryRow(ctx,
		"UPDATE seo_pages SET path = $2, translations = $3, updated_at = now() WHERE id = $1 RETURNING "+seoColumns,
		p.ID, p.Path, tr)
	return scanSEOPage(row)
}

// DeleteSEOPage removes a SEO entry.
func (r *Repository) DeleteSEOPage(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM seo_pages WHERE id = $1", id)
	return err
}
