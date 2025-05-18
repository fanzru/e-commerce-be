package errs

import (
	"errors"
	"fmt"
)

// UserNotFoundError represents an error when a user is not found
type UserNotFoundError struct {
	ID    string
	Email string
}

func (e UserNotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("user with ID %s not found", e.ID)
	}
	return fmt.Sprintf("user with email %s not found", e.Email)
}

// AuthenticationError represents an error during authentication
type AuthenticationError struct {
	Message string
}

func (e AuthenticationError) Error() string {
	if e.Message == "" {
		return "authentication failed"
	}
	return e.Message
}

// TokenError represents an error related to tokens
type TokenError struct {
	Message string
}

func (e TokenError) Error() string {
	if e.Message == "" {
		return "token error"
	}
	return e.Message
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

// InvalidCredentialsError represents invalid login credentials
type InvalidCredentialsError struct{}

func (e InvalidCredentialsError) Error() string {
	return "invalid email or password"
}

// EmailAlreadyExistsError indicates a duplicate email
type EmailAlreadyExistsError struct {
	Email string
}

func (e EmailAlreadyExistsError) Error() string {
	return fmt.Sprintf("email %s is already registered", e.Email)
}

// User domain specific errors
var (
	// ErrEmailAlreadyExists is returned when trying to register with an email that already exists
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")

	// ErrInvalidCurrentPassword is returned when the current password is incorrect
	ErrInvalidCurrentPassword = errors.New("current password is incorrect")

	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrInvalidRefreshToken is returned when the refresh token is invalid
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	// ErrRefreshTokenExpired is returned when the refresh token has expired
	ErrRefreshTokenExpired = errors.New("refresh token expired")

	// ErrUnauthorized is returned when a user is not authorized to perform an action
	ErrUnauthorized = errors.New("unauthorized")
)
