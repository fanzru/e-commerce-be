package params

import (
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/entity"
)

// RegisterUserParams defines parameters for user registration
type RegisterUserParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

// LoginUserParams defines parameters for user login
type LoginUserParams struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserParams defines parameters for updating user details
type UpdateUserParams struct {
	Name  *string `json:"name"`
	Email *string `json:"email" validate:"omitempty,email"`
}

// UpdatePasswordParams defines parameters for updating user password
type UpdatePasswordParams struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// RefreshTokenParams defines parameters for refreshing an access token
type RefreshTokenParams struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // Seconds until access token expires
}

// TokenClaims represents the claims in a JWT token
type TokenClaims struct {
	UserID string          `json:"user_id"`
	Email  string          `json:"email"`
	Role   entity.UserRole `json:"role"`
}
