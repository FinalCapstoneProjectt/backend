package errors

import "errors"

var (
	ErrNotFound        = errors.New("resource not found")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrBadRequest      = errors.New("bad request")
	ErrInternalServer  = errors.New("internal server error")
	ErrConflict        = errors.New("resource conflict")
	ErrValidation      = errors.New("validation failed")
	ErrUnauthenticated = errors.New("authentication required")
	ErrInvalidToken    = errors.New("invalid or expired token")
)
