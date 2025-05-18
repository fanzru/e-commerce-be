package errs

import (
	"fmt"

	appErrors "github.com/fanzru/e-commerce-be/pkg/errors"
)

// Promotion domain error messages
const (
	ErrPromotionNotFoundMsg      = "promotion not found"
	ErrInvalidPromotionTypeMsg   = "invalid promotion type"
	ErrPromotionRuleParsingMsg   = "failed to parse promotion rule"
	ErrPromotionApplicationMsg   = "failed to apply promotion"
	ErrPromotionAlreadyExistsMsg = "promotion already exists"
)

// NewPromotionNotFoundError creates a new promotion not found error
func NewPromotionNotFoundError(id string) error {
	return appErrors.NewNotFound(fmt.Sprintf("%s: %s", ErrPromotionNotFoundMsg, id))
}

// NewInvalidPromotionTypeError creates a new invalid promotion type error
func NewInvalidPromotionTypeError(promotionType string) error {
	return appErrors.NewBadRequest(fmt.Sprintf("%s: %s", ErrInvalidPromotionTypeMsg, promotionType))
}

// NewPromotionRuleParsingError creates a new promotion rule parsing error
func NewPromotionRuleParsingError(promotionType string, err error) error {
	return appErrors.NewBadRequest(fmt.Sprintf("%s for type %s: %v", ErrPromotionRuleParsingMsg, promotionType, err))
}

// NewPromotionApplicationError creates a new promotion application error
func NewPromotionApplicationError(promotionID string, err error) error {
	return appErrors.NewInternalServerError(fmt.Errorf("%s for promotion %s: %w", ErrPromotionApplicationMsg, promotionID, err))
}

// NewPromotionAlreadyExistsError creates a new promotion already exists error
func NewPromotionAlreadyExistsError(description string) error {
	return appErrors.NewConflict(fmt.Sprintf("%s: %s", ErrPromotionAlreadyExistsMsg, description))
}
