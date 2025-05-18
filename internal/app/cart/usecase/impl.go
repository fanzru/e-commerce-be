package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	cartEntity "github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	domainErrors "github.com/fanzru/e-commerce-be/internal/app/cart/domain/errs"
	cartRepo "github.com/fanzru/e-commerce-be/internal/app/cart/repo"
	productRepo "github.com/fanzru/e-commerce-be/internal/app/product/repo"
	promotionUseCase "github.com/fanzru/e-commerce-be/internal/app/promotion/usecase"
	"github.com/fanzru/e-commerce-be/internal/common/errs"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/google/uuid"
)

// Ensure cartUseCase implements CartUseCase
var _ CartUseCase = (*cartUseCase)(nil)

// cartUseCase implements the CartUseCase interface
type cartUseCase struct {
	cartRepo         cartRepo.CartRepository
	productRepo      productRepo.ProductRepository
	promotionUseCase promotionUseCase.PromotionUseCase
}

// NewCartUseCase creates a new instance of cartUseCase
func NewCartUseCase(
	cartRepo cartRepo.CartRepository,
	productRepo productRepo.ProductRepository,
	promotionUseCase promotionUseCase.PromotionUseCase,
) CartUseCase {
	return &cartUseCase{
		cartRepo:         cartRepo,
		productRepo:      productRepo,
		promotionUseCase: promotionUseCase,
	}
}

// GetUserCart retrieves the cart for a user
func (u *cartUseCase) GetUserCart(ctx context.Context, userID uuid.UUID) (*cartEntity.Cart, error) {
	logger := middleware.Logger.With(
		"method", "CartUseCase.GetUserCart",
		"user_id", userID.String(),
	)
	logger.Info("Retrieving user cart")
	startTime := time.Now()

	if userID == uuid.Nil {
		logger.Warn("Invalid user ID")
		return nil, errors.New("invalid user ID")
	}

	cart, err := u.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get cart", "error", err.Error())
		return nil, err
	}

	itemCount := len(cart.Items)
	duration := time.Since(startTime)
	logger.Info("Successfully retrieved cart",
		"item_count", itemCount,
		"subtotal", cart.Subtotal(),
		"duration_ms", duration.Milliseconds())

	return cart, nil
}

// AddItemToUserCart adds a product to a user's cart
func (u *cartUseCase) AddItemToUserCart(ctx context.Context, userID, productID uuid.UUID, quantity int) (*cartEntity.CartItem, error) {
	logger := middleware.Logger.With(
		"method", "CartUseCase.AddItemToUserCart",
		"user_id", userID.String(),
		"product_id", productID.String(),
		"quantity", quantity,
	)
	logger.Info("Adding item to user cart")
	startTime := time.Now()

	if userID == uuid.Nil {
		logger.Warn("Invalid user ID")
		return nil, errors.New("invalid user ID")
	}
	if productID == uuid.Nil {
		logger.Warn("Invalid product ID")
		return nil, errors.New("invalid product ID")
	}
	if quantity <= 0 {
		logger.Warn("Invalid quantity", "quantity", quantity)
		return nil, errors.New("quantity must be greater than zero")
	}

	// Check if the user already has this product in cart
	logger.Debug("Checking if user already has this product in cart")
	existingItem, err := u.cartRepo.GetItemByProductID(ctx, userID, productID)
	if err != nil && !errors.Is(err, domainErrors.ErrItemNotFound) {
		logger.Error("Failed to check for existing item", "error", err.Error())
		return nil, fmt.Errorf("failed to check for existing item: %w", err)
	}

	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		logger.Error("Failed to get product", "error", err.Error())
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if existingItem != nil {

		// Check if there's enough inventory
		if !product.HasEnoughInventory(existingItem.Quantity + quantity) {
			logger.Warn("Not enough inventory",
				"product_id", productID.String(),
				"requested", quantity,
				"available", product.Inventory)
			return nil, errs.New(nil, errs.CodeOutOfStock, 400, "Insufficient stock")
		}

		// Update existing item quantity
		logger.Debug("Product already in cart, updating quantity", "existing_quantity", existingItem.Quantity)
		err = u.cartRepo.UpdateItem(ctx, existingItem.ID, existingItem.Quantity+quantity)
		if err != nil {
			logger.Error("Failed to update existing cart item", "error", err.Error())
			return nil, fmt.Errorf("failed to update existing cart item: %w", err)
		}

		// Get the updated item
		updatedItem, err := u.cartRepo.GetItemByProductID(ctx, userID, productID)
		if err != nil {
			logger.Error("Failed to get updated cart item", "error", err.Error())
			return nil, fmt.Errorf("failed to get updated cart item: %w", err)
		}
		return updatedItem, nil
	}

	// Check if there's enough inventory
	if !product.HasEnoughInventory(quantity) {
		logger.Warn("Not enough inventory",
			"product_id", productID.String(),
			"requested", quantity,
			"available", product.Inventory)
		return nil, errs.New(nil, errs.CodeOutOfStock, 400, "Insufficient stock")
	}

	// Create a new cart item
	item := &cartEntity.CartItem{
		ID:        uuid.New(),
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add the item to the cart
	err = u.cartRepo.AddItem(ctx, item)
	if err != nil {
		logger.Error("Failed to add item to cart", "error", err.Error())
		return nil, fmt.Errorf("failed to add item to cart: %w", err)
	}

	// Get the updated item
	updatedItem, err := u.cartRepo.GetItemByProductID(ctx, userID, productID)
	if err != nil {
		logger.Error("Failed to get updated cart item", "error", err.Error())
		return nil, fmt.Errorf("failed to get updated cart item: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully added item to user cart",
		"item_id", updatedItem.ID.String(),
		"product_id", updatedItem.ProductID,
		"quantity", updatedItem.Quantity,
		"duration_ms", duration.Milliseconds())

	return updatedItem, nil
}

// UpdateItemQuantity updates the quantity of a cart item
func (u *cartUseCase) UpdateItemQuantity(ctx context.Context, userID, itemID uuid.UUID, quantity int) error {
	logger := middleware.Logger.With(
		"method", "CartUseCase.UpdateItemQuantity",
		"user_id", userID.String(),
		"item_id", itemID.String(),
		"quantity", quantity,
	)
	logger.Info("Updating item quantity")
	startTime := time.Now()

	if userID == uuid.Nil {
		logger.Warn("Invalid user ID")
		return errors.New("invalid user ID")
	}
	if itemID == uuid.Nil {
		logger.Warn("Invalid item ID")
		return errors.New("invalid item ID")
	}

	// If quantity is 0 or negative, remove the item
	if quantity <= 0 {
		logger.Debug("Quantity is zero or negative, removing item")
		return u.RemoveItem(ctx, userID, itemID)
	}

	// Get the item to check the product ID
	item, err := u.cartRepo.GetItem(ctx, userID, itemID)
	if err != nil {
		logger.Error("Failed to get cart item", "error", err.Error())
		return fmt.Errorf("failed to get cart item: %w", err)
	}

	// Check inventory before updating
	product, err := u.productRepo.GetByID(ctx, item.ProductID)
	if err != nil {
		logger.Error("Failed to get product", "error", err.Error())
		return fmt.Errorf("failed to get product: %w", err)
	}

	if !product.HasEnoughInventory(quantity) {
		logger.Warn("Not enough inventory",
			"product_id", product.ID.String(),
			"requested", quantity,
			"available", product.Inventory)
		return errs.New(nil, errs.CodeOutOfStock, 400, "Insufficient stock")
	}

	// Update the item quantity
	err = u.cartRepo.UpdateItem(ctx, itemID, quantity)
	if err != nil {
		logger.Error("Failed to update cart item", "error", err.Error())
		return fmt.Errorf("failed to update cart item: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully updated cart item quantity",
		"quantity", quantity,
		"duration_ms", duration.Milliseconds())

	return nil
}

// RemoveItem removes an item from a user's cart
func (u *cartUseCase) RemoveItem(ctx context.Context, userID, itemID uuid.UUID) error {
	logger := middleware.Logger.With(
		"method", "CartUseCase.RemoveItem",
		"user_id", userID.String(),
		"item_id", itemID.String(),
	)
	logger.Info("Removing item from cart")
	startTime := time.Now()

	if userID == uuid.Nil {
		logger.Warn("Invalid user ID")
		return errors.New("invalid user ID")
	}
	if itemID == uuid.Nil {
		logger.Warn("Invalid item ID")
		return errors.New("invalid item ID")
	}

	err := u.cartRepo.DeleteItem(ctx, userID, itemID)
	if err != nil {
		logger.Error("Failed to remove cart item", "error", err.Error())
		return fmt.Errorf("failed to remove cart item: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully removed cart item",
		"duration_ms", duration.Milliseconds())

	return nil
}

// ClearUserCart removes all items from a user's cart
func (u *cartUseCase) ClearUserCart(ctx context.Context, userID uuid.UUID) error {
	logger := middleware.Logger.With(
		"method", "CartUseCase.ClearUserCart",
		"user_id", userID.String(),
	)
	logger.Info("Clearing user cart")
	startTime := time.Now()

	if userID == uuid.Nil {
		logger.Warn("Invalid user ID")
		return errors.New("invalid user ID")
	}

	err := u.cartRepo.ClearUserCart(ctx, userID)
	if err != nil {
		logger.Error("Failed to clear user cart", "error", err.Error())
		return fmt.Errorf("failed to clear user cart: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully cleared user cart",
		"duration_ms", duration.Milliseconds())

	return nil
}

// GetUserCartInfo retrieves the cart with product details for a user
func (u *cartUseCase) GetUserCartInfo(ctx context.Context, userID uuid.UUID) (*cartEntity.CartInfo, error) {
	logger := middleware.Logger.With(
		"method", "CartUseCase.GetUserCartInfo",
		"user_id", userID.String(),
	)
	logger.Info("Retrieving user cart with product details")
	startTime := time.Now()

	if userID == uuid.Nil {
		logger.Warn("Invalid user ID")
		return nil, errors.New("invalid user ID")
	}

	cartInfo, err := u.cartRepo.GetCartInfo(ctx, userID)
	if err != nil {
		logger.Error("Failed to get cart info", "error", err.Error())
		return nil, err
	}

	// Apply promotions if cart is not empty
	if len(cartInfo.Items) > 0 {
		// Get applicable promotions from promotion service
		promotions, totalDiscount, err := u.promotionUseCase.ApplyPromotions(ctx, cartInfo)
		if err != nil {
			logger.Error("Failed to apply promotions", "error", err.Error())
			// Continue without promotions if there's an error
		} else {
			// Convert promotion discounts to cart's ApplicablePromotion type
			applicablePromotions := make([]cartEntity.ApplicablePromotion, 0, len(promotions))
			for _, p := range promotions {
				applicablePromotions = append(applicablePromotions, cartEntity.ApplicablePromotion{
					ID:          p.PromotionID,
					Type:        p.PromotionType,
					Description: p.Description,
					Discount:    p.Discount,
				})
			}

			// Apply promotions to cart
			cartInfo.ApplicablePromotions = applicablePromotions
			cartInfo.PotentialDiscount = totalDiscount
			cartInfo.PotentialTotal = cartInfo.Subtotal - totalDiscount
			if cartInfo.PotentialTotal < 0 {
				cartInfo.PotentialTotal = 0
			}

			logger.Info("Applied promotions to cart",
				"applicable_promotions_count", len(applicablePromotions),
				"total_discount", totalDiscount)
		}
	}

	itemCount := len(cartInfo.Items)
	duration := time.Since(startTime)
	logger.Info("Successfully retrieved cart info",
		"item_count", itemCount,
		"subtotal", cartInfo.Subtotal,
		"potential_total", cartInfo.PotentialTotal,
		"duration_ms", duration.Milliseconds())

	return cartInfo, nil
}
