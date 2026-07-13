package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/piplos-media/site/internal/models"
)

const leadColumns = `id, types, project_name, description, stack, reference_urls, budget, currency,
	timeline, stage, first_name, last_name, email, company, phone, how_found, notes, lang, status,
	created_at, updated_at`

func scanLead(row pgx.Row) (*models.Lead, error) {
	var l models.Lead
	err := row.Scan(&l.ID, &l.Types, &l.ProjectName, &l.Description, &l.Stack, &l.ReferenceURLs,
		&l.Budget, &l.Currency, &l.Timeline, &l.Stage, &l.FirstName, &l.LastName, &l.Email,
		&l.Company, &l.Phone, &l.HowFound, &l.Notes, &l.Lang, &l.Status, &l.CreatedAt, &l.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan lead: %w", err)
	}
	return &l, nil
}

// CreateLead stores a request from the site order form.
func (r *Repository) CreateLead(ctx context.Context, l *models.Lead) (*models.Lead, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO leads (types, project_name, description, stack, reference_urls, budget, currency,
			timeline, stage, first_name, last_name, email, company, phone, how_found, notes, lang)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
		 RETURNING `+leadColumns,
		l.Types, l.ProjectName, l.Description, l.Stack, l.ReferenceURLs, l.Budget, l.Currency,
		l.Timeline, l.Stage, l.FirstName, l.LastName, l.Email, l.Company, l.Phone, l.HowFound,
		l.Notes, l.Lang)
	return scanLead(row)
}

// ListLeads returns leads, optionally filtered by status, newest first.
func (r *Repository) ListLeads(ctx context.Context, status string, limit, offset int) ([]models.Lead, int, error) {
	var total int
	countQuery := "SELECT count(*) FROM leads"
	listQuery := "SELECT " + leadColumns + " FROM leads"
	args := []any{}
	if status != "" {
		countQuery += " WHERE status = $1"
		listQuery += " WHERE status = $1"
		args = append(args, status)
	}
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count leads: %w", err)
	}

	listQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list leads: %w", err)
	}
	defer rows.Close()

	items := []models.Lead{}
	for rows.Next() {
		l, err := scanLead(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *l)
	}
	return items, total, rows.Err()
}

// GetLead returns a lead by id or nil.
func (r *Repository) GetLead(ctx context.Context, id string) (*models.Lead, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+leadColumns+" FROM leads WHERE id = $1", id)
	return scanLead(row)
}

// UpdateLeadStatus changes the processing status.
func (r *Repository) UpdateLeadStatus(ctx context.Context, id string, status models.LeadStatus) (*models.Lead, error) {
	row := r.pool.QueryRow(ctx,
		"UPDATE leads SET status = $2, updated_at = now() WHERE id = $1 RETURNING "+leadColumns,
		id, status)
	return scanLead(row)
}

// DeleteLead removes a lead.
func (r *Repository) DeleteLead(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM leads WHERE id = $1", id)
	return err
}
