package errs

import "errors"

// Checkout domain errors
var (
	ErrCheckoutNotFound        = errors.New("checkout not found")
	ErrCartNotFound            = errors.New("cart not found")
	ErrEmptyCart               = errors.New("cart is empty")
	ErrCartAlreadyCheckedOut   = errors.New("cart has already been checked out")
	ErrInsufficientStock       = errors.New("insufficient stock for one or more products")
	ErrInvalidPaymentStatus    = errors.New("invalid payment status")
	ErrInvalidOrderStatus      = errors.New("invalid order status")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrPaymentRequired         = errors.New("payment required for this operation")
)
