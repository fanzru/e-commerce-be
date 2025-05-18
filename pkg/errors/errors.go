package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error types
var (
	// ErrInternalServer indicates an internal server error
	ErrInternalServer = errors.New("internal server error")

	// ErrNotFound indicates a requested resource was not found
	ErrNotFound = errors.New("resource not found")

	// ErrBadRequest indicates a bad request from the client
	ErrBadRequest = errors.New("bad request")

	// ErrInvalidParameter indicates an invalid parameter in a request
	ErrInvalidParameter = errors.New("invalid parameter")

	// ErrUnauthorized indicates an unauthorized request
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates a forbidden request
	ErrForbidden = errors.New("forbidden")

	// ErrConflict indicates a conflict with the current state
	ErrConflict = errors.New("conflict")
)

// AppError represents an application error with HTTP status code
type AppError struct {
	Err     error
	Status  int
	Message string
	Data    map[string]interface{}
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(err error, status int, message string) *AppError {
	return &AppError{
		Err:     err,
		Status:  status,
		Message: message,
	}
}

// NewWithData creates a new AppError with additional data
func NewWithData(err error, status int, message string, data map[string]interface{}) *AppError {
	return &AppError{
		Err:     err,
		Status:  status,
		Message: message,
		Data:    data,
	}
}

// NewNotFound creates a new not found error
func NewNotFound(message string) *AppError {
	return &AppError{
		Err:     ErrNotFound,
		Status:  http.StatusNotFound,
		Message: message,
	}
}

// NewBadRequest creates a new bad request error
func NewBadRequest(message string) *AppError {
	return &AppError{
		Err:     ErrBadRequest,
		Status:  http.StatusBadRequest,
		Message: message,
	}
}

// NewInternalServerError creates a new internal server error
func NewInternalServerError(err error) *AppError {
	message := "internal server error"
	if err != nil {
		message = err.Error()
	}

	return &AppError{
		Err:     ErrInternalServer,
		Status:  http.StatusInternalServerError,
		Message: message,
	}
}

// NewConflict creates a new conflict error
func NewConflict(message string) *AppError {
	return &AppError{
		Err:     ErrConflict,
		Status:  http.StatusConflict,
		Message: message,
	}
}

// NewUnauthorized creates a new unauthorized error
func NewUnauthorized(message string) *AppError {
	return &AppError{
		Err:     ErrUnauthorized,
		Status:  http.StatusUnauthorized,
		Message: message,
	}
}

// NewForbidden creates a new forbidden error
func NewForbidden(message string) *AppError {
	return &AppError{
		Err:     ErrForbidden,
		Status:  http.StatusForbidden,
		Message: message,
	}
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, ErrNotFound)
	}
	return errors.Is(err, ErrNotFound)
}

// IsBadRequest checks if the error is a bad request error
func IsBadRequest(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, ErrBadRequest)
	}
	return errors.Is(err, ErrBadRequest)
}

// IsInternalServerError checks if the error is an internal server error
func IsInternalServerError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return errors.Is(appErr.Err, ErrInternalServer)
	}
	return errors.Is(err, ErrInternalServer)
}

// NewValidationError creates a new validation error with field errors
func NewValidationError(fieldErrors map[string]string) *AppError {
	return &AppError{
		Err:     ErrInvalidParameter,
		Status:  http.StatusBadRequest,
		Message: "validation error",
		Data:    map[string]interface{}{"fields": fieldErrors},
	}
}

// Wrap wraps an error with a message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
