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

// ClaimRefreshSession atomically revokes the session if it is still active
// (never revoked and not expired) and reports whether this call won the claim.
// It is the serialization point of refresh-token rotation: concurrent refresh
// requests presenting the same token run this UPDATE against the same row, so
// exactly one of them can see affected>0; every loser means a replay.
func (r *Repository) ClaimRefreshSession(ctx context.Context, sessionID string) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE refresh_sessions SET revoked_at = now()
		 WHERE id = $1 AND revoked_at IS NULL AND expires_at > now()`, sessionID)
	if err != nil {
		return false, fmt.Errorf("claim refresh session: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

// RevokeSessionChain revokes every active session in the rotation subtree
// rooted at sessionID: the session itself plus everything rotated from it,
// directly or transitively (refresh_sessions.rotated_from forms a tree).
// Used by reuse detection: an already-revoked token coming back means its
// rotation chain leaked, so all live tokens derived from it must die.
func (r *Repository) RevokeSessionChain(ctx context.Context, sessionID string) error {
	_, err := r.pool.Exec(ctx,
		`WITH RECURSIVE chain(id) AS (
			SELECT id FROM refresh_sessions WHERE id = $1
			UNION
			SELECT s.id FROM refresh_sessions s JOIN chain c ON s.rotated_from = c.id
		)
		UPDATE refresh_sessions r SET revoked_at = now()
		WHERE r.revoked_at IS NULL AND r.id IN (SELECT id FROM chain)`, sessionID)
	if err != nil {
		return fmt.Errorf("revoke session chain: %w", err)
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
