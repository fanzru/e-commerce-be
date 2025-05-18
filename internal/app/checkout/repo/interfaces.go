package repo

import (
	"context"
	"errors"

	"github.com/fanzru/e-commerce-be/internal/app/checkout/domain/entity"
	"github.com/google/uuid"
)

// Error constants for checkout repository
var (
	ErrCheckoutNotFound = errors.New("checkout not found")
)

// CheckoutRepository defines the interface for checkout repository
type CheckoutRepository interface {
	// GetByID retrieves a checkout by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Checkout, error)

	// Create creates a new checkout
	Create(ctx context.Context, checkout *entity.Checkout) error

	// List retrieves a list of checkouts with pagination
	List(ctx context.Context, page, limit int) ([]*entity.Checkout, int, error)

	// GetByUserID retrieves a list of checkouts for a specific user
	GetByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.Checkout, int, error)

	// UpdatePaymentStatus updates the payment status of a checkout
	UpdatePaymentStatus(ctx context.Context, checkoutID uuid.UUID, status entity.PaymentStatus, paymentMethod, paymentReference string) error

	// UpdateOrderStatus updates the order status of a checkout
	UpdateOrderStatus(ctx context.Context, checkoutID uuid.UUID, status entity.OrderStatus) error
}
