package handlers

import (
	apperrors "github.com/piplos/piplos.media/internal/errors"
)

// withCause attaches the underlying error to an AppError so the request
// logger (middleware.ErrorHandler) records the root cause, not just the
// generic message. The HTTP response itself only carries code and message.
func withCause(e *apperrors.AppError, cause error) *apperrors.AppError {
	e.Cause = cause
	return e
}

// internalErr builds a 500 AppError keeping the underlying cause for logging.
func internalErr(msg string, cause error) *apperrors.AppError {
	return withCause(apperrors.ErrInternal(msg), cause)
}
