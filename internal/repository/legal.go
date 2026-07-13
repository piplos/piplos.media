package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/piplos/piplos.media/internal/models"
)

const legalColumns = "id, slug, path, sort_order, translations, created_at, updated_at"

func legalTranslationsFromJSON(raw []byte) models.LegalTranslations {
	t := models.LegalTranslations{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &t)
	}
	return t
}

func scanLegalPage(row pgx.Row) (*models.LegalPage, error) {
	var p models.LegalPage
	var raw []byte
	err := row.Scan(&p.ID, &p.Slug, &p.Path, &p.SortOrder, &raw, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan legal page: %w", err)
	}
	p.Translations = legalTranslationsFromJSON(raw)
	return &p, nil
}

// ListLegalPages returns all legal documents ordered by sort_order.
func (r *Repository) ListLegalPages(ctx context.Context) ([]models.LegalPage, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+legalColumns+" FROM legal_pages ORDER BY sort_order, path")
	if err != nil {
		return nil, fmt.Errorf("list legal pages: %w", err)
	}
	defer rows.Close()

	items := []models.LegalPage{}
	for rows.Next() {
		p, err := scanLegalPage(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *p)
	}
	return items, rows.Err()
}

// GetLegalPage returns a legal document by id or nil.
func (r *Repository) GetLegalPage(ctx context.Context, id string) (*models.LegalPage, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+legalColumns+" FROM legal_pages WHERE id = $1", id)
	return scanLegalPage(row)
}

// GetLegalPageBySlug returns a legal document by slug or nil.
func (r *Repository) GetLegalPageBySlug(ctx context.Context, slug string) (*models.LegalPage, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+legalColumns+" FROM legal_pages WHERE slug = $1", slug)
	return scanLegalPage(row)
}

// UpdateLegalPage updates translations for a legal document (slug/path are fixed).
func (r *Repository) UpdateLegalPage(ctx context.Context, p *models.LegalPage) (*models.LegalPage, error) {
	tr, err := json.Marshal(p.Translations)
	if err != nil {
		return nil, fmt.Errorf("marshal legal translations: %w", err)
	}
	row := r.pool.QueryRow(ctx,
		"UPDATE legal_pages SET translations = $2, updated_at = now() WHERE id = $1 RETURNING "+legalColumns,
		p.ID, tr)
	return scanLegalPage(row)
}
