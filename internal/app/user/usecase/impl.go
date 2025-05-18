package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/user/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/errs"
	"github.com/fanzru/e-commerce-be/internal/app/user/domain/params"
	"github.com/fanzru/e-commerce-be/internal/app/user/repo"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Define JWT configuration struct to avoid import cycle
type jwtConfig struct {
	SecretKey       string
	ExpirationHours int
}

// UserUseCaseImpl implements the UserUseCase interface
type UserUseCaseImpl struct {
	userRepo  repo.UserRepository
	jwtConfig jwtConfig
}

// NewUserUseCase creates a new instance of UserUseCaseImpl
func NewUserUseCase(userRepo repo.UserRepository, secretKey string, expirationHours int) UserUseCase {
	return &UserUseCaseImpl{
		userRepo: userRepo,
		jwtConfig: jwtConfig{
			SecretKey:       secretKey,
			ExpirationHours: expirationHours,
		},
	}
}

// Register registers a new user
func (uc *UserUseCaseImpl) Register(ctx context.Context, registerParams params.RegisterUserParams) (*entity.User, error) {
	// Check if email already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, registerParams.Email)
	if err == nil && existingUser != nil {
		return nil, errs.ErrEmailAlreadyExists
	}

	// Create user with helper function from entity
	user, err := entity.NewUser(
		registerParams.Email,
		registerParams.Password,
		registerParams.Name,
		entity.RoleCustomer, // Default role is customer
	)
	if err != nil {
		return nil, err
	}

	// Save user to repository
	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (uc *UserUseCaseImpl) Login(ctx context.Context, loginParams params.LoginUserParams) (*params.TokenPair, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, loginParams.Email)
	if err != nil {
		return nil, errs.ErrInvalidCredentials
	}

	// Compare password using entity method
	if !user.ComparePassword(loginParams.Password) {
		return nil, errs.ErrInvalidCredentials
	}

	// Generate JWT token (implemented below)
	accessToken, expiresIn, err := uc.generateJWT(user)
	if err != nil {
		return nil, err
	}

	// Generate refresh token (in a real implementation, you'd store this in a repository)
	refreshTokenID := uuid.New()
	refreshToken := refreshTokenID.String()

	// For a complete implementation, you would store the refresh token in a repository
	// Example:
	// refreshTokenEntity := entity.NewRefreshToken(user.ID, refreshToken, 7) // 7 days
	// err = uc.tokenRepo.SaveRefreshToken(ctx, refreshTokenEntity)
	// if err != nil {
	//     return nil, err
	// }

	return &params.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (uc *UserUseCaseImpl) RefreshToken(ctx context.Context, refreshParams params.RefreshTokenParams) (*params.TokenPair, error) {
	// For a complete implementation, you would validate the refresh token from repository
	// Example:
	// refreshTokenEntity, err := uc.tokenRepo.GetRefreshToken(ctx, refreshParams.RefreshToken)
	// if err != nil {
	//     return nil, errs.ErrInvalidRefreshToken
	// }
	//
	// if refreshTokenEntity.IsExpired() {
	//     return nil, errs.ErrRefreshTokenExpired
	// }
	//
	// user, err := uc.userRepo.GetByID(ctx, refreshTokenEntity.UserID)
	// if err != nil {
	//     return nil, err
	// }

	// For now, simulate by extracting user ID from token directly (not secure for production)
	userID, err := uuid.Parse(refreshParams.RefreshToken)
	if err != nil {
		return nil, errs.ErrInvalidRefreshToken
	}

	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errs.ErrInvalidRefreshToken
	}

	// Generate new access token
	accessToken, expiresIn, err := uc.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &params.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshParams.RefreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// Logout invalidates a refresh token
func (uc *UserUseCaseImpl) Logout(ctx context.Context, refreshToken string) error {
	// For a complete implementation, you would delete the refresh token from repository
	// Example:
	// return uc.tokenRepo.DeleteRefreshToken(ctx, refreshToken)

	// Placeholder for now
	return nil
}

// GetUserByID retrieves a user by ID
func (uc *UserUseCaseImpl) GetUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

// GetUserByEmail retrieves a user by email
func (uc *UserUseCaseImpl) GetUserByEmail(ctx context.Context, email string) (*entity.User, error) {
	return uc.userRepo.GetByEmail(ctx, email)
}

// UpdateUser updates a user's details
func (uc *UserUseCaseImpl) UpdateUser(ctx context.Context, id uuid.UUID, updateParams params.UpdateUserParams) (*entity.User, error) {
	// Get existing user
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if updateParams.Name != nil {
		user.Name = *updateParams.Name
	}
	if updateParams.Email != nil && *updateParams.Email != user.Email {
		// Check if email is already used by another user
		existingUser, err := uc.userRepo.GetByEmail(ctx, *updateParams.Email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			return nil, errs.ErrEmailAlreadyExists
		}
		user.Email = *updateParams.Email
	}

	user.UpdatedAt = time.Now()

	// Save updated user
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdatePassword updates a user's password
func (uc *UserUseCaseImpl) UpdatePassword(ctx context.Context, id uuid.UUID, updateParams params.UpdatePasswordParams) error {
	// Get existing user
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify current password
	if !user.ComparePassword(updateParams.CurrentPassword) {
		return errs.ErrInvalidCurrentPassword
	}

	// Update password
	return user.UpdatePassword(updateParams.NewPassword)
}

// DeleteUser deletes a user
func (uc *UserUseCaseImpl) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return uc.userRepo.Delete(ctx, id)
}

// ListUsers lists users with pagination and filters
func (uc *UserUseCaseImpl) ListUsers(ctx context.Context, page, limit int, role *entity.UserRole) ([]*entity.User, int, error) {
	// Use the List method from interface
	users, count, err := uc.userRepo.List(ctx, page, limit, role)
	if err != nil {
		return nil, 0, err
	}

	return users, count, nil
}

// ValidateToken validates and extracts claims from a token
func (uc *UserUseCaseImpl) ValidateToken(token string) (*params.TokenClaims, error) {
	// Use the ValidateJWT function from middleware package
	return middleware.ValidateJWT(token, uc.jwtConfig.SecretKey)
}

// generateJWT generates a JWT token for a user
func (uc *UserUseCaseImpl) generateJWT(user *entity.User) (string, int, error) {
	// Set expiration time
	expiresIn := uc.jwtConfig.ExpirationHours * 3600 // Convert hours to seconds

	// Create claims with user data
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Duration(expiresIn) * time.Second).Unix(),
		"iat":     time.Now().Unix(),
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(uc.jwtConfig.SecretKey))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresIn, nil
}
