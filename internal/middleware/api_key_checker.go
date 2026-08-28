package middleware

import (
	"context"

	"github.com/piplos/piplos.media/internal/models"
)

// APIKeyChecker resolves agent API keys by SHA-256 hash and records usage.
type APIKeyChecker interface {
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error)
	TouchAPIKeyLastUsed(ctx context.Context, id string) error
}
