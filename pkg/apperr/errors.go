package apperr

import "errors"

// Sentinel errors for domain-level failure cases.
// HTTP layers map these to appropriate status codes.
var (
	ErrNotFound          = errors.New("not found")
	ErrConflict          = errors.New("conflict")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrValidation        = errors.New("validation error")
	ErrInternal          = errors.New("internal error")
	ErrInvalidCredential = errors.New("invalid credentials")
)

// AppError wraps a sentinel with a human-readable message and optional detail.
type AppError struct {
	Code    error
	Message string
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Code }

// New creates an AppError with the given sentinel and message.
func New(code error, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Is satisfies errors.Is for sentinel matching.
func Is(err, target error) bool { return errors.Is(err, target) }
