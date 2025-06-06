package usecase

import (
	"context"

	checkoutEntity "github.com/fanzru/e-commerce-be/internal/app/checkout/domain/entity"
	"github.com/google/uuid"
)

// CheckoutUseCase defines the interface for checkout use cases
type CheckoutUseCase interface {
	// GetByID retrieves a checkout by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*checkoutEntity.Checkout, error)

	// ProcessCart processes a cart and creates a checkout
	ProcessCart(ctx context.Context, userID uuid.UUID) (*checkoutEntity.Checkout, error)

	// ListCheckouts retrieves a list of checkouts with pagination
	ListCheckouts(ctx context.Context, page, limit int) ([]*checkoutEntity.Checkout, int, error)

	// GetUserOrders retrieves a list of checkouts for a specific user
	GetUserOrders(ctx context.Context, userID uuid.UUID, page, limit int) ([]*checkoutEntity.Checkout, int, error)

	// UpdatePaymentStatus updates the payment status of a checkout
	UpdatePaymentStatus(ctx context.Context, checkoutID uuid.UUID, status checkoutEntity.PaymentStatus, paymentMethod, paymentReference string) error

	// UpdateOrderStatus updates the order status of a checkout
	UpdateOrderStatus(ctx context.Context, checkoutID uuid.UUID, status checkoutEntity.OrderStatus) error
}
