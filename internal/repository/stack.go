package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/piplos/site/internal/models"
)

const stackColumns = "id, slug, label, group_id, published, sort_order, created_at, updated_at"

func scanStackItem(row pgx.Row) (*models.StackItem, error) {
	var s models.StackItem
	err := row.Scan(&s.ID, &s.Slug, &s.Label, &s.GroupID, &s.Published, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan stack item: %w", err)
	}
	return &s, nil
}

// ListStackItems returns all stack items ordered by group and sort order.
func (r *Repository) ListStackItems(ctx context.Context) ([]models.StackItem, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+stackColumns+" FROM stack_items ORDER BY group_id, sort_order, label")
	if err != nil {
		return nil, fmt.Errorf("list stack items: %w", err)
	}
	defer rows.Close()

	items := []models.StackItem{}
	for rows.Next() {
		s, err := scanStackItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	return items, rows.Err()
}

// CreateStackItem inserts a stack item.
func (r *Repository) CreateStackItem(ctx context.Context, s *models.StackItem) (*models.StackItem, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO stack_items (slug, label, group_id, published, sort_order)
		 VALUES ($1, $2, $3, $4, $5) RETURNING `+stackColumns,
		s.Slug, s.Label, s.GroupID, s.Published, s.SortOrder)
	return scanStackItem(row)
}

// UpdateStackItem updates a stack item.
func (r *Repository) UpdateStackItem(ctx context.Context, s *models.StackItem) (*models.StackItem, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE stack_items SET slug = $2, label = $3, group_id = $4, published = $5,
			sort_order = $6, updated_at = now()
		 WHERE id = $1 RETURNING `+stackColumns,
		s.ID, s.Slug, s.Label, s.GroupID, s.Published, s.SortOrder)
	return scanStackItem(row)
}

// DeleteStackItem removes a stack item.
func (r *Repository) DeleteStackItem(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM stack_items WHERE id = $1", id)
	return err
}

// StackGroupOrder is one service group with ordered stack item ids.
type StackGroupOrder struct {
	GroupID string
	IDs     []string
}

// ReorderStack updates group_id and sort_order from grouped id lists.
func (r *Repository) ReorderStack(ctx context.Context, groups []StackGroupOrder) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin reorder stack tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, group := range groups {
		for i, id := range group.IDs {
			if _, err := tx.Exec(ctx,
				`UPDATE stack_items SET group_id = $2, sort_order = $3, updated_at = now() WHERE id = $1`,
				id, group.GroupID, i,
			); err != nil {
				return fmt.Errorf("reorder stack item %s: %w", id, err)
			}
		}
	}

	return tx.Commit(ctx)
}
