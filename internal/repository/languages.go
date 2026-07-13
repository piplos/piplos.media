package repository

import (
	"context"
	"fmt"

	"github.com/piplos-media/site/internal/models"
)

// ListLanguages returns all languages ordered for display.
func (r *Repository) ListLanguages(ctx context.Context) ([]models.Language, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT code, name, is_default, enabled, sort_order FROM languages ORDER BY sort_order, code")
	if err != nil {
		return nil, fmt.Errorf("list languages: %w", err)
	}
	defer rows.Close()

	langs := []models.Language{}
	for rows.Next() {
		var l models.Language
		if err := rows.Scan(&l.Code, &l.Name, &l.IsDefault, &l.Enabled, &l.SortOrder); err != nil {
			return nil, fmt.Errorf("scan language: %w", err)
		}
		langs = append(langs, l)
	}
	return langs, rows.Err()
}

// UpsertLanguage creates or updates a language.
func (r *Repository) UpsertLanguage(ctx context.Context, l models.Language) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO languages (code, name, is_default, enabled, sort_order)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name, is_default = EXCLUDED.is_default,
			enabled = EXCLUDED.enabled, sort_order = EXCLUDED.sort_order`,
		l.Code, l.Name, l.IsDefault, l.Enabled, l.SortOrder)
	return err
}

// DeleteLanguage removes a language (default language is protected by the handler).
func (r *Repository) DeleteLanguage(ctx context.Context, code string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM languages WHERE code = $1", code)
	return err
}
