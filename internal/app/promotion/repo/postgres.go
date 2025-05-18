package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fanzru/e-commerce-be/internal/app/promotion/domain/entity"
	domainErrors "github.com/fanzru/e-commerce-be/internal/app/promotion/domain/errs"
	"github.com/google/uuid"
)

// PromotionPostgresRepository implements PromotionRepository using PostgreSQL
type PromotionPostgresRepository struct {
	db *sql.DB
}

// NewPromotionRepository creates a new promotion repository
func NewPromotionRepository(db *sql.DB) PromotionRepository {
	return &PromotionPostgresRepository{
		db: db,
	}
}

// GetByID retrieves a promotion by its ID
func (r *PromotionPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Promotion, error) {
	query := `
		SELECT id, type, description, active, created_at, updated_at
		FROM promotions
		WHERE id = $1 AND deleted_at IS NULL
	`

	var promotion entity.Promotion
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&promotion.ID,
		&promotion.Type,
		&promotion.Description,
		&promotion.Active,
		&promotion.CreatedAt,
		&promotion.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrPromotionNotFound
		}
		return nil, fmt.Errorf("error querying promotion by ID: %w", err)
	}

	return &promotion, nil
}

// List retrieves a list of promotions with pagination and filtering
func (r *PromotionPostgresRepository) List(ctx context.Context, page, limit int, active *bool) ([]*entity.Promotion, int, error) {
	offset := (page - 1) * limit

	// Base query for filtering
	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argPos := 1

	// Add active filter if provided
	if active != nil {
		whereClause += fmt.Sprintf(" AND active = $%d", argPos)
		args = append(args, *active)
		argPos++
	}

	// Count total matches first
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM promotions %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting promotions: %w", err)
	}

	// Now fetch the actual data with pagination
	query := fmt.Sprintf(`
		SELECT id, type, description, active, created_at, updated_at
		FROM promotions
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying promotions: %w", err)
	}
	defer rows.Close()

	promotions := []*entity.Promotion{}
	for rows.Next() {
		var promotion entity.Promotion
		err := rows.Scan(
			&promotion.ID,
			&promotion.Type,
			&promotion.Description,
			&promotion.Active,
			&promotion.CreatedAt,
			&promotion.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning promotion row: %w", err)
		}
		promotions = append(promotions, &promotion)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating promotion rows: %w", err)
	}

	return promotions, total, nil
}

// Create creates a new promotion
func (r *PromotionPostgresRepository) Create(ctx context.Context, promotion *entity.Promotion) error {
	// Generate a new UUID if not provided
	if promotion.ID == uuid.Nil {
		promotion.ID = uuid.New()
	}

	query := `
		INSERT INTO promotions (id, type, description, active)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		promotion.ID,
		promotion.Type,
		promotion.Description,
		promotion.Active,
	).Scan(
		&promotion.CreatedAt,
		&promotion.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("error creating promotion: %w", err)
	}

	return nil
}

// UpdateStatus updates a promotion's active status
func (r *PromotionPostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, active bool) error {
	query := `
		UPDATE promotions
		SET active = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id
	`

	var promotionID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, active, id).Scan(&promotionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainErrors.ErrPromotionNotFound
		}
		return fmt.Errorf("error updating promotion status: %w", err)
	}

	return nil
}

// Delete deletes a promotion by its ID (soft delete)
func (r *PromotionPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE promotions
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`

	var promotionID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, id).Scan(&promotionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainErrors.ErrPromotionNotFound
		}
		return fmt.Errorf("error deleting promotion: %w", err)
	}

	return nil
}

// GetByType retrieves promotions by type
func (r *PromotionPostgresRepository) GetByType(ctx context.Context, promotionType entity.PromotionType) ([]*entity.Promotion, error) {
	query := `
		SELECT id, type, description, active, created_at, updated_at
		FROM promotions
		WHERE type = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, promotionType)
	if err != nil {
		return nil, fmt.Errorf("error querying promotions by type: %w", err)
	}
	defer rows.Close()

	promotions := []*entity.Promotion{}
	for rows.Next() {
		var promotion entity.Promotion
		err := rows.Scan(
			&promotion.ID,
			&promotion.Type,
			&promotion.Description,
			&promotion.Active,
			&promotion.CreatedAt,
			&promotion.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning promotion row: %w", err)
		}
		promotions = append(promotions, &promotion)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating promotion rows: %w", err)
	}

	return promotions, nil
}

// GetActive retrieves all active promotions
func (r *PromotionPostgresRepository) GetActive(ctx context.Context) ([]*entity.Promotion, error) {
	query := `
		SELECT id, type, description, active, created_at, updated_at
		FROM promotions
		WHERE active = true AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying active promotions: %w", err)
	}
	defer rows.Close()

	promotions := []*entity.Promotion{}
	for rows.Next() {
		var promotion entity.Promotion
		err := rows.Scan(
			&promotion.ID,
			&promotion.Type,
			&promotion.Description,
			&promotion.Active,
			&promotion.CreatedAt,
			&promotion.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning promotion row: %w", err)
		}
		promotions = append(promotions, &promotion)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating promotion rows: %w", err)
	}

	return promotions, nil
}
