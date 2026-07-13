// Package errors provides structured application errors with HTTP status mapping.
package errors

import (
	"errors"
	"net/http"
)

// Error codes used in API responses.
const (
	CodeInternal        = "internal_error"
	CodeInvalidRequest  = "invalid_request"
	CodeUnauthorized    = "unauthorized"
	CodeForbidden       = "forbidden"
	CodeAccountDisabled = "account_disabled"
	CodeNotFound        = "not_found"
	CodeConflict        = "conflict"
	CodeServiceError    = "service_error"
)

var codeToStatus = map[string]int{
	CodeInvalidRequest:  http.StatusBadRequest,
	CodeUnauthorized:    http.StatusUnauthorized,
	CodeForbidden:       http.StatusForbidden,
	CodeAccountDisabled: http.StatusForbidden,
	CodeNotFound:        http.StatusNotFound,
	CodeConflict:        http.StatusConflict,
	CodeServiceError:    http.StatusServiceUnavailable,
	CodeInternal:        http.StatusInternalServerError,
}

// AppError is a structured application error.
type AppError struct {
	Code    string
	Message string
	Cause   error
	Status  int
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Cause }

// New creates an AppError with the status derived from code.
func New(code, message string) *AppError {
	status, ok := codeToStatus[code]
	if !ok {
		status = http.StatusInternalServerError
	}
	return &AppError{Code: code, Message: message, Status: status}
}

// From converts any error into an AppError (default: internal_error).
func From(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return &AppError{Code: CodeInternal, Message: err.Error(), Cause: err, Status: http.StatusInternalServerError}
}

func ErrInvalidRequest(msg string) *AppError  { return New(CodeInvalidRequest, msg) }
func ErrUnauthorized(msg string) *AppError    { return New(CodeUnauthorized, msg) }
func ErrForbidden(msg string) *AppError       { return New(CodeForbidden, msg) }
func ErrAccountDisabled(msg string) *AppError { return New(CodeAccountDisabled, msg) }
func ErrNotFound(msg string) *AppError        { return New(CodeNotFound, msg) }
func ErrConflict(msg string) *AppError        { return New(CodeConflict, msg) }
func ErrInternal(msg string) *AppError        { return New(CodeInternal, msg) }
func ErrService(msg string) *AppError         { return New(CodeServiceError, msg) }
