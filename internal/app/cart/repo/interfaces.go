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
)

// CartRepository defines the interface for cart repository
type CartRepository interface {
	// GetByID retrieves a cart by its ID with all items
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Cart, error)

	// GetByUserID retrieves a cart by user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Cart, error)

	// Create creates a new empty cart
	Create(ctx context.Context, cart *entity.Cart) error

	// Delete deletes a cart by its ID
	Delete(ctx context.Context, id uuid.UUID) error

	// AddItem adds an item to a cart or updates its quantity if already exists
	AddItem(ctx context.Context, item *entity.CartItem) error

	// UpdateItem updates a cart item's quantity
	UpdateItem(ctx context.Context, itemID uuid.UUID, quantity int) error

	// DeleteItem removes an item from a cart
	DeleteItem(ctx context.Context, cartID, itemID uuid.UUID) error

	// GetItem gets a specific item from a cart
	GetItem(ctx context.Context, cartID, itemID uuid.UUID) (*entity.CartItem, error)
}
