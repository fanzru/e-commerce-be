package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	cartEntity "github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	promotionEntity "github.com/fanzru/e-commerce-be/internal/app/promotion/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/promotion/repo"
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
	if id == uuid.Nil {
		return nil, errors.New("invalid promotion ID")
	}
	return u.repo.GetByID(ctx, id)
}

// List retrieves a list of promotions with pagination and filtering
func (u *promotionUseCase) List(ctx context.Context, page, limit int, active *bool) ([]*promotionEntity.Promotion, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return u.repo.List(ctx, page, limit, active)
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
	if description == "" {
		return nil, errors.New("description is required")
	}
	if triggerSKU == "" {
		return nil, errors.New("trigger SKU is required")
	}
	if freeSKU == "" {
		return nil, errors.New("free SKU is required")
	}
	if triggerQuantity < 1 {
		return nil, errors.New("trigger quantity must be greater than zero")
	}
	if freeQuantity < 1 {
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
		return nil, fmt.Errorf("failed to marshal promotion rule: %w", err)
	}

	promotion := &promotionEntity.Promotion{
		ID:          uuid.New(),
		Type:        promotionEntity.BuyOneGetOneFree,
		Description: description,
		Rule:        ruleJSON,
		Active:      active,
	}

	err = u.repo.Create(ctx, promotion)
	if err != nil {
		return nil, fmt.Errorf("failed to create promotion: %w", err)
	}

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
	if description == "" {
		return nil, errors.New("description is required")
	}
	if sku == "" {
		return nil, errors.New("SKU is required")
	}
	if minQuantity < 1 {
		return nil, errors.New("minimum quantity must be greater than zero")
	}
	if paidQuantityDivisor < 1 {
		return nil, errors.New("paid quantity divisor must be greater than zero")
	}
	if freeQuantityDivisor < 1 {
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
		return nil, fmt.Errorf("failed to marshal promotion rule: %w", err)
	}

	promotion := &promotionEntity.Promotion{
		ID:          uuid.New(),
		Type:        promotionEntity.Buy3Pay2,
		Description: description,
		Rule:        ruleJSON,
		Active:      active,
	}

	err = u.repo.Create(ctx, promotion)
	if err != nil {
		return nil, fmt.Errorf("failed to create promotion: %w", err)
	}

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
	if description == "" {
		return nil, errors.New("description is required")
	}
	if sku == "" {
		return nil, errors.New("SKU is required")
	}
	if minQuantity < 1 {
		return nil, errors.New("minimum quantity must be greater than zero")
	}
	if discountPercentage <= 0 || discountPercentage > 100 {
		return nil, errors.New("discount percentage must be between 0 and 100")
	}

	rule := promotionEntity.BulkDiscountPromotion{
		SKU:                sku,
		MinQuantity:        minQuantity,
		DiscountPercentage: discountPercentage,
	}

	ruleJSON, err := json.Marshal(rule)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal promotion rule: %w", err)
	}

	promotion := &promotionEntity.Promotion{
		ID:          uuid.New(),
		Type:        promotionEntity.BulkDiscount,
		Description: description,
		Rule:        ruleJSON,
		Active:      active,
	}

	err = u.repo.Create(ctx, promotion)
	if err != nil {
		return nil, fmt.Errorf("failed to create promotion: %w", err)
	}

	return promotion, nil
}

// UpdateStatus updates a promotion's active status
func (u *promotionUseCase) UpdateStatus(ctx context.Context, id uuid.UUID, active bool) error {
	if id == uuid.Nil {
		return errors.New("invalid promotion ID")
	}
	return u.repo.UpdateStatus(ctx, id, active)
}

// Delete deletes a promotion
func (u *promotionUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid promotion ID")
	}
	return u.repo.Delete(ctx, id)
}

// ApplyPromotions applies promotions to a cart and returns the discounts
func (u *promotionUseCase) ApplyPromotions(ctx context.Context, cart *cartEntity.Cart) ([]PromotionDiscount, float64, error) {
	if cart == nil || cart.IsEmpty() {
		return nil, 0, nil
	}

	// Get all active promotions
	promotions, err := u.repo.GetActive(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get active promotions: %w", err)
	}

	if len(promotions) == 0 {
		return nil, 0, nil
	}

	// Convert cart items to promotion cart items
	promotionItems := make([]promotionEntity.CartItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		promotionItems = append(promotionItems, promotionEntity.CartItem{
			ProductID:   item.ProductID,
			ProductSKU:  item.ProductSKU,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}

	// Apply promotions
	discounts := make([]PromotionDiscount, 0, len(promotions))
	totalDiscount := 0.0

	for _, promotion := range promotions {
		discount, err := promotion.ApplyToCart(promotionItems)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to apply promotion: %w", err)
		}

		if discount > 0 {
			discounts = append(discounts, PromotionDiscount{
				PromotionID:   promotion.ID,
				PromotionType: string(promotion.Type),
				Description:   promotion.Description,
				Discount:      discount,
			})
			totalDiscount += discount
		}
	}

	return discounts, totalDiscount, nil
}
