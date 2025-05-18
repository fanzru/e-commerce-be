package errs

import "errors"

// Cart domain errors
var (
	ErrCartNotFound      = errors.New("cart not found")
	ErrCartItemNotFound  = errors.New("cart item not found")
	ErrProductNotFound   = errors.New("product not found")
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrInsufficientStock = errors.New("insufficient stock")
)
