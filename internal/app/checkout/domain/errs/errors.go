package errs

import "errors"

// Checkout domain errors
var (
	ErrCheckoutNotFound      = errors.New("checkout not found")
	ErrCartNotFound          = errors.New("cart not found")
	ErrEmptyCart             = errors.New("cart is empty")
	ErrCartAlreadyCheckedOut = errors.New("cart has already been checked out")
	ErrInsufficientStock     = errors.New("insufficient stock for one or more products")
)
