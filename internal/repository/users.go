package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/piplos/piplos.media/internal/models"
)

const userColumns = "id, email, password_hash, full_name, role, is_active, created_at, updated_at"

func scanUser(row pgx.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
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
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
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

// CreateUser inserts a user and returns it.
func (r *Repository) CreateUser(ctx context.Context, email, passwordHash, fullName string, role models.UserRole) (*models.User, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, full_name, role)
		 VALUES ($1, $2, $3, $4) RETURNING `+userColumns,
		email, passwordHash, fullName, role)
	return scanUser(row)
}

// UpdateUser updates mutable user fields. Empty passwordHash keeps the old password.
func (r *Repository) UpdateUser(ctx context.Context, id, fullName string, role models.UserRole, isActive bool, passwordHash string) (*models.User, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET
			full_name = $2,
			role = $3,
			is_active = $4,
			password_hash = CASE WHEN $5 = '' THEN password_hash ELSE $5 END,
			updated_at = now()
		 WHERE id = $1 RETURNING `+userColumns,
		id, fullName, role, isActive, passwordHash)
	return scanUser(row)
}

// DeleteUser removes a user.
func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}
