package repo

import (
	"context"

	"github.com/fanzru/e-commerce-be/internal/app/user/domain/entity"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user repositories
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update updates a user's details
	Update(ctx context.Context, user *entity.User) error

	// Delete soft-deletes a user
	Delete(ctx context.Context, id uuid.UUID) error

	// List lists users with pagination and filters
	List(ctx context.Context, page, limit int, role *entity.UserRole) ([]*entity.User, int, error)
}

// TokenRepository defines the interface for token repositories
type TokenRepository interface {
	// SaveRefreshToken saves a refresh token
	SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error

	// GetRefreshToken retrieves a refresh token
	GetRefreshToken(ctx context.Context, tokenStr string) (*entity.RefreshToken, error)

	// DeleteRefreshToken deletes a refresh token
	DeleteRefreshToken(ctx context.Context, tokenStr string) error

	// DeleteUserTokens deletes all tokens for a user
	DeleteUserTokens(ctx context.Context, userID uuid.UUID) error
}
