package errs

import "errors"

// Product domain errors
var (
	ErrProductNotFound         = errors.New("product not found")
	ErrProductSKUAlreadyExists = errors.New("product with this SKU already exists")
	ErrInvalidProductPrice     = errors.New("invalid product price")
	ErrInvalidProductInventory = errors.New("invalid product inventory")
	ErrInvalidInput            = errors.New("invalid input")
)
