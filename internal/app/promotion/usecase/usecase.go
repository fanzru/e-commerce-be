package usecase

import (
	"context"

	cartEntity "github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	promotionEntity "github.com/fanzru/e-commerce-be/internal/app/promotion/domain/entity"
	"github.com/google/uuid"
)

// PromotionDiscount represents a promotion applied to a cart with discount amount
type PromotionDiscount struct {
	PromotionID   uuid.UUID `json:"promotion_id"`
	PromotionType string    `json:"promotion_type"`
	Description   string    `json:"description"`
	Discount      float64   `json:"discount"`
}

// PromotionUseCase defines the interface for promotion use cases
type PromotionUseCase interface {
	// GetByID retrieves a promotion by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*promotionEntity.Promotion, error)

	// List retrieves a list of promotions with pagination and filtering
	List(ctx context.Context, page, limit int, active *bool) ([]*promotionEntity.Promotion, int, error)

	// Create creates a new promotion
	CreateBuyOneGetOneFree(
		ctx context.Context,
		description string,
		triggerSKU string,
		freeSKU string,
		triggerQuantity int,
		freeQuantity int,
		active bool,
	) (*promotionEntity.Promotion, error)

	// Create creates a new Buy3Pay2 promotion
	CreateBuy3Pay2(
		ctx context.Context,
		description string,
		sku string,
		minQuantity int,
		paidQuantityDivisor int,
		freeQuantityDivisor int,
		active bool,
	) (*promotionEntity.Promotion, error)

	// Create creates a new BulkDiscount promotion
	CreateBulkDiscount(
		ctx context.Context,
		description string,
		sku string,
		minQuantity int,
		discountPercentage float64,
		active bool,
	) (*promotionEntity.Promotion, error)

	// UpdateStatus updates a promotion's active status
	UpdateStatus(ctx context.Context, id uuid.UUID, active bool) error

	// Delete deletes a promotion
	Delete(ctx context.Context, id uuid.UUID) error

	// ApplyPromotions applies promotions to a cart and returns the discounts
	ApplyPromotions(ctx context.Context, cart *cartEntity.Cart) ([]PromotionDiscount, float64, error)
}
