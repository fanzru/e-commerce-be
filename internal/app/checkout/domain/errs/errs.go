package errs

import (
	"fmt"

	appErrors "github.com/fanzru/e-commerce-be/pkg/errors"
)

// Checkout domain error messages
const (
	ErrCheckoutNotFoundMsg      = "checkout not found"
	ErrCheckoutFailedMsg        = "checkout failed"
	ErrCartEmptyMsg             = "cart is empty"
	ErrInsufficientInventoryMsg = "insufficient product inventory"
)

// NewCheckoutNotFoundError creates a new checkout not found error
func NewCheckoutNotFoundError(id string) error {
	return appErrors.NewNotFound(fmt.Sprintf("%s: %s", ErrCheckoutNotFoundMsg, id))
}

// NewCheckoutFailedError creates a new checkout failed error
func NewCheckoutFailedError(cartID string, err error) error {
	return appErrors.NewBadRequest(fmt.Sprintf("%s for cart %s: %v", ErrCheckoutFailedMsg, cartID, err))
}

// NewCartEmptyError creates a new cart empty error
func NewCartEmptyError(cartID string) error {
	return appErrors.NewBadRequest(fmt.Sprintf("%s: %s", ErrCartEmptyMsg, cartID))
}

// NewInsufficientInventoryError creates a new insufficient inventory error
func NewInsufficientInventoryError(productID, sku string, requested, available int) error {
	return appErrors.NewBadRequest(
		fmt.Sprintf("%s: product %s (SKU: %s, requested: %d, available: %d)",
			ErrInsufficientInventoryMsg,
			productID,
			sku,
			requested,
			available,
		),
	)
}
