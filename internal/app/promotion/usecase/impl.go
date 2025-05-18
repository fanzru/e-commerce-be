package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cartEntity "github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	promotionEntity "github.com/fanzru/e-commerce-be/internal/app/promotion/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/promotion/repo"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/google/uuid"
)

// Ensure promotionUseCase implements PromotionUseCase
var _ PromotionUseCase = (*promotionUseCase)(nil)

// promotionUseCase implements the PromotionUseCase interface
type promotionUseCase struct {
	repo repo.PromotionRepository
}

// NewPromotionUseCase creates a new instance of promotionUseCase
func NewPromotionUseCase(repo repo.PromotionRepository) PromotionUseCase {
	return &promotionUseCase{
		repo: repo,
	}
}

// GetByID retrieves a promotion by its ID
func (u *promotionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*promotionEntity.Promotion, error) {
	logger := middleware.Logger.With(
		"method", "PromotionUseCase.GetByID",
		"promotion_id", id.String(),
	)
	logger.Info("Getting promotion by ID")
	startTime := time.Now()

	if id == uuid.Nil {
		logger.Warn("Invalid promotion ID", "error", "ErrInvalidInput")
		return nil, errors.New("invalid promotion ID")
	}

	promotion, err := u.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get promotion", "error", err.Error())
		return nil, fmt.Errorf("error getting promotion: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved promotion",
		"type", promotion.Type,
		"active", promotion.Active,
		"duration_ms", duration.Milliseconds())

	return promotion, nil
}

// List retrieves a list of promotions with pagination and filtering
func (u *promotionUseCase) List(ctx context.Context, page, limit int, active *bool) ([]*promotionEntity.Promotion, int, error) {
	logger := middleware.Logger.With(
		"method", "PromotionUseCase.List",
		"page", page,
		"limit", limit,
	)
	if active != nil {
		logger = logger.With("active", *active)
	}
	logger.Info("Listing promotions with filters")
	startTime := time.Now()

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	promotions, total, err := u.repo.List(ctx, page, limit, active)
	if err != nil {
		logger.Error("Failed to list promotions", "error", err.Error())
		return nil, 0, fmt.Errorf("error listing promotions: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully listed promotions",
		"total", total,
		"returned", len(promotions),
		"duration_ms", duration.Milliseconds())

	return promotions, total, nil
}

// CreateBuyOneGetOneFree creates a new BuyOneGetOneFree promotion
func (u *promotionUseCase) CreateBuyOneGetOneFree(
	ctx context.Context,
	description string,
	triggerSKU string,
	freeSKU string,
	triggerQuantity int,
	freeQuantity int,
	active bool,
) (*promotionEntity.Promotion, error) {
	logger := middleware.Logger.With(
		"method", "PromotionUseCase.CreateBuyOneGetOneFree",
		"trigger_sku", triggerSKU,
		"free_sku", freeSKU,
		"trigger_quantity", triggerQuantity,
		"free_quantity", freeQuantity,
	)
	logger.Info("Creating BuyOneGetOneFree promotion")
	startTime := time.Now()

	if description == "" {
		logger.Warn("Invalid input: Empty description", "error", "ErrInvalidInput")
		return nil, errors.New("description is required")
	}
	if triggerSKU == "" {
		logger.Warn("Invalid input: Empty trigger SKU", "error", "ErrInvalidInput")
		return nil, errors.New("trigger SKU is required")
	}
	if freeSKU == "" {
		logger.Warn("Invalid input: Empty free SKU", "error", "ErrInvalidInput")
		return nil, errors.New("free SKU is required")
	}
	if triggerQuantity < 1 {
		logger.Warn("Invalid input: Invalid trigger quantity", "error", "ErrInvalidInput")
		return nil, errors.New("trigger quantity must be greater than zero")
	}
	if freeQuantity < 1 {
		logger.Warn("Invalid input: Invalid free quantity", "error", "ErrInvalidInput")
		return nil, errors.New("free quantity must be greater than zero")
	}

	rule := promotionEntity.BuyOneGetOneFreePromotion{
		TriggerSKU:      triggerSKU,
		FreeSKU:         freeSKU,
		TriggerQuantity: triggerQuantity,
		FreeQuantity:    freeQuantity,
	}

	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		logger.Error("Failed to marshal promotion rule", "error", err.Error())
		return nil, fmt.Errorf("failed to marshal promotion rule: %w", err)
	}

	promotion := &promotionEntity.Promotion{
		ID:          uuid.New(),
		Type:        promotionEntity.BuyOneGetOneFree,
		Description: description,
		Rule:        ruleJSON,
		Active:      active,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = u.repo.Create(ctx, promotion)
	if err != nil {
		logger.Error("Failed to create promotion", "error", err.Error())
		return nil, fmt.Errorf("failed to create promotion: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully created BuyOneGetOneFree promotion",
		"promotion_id", promotion.ID.String(),
		"duration_ms", duration.Milliseconds())

	return promotion, nil
}

// CreateBuy3Pay2 creates a new Buy3Pay2 promotion
func (u *promotionUseCase) CreateBuy3Pay2(
	ctx context.Context,
	description string,
	sku string,
	minQuantity int,
	paidQuantityDivisor int,
	freeQuantityDivisor int,
	active bool,
) (*promotionEntity.Promotion, error) {
	logger := middleware.Logger.With(
		"method", "PromotionUseCase.CreateBuy3Pay2",
		"sku", sku,
		"min_quantity", minQuantity,
		"paid_quantity_divisor", paidQuantityDivisor,
		"free_quantity_divisor", freeQuantityDivisor,
	)
	logger.Info("Creating Buy3Pay2 promotion")
	startTime := time.Now()

	if description == "" {
		logger.Warn("Invalid input: Empty description", "error", "ErrInvalidInput")
		return nil, errors.New("description is required")
	}
	if sku == "" {
		logger.Warn("Invalid input: Empty SKU", "error", "ErrInvalidInput")
		return nil, errors.New("SKU is required")
	}
	if minQuantity < 1 {
		logger.Warn("Invalid input: Invalid minimum quantity", "error", "ErrInvalidInput")
		return nil, errors.New("minimum quantity must be greater than zero")
	}
	if paidQuantityDivisor < 1 {
		logger.Warn("Invalid input: Invalid paid quantity divisor", "error", "ErrInvalidInput")
		return nil, errors.New("paid quantity divisor must be greater than zero")
	}
	if freeQuantityDivisor < 1 {
		logger.Warn("Invalid input: Invalid free quantity divisor", "error", "ErrInvalidInput")
		return nil, errors.New("free quantity divisor must be greater than zero")
	}

	rule := promotionEntity.Buy3Pay2Promotion{
		SKU:                 sku,
		MinQuantity:         minQuantity,
		PaidQuantityDivisor: paidQuantityDivisor,
		FreeQuantityDivisor: freeQuantityDivisor,
	}

	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		logger.Error("Failed to marshal promotion rule", "error", err.Error())
		return nil, fmt.Errorf("failed to marshal promotion rule: %w", err)
	}

	promotion := &promotionEntity.Promotion{
		ID:          uuid.New(),
		Type:        promotionEntity.Buy3Pay2,
		Description: description,
		Rule:        ruleJSON,
		Active:      active,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = u.repo.Create(ctx, promotion)
	if err != nil {
		logger.Error("Failed to create promotion", "error", err.Error())
		return nil, fmt.Errorf("failed to create promotion: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully created Buy3Pay2 promotion",
		"promotion_id", promotion.ID.String(),
		"duration_ms", duration.Milliseconds())

	return promotion, nil
}

// CreateBulkDiscount creates a new BulkDiscount promotion
func (u *promotionUseCase) CreateBulkDiscount(
	ctx context.Context,
	description string,
	sku string,
	minQuantity int,
	discountPercentage float64,
	active bool,
) (*promotionEntity.Promotion, error) {
	logger := middleware.Logger.With(
		"method", "PromotionUseCase.CreateBulkDiscount",
		"sku", sku,
		"min_quantity", minQuantity,
		"discount_percentage", discountPercentage,
	)
	logger.Info("Creating BulkDiscount promotion")
	startTime := time.Now()

	if description == "" {
		logger.Warn("Invalid input: Empty description", "error", "ErrInvalidInput")
		return nil, errors.New("description is required")
	}
	if sku == "" {
		logger.Warn("Invalid input: Empty SKU", "error", "ErrInvalidInput")
		return nil, errors.New("SKU is required")
	}
	if minQuantity < 1 {
		logger.Warn("Invalid input: Invalid minimum quantity", "error", "ErrInvalidInput")
		return nil, errors.New("minimum quantity must be greater than zero")
	}
	if discountPercentage <= 0 || discountPercentage > 100 {
		logger.Warn("Invalid input: Invalid discount percentage", "error", "ErrInvalidInput")
		return nil, errors.New("discount percentage must be between 0 and 100")
	}

	rule := promotionEntity.BulkDiscountPromotion{
		SKU:                sku,
		MinQuantity:        minQuantity,
		DiscountPercentage: discountPercentage,
	}

	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		logger.Error("Failed to marshal promotion rule", "error", err.Error())
		return nil, fmt.Errorf("failed to marshal promotion rule: %w", err)
	}

	promotion := &promotionEntity.Promotion{
		ID:          uuid.New(),
		Type:        promotionEntity.BulkDiscount,
		Description: description,
		Rule:        ruleJSON,
		Active:      active,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = u.repo.Create(ctx, promotion)
	if err != nil {
		logger.Error("Failed to create promotion", "error", err.Error())
		return nil, fmt.Errorf("failed to create promotion: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully created BulkDiscount promotion",
		"promotion_id", promotion.ID.String(),
		"duration_ms", duration.Milliseconds())

	return promotion, nil
}

// UpdateStatus updates a promotion's active status
func (u *promotionUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, active bool) error {
	logger := middleware.Logger.With(
		"method", "PromotionUseCase.UpdateStatus",
		"promotion_id", id.String(),
		"active", active,
	)
	logger.Info("Updating promotion status")
	startTime := time.Now()

	if id == uuid.Nil {
		logger.Warn("Invalid promotion ID", "error", "ErrInvalidInput")
		return errors.New("invalid promotion ID")
	}

	err := u.repo.UpdateStatus(ctx, id, active)
	if err != nil {
		logger.Error("Failed to update promotion status", "error", err.Error())
		return fmt.Errorf("error updating promotion status: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully updated promotion status",
		"duration_ms", duration.Milliseconds())

	return nil
}

// Delete deletes a promotion
func (u *promotionUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	logger := middleware.Logger.With(
		"method", "PromotionUseCase.Delete",
		"promotion_id", id.String(),
	)
	logger.Info("Deleting promotion")
	startTime := time.Now()

	if id == uuid.Nil {
		logger.Warn("Invalid promotion ID", "error", "ErrInvalidInput")
		return errors.New("invalid promotion ID")
	}

	err := u.repo.Delete(ctx, id)
	if err != nil {
		logger.Error("Failed to delete promotion", "error", err.Error())
		return fmt.Errorf("error deleting promotion: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully deleted promotion",
		"duration_ms", duration.Milliseconds())

	return nil
}

// ApplyPromotions applies promotions to a cart and returns the discounts
func (u *promotionUseCase) ApplyPromotions(ctx context.Context, cart *cartEntity.CartInfo) ([]PromotionDiscount, float64, error) {
	logger := middleware.Logger.With(
		"method", "PromotionUseCase.ApplyPromotions",
		"user_id", cart.UserID.String(),
	)
	logger.Info("Applying promotions to cart")
	startTime := time.Now()

	if cart == nil || len(cart.Items) == 0 {
		logger.Info("Cart is empty, no promotions applied")
		return nil, 0, nil
	}

	// Get all active promotions
	promotions, err := u.repo.GetActive(ctx)
	if err != nil {
		logger.Error("Failed to get active promotions", "error", err.Error())
		return nil, 0, fmt.Errorf("failed to get active promotions: %w", err)
	}

	if len(promotions) == 0 {
		logger.Info("No active promotions found")
		return nil, 0, nil
	}

	// Convert cart items to promotion cart items
	promotionItems := promotionEntity.ConvertCartToPromotionItems(cart.Items)

	// Get applicable promotions
	applicablePromotions := promotionEntity.GetApplicablePromotions(promotions, promotionItems)

	// Convert to PromotionDiscount format for response
	discounts := make([]PromotionDiscount, 0, len(applicablePromotions))
	for _, promo := range applicablePromotions {
		discounts = append(discounts, PromotionDiscount{
			PromotionID:   promo.ID,
			PromotionType: string(promo.Type),
			Description:   promo.Description,
			Discount:      promo.Discount,
		})
	}

	// Calculate total discount
	totalDiscount := promotionEntity.CalculateTotalDiscount(applicablePromotions)

	duration := time.Since(startTime)
	logger.Info("Successfully processed promotions for cart",
		"applicable_promotions", len(discounts),
		"active_promotions", len(discounts)-0, // Count of promotions with actual discounts
		"total_discount", totalDiscount,
		"duration_ms", duration.Milliseconds())

	return discounts, totalDiscount, nil
}
