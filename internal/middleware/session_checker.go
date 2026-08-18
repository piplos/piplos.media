package middleware

import (
	"context"
)

// SessionChecker validates refresh session rows referenced by JWT sid claims.
type SessionChecker interface {
	IsSessionValid(ctx context.Context, sessionID string) (bool, error)
}
