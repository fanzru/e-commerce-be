package errs

import (
	"fmt"

	appErrors "github.com/fanzru/e-commerce-be/pkg/errors"
)

// Product domain error messages
const (
	ErrProductNotFoundMsg       = "product not found"
	ErrProductAlreadyExistsMsg  = "product with this SKU already exists"
	ErrInsufficientInventoryMsg = "insufficient product inventory"
)

// NewProductNotFoundError creates a new product not found error
func NewProductNotFoundError(id string) error {
	return appErrors.NewNotFound(fmt.Sprintf("%s: %s", ErrProductNotFoundMsg, id))
}

// NewProductAlreadyExistsError creates a new product already exists error
func NewProductAlreadyExistsError(sku string) error {
	return appErrors.NewConflict(fmt.Sprintf("%s: %s", ErrProductAlreadyExistsMsg, sku))
}

// NewInsufficientInventoryError creates a new insufficient inventory error
func NewInsufficientInventoryError(sku string, requested, available int) error {
	return appErrors.NewBadRequest(
		fmt.Sprintf("%s: %s (requested: %d, available: %d)",
			ErrInsufficientInventoryMsg,
			sku,
			requested,
			available,
		),
	)
}
