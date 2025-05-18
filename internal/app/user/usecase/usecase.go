package usecase

import (
	"context"

	"github.com/fanzru/e-commerce-be/internal/app/user/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/params"
	"github.com/google/uuid"
)

// UserUseCase defines the interface for user use cases
type UserUseCase interface {
	// Register registers a new user
	Register(ctx context.Context, registerParams params.RegisterUserParams) (*entity.User, error)

	// Login authenticates a user and returns tokens
	Login(ctx context.Context, loginParams params.LoginUserParams) (*params.TokenPair, error)

	// RefreshToken refreshes an access token using a refresh token
	RefreshToken(ctx context.Context, refreshParams params.RefreshTokenParams) (*params.TokenPair, error)

	// Logout invalidates a refresh token
	Logout(ctx context.Context, refreshToken string) error

	// GetUserByID retrieves a user by ID
	GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*entity.User, error)

	// UpdateUser updates a user's details
	UpdateUser(ctx context.Context, id uuid.UUID, updateParams params.UpdateUserParams) (*entity.User, error)

	// UpdatePassword updates a user's password
	UpdatePassword(ctx context.Context, id uuid.UUID, updateParams params.UpdatePasswordParams) error

	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// ListUsers lists users with pagination and filters
	ListUsers(ctx context.Context, page, limit int, role *entity.UserRole) ([]*entity.User, int, error)

	// ValidateToken validates and extracts claims from a token
	ValidateToken(token string) (*params.TokenClaims, error)
}
