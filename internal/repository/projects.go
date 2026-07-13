package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/piplos-media/site/internal/models"
)

const projectColumns = "id, slug, category, categories, tags, year, featured, published, sort_order, translations, created_at, updated_at"

func scanProject(row pgx.Row) (*models.Project, error) {
	var p models.Project
	var raw []byte
	err := row.Scan(&p.ID, &p.Slug, &p.Category, &p.Categories, &p.Tags, &p.Year,
		&p.Featured, &p.Published, &p.SortOrder, &raw, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan project: %w", err)
	}
	p.Translations = translationsFromJSON(raw)
	return &p, nil
}

// ListProjects returns all projects (admin view, including unpublished).
func (r *Repository) ListProjects(ctx context.Context) ([]models.Project, error) {
	rows, err := r.pool.Query(ctx,
		"SELECT "+projectColumns+" FROM projects ORDER BY sort_order, year DESC, created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	items := []models.Project{}
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *p)
	}
	return items, rows.Err()
}

// GetProject returns a project by UUID or slug, or nil.
func (r *Repository) GetProject(ctx context.Context, ref string) (*models.Project, error) {
	row := r.pool.QueryRow(ctx,
		"SELECT "+projectColumns+" FROM projects WHERE id::text = $1 OR slug = $1", ref)
	return scanProject(row)
}

// CreateProject inserts a project.
func (r *Repository) CreateProject(ctx context.Context, p *models.Project) (*models.Project, error) {
	tr, err := p.Translations.JSON()
	if err != nil {
		return nil, fmt.Errorf("marshal translations: %w", err)
	}
	row := r.pool.QueryRow(ctx,
		`INSERT INTO projects (slug, category, categories, tags, year, featured, published, sort_order, translations)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING `+projectColumns,
		p.Slug, p.Category, p.Categories, p.Tags, p.Year, p.Featured, p.Published, p.SortOrder, tr)
	return scanProject(row)
}

// UpdateProject updates a project.
func (r *Repository) UpdateProject(ctx context.Context, p *models.Project) (*models.Project, error) {
	tr, err := p.Translations.JSON()
	if err != nil {
		return nil, fmt.Errorf("marshal translations: %w", err)
	}
	row := r.pool.QueryRow(ctx,
		`UPDATE projects SET slug = $2, category = $3, categories = $4, tags = $5, year = $6,
			featured = $7, published = $8, sort_order = $9, translations = $10, updated_at = now()
		 WHERE id = $1 RETURNING `+projectColumns,
		p.ID, p.Slug, p.Category, p.Categories, p.Tags, p.Year, p.Featured, p.Published, p.SortOrder, tr)
	return scanProject(row)
}

// DeleteProject removes a project by UUID or slug.
func (r *Repository) DeleteProject(ctx context.Context, ref string) error {
	p, err := r.GetProject(ctx, ref)
	if err != nil {
		return err
	}
	if p == nil {
		return nil
	}
	_, err = r.pool.Exec(ctx, "DELETE FROM projects WHERE id = $1", p.ID)
	return err
}

// ProjectGroupOrder is one service category with ordered project ids.
type ProjectGroupOrder struct {
	GroupID string
	IDs     []string
}

func syncProjectCategories(oldCategory, newCategory string, categories []string) []string {
	if newCategory == "" {
		return categories
	}
	out := make([]string, 0, len(categories)+1)
	seen := map[string]struct{}{}
	for _, c := range categories {
		if c == "" || c == oldCategory {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	if _, ok := seen[newCategory]; !ok {
		out = append([]string{newCategory}, out...)
	}
	return out
}

// ReorderProjects updates category, categories and sort_order from grouped id lists.
func (r *Repository) ReorderProjects(ctx context.Context, groups []ProjectGroupOrder) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin reorder projects tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, group := range groups {
		for i, id := range group.IDs {
			var oldCategory string
			var categories []string
			if err := tx.QueryRow(ctx,
				"SELECT category, categories FROM projects WHERE id = $1",
				id,
			).Scan(&oldCategory, &categories); err != nil {
				return fmt.Errorf("load project %s for reorder: %w", id, err)
			}

			nextCategories := syncProjectCategories(oldCategory, group.GroupID, categories)
			if _, err := tx.Exec(ctx,
				`UPDATE projects SET category = $2, categories = $3, sort_order = $4, updated_at = now() WHERE id = $1`,
				id, group.GroupID, nextCategories, i,
			); err != nil {
				return fmt.Errorf("reorder project %s: %w", id, err)
			}
		}
	}

	return tx.Commit(ctx)
}
