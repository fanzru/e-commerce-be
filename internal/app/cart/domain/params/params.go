package params

import (
	"github.com/google/uuid"
)

// AddItemParams defines the parameters for adding an item to a cart
type AddItemParams struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,gt=0"`
}

// UpdateItemQuantityParams defines the parameters for updating an item quantity
type UpdateItemQuantityParams struct {
	Quantity int `json:"quantity" binding:"required,gte=0"`
}

// CartItemResponse defines the response structure for a cart item
type CartItemResponse struct {
	ID        uuid.UUID `json:"id"`
	ProductID uuid.UUID `json:"product_id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Subtotal  float64   `json:"subtotal"`
}

// CartResponse defines the response structure for a cart
type CartResponse struct {
	ID         uuid.UUID          `json:"id"`
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	Subtotal   float64            `json:"subtotal"`
}
