package repo

import (
	"context"
	"errors"

	"github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	"github.com/google/uuid"
)

// Error constants for cart repository
var (
	ErrCartNotFound = errors.New("cart not found")
	ErrItemNotFound = errors.New("cart item not found")
)

// CartRepository defines the interface for cart repository
type CartRepository interface {
	// GetByUserID retrieves all cart items for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Cart, error)

	// GetCartInfo retrieves cart with product details for display
	GetCartInfo(ctx context.Context, userID uuid.UUID) (*entity.CartInfo, error)

	// AddItem adds an item to a user's cart or updates its quantity if already exists
	AddItem(ctx context.Context, item *entity.CartItem) error

	// UpdateItem updates a cart item's quantity
	UpdateItem(ctx context.Context, itemID uuid.UUID, quantity int) error

	// DeleteItem removes an item from a user's cart
	DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error

	// GetItem gets a specific item from a user's cart
	GetItem(ctx context.Context, userID, itemID uuid.UUID) (*entity.CartItem, error)

	// GetItemByProductID gets a specific item by product ID from a user's cart
	GetItemByProductID(ctx context.Context, userID, productID uuid.UUID) (*entity.CartItem, error)

	// ClearUserCart removes all items from a user's cart
	ClearUserCart(ctx context.Context, userID uuid.UUID) error
}
