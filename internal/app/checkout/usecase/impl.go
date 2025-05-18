package usecase

import (
	"context"
	"errors"
	"fmt"

	cartRepo "github.com/fanzru/e-commerce-be/internal/app/cart/repo"
	checkoutEntity "github.com/fanzru/e-commerce-be/internal/app/checkout/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/checkout/repo"
	promotionUseCase "github.com/fanzru/e-commerce-be/internal/app/promotion/usecase"
	"github.com/google/uuid"
)

// Ensure checkoutUseCase implements CheckoutUseCase
var _ CheckoutUseCase = (*checkoutUseCase)(nil)

// checkoutUseCase implements the CheckoutUseCase interface
type checkoutUseCase struct {
	repo             repo.CheckoutRepository
	cartRepo         cartRepo.CartRepository
	promotionUseCase promotionUseCase.PromotionUseCase
}

// NewCheckoutUseCase creates a new instance of checkoutUseCase
func NewCheckoutUseCase(
	repo repo.CheckoutRepository,
	cartRepo cartRepo.CartRepository,
	promotionUseCase promotionUseCase.PromotionUseCase,
) CheckoutUseCase {
	return &checkoutUseCase{
		repo:             repo,
		cartRepo:         cartRepo,
		promotionUseCase: promotionUseCase,
	}
}

// GetByID retrieves a checkout by its ID
func (u *checkoutUseCase) GetByID(ctx context.Context, id uuid.UUID) (*checkoutEntity.Checkout, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid checkout ID")
	}
	return u.repo.GetByID(ctx, id)
}

// ProcessCart processes a cart and creates a checkout
func (u *checkoutUseCase) ProcessCart(ctx context.Context, cartID uuid.UUID) (*checkoutEntity.Checkout, error) {
	if cartID == uuid.Nil {
		return nil, errors.New("invalid cart ID")
	}

	// Check if cart already has a checkout
	existingCheckout, err := u.repo.GetByCartID(ctx, cartID)
	if err == nil && existingCheckout != nil {
		return nil, errors.New("cart has already been checked out")
	}

	// Get the cart
	cart, err := u.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	if cart.IsEmpty() {
		return nil, errors.New("cannot checkout an empty cart")
	}

	// Apply promotions to the cart
	promotionDiscounts, totalDiscount, err := u.promotionUseCase.ApplyPromotions(ctx, cart)
	if err != nil {
		return nil, fmt.Errorf("failed to apply promotions: %w", err)
	}

	// Calculate cart subtotal
	subtotal := cart.Subtotal()
	total := subtotal - totalDiscount

	// Create checkout items
	checkoutItems := make([]*checkoutEntity.CheckoutItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		// Calculate item discount (proportional to the total discount)
		itemSubtotal := item.UnitPrice * float64(item.Quantity)
		itemDiscountRatio := 0.0
		if subtotal > 0 {
			itemDiscountRatio = itemSubtotal / subtotal
		}
		itemDiscount := totalDiscount * itemDiscountRatio
		itemTotal := itemSubtotal - itemDiscount

		checkoutItem := &checkoutEntity.CheckoutItem{
			ID:          uuid.New(),
			ProductID:   item.ProductID,
			ProductSKU:  item.ProductSKU,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    itemSubtotal,
			Discount:    itemDiscount,
			Total:       itemTotal,
		}
		checkoutItems = append(checkoutItems, checkoutItem)
	}

	// Create the checkout
	checkout := &checkoutEntity.Checkout{
		ID:            uuid.New(),
		CartID:        cartID,
		Items:         checkoutItems,
		Subtotal:      subtotal,
		TotalDiscount: totalDiscount,
		Total:         total,
	}

	// Add promotions to the checkout
	promotionApplied := make([]*checkoutEntity.PromotionApplied, 0, len(promotionDiscounts))
	for _, pd := range promotionDiscounts {
		promotion := &checkoutEntity.PromotionApplied{
			ID:          uuid.New(),
			CheckoutID:  checkout.ID,
			PromotionID: pd.PromotionID,
			Description: pd.Description,
			Discount:    pd.Discount,
		}
		promotionApplied = append(promotionApplied, promotion)
	}
	checkout.Promotions = promotionApplied

	// Set checkout items' checkout ID
	for _, item := range checkout.Items {
		item.CheckoutID = checkout.ID
	}

	// Create the checkout in the repository
	err = u.repo.Create(ctx, checkout)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout: %w", err)
	}

	return checkout, nil
}

// ListCheckouts retrieves a list of checkouts with pagination
func (u *checkoutUseCase) ListCheckouts(ctx context.Context, page, limit int) ([]*checkoutEntity.Checkout, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return u.repo.ListCheckouts(ctx, page, limit)
}
