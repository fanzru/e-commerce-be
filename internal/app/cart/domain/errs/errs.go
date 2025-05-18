package errs

import (
	"errors"
	"fmt"

	"github.com/fanzru/e-commerce-be/internal/common/errs"
)

// Cart domain errors
var (
	ErrCartNotFound      = errs.NewNotFound("Cart not found")
	ErrCartItemNotFound  = errs.NewNotFound("Cart item not found")
	ErrItemNotFound      = errs.NewBadRequest("Cart item not found")
	ErrProductNotFound   = errs.NewNotFound("Product not found")
	ErrInvalidQuantity   = errs.NewBadRequest("Invalid quantity")
	ErrInsufficientStock = errs.New(nil, errs.CodeOutOfStock, 400, "Insufficient stock")
)

// Cart domain error messages
const (
	ErrCartNotFoundMsg         = "cart not found"
	ErrCartItemNotFoundMsg     = "cart item not found"
	ErrProductNotInCartMsg     = "product not in cart"
	ErrQuantityUpdateFailedMsg = "failed to update item quantity"
	ErrCartEmptyMsg            = "cart is empty"
)

// NewCartNotFoundError creates a new cart not found error
func NewCartNotFoundError(id string) error {
	return errs.NewNotFound(fmt.Sprintf("%s: %s", ErrCartNotFoundMsg, id))
}

// NewCartItemNotFoundError creates a new cart item not found error
func NewCartItemNotFoundError(cartID, itemID string) error {
	return errs.NewNotFound(fmt.Sprintf("%s: cart %s, item %s", ErrCartItemNotFoundMsg, cartID, itemID))
}

// NewProductNotInCartError creates a new product not in cart error
func NewProductNotInCartError(cartID, productID string) error {
	return errs.NewNotFound(fmt.Sprintf("%s: cart %s, product %s", ErrProductNotInCartMsg, cartID, productID))
}

// NewQuantityUpdateFailedError creates a new quantity update failed error
func NewQuantityUpdateFailedError(cartID, productID string) error {
	return errs.NewBadRequest(fmt.Sprintf("%s: cart %s, product %s", ErrQuantityUpdateFailedMsg, cartID, productID))
}

// NewCartEmptyError creates a new cart empty error
func NewCartEmptyError(cartID string) error {
	return errs.NewBadRequest(fmt.Sprintf("%s: %s", ErrCartEmptyMsg, cartID))
}

// NewInsufficientStockError creates an insufficient stock error that will capture the caller's file and line
func NewInsufficientStockError(productID string, requested, available int) error {
	message := fmt.Sprintf("Insufficient stock for product %s (requested: %d, available: %d)",
		productID, requested, available)
	return errs.New(errors.New("insufficient stock"), errs.CodeOutOfStock, 400, message)
}

// IsItemNotFound checks if the error is an item not found error
func IsItemNotFound(err error) bool {
	return errors.Is(err, ErrItemNotFound)
}
