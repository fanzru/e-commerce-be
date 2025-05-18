package usecase

import (
	"context"

	cartEntity "github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	"github.com/google/uuid"
)

// CartUseCase defines the interface for cart use cases
type CartUseCase interface {
	// Create creates a new empty cart
	Create(ctx context.Context) (*cartEntity.Cart, error)

	// CreateForUser creates a new empty cart for a specific user
	CreateForUser(ctx context.Context, userID uuid.UUID) (*cartEntity.Cart, error)

	// GetByID retrieves a cart by its ID with all items
	GetByID(ctx context.Context, id uuid.UUID) (*cartEntity.Cart, error)

	// GetByUserID retrieves a cart by user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*cartEntity.Cart, error)

	// Delete deletes a cart by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// AddItem adds a product to a cart
	AddItem(ctx context.Context, cartID, productID uuid.UUID, quantity int) (*cartEntity.CartItem, error)

	// AddItemToUserCart adds a product to a user's cart (creates cart if needed)
	AddItemToUserCart(ctx context.Context, userID, productID uuid.UUID, quantity int) (*cartEntity.CartItem, error)

	// UpdateItemQuantity updates the quantity of a cart item
	UpdateItemQuantity(ctx context.Context, cartID, itemID uuid.UUID, quantity int) error

	// RemoveItem removes an item from a cart
	RemoveItem(ctx context.Context, cartID, itemID uuid.UUID) error
}
