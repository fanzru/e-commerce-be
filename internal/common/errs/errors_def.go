package errs

import (
	"errors"
	"net/http"
)

// Domain-specific error definitions
var (
	// Cart errors
	ErrCartNotFound     = NewNotFound("Cart not found")
	ErrCartItemNotFound = NewNotFound("Cart item not found")
	ErrCartEmpty        = NewBadRequest("Cart is empty")

	// Product errors
	ErrProductNotFound   = NewNotFound("Product not found")
	ErrProductOutOfStock = New(
		errors.New("product out of stock"),
		"product_out_of_stock",
		http.StatusConflict,
		"The requested product is out of stock",
	)

	// User errors
	ErrUserNotFound       = NewNotFound("User not found")
	ErrUserAlreadyExists  = NewConflict("User already exists with this email")
	ErrInvalidCredentials = NewUnauthorized("Invalid username or password")

	// Checkout errors
	ErrCheckoutFailed = New(
		errors.New("checkout failed"),
		"checkout_failed",
		http.StatusBadRequest,
		"Failed to process checkout",
	)

	// Promotion errors
	ErrPromotionNotFound = NewNotFound("Promotion not found")
	ErrPromotionExpired  = New(
		errors.New("promotion expired"),
		"promotion_expired",
		http.StatusBadRequest,
		"The promotion has expired",
	)
	ErrPromotionNotApplicable = New(
		errors.New("promotion not applicable"),
		"promotion_not_applicable",
		http.StatusBadRequest,
		"This promotion is not applicable to your order",
	)
)
