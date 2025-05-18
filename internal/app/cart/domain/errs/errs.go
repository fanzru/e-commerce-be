package errs

import (
	"fmt"

	appErrors "github.com/fanzru/e-commerce-be/pkg/errors"
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
	return appErrors.NewNotFound(fmt.Sprintf("%s: %s", ErrCartNotFoundMsg, id))
}

// NewCartItemNotFoundError creates a new cart item not found error
func NewCartItemNotFoundError(cartID, itemID string) error {
	return appErrors.NewNotFound(fmt.Sprintf("%s: cart %s, item %s", ErrCartItemNotFoundMsg, cartID, itemID))
}

// NewProductNotInCartError creates a new product not in cart error
func NewProductNotInCartError(cartID, productID string) error {
	return appErrors.NewNotFound(fmt.Sprintf("%s: cart %s, product %s", ErrProductNotInCartMsg, cartID, productID))
}

// NewQuantityUpdateFailedError creates a new quantity update failed error
func NewQuantityUpdateFailedError(cartID, productID string) error {
	return appErrors.NewBadRequest(fmt.Sprintf("%s: cart %s, product %s", ErrQuantityUpdateFailedMsg, cartID, productID))
}

// NewCartEmptyError creates a new cart empty error
func NewCartEmptyError(cartID string) error {
	return appErrors.NewBadRequest(fmt.Sprintf("%s: %s", ErrCartEmptyMsg, cartID))
}
