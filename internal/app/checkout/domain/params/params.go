package params

import (
	"github.com/google/uuid"
)

// CheckoutRequest defines the parameters for creating a checkout
type CheckoutRequest struct {
	CartID uuid.UUID `json:"cart_id" binding:"required"`
}

// CheckoutItemResponse defines the response structure for a checkout item
type CheckoutItemResponse struct {
	ProductID uuid.UUID `json:"product_id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Subtotal  float64   `json:"subtotal"`
	Discount  float64   `json:"discount"`
	Total     float64   `json:"total"`
}

// PromotionAppliedResponse defines the response structure for an applied promotion
type PromotionAppliedResponse struct {
	Description string  `json:"description"`
	Discount    float64 `json:"discount"`
}

// CheckoutResponse defines the response structure for a checkout
type CheckoutResponse struct {
	ID            uuid.UUID                  `json:"id"`
	CartID        uuid.UUID                  `json:"cart_id"`
	Items         []CheckoutItemResponse     `json:"items"`
	Promotions    []PromotionAppliedResponse `json:"promotions"`
	Subtotal      float64                    `json:"subtotal"`
	TotalDiscount float64                    `json:"total_discount"`
	Total         float64                    `json:"total"`
	CreatedAt     string                     `json:"created_at"`
}
