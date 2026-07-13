package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/piplos-media/site/internal/models"
)

const aiProviderModelColumns = `id, provider, model_id, display_name, enabled`

func scanAIProviderModel(row pgx.Row, m *models.AIProviderModel) error {
	return row.Scan(&m.ID, &m.Provider, &m.ModelID, &m.DisplayName, &m.Enabled)
}

// ListAIProviderModels returns all catalog models.
func (r *Repository) ListAIProviderModels(ctx context.Context) ([]models.AIProviderModel, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+aiProviderModelColumns+` FROM ai_provider_models ORDER BY provider, display_name`)
	if err != nil {
		return nil, fmt.Errorf("list ai provider models: %w", err)
	}
	defer rows.Close()

	var list []models.AIProviderModel
	for rows.Next() {
		var m models.AIProviderModel
		if err := scanAIProviderModel(rows, &m); err != nil {
			return nil, fmt.Errorf("scan ai provider model: %w", err)
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

// ListEnabledAIProviderModels returns enabled models for a provider.
func (r *Repository) ListEnabledAIProviderModels(ctx context.Context, provider string) ([]models.AIProviderModel, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+aiProviderModelColumns+` FROM ai_provider_models WHERE provider = $1 AND enabled = true ORDER BY display_name`,
		provider)
	if err != nil {
		return nil, fmt.Errorf("list enabled ai models: %w", err)
	}
	defer rows.Close()

	var list []models.AIProviderModel
	for rows.Next() {
		var m models.AIProviderModel
		if err := scanAIProviderModel(rows, &m); err != nil {
			return nil, fmt.Errorf("scan ai provider model: %w", err)
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

// GetAIProviderModelByID returns one model or nil.
func (r *Repository) GetAIProviderModelByID(ctx context.Context, id uuid.UUID) (*models.AIProviderModel, error) {
	var m models.AIProviderModel
	err := scanAIProviderModel(
		r.pool.QueryRow(ctx, `SELECT `+aiProviderModelColumns+` FROM ai_provider_models WHERE id = $1`, id),
		&m,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get ai provider model: %w", err)
	}
	return &m, nil
}

// CreateAIProviderModel inserts a catalog model.
func (r *Repository) CreateAIProviderModel(ctx context.Context, provider, modelID, displayName string) (*models.AIProviderModel, error) {
	var m models.AIProviderModel
	err := scanAIProviderModel(
		r.pool.QueryRow(ctx,
			`INSERT INTO ai_provider_models (provider, model_id, display_name) VALUES ($1, $2, $3)
			 RETURNING `+aiProviderModelColumns,
			provider, modelID, displayName),
		&m,
	)
	if err != nil {
		return nil, fmt.Errorf("create ai provider model: %w", err)
	}
	return &m, nil
}

// UpdateAIProviderModel updates display name and enabled flag.
func (r *Repository) UpdateAIProviderModel(ctx context.Context, id uuid.UUID, displayName string, enabled bool) (*models.AIProviderModel, error) {
	var m models.AIProviderModel
	err := scanAIProviderModel(
		r.pool.QueryRow(ctx,
			`UPDATE ai_provider_models SET display_name = $1, enabled = $2, updated_at = NOW() WHERE id = $3
			 RETURNING `+aiProviderModelColumns,
			displayName, enabled, id),
		&m,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("update ai provider model: %w", err)
	}
	return &m, nil
}

// DeleteAIProviderModel removes a catalog model.
func (r *Repository) DeleteAIProviderModel(ctx context.Context, id uuid.UUID) (bool, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM ai_provider_models WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete ai provider model: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}
