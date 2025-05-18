package repo

import (
	"context"

	"github.com/fanzru/e-commerce-be/internal/app/promotion/domain/entity"
	"github.com/google/uuid"
)

// PromotionRepository defines the interface for promotion repository
type PromotionRepository interface {
	// GetByID retrieves a promotion by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Promotion, error)

	// List retrieves a list of promotions with pagination and filtering
	List(ctx context.Context, page, limit int, active *bool) ([]*entity.Promotion, int, error)

	// Create creates a new promotion
	Create(ctx context.Context, promotion *entity.Promotion) error

	// UpdateStatus updates a promotion's active status
	UpdateStatus(ctx context.Context, id uuid.UUID, active bool) error

	// Delete deletes a promotion by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// GetByType retrieves promotions by type
	GetByType(ctx context.Context, promotionType entity.PromotionType) ([]*entity.Promotion, error)

	// GetActive retrieves all active promotions
	GetActive(ctx context.Context) ([]*entity.Promotion, error)
}
