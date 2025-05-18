package usecase

import (
	"context"
	"errors"
	"fmt"

	cartEntity "github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	cartRepo "github.com/fanzru/e-commerce-be/internal/app/cart/repo"
	productRepo "github.com/fanzru/e-commerce-be/internal/app/product/repo"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/google/uuid"
)

// Ensure cartUseCase implements CartUseCase
var _ CartUseCase = (*cartUseCase)(nil)

// cartUseCase implements the CartUseCase interface
type cartUseCase struct {
	cartRepo    cartRepo.CartRepository
	productRepo productRepo.ProductRepository
}

// NewCartUseCase creates a new instance of cartUseCase
func NewCartUseCase(
	cartRepo cartRepo.CartRepository,
	productRepo productRepo.ProductRepository,
) CartUseCase {
	return &cartUseCase{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

// Create creates a new empty cart
func (u *cartUseCase) Create(ctx context.Context) (*cartEntity.Cart, error) {
	cart := cartEntity.NewCart()
	err := u.cartRepo.Create(ctx, cart)
	if err != nil {
		return nil, fmt.Errorf("failed to create cart: %w", err)
	}
	return cart, nil
}

// GetByID retrieves a cart by its ID with all items
func (u *cartUseCase) GetByID(ctx context.Context, id uuid.UUID) (*cartEntity.Cart, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid cart ID")
	}
	return u.cartRepo.GetByID(ctx, id)
}

// Delete deletes a cart by its ID
func (u *cartUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid cart ID")
	}
	return u.cartRepo.Delete(ctx, id)
}

// AddItem adds a product to a cart
func (u *cartUseCase) AddItem(ctx context.Context, cartID, productID uuid.UUID, quantity int) (*cartEntity.CartItem, error) {
	if cartID == uuid.Nil {
		return nil, errors.New("invalid cart ID")
	}
	if productID == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}

	// Get the cart
	cart, err := u.cartRepo.GetByID(ctx, cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Get the product
	product, err := u.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Check if there's enough inventory
	if !product.HasEnoughInventory(quantity) {
		return nil, errors.New("not enough inventory")
	}

	// Check if the item already exists in the cart
	item := cart.GetItem(productID)
	if item != nil {
		// Update quantity if item exists
		newQuantity := item.Quantity + quantity
		if !product.HasEnoughInventory(newQuantity) {
			return nil, errors.New("not enough inventory")
		}

		err = u.cartRepo.UpdateItem(ctx, item.ID, newQuantity)
		if err != nil {
			return nil, fmt.Errorf("failed to update cart item: %w", err)
		}

		// Get updated item
		updatedItem, err := u.cartRepo.GetItem(ctx, cartID, item.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get updated cart item: %w", err)
		}
		return updatedItem, nil
	}

	// Add new item to cart
	cartItem := &cartEntity.CartItem{
		ID:          uuid.New(),
		CartID:      cartID,
		ProductID:   productID,
		ProductSKU:  product.SKU,
		ProductName: product.Name,
		UnitPrice:   product.Price,
		Quantity:    quantity,
	}

	err = u.cartRepo.AddItem(ctx, cartItem)
	if err != nil {
		return nil, fmt.Errorf("failed to add item to cart: %w", err)
	}

	return cartItem, nil
}

// UpdateItemQuantity updates the quantity of a cart item
func (u *cartUseCase) UpdateItemQuantity(ctx context.Context, cartID, itemID uuid.UUID, quantity int) error {
	if cartID == uuid.Nil {
		return errors.New("invalid cart ID")
	}
	if itemID == uuid.Nil {
		return errors.New("invalid item ID")
	}
	if quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	// Get the cart item
	item, err := u.cartRepo.GetItem(ctx, cartID, itemID)
	if err != nil {
		return fmt.Errorf("failed to get cart item: %w", err)
	}

	// Get the product to check inventory
	product, err := u.productRepo.GetByID(ctx, item.ProductID)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// If quantity is 0, remove the item
	if quantity == 0 {
		return u.cartRepo.DeleteItem(ctx, cartID, itemID)
	}

	// Check if there's enough inventory
	if !product.HasEnoughInventory(quantity) {
		return errors.New("not enough inventory")
	}

	// Update the item quantity
	return u.cartRepo.UpdateItem(ctx, itemID, quantity)
}

// RemoveItem removes an item from a cart
func (u *cartUseCase) RemoveItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	if cartID == uuid.Nil {
		return errors.New("invalid cart ID")
	}
	if itemID == uuid.Nil {
		return errors.New("invalid item ID")
	}

	return u.cartRepo.DeleteItem(ctx, cartID, itemID)
}

// CreateForUser creates a new empty cart for a specific user
func (u *cartUseCase) CreateForUser(ctx context.Context, userID uuid.UUID) (*cartEntity.Cart, error) {
	// Check if user already has a cart
	existingCart, err := u.cartRepo.GetByUserID(ctx, userID)
	if err == nil && existingCart != nil {
		return existingCart, nil
	}

	// Create a new cart with user ID
	cart := cartEntity.NewCartWithUser(userID)
	err = u.cartRepo.Create(ctx, cart)
	if err != nil {
		return nil, fmt.Errorf("failed to create cart: %w", err)
	}
	return cart, nil
}

// GetByUserID retrieves a cart by user ID
func (u *cartUseCase) GetByUserID(ctx context.Context, userID uuid.UUID) (*cartEntity.Cart, error) {
	if userID == uuid.Nil {
		return nil, errors.New("invalid user ID")
	}

	cart, err := u.cartRepo.GetByUserID(ctx, userID)
	if err != nil {
		// If no cart found, create a new one for the user
		if errors.Is(err, cartRepo.ErrCartNotFound) {
			return u.CreateForUser(ctx, userID)
		}
		return nil, err
	}

	return cart, nil
}

// AddItemToUserCart adds a product to a user's cart (creates cart if needed)
func (u *cartUseCase) AddItemToUserCart(ctx context.Context, userID, productID uuid.UUID, quantity int) (*cartEntity.CartItem, error) {
	// Create a logger with context information
	logger := middleware.Logger.With(
		"method", "AddItemToUserCart",
		"user_id", userID.String(),
		"product_id", productID.String(),
		"quantity", quantity,
	)

	logger.Info("Starting to add item to user cart")

	if userID == uuid.Nil {
		logger.Error("Invalid user ID")
		return nil, errors.New("invalid user ID")
	}
	if productID == uuid.Nil {
		logger.Error("Invalid product ID")
		return nil, errors.New("invalid product ID")
	}
	if quantity <= 0 {
		logger.Error("Invalid quantity", "quantity", quantity)
		return nil, errors.New("quantity must be greater than zero")
	}

	// Get or create cart for user
	logger.Debug("Fetching or creating cart for user")
	cart, err := u.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get or create cart", "error", err.Error())
		return nil, fmt.Errorf("failed to get or create cart: %w", err)
	}

	logger.Info("Successfully retrieved cart", "cart_id", cart.ID.String())

	// Add the item to the cart
	logger.Debug("Adding item to cart", "cart_id", cart.ID.String())
	item, err := u.AddItem(ctx, cart.ID, productID, quantity)
	if err != nil {
		logger.Error("Failed to add item to cart", "error", err.Error())
		return nil, err
	}

	logger.Info("Successfully added item to cart",
		"cart_id", cart.ID.String(),
		"item_id", item.ID.String(),
		"product_id", item.ProductID.String(),
		"quantity", item.Quantity)

	return item, nil
}
