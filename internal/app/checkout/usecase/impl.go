package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cartRepo "github.com/fanzru/e-commerce-be/internal/app/cart/repo"
	checkoutEntity "github.com/fanzru/e-commerce-be/internal/app/checkout/domain/entity"
	checkoutErrors "github.com/fanzru/e-commerce-be/internal/app/checkout/domain/errs"
	checkoutRepo "github.com/fanzru/e-commerce-be/internal/app/checkout/repo"
	"github.com/fanzru/e-commerce-be/internal/app/promotion/domain/entity"
	promotionRepo "github.com/fanzru/e-commerce-be/internal/app/promotion/repo"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/persistence"
	"github.com/google/uuid"
)

// Ensure checkoutUseCase implements CheckoutUseCase
var _ CheckoutUseCase = (*checkoutUseCase)(nil)

// checkoutUseCase implements the CheckoutUseCase interface
type checkoutUseCase struct {
	checkoutRepo  checkoutRepo.CheckoutRepository
	cartRepo      cartRepo.CartRepository
	promotionRepo promotionRepo.PromotionRepository
	txManager     *persistence.TransactionManager
}

// NewCheckoutUseCase creates a new instance of checkoutUseCase
func NewCheckoutUseCase(
	checkoutRepo checkoutRepo.CheckoutRepository,
	cartRepo cartRepo.CartRepository,
	promotionRepo promotionRepo.PromotionRepository,
	txManager *persistence.TransactionManager,
) CheckoutUseCase {
	return &checkoutUseCase{
		checkoutRepo:  checkoutRepo,
		cartRepo:      cartRepo,
		promotionRepo: promotionRepo,
		txManager:     txManager,
	}
}

// GetByID retrieves a checkout by its ID
func (u *checkoutUseCase) GetByID(ctx context.Context, id uuid.UUID) (*checkoutEntity.Checkout, error) {
	logger := middleware.Logger.With(
		"method", "CheckoutUseCase.GetByID",
		"checkout_id", id.String(),
	)
	logger.Info("Getting checkout by ID")
	startTime := time.Now()

	checkout, err := u.checkoutRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get checkout", "error", err.Error())
		return nil, fmt.Errorf("error getting checkout: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved checkout",
		"user_id", checkout.UserID,
		"subtotal", checkout.Subtotal,
		"total_discount", checkout.TotalDiscount,
		"total", checkout.Total,
		"duration_ms", duration.Milliseconds())

	return checkout, nil
}

// ProcessCart processes a cart and creates a checkout
func (u *checkoutUseCase) ProcessCart(ctx context.Context, userID uuid.UUID) (*checkoutEntity.Checkout, error) {
	logger := middleware.Logger.With(
		"method", "CheckoutUseCase.ProcessCart",
		"user_id", userID.String(),
	)
	logger.Info("Processing cart for checkout")
	startTime := time.Now()

	// Create a checkout object that will be populated
	var checkout *checkoutEntity.Checkout

	// Execute all checkout operations in a transaction
	err := u.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		// Get cart with items by user ID
		cartInfo, err := u.cartRepo.GetCartInfo(txCtx, userID)
		if err != nil {
			logger.Error("Failed to get cart", "error", err.Error())
			return fmt.Errorf("error getting cart: %w", err)
		}

		// Check if cart has items
		if len(cartInfo.Items) == 0 {
			logger.Warn("Cart is empty", "error", "ErrEmptyCart")
			return checkoutErrors.ErrEmptyCart
		}

		// Get active promotions
		activePromotions, err := u.getActivePromotions(txCtx)
		if err != nil {
			logger.Error("Failed to get active promotions", "error", err.Error())
			return fmt.Errorf("error getting active promotions: %w", err)
		}

		// Create checkout
		checkout = &checkoutEntity.Checkout{
			ID:            uuid.New(),
			UserID:        getUserIDPointer(userID),
			Items:         []*checkoutEntity.CheckoutItem{},
			Promotions:    []*checkoutEntity.PromotionApplied{},
			Subtotal:      0,
			TotalDiscount: 0,
			Total:         0,
			PaymentStatus: checkoutEntity.PaymentStatusPending,
			Status:        checkoutEntity.OrderStatusCreated,
		}

		// Process items
		for _, cartItem := range cartInfo.Items {
			// Create checkout item
			checkoutItem := &checkoutEntity.CheckoutItem{
				ID:          uuid.New(),
				CheckoutID:  checkout.ID,
				ProductID:   cartItem.ProductID,
				ProductSKU:  cartItem.ProductSKU,
				ProductName: cartItem.ProductName,
				Quantity:    cartItem.Quantity,
				UnitPrice:   cartItem.UnitPrice,
				Subtotal:    cartItem.UnitPrice * float64(cartItem.Quantity),
				Discount:    0, // Will be calculated later
				Total:       cartItem.UnitPrice * float64(cartItem.Quantity),
			}

			checkout.Items = append(checkout.Items, checkoutItem)
			checkout.Subtotal += checkoutItem.Subtotal
		}

		// Apply promotions
		u.applyPromotions(checkout, activePromotions)

		// Calculate totals
		checkout.Total = checkout.Subtotal - checkout.TotalDiscount

		// Save checkout - this will be part of the transaction
		err = u.checkoutRepo.Create(txCtx, checkout)
		if err != nil {
			logger.Error("Failed to create checkout", "error", err.Error())
			return fmt.Errorf("error creating checkout: %w", err)
		}

		// Clear the user's cart after successful checkout
		err = u.cartRepo.ClearUserCart(txCtx, userID)
		if err != nil {
			logger.Error("Failed to clear user cart", "error", err.Error())
			return fmt.Errorf("error clearing user cart: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	logger.Info("Successfully processed cart checkout",
		"checkout_id", checkout.ID.String(),
		"user_id", checkout.UserID,
		"payment_status", checkout.PaymentStatus,
		"status", checkout.Status,
		"subtotal", checkout.Subtotal,
		"total_discount", checkout.TotalDiscount,
		"total", checkout.Total,
		"item_count", len(checkout.Items),
		"promotion_count", len(checkout.Promotions),
		"duration_ms", duration.Milliseconds())

	return checkout, nil
}

// ListCheckouts retrieves a list of checkouts with pagination
func (u *checkoutUseCase) ListCheckouts(ctx context.Context, page, limit int) ([]*checkoutEntity.Checkout, int, error) {
	logger := middleware.Logger.With(
		"method", "CheckoutUseCase.ListCheckouts",
		"page", page,
		"limit", limit,
	)
	logger.Info("Listing checkouts")
	startTime := time.Now()

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	checkouts, total, err := u.checkoutRepo.List(ctx, page, limit)
	if err != nil {
		logger.Error("Failed to list checkouts", "error", err.Error())
		return nil, 0, fmt.Errorf("error listing checkouts: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully listed checkouts",
		"total", total,
		"returned", len(checkouts),
		"duration_ms", duration.Milliseconds())

	return checkouts, total, nil
}

// GetUserOrders retrieves a list of checkouts for a specific user
func (u *checkoutUseCase) GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int) ([]*checkoutEntity.Checkout, int, error) {
	logger := middleware.Logger.With(
		"method", "CheckoutUseCase.GetUserOrders",
		"user_id", userID.String(),
		"page", page,
		"limit", limit,
	)
	logger.Info("Getting user orders")
	startTime := time.Now()

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	// Get user orders from repository
	orders, total, err := u.checkoutRepo.GetByUserID(ctx, userID, page, limit)
	if err != nil {
		logger.Error("Failed to get user orders", "error", err.Error())
		return nil, 0, fmt.Errorf("error getting user orders: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved user orders",
		"user_id", userID,
		"total", total,
		"returned", len(orders),
		"duration_ms", duration.Milliseconds())

	return orders, total, nil
}

// UpdatePaymentStatus updates the payment status of a checkout
func (u *checkoutUseCase) UpdatePaymentStatus(ctx context.Context, checkoutID uuid.UUID, status checkoutEntity.PaymentStatus, paymentMethod, paymentReference string) error {
	logger := middleware.Logger.With(
		"method", "CheckoutUseCase.UpdatePaymentStatus",
		"checkout_id", checkoutID.String(),
		"payment_status", status,
		"payment_method", paymentMethod,
	)
	logger.Info("Updating payment status")
	startTime := time.Now()

	// Validate inputs
	if checkoutID == uuid.Nil {
		return fmt.Errorf("invalid checkout ID")
	}

	if status == "" {
		return fmt.Errorf("payment status cannot be empty")
	}

	// Check if checkout exists
	_, err := u.checkoutRepo.GetByID(ctx, checkoutID)
	if err != nil {
		logger.Error("Failed to get checkout", "error", err.Error())
		return fmt.Errorf("error getting checkout: %w", err)
	}

	// Update payment status
	err = u.checkoutRepo.UpdatePaymentStatus(ctx, checkoutID, status, paymentMethod, paymentReference)
	if err != nil {
		logger.Error("Failed to update payment status", "error", err.Error())
		return fmt.Errorf("error updating payment status: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully updated payment status",
		"checkout_id", checkoutID,
		"payment_status", status,
		"payment_method", paymentMethod,
		"duration_ms", duration.Milliseconds())

	return nil
}

// UpdateOrderStatus updates the order status of a checkout
func (u *checkoutUseCase) UpdateOrderStatus(ctx context.Context, checkoutID uuid.UUID, status checkoutEntity.OrderStatus) error {
	logger := middleware.Logger.With(
		"method", "CheckoutUseCase.UpdateOrderStatus",
		"checkout_id", checkoutID.String(),
		"order_status", status,
	)
	logger.Info("Updating order status")
	startTime := time.Now()

	// Validate inputs
	if checkoutID == uuid.Nil {
		return fmt.Errorf("invalid checkout ID")
	}

	if status == "" {
		return fmt.Errorf("order status cannot be empty")
	}

	// Check if checkout exists
	checkout, err := u.checkoutRepo.GetByID(ctx, checkoutID)
	if err != nil {
		logger.Error("Failed to get checkout", "error", err.Error())
		return fmt.Errorf("error getting checkout: %w", err)
	}

	// Validate status transition
	if !isValidOrderStatusTransition(checkout.Status, status) {
		logger.Warn("Invalid order status transition",
			"current_status", checkout.Status,
			"requested_status", status)
		return fmt.Errorf("invalid order status transition from %s to %s", checkout.Status, status)
	}

	// Check if payment is required for this status
	if requiresPayment(status) && checkout.PaymentStatus != checkoutEntity.PaymentStatusPaid {
		logger.Warn("Order status update requires payment",
			"current_payment_status", checkout.PaymentStatus,
			"requested_order_status", status)
		return fmt.Errorf("order must be paid before changing status to %s", status)
	}

	// Update order status
	err = u.checkoutRepo.UpdateOrderStatus(ctx, checkoutID, status)
	if err != nil {
		logger.Error("Failed to update order status", "error", err.Error())
		return fmt.Errorf("error updating order status: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully updated order status",
		"checkout_id", checkoutID,
		"order_status", status,
		"duration_ms", duration.Milliseconds())

	return nil
}

// Helper functions

// getActivePromotions retrieves all active promotions
func (u *checkoutUseCase) getActivePromotions(ctx context.Context) ([]*entity.Promotion, error) {
	// Get active promotions
	activeFlag := true
	promotions, _, err := u.promotionRepo.List(ctx, 1, 100, &activeFlag)
	if err != nil {
		return nil, fmt.Errorf("error getting active promotions: %w", err)
	}
	return promotions, nil
}

// applyPromotions applies promotions to the checkout
func (u *checkoutUseCase) applyPromotions(checkout *checkoutEntity.Checkout, promotions []*entity.Promotion) {
	logger := middleware.Logger.With(
		"method", "CheckoutUseCase.applyPromotions",
		"checkout_id", checkout.ID.String(),
	)

	// Create a map for quick lookup of items by SKU
	itemsBySKU := make(map[string][]*checkoutEntity.CheckoutItem)
	skuQuantities := make(map[string]int)

	// Group items by SKU and calculate total quantities
	for _, item := range checkout.Items {
		itemsBySKU[item.ProductSKU] = append(itemsBySKU[item.ProductSKU], item)
		skuQuantities[item.ProductSKU] += item.Quantity
	}

	// Process each promotion
	for _, promotion := range promotions {
		var ruleMap map[string]interface{}
		if err := json.Unmarshal(promotion.Rule, &ruleMap); err != nil {
			logger.Error("Failed to parse promotion rule",
				"promotion_id", promotion.ID.String(),
				"error", err.Error())
			continue
		}

		switch promotion.Type {
		case entity.BuyOneGetOneFree:
			// Apply buy one get one free promotion
			triggerSKU, _ := ruleMap["trigger_sku"].(string)
			freeSKU, _ := ruleMap["free_sku"].(string)
			triggerQty, _ := ruleMap["trigger_quantity"].(float64)
			freeQty, _ := ruleMap["free_quantity"].(float64)

			// Check if we have enough trigger items
			if triggerSKU == "" || freeSKU == "" || triggerQty <= 0 || freeQty <= 0 {
				logger.Warn("Invalid BuyOneGetOneFree promotion rule", "promotion_id", promotion.ID.String())
				continue
			}

			triggerQuantity := int(triggerQty)
			freeQuantity := int(freeQty)

			// Check if we have both the trigger and free items in the cart
			_, hasTriggerItems := itemsBySKU[triggerSKU]
			freeItems, hasFreeItems := itemsBySKU[freeSKU]

			if !hasTriggerItems || !hasFreeItems {
				continue // Skip this promotion if items not found
			}

			// Calculate how many free items to give
			triggerItemsCount := skuQuantities[triggerSKU]
			freeItemsCount := skuQuantities[freeSKU]

			// Calculate how many free items we're entitled to
			entitledFreeItems := (triggerItemsCount / triggerQuantity) * freeQuantity

			// Cap at the actual number of items
			discountItemCount := min(entitledFreeItems, freeItemsCount)

			if discountItemCount <= 0 {
				continue
			}

			// Find the total value of discounted items
			var totalDiscount float64
			remainingDiscountItems := discountItemCount

			// Apply discount to free items
			for _, item := range freeItems {
				if remainingDiscountItems <= 0 {
					break
				}

				itemDiscountQty := min(remainingDiscountItems, item.Quantity)
				itemDiscount := item.UnitPrice * float64(itemDiscountQty)

				// Update the discount for this item
				item.Discount += itemDiscount
				item.Total = item.Subtotal - item.Discount

				totalDiscount += itemDiscount
				remainingDiscountItems -= itemDiscountQty
			}

			if totalDiscount > 0 {
				// Add to total discount
				checkout.TotalDiscount += totalDiscount

				// Record the applied promotion
				appliedPromotion := &checkoutEntity.PromotionApplied{
					ID:          uuid.New(),
					CheckoutID:  checkout.ID,
					PromotionID: promotion.ID,
					Description: fmt.Sprintf("%s (Free: %d x %s)",
						promotion.Description, discountItemCount, freeSKU),
					Discount: totalDiscount,
				}
				checkout.Promotions = append(checkout.Promotions, appliedPromotion)

				logger.Info("Applied BuyOneGetOneFree promotion",
					"promotion_id", promotion.ID.String(),
					"discount_amount", totalDiscount,
					"trigger_sku", triggerSKU,
					"free_sku", freeSKU)
			}

		case entity.Buy3Pay2:
			// Apply buy 3 pay 2 promotion
			sku, _ := ruleMap["sku"].(string)
			minQuantity, _ := ruleMap["min_quantity"].(float64)
			paidDivisor, _ := ruleMap["paid_quantity_divisor"].(float64)
			freeDivisor, _ := ruleMap["free_quantity_divisor"].(float64)

			// Validate rule
			if sku == "" || minQuantity <= 0 || paidDivisor <= 0 || freeDivisor <= 0 {
				logger.Warn("Invalid Buy3Pay2 promotion rule", "promotion_id", promotion.ID.String())
				continue
			}

			// Check if we have the items and enough quantity
			items, hasItems := itemsBySKU[sku]
			totalQuantity := skuQuantities[sku]

			if !hasItems || totalQuantity < int(minQuantity) {
				continue // Skip if we don't have enough items
			}

			// Calculate discount rate (e.g., for 3 pay 2, the rate is 1/3 = 33.33%)
			discountRate := freeDivisor / (paidDivisor + freeDivisor)

			// Calculate total discount amount
			var totalDiscount float64

			// Apply discount proportionally to all items of this SKU
			for _, item := range items {
				itemDiscount := item.Subtotal * discountRate
				item.Discount += itemDiscount
				item.Total = item.Subtotal - item.Discount
				totalDiscount += itemDiscount
			}

			if totalDiscount > 0 {
				// Add to total discount
				checkout.TotalDiscount += totalDiscount

				// Record the applied promotion
				appliedPromotion := &checkoutEntity.PromotionApplied{
					ID:          uuid.New(),
					CheckoutID:  checkout.ID,
					PromotionID: promotion.ID,
					Description: fmt.Sprintf("%s (Discount: %.1f%%)",
						promotion.Description, discountRate*100),
					Discount: totalDiscount,
				}
				checkout.Promotions = append(checkout.Promotions, appliedPromotion)

				logger.Info("Applied Buy3Pay2 promotion",
					"promotion_id", promotion.ID.String(),
					"discount_amount", totalDiscount,
					"discount_rate", discountRate,
					"sku", sku)
			}

		case entity.BulkDiscount:
			// Apply bulk discount promotion
			sku, _ := ruleMap["sku"].(string)
			minQuantity, _ := ruleMap["min_quantity"].(float64)
			discountPercentage, _ := ruleMap["discount_percentage"].(float64)

			// Validate rule
			if sku == "" || minQuantity <= 0 || discountPercentage <= 0 || discountPercentage > 100 {
				logger.Warn("Invalid BulkDiscount promotion rule", "promotion_id", promotion.ID.String())
				continue
			}

			// Check if we have the items and enough quantity
			items, hasItems := itemsBySKU[sku]
			totalQuantity := skuQuantities[sku]

			if !hasItems || totalQuantity < int(minQuantity) {
				continue // Skip if we don't have enough items
			}

			// Calculate discount rate
			discountRate := discountPercentage / 100.0

			// Calculate total discount amount
			var totalDiscount float64

			// Apply discount to all items of this SKU
			for _, item := range items {
				itemDiscount := item.Subtotal * discountRate
				item.Discount += itemDiscount
				item.Total = item.Subtotal - item.Discount
				totalDiscount += itemDiscount
			}

			if totalDiscount > 0 {
				// Add to total discount
				checkout.TotalDiscount += totalDiscount

				// Record the applied promotion
				appliedPromotion := &checkoutEntity.PromotionApplied{
					ID:          uuid.New(),
					CheckoutID:  checkout.ID,
					PromotionID: promotion.ID,
					Description: fmt.Sprintf("%s (%.1f%% off)",
						promotion.Description, discountPercentage),
					Discount: totalDiscount,
				}
				checkout.Promotions = append(checkout.Promotions, appliedPromotion)

				logger.Info("Applied BulkDiscount promotion",
					"promotion_id", promotion.ID.String(),
					"discount_amount", totalDiscount,
					"discount_percentage", discountPercentage,
					"sku", sku)
			}
		}
	}

	// Update all item totals after all discounts are applied
	for _, item := range checkout.Items {
		item.Total = item.Subtotal - item.Discount
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper functions for order status validation

// isValidOrderStatusTransition checks if a status transition is valid
func isValidOrderStatusTransition(current, next checkoutEntity.OrderStatus) bool {
	// Define allowed transitions
	transitions := map[checkoutEntity.OrderStatus][]checkoutEntity.OrderStatus{
		checkoutEntity.OrderStatusCreated: {
			checkoutEntity.OrderStatusProcessing,
			checkoutEntity.OrderStatusCancelled,
		},
		checkoutEntity.OrderStatusProcessing: {
			checkoutEntity.OrderStatusShipped,
			checkoutEntity.OrderStatusCancelled,
		},
		checkoutEntity.OrderStatusShipped: {
			checkoutEntity.OrderStatusDelivered,
			checkoutEntity.OrderStatusCancelled,
		},
		checkoutEntity.OrderStatusDelivered: {
			// Terminal state, no further transitions
		},
		checkoutEntity.OrderStatusCancelled: {
			// Terminal state, no further transitions
		},
	}

	// Check if the transition is allowed
	allowedTransitions, exists := transitions[current]
	if !exists {
		return false
	}

	// Allow same status to handle idempotent updates
	if current == next {
		return true
	}

	// Check if next status is in the allowed transitions
	for _, status := range allowedTransitions {
		if status == next {
			return true
		}
	}

	return false
}

// requiresPayment checks if a status requires payment
func requiresPayment(status checkoutEntity.OrderStatus) bool {
	// These statuses require payment
	paymentRequiredStatuses := []checkoutEntity.OrderStatus{
		checkoutEntity.OrderStatusProcessing,
		checkoutEntity.OrderStatusShipped,
		checkoutEntity.OrderStatusDelivered,
	}

	for _, s := range paymentRequiredStatuses {
		if status == s {
			return true
		}
	}

	return false
}

// Helper function to convert a UUID to a pointer
func getUserIDPointer(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
