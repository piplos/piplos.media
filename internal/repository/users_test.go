package repository

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsUniqueViolation(t *testing.T) {
	pgDup := &pgconn.PgError{Code: "23505", Message: "duplicate key value violates unique constraint"}
	pgOther := &pgconn.PgError{Code: "42P01", Message: "relation does not exist"}

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"direct pg unique violation", pgDup, true},
		{"wrapped pg unique violation", fmt.Errorf("create user: %w", pgDup), true},
		{"other pg error", pgOther, false},
		{"plain error", errors.New("boom"), false},
		{"nil", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isUniqueViolation(tc.err); got != tc.want {
				t.Fatalf("isUniqueViolation(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
