package usecase

import (
	"context"

	cartEntity "github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	"github.com/google/uuid"
)

// CartUseCase defines the interface for cart use cases
type CartUseCase interface {
	// GetUserCart retrieves the cart for a user
	GetUserCart(ctx context.Context, userID uuid.UUID) (*cartEntity.Cart, error)

	// GetUserCartInfo retrieves the cart with product details for a user
	GetUserCartInfo(ctx context.Context, userID uuid.UUID) (*cartEntity.CartInfo, error)

	// AddItemToUserCart adds a product to a user's cart
	AddItemToUserCart(ctx context.Context, userID, productID uuid.UUID, quantity int) (*cartEntity.CartItem, error)

	// UpdateItemQuantity updates the quantity of a cart item
	UpdateItemQuantity(ctx context.Context, userID, itemID uuid.UUID, quantity int) error

	// RemoveItem removes an item from a user's cart
	RemoveItem(ctx context.Context, userID, itemID uuid.UUID) error

	// ClearUserCart removes all items from a user's cart
	ClearUserCart(ctx context.Context, userID uuid.UUID) error
}
