package middleware

import (
	"context"

	"github.com/piplos-media/site/internal/models"
)

// UserLookup loads users for JWT authentication (implemented by repository.Repository).
type UserLookup interface {
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}
