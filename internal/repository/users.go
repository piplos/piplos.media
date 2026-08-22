package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/piplos/piplos.media/internal/models"
)

const userColumns = "id, email, password_hash, full_name, role, is_active, notify_leads, created_at, updated_at"

func scanUser(row pgx.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.NotifyLeads, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}

// GetUserByID returns a user or nil when not found.
func (r *Repository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+userColumns+" FROM users WHERE id = $1", id)
	return scanUser(row)
}

// GetUserByEmail returns a user or nil when not found.
func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.pool.QueryRow(ctx, "SELECT "+userColumns+" FROM users WHERE lower(email) = lower($1)", email)
	return scanUser(row)
}

// ListUsers returns all users ordered by creation date.
func (r *Repository) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := r.pool.Query(ctx, "SELECT "+userColumns+" FROM users ORDER BY created_at")
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

// CountUsers returns the total number of users.
func (r *Repository) CountUsers(ctx context.Context) (int, error) {
	var n int
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&n); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return n, nil
}

// isUniqueViolation reports whether err is (or wraps) a Postgres unique_violation.
// The only unique constraint on users is email.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// lockActiveAdminIDs locks all active admin rows until the enclosing transaction
// ends, so concurrent role/deactivation changes serialize instead of racing to
// demote each other into zero-admin state (aggregates cannot take row locks).
// Taking this lock BEFORE the mutation also prevents deadlocks between two
// transactions demoting different admins concurrently.
func lockActiveAdminIDs(ctx context.Context, tx pgx.Tx) ([]string, error) {
	rows, err := tx.Query(ctx,
		"SELECT id FROM users WHERE role = $1 AND is_active FOR UPDATE", models.RoleAdmin)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// CreateUser inserts a user and returns it. Concurrent inserts of the same
// email return ErrDuplicateEmail instead of a raw constraint error.
func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, fullName string, role models.UserRole, notifyLeads bool) (*models.User, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, full_name, role, notify_leads)
		 VALUES ($1, $2, $3, $4, $5) RETURNING `+userColumns,
		email, passwordHash, fullName, role, notifyLeads)
	u, err := scanUser(row)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("create user: %w", ErrDuplicateEmail)
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// updateUserTx applies the change inside tx and returns the updated user
// (nil when the id does not exist).
func updateUserTx(ctx context.Context, tx pgx.Tx, id, fullName string, role models.UserRole, isActive bool, notifyLeads bool, passwordHash string) (*models.User, error) {
	row := tx.QueryRow(ctx,
		`UPDATE users SET
			full_name = $2,
			role = $3,
			is_active = $4,
			notify_leads = $5,
			password_hash = CASE WHEN $6 = '' THEN password_hash ELSE $6 END,
			updated_at = now()
		 WHERE id = $1 RETURNING `+userColumns,
		id, fullName, role, isActive, notifyLeads, passwordHash)
	return scanUser(row)
}

// guardLastActiveAdmin fails the transaction when no active administrator
// remains after the pending change. Must run while lockActiveAdminIDs still
// holds its locks. Degenerate databases that already have zero active admins
// reject every mutation here — fail-closed by design.
func guardLastActiveAdmin(ctx context.Context, tx pgx.Tx) error {
	admins, err := lockActiveAdminIDs(ctx, tx)
	if err != nil {
		return fmt.Errorf("lock active admins: %w", err)
	}
	if len(admins) == 0 {
		return ErrLastActiveAdmin
	}
	return nil
}

// UpdateUser updates mutable user fields. Empty passwordHash keeps the old
// password. Removing the last active admin (demotion or deactivation) rolls
// back and returns ErrLastActiveAdmin.
func (r *Repository) UpdateUser(ctx context.Context, id, fullName string, role models.UserRole, isActive bool, notifyLeads bool, passwordHash string) (*models.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update user tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Сериализуем мутации админ-набора до изменения, чтобы конкурентные
	// понижения разных админов не сталкивались дедлоком.
	if _, err := lockActiveAdminIDs(ctx, tx); err != nil {
		return nil, fmt.Errorf("lock active admins: %w", err)
	}

	u, err := updateUserTx(ctx, tx, id, fullName, role, isActive, notifyLeads, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}
	if u == nil {
		return nil, nil
	}
	if err := guardLastActiveAdmin(ctx, tx); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit update user: %w", err)
	}
	return u, nil
}

// DeleteUser removes a user. Deleting the last active admin rolls back and
// returns ErrLastActiveAdmin.
func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin delete user tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := lockActiveAdminIDs(ctx, tx); err != nil {
		return fmt.Errorf("lock active admins: %w", err)
	}

	tag, err := tx.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	if err := guardLastActiveAdmin(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
