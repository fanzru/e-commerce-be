package errs

import "errors"

// Promotion domain errors
var (
	ErrPromotionNotFound         = errors.New("promotion not found")
	ErrInvalidPromotionType      = errors.New("invalid promotion type")
	ErrInvalidDiscountPercentage = errors.New("invalid discount percentage")
	ErrInvalidMinQuantity        = errors.New("invalid minimum quantity")
	ErrDuplicatePromotion        = errors.New("promotion with this configuration already exists")
)
