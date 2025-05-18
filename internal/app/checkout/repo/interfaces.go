package repo

import (
	"context"

	"github.com/fanzru/e-commerce-be/internal/app/checkout/domain/entity"
	"github.com/google/uuid"
)

// CheckoutRepository defines the interface for checkout repository
type CheckoutRepository interface {
	// GetByID retrieves a checkout by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Checkout, error)

	// GetByCartID retrieves a checkout by its cart ID
	GetByCartID(ctx context.Context, cartID uuid.UUID) (*entity.Checkout, error)

	// Create creates a new checkout
	Create(ctx context.Context, checkout *entity.Checkout) error

	// ListCheckouts retrieves a list of checkouts with pagination
	ListCheckouts(ctx context.Context, page, limit int) ([]*entity.Checkout, int, error)
}
