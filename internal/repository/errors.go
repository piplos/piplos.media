package repository

import "errors"

// Domain errors for user integrity; handlers render them as HTTP 409.
var (
	// ErrLastActiveAdmin is returned when a mutation would leave the system
	// without any active administrator.
	ErrLastActiveAdmin = errors.New("cannot remove the last active administrator")

	// ErrDuplicateEmail is returned when inserting a user violates the
	// unique constraint on users.email.
	ErrDuplicateEmail = errors.New("user with this email already exists")

	// ErrDuplicateSlug is returned when inserting/updating a page violates
	// the unique constraint on pages.slug.
	ErrDuplicateSlug = errors.New("page with this slug already exists")
)
