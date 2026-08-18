package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/piplos/piplos.media/internal/models"
)

func scanRefreshSession(row pgx.Row) (*models.RefreshSession, error) {
	var s models.RefreshSession
	err := row.Scan(&s.ID, &s.UserID, &s.TokenHash, &s.ExpiresAt, &s.RevokedAt, &s.RotatedFrom, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan refresh session: %w", err)
	}
	return &s, nil
}

// CreateRefreshSession inserts a new refresh session and returns its ID.
func (r *Repository) CreateRefreshSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time, rotatedFrom *string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`INSERT INTO refresh_sessions (user_id, token_hash, expires_at, rotated_from)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		userID, tokenHash, expiresAt, rotatedFrom,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("create refresh session: %w", err)
	}
	return id, nil
}

// GetSessionByTokenHash returns a session by SHA-256 hash of the opaque refresh token.
func (r *Repository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, rotated_from, created_at
		 FROM refresh_sessions WHERE token_hash = $1`, tokenHash)
	return scanRefreshSession(row)
}

// GetSessionByID returns a session by primary key.
func (r *Repository) GetSessionByID(ctx context.Context, sessionID string) (*models.RefreshSession, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, rotated_from, created_at
		 FROM refresh_sessions WHERE id = $1`, sessionID)
	return scanRefreshSession(row)
}

// IsSessionValid reports whether the session exists, is not revoked, and not expired.
func (r *Repository) IsSessionValid(ctx context.Context, sessionID string) (bool, error) {
	var valid bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM refresh_sessions
			WHERE id = $1 AND revoked_at IS NULL AND expires_at > now()
		)`, sessionID,
	).Scan(&valid)
	if err != nil {
		return false, fmt.Errorf("check session valid: %w", err)
	}
	return valid, nil
}

// RevokeSession marks a session as revoked by ID.
func (r *Repository) RevokeSession(ctx context.Context, sessionID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

// RevokeSessionByTokenHash marks a session revoked by token hash.
func (r *Repository) RevokeSessionByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`, tokenHash)
	if err != nil {
		return fmt.Errorf("revoke session by hash: %w", err)
	}
	return nil
}

// RevokeAllUserSessions revokes every active session for a user.
func (r *Repository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	if err != nil {
		return fmt.Errorf("revoke all user sessions: %w", err)
	}
	return nil
}

// PurgeExpiredSessions deletes expired or revoked sessions older than cutoff.
func (r *Repository) PurgeExpiredSessions(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM refresh_sessions
		 WHERE (expires_at < now() OR revoked_at IS NOT NULL) AND created_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("purge expired sessions: %w", err)
	}
	return tag.RowsAffected(), nil
}
