package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/user/domain/entity"
	userErrs "github.com/fanzru/e-commerce-be/internal/app/user/domain/errs"
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
	// Check if email already exists
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", user.Email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}

	if count > 0 {
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
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `
		SELECT id, email, password, name, role, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.scanUser(ctx, query, id)
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, email, password, name, role, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`
	return r.scanUser(ctx, query, email)
}

// Update updates a user's details
func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	query := `
		UPDATE users
		SET name = $1, password = $2, role = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query,
		user.Name, user.Password, user.Role, time.Now(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return &userErrs.UserNotFoundError{ID: user.ID.String()}
	}

	return nil
}

// Delete soft-deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return &userErrs.UserNotFoundError{ID: id.String()}
	}

	return nil
}

// List lists users with pagination and filters
func (r *userRepository) List(ctx context.Context, page, limit int, role *entity.UserRole) ([]*entity.User, int, error) {
	// Calculate offset
	offset := (page - 1) * limit

	// Build query
	countQuery := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	query := `
		SELECT id, email, password, name, role, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
	`

	// Add role filter if provided
	var args []interface{}
	if role != nil {
		countQuery += ` AND role = $1`
		query += ` AND role = $1`
		args = append(args, *role)
	}

	// Add pagination
	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, limit, offset)

	// Get total count
	var total int
	var err error
	if role != nil {
		err = r.db.QueryRowContext(ctx, countQuery, *role).Scan(&total)
	} else {
		err = r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get users count: %w", err)
	}

	// Execute query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	// Scan results
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
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}

		if deletedAt.Valid {
			user.DeletedAt = &deletedAt.Time
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, total, nil
}

// scanUser scans a single user from a query
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
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves a refresh token
func (r *tokenRepository) GetRefreshToken(ctx context.Context, tokenStr string) (*entity.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1
	`
	row := r.db.QueryRowContext(ctx, query, tokenStr)

	token := &entity.RefreshToken{}
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &userErrs.TokenError{Message: "refresh token not found"}
		}
		return nil, fmt.Errorf("failed to scan refresh token: %w", err)
	}

	return token, nil
}

// DeleteRefreshToken deletes a refresh token
func (r *tokenRepository) DeleteRefreshToken(ctx context.Context, tokenStr string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	_, err := r.db.ExecContext(ctx, query, tokenStr)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

// DeleteUserTokens deletes all tokens for a user
func (r *tokenRepository) DeleteUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user tokens: %w", err)
	}

	return nil
}
