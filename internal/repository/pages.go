package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/piplos/piplos.media/internal/models"
)

const pageColumns = "id, slug, published, publish_at, image, tags, translations, created_at, updated_at"

func scanPage(row pgx.Row) (*models.Page, error) {
	var p models.Page
	var raw []byte
	err := row.Scan(&p.ID, &p.Slug, &p.Published, &p.PublishAt, &p.Image, &p.Tags, &raw, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan page: %w", err)
	}
	if p.Tags == nil {
		p.Tags = []string{}
	}
	p.Translations = translationsFromJSON(raw)
	return &p, nil
}

// ListPages returns all custom pages, newest first (admin view, including drafts).
func (r *Repository) ListPages(ctx context.Context) ([]models.Page, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT "+pageColumns+" FROM pages ORDER BY COALESCE(publish_at, created_at) DESC, created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list pages: %w", err)
	}
	defer rows.Close()

	items := []models.Page{}
	for rows.Next() {
		p, err := scanPage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *p)
	}
	return items, rows.Err()
}

// GetPage returns a page by id or nil.
func (r *Repository) GetPage(ctx context.Context, id string) (*models.Page, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+pageColumns+" FROM pages WHERE id = $1", id)
	return scanPage(row)
}

// GetPageBySlug returns a page by slug or nil.
func (r *Repository) GetPageBySlug(ctx context.Context, slug string) (*models.Page, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+pageColumns+" FROM pages WHERE slug = $1", slug)
	return scanPage(row)
}

// CreatePage inserts a page.
func (r *Repository) CreatePage(ctx context.Context, p *models.Page) (*models.Page, error) {
	tr, err := p.Translations.JSON()
	if err != nil {
		return nil, fmt.Errorf("marshal page translations: %w", err)
	}
	if p.Tags == nil {
		p.Tags = []string{}
	}
	row := r.pool.QueryRow(ctx,
		`INSERT INTO pages (slug, published, publish_at, image, tags, translations)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+pageColumns,
		p.Slug, p.Published, p.PublishAt, p.Image, p.Tags, tr)
	return scanPage(row)
}

// UpdatePage updates a page.
func (r *Repository) UpdatePage(ctx context.Context, p *models.Page) (*models.Page, error) {
	tr, err := p.Translations.JSON()
	if err != nil {
		return nil, fmt.Errorf("marshal page translations: %w", err)
	}
	if p.Tags == nil {
		p.Tags = []string{}
	}
	row := r.pool.QueryRow(ctx,
		`UPDATE pages SET slug = $2, published = $3, publish_at = $4, image = $5, tags = $6, translations = $7, updated_at = now()
		 WHERE id = $1 RETURNING `+pageColumns,
		p.ID, p.Slug, p.Published, p.PublishAt, p.Image, p.Tags, tr)
	return scanPage(row)
}

// DeletePage removes a page by id.
func (r *Repository) DeletePage(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM pages WHERE id = $1", id)
	return err
}
