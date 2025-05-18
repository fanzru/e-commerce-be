package entity

import (
	"encoding/json"

	cartEntity "github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	"github.com/google/uuid"
)

// ApplicablePromotion represents a promotion that can be applied to a cart
type ApplicablePromotion struct {
	ID          uuid.UUID     `json:"id"`
	Type        PromotionType `json:"type"`
	Description string        `json:"description"`
	Discount    float64       `json:"discount"`
}

// ConvertCartToPromotionItems converts cart items to promotion cart items
func ConvertCartToPromotionItems(cartItems []*cartEntity.CartItemInfo) []CartItem {
	promotionItems := make([]CartItem, 0, len(cartItems))
	for _, item := range cartItems {
		promotionItems = append(promotionItems, CartItem{
			ProductID:   item.ProductID,
			ProductSKU:  item.ProductSKU,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		})
	}
	return promotionItems
}

// BuildSKUMap creates a map of SKUs in the cart for quick lookup
func BuildSKUMap(items []CartItem) map[string]bool {
	skuMap := make(map[string]bool)
	for _, item := range items {
		skuMap[item.ProductSKU] = true
	}
	return skuMap
}

// IsPromotionApplicableToCart checks if a promotion is applicable to the cart based on SKUs
func IsPromotionApplicableToCart(promotion *Promotion, skuMap map[string]bool) bool {
	switch promotion.Type {
	case BuyOneGetOneFree:
		var rule struct {
			TriggerSKU      string `json:"trigger_sku"`
			FreeSKU         string `json:"free_sku"`
			TriggerQuantity int    `json:"trigger_quantity"`
			FreeQuantity    int    `json:"free_quantity"`
		}
		if err := json.Unmarshal(promotion.Rule, &rule); err == nil {
			return skuMap[rule.TriggerSKU] && skuMap[rule.FreeSKU]
		}
	case Buy3Pay2:
		var rule struct {
			SKU                 string `json:"sku"`
			MinQuantity         int    `json:"min_quantity"`
			PaidQuantityDivisor int    `json:"paid_quantity_divisor"`
			FreeQuantityDivisor int    `json:"free_quantity_divisor"`
		}

		err := json.Unmarshal(promotion.Rule, &rule)
		if err != nil {
			return false
		}
		return skuMap[rule.SKU]
	case BulkDiscount:
		var rule struct {
			SKU                string  `json:"sku"`
			MinQuantity        int     `json:"min_quantity"`
			DiscountPercentage float64 `json:"discount_percentage"`
		}
		if err := json.Unmarshal(promotion.Rule, &rule); err == nil {
			return skuMap[rule.SKU]
		}
	}
	return false
}

// GetApplicablePromotions returns all promotions that are applicable to the cart
func GetApplicablePromotions(promotions []*Promotion, cartItems []CartItem) []ApplicablePromotion {
	if len(cartItems) == 0 {
		return nil
	}

	skuMap := BuildSKUMap(cartItems)
	applicablePromotions := make([]ApplicablePromotion, 0)

	for _, promotion := range promotions {
		if !promotion.Active {
			continue
		}

		if IsPromotionApplicableToCart(promotion, skuMap) {
			discount, err := promotion.ApplyToCart(cartItems)
			if err != nil {
				continue
			}

			applicablePromotions = append(applicablePromotions, ApplicablePromotion{
				ID:          promotion.ID,
				Type:        promotion.Type,
				Description: promotion.Description,
				Discount:    discount,
			})
		}
	}

	return applicablePromotions
}

// CalculateTotalDiscount calculates the total discount from applicable promotions
func CalculateTotalDiscount(applicablePromotions []ApplicablePromotion) float64 {
	totalDiscount := 0.0
	for _, promo := range applicablePromotions {
		if promo.Discount > 0 {
			totalDiscount += promo.Discount
		}
	}
	return totalDiscount
}
