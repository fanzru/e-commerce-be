package entity

import (
	"github.com/google/uuid"
)

// ApplyPromotionsToCart adds promotion information to a cart
func ApplyPromotionsToCart(
	cart *CartInfo,
	promotions []ApplicablePromotion,
	totalDiscount float64,
) *CartInfo {
	// Apply promotion data to cart
	cart.ApplicablePromotions = promotions
	cart.PotentialDiscount = totalDiscount
	cart.PotentialTotal = cart.Subtotal - totalDiscount
	if cart.PotentialTotal < 0 {
		cart.PotentialTotal = 0
	}
	return cart
}

// ConvertPromotionData converts promotion data from promotion domain to cart domain
func ConvertPromotionData(
	promotionID uuid.UUID,
	promotionType string,
	description string,
	discount float64,
) ApplicablePromotion {
	return ApplicablePromotion{
		ID:          promotionID,
		Type:        promotionType,
		Description: description,
		Discount:    discount,
	}
}

// PromotionDiscount represents the promotion type from promotion domain
type PromotionDiscount struct {
	PromotionID   uuid.UUID `json:"promotion_id"`
	PromotionType string    `json:"promotion_type"`
	Description   string    `json:"description"`
	Discount      float64   `json:"discount"`
}

// ConvertPromotionDiscounts converts a list of promotion discounts to cart applicable promotions
func ConvertPromotionDiscounts(discounts []PromotionDiscount) []ApplicablePromotion {
	promotions := make([]ApplicablePromotion, 0, len(discounts))
	for _, p := range discounts {
		promotions = append(promotions, ConvertPromotionData(
			p.PromotionID,
			p.PromotionType,
			p.Description,
			p.Discount,
		))
	}
	return promotions
}
