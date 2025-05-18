package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/user/domain/entity"
	userErrs "github.com/fanzru/e-commerce-be/internal/app/user/domain/errs"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/google/uuid"
)

// userRepository implements UserRepository using PostgreSQL
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	logger := middleware.Logger.With(
		"method", "UserRepository.Create",
		"email", user.Email,
		"role", user.Role,
	)
	logger.Debug("Creating new user")
	startTime := time.Now()

	// Check if email already exists
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", user.Email).Scan(&count)
	if err != nil {
		logger.Error("Failed to check email existence", "error", err.Error())
		return fmt.Errorf("failed to check email existence: %w", err)
	}

	if count > 0 {
		logger.Warn("Email already exists", "error", "EmailAlreadyExistsError")
		return &userErrs.EmailAlreadyExistsError{Email: user.Email}
	}

	// Insert new user
	query := `
		INSERT INTO users (id, email, password, name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.Password, user.Name, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		logger.Error("Failed to create user", "error", err.Error())
		return fmt.Errorf("failed to create user: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully created user",
		"user_id", user.ID.String(),
		"name", user.Name,
		"duration_ms", duration.Milliseconds())

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	logger := middleware.Logger.With(
		"method", "UserRepository.GetByID",
		"user_id", id.String(),
	)
	logger.Debug("Fetching user by ID")
	startTime := time.Now()

	query := `
		SELECT id, email, password, name, role, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	user, err := r.scanUser(ctx, query, id)
	if err != nil {
		if _, ok := err.(*userErrs.UserNotFoundError); ok {
			logger.Warn("User not found", "error", "UserNotFoundError")
		} else {
			logger.Error("Failed to get user by ID", "error", err.Error())
		}
		return nil, err
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved user",
		"email", user.Email,
		"name", user.Name,
		"role", user.Role,
		"duration_ms", duration.Milliseconds())

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	logger := middleware.Logger.With(
		"method", "UserRepository.GetByEmail",
		"email", email,
	)
	logger.Debug("Fetching user by email")
	startTime := time.Now()

	query := `
		SELECT id, email, password, name, role, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`
	user, err := r.scanUser(ctx, query, email)
	if err != nil {
		if _, ok := err.(*userErrs.UserNotFoundError); ok {
			logger.Warn("User not found", "error", "UserNotFoundError")
		} else {
			logger.Error("Failed to get user by email", "error", err.Error())
		}
		return nil, err
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved user",
		"user_id", user.ID.String(),
		"name", user.Name,
		"role", user.Role,
		"duration_ms", duration.Milliseconds())

	return user, nil
}

// Update updates a user's details
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	logger := middleware.Logger.With(
		"method", "UserRepository.Update",
		"user_id", user.ID.String(),
	)
	logger.Debug("Updating user")
	startTime := time.Now()

	query := `
		UPDATE users
		SET email = $1, password = $2, name = $3, role = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
		RETURNING id
	`
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query,
		user.Email, user.Password, user.Name, user.Role, time.Now(), user.ID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("User not found", "error", "UserNotFoundError")
			return &userErrs.UserNotFoundError{ID: user.ID.String()}
		}
		logger.Error("Failed to update user", "error", err.Error())
		return fmt.Errorf("failed to update user: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully updated user",
		"email", user.Email,
		"name", user.Name,
		"role", user.Role,
		"duration_ms", duration.Milliseconds())

	return nil
}

// Delete soft-deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	logger := middleware.Logger.With(
		"method", "UserRepository.Delete",
		"user_id", id.String(),
	)
	logger.Debug("Soft-deleting user")
	startTime := time.Now()

	query := `
		UPDATE users
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id
	`
	var userID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, time.Now(), id).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("User not found", "error", "UserNotFoundError")
			return &userErrs.UserNotFoundError{ID: id.String()}
		}
		logger.Error("Failed to delete user", "error", err.Error())
		return fmt.Errorf("failed to delete user: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully deleted user",
		"duration_ms", duration.Milliseconds())

	return nil
}

// List lists users with pagination and filters
func (r *userRepository) List(ctx context.Context, page, limit int, role *entity.UserRole) ([]*entity.User, int, error) {
	logger := middleware.Logger.With(
		"method", "UserRepository.List",
		"page", page,
		"limit", limit,
	)
	if role != nil {
		logger = logger.With("role", *role)
	}
	logger.Debug("Listing users with filters")
	startTime := time.Now()

	// Calculate offset for pagination
	offset := (page - 1) * limit

	// Build query with optional role filter
	countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at IS NULL"
	listQuery := `
		SELECT id, email, password, name, role, created_at, updated_at, NULL as deleted_at
		FROM users
		WHERE deleted_at IS NULL
	`
	var args []interface{}
	if role != nil {
		countQuery += " AND role = $1"
		listQuery += " AND role = $1"
		args = append(args, *role)
	}

	// Add pagination to list query
	listQuery += " ORDER BY created_at DESC"
	if len(args) > 0 {
		listQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	} else {
		listQuery += " LIMIT $1 OFFSET $2"
		args = append(args, limit, offset)
	}

	// Get total count
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		logger.Error("Failed to count users", "error", err.Error())
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users
	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		logger.Error("Failed to query users", "error", err.Error())
		return nil, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	// Scan users
	users := []*entity.User{}
	for rows.Next() {
		user := &entity.User{}
		var deletedAt sql.NullTime
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Password,
			&user.Name,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			logger.Error("Failed to scan user row", "error", err.Error())
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		if deletedAt.Valid {
			user.DeletedAt = &deletedAt.Time
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Failed to iterate user rows", "error", err.Error())
		return nil, 0, fmt.Errorf("failed to iterate user rows: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully listed users",
		"total_count", total,
		"returned_count", len(users),
		"duration_ms", duration.Milliseconds())

	return users, total, nil
}

func (r *userRepository) scanUser(ctx context.Context, query string, args ...interface{}) (*entity.User, error) {
	row := r.db.QueryRowContext(ctx, query, args...)

	user := &entity.User{}
	var deletedAt sql.NullTime
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if len(args) > 0 {
				switch arg := args[0].(type) {
				case uuid.UUID:
					return nil, &userErrs.UserNotFoundError{ID: arg.String()}
				case string:
					return nil, &userErrs.UserNotFoundError{Email: arg}
				}
			}
			return nil, &userErrs.UserNotFoundError{}
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return user, nil
}

// tokenRepository implements TokenRepository using PostgreSQL
type tokenRepository struct {
	db *sql.DB
}

// NewTokenRepository creates a new PostgreSQL token repository
func NewTokenRepository(db *sql.DB) TokenRepository {
	return &tokenRepository{
		db: db,
	}
}

// SaveRefreshToken saves a refresh token
func (r *tokenRepository) SaveRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	logger := middleware.Logger.With(
		"method", "TokenRepository.SaveRefreshToken",
		"user_id", token.UserID.String(),
	)
	logger.Debug("Saving refresh token")
	startTime := time.Now()

	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt)
	if err != nil {
		logger.Error("Failed to save refresh token", "error", err.Error())
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully saved refresh token",
		"token_id", token.ID.String(),
		"expires_at", token.ExpiresAt,
		"duration_ms", duration.Milliseconds())

	return nil
}

// GetRefreshToken retrieves a refresh token
func (r *tokenRepository) GetRefreshToken(ctx context.Context, token string) (*entity.RefreshToken, error) {
	logger := middleware.Logger.With(
		"method", "TokenRepository.GetRefreshToken",
	)
	logger.Debug("Fetching refresh token")
	startTime := time.Now()

	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1
	`
	var refreshToken entity.RefreshToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Refresh token not found", "error", "TokenError")
			return nil, &userErrs.TokenError{Message: "refresh token not found"}
		}
		logger.Error("Failed to get refresh token", "error", err.Error())
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved refresh token",
		"token_id", refreshToken.ID.String(),
		"user_id", refreshToken.UserID.String(),
		"expires_at", refreshToken.ExpiresAt,
		"duration_ms", duration.Milliseconds())

	return &refreshToken, nil
}

// DeleteRefreshToken deletes a refresh token
func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	logger := middleware.Logger.With(
		"method", "TokenRepository.DeleteRefreshToken",
	)
	logger.Debug("Deleting refresh token")
	startTime := time.Now()

	query := `
		DELETE FROM refresh_tokens
		WHERE token = $1
		RETURNING id
	`
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, token).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Refresh token not found", "error", "TokenError")
			return &userErrs.TokenError{Message: "refresh token not found"}
		}
		logger.Error("Failed to delete refresh token", "error", err.Error())
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully deleted refresh token",
		"token_id", id.String(),
		"duration_ms", duration.Milliseconds())

	return nil
}

// DeleteUserTokens deletes all tokens for a user
func (r *tokenRepository) DeleteUserTokens(ctx context.Context, userID uuid.UUID) error {
	logger := middleware.Logger.With(
		"method", "TokenRepository.DeleteUserTokens",
		"user_id", userID.String(),
	)
	logger.Debug("Deleting all user tokens")
	startTime := time.Now()

	query := `
		DELETE FROM refresh_tokens
		WHERE user_id = $1
	`
	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		logger.Error("Failed to delete user tokens", "error", err.Error())
		return fmt.Errorf("failed to delete user tokens: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err.Error())
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully deleted user tokens",
		"tokens_deleted", rowsAffected,
		"duration_ms", duration.Milliseconds())

	return nil
}
