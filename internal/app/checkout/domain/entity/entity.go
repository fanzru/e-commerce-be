package entity

import (
	"time"

	"github.com/google/uuid"
)

// Checkout represents a completed checkout
type Checkout struct {
	ID            uuid.UUID           `json:"id"`
	CartID        uuid.UUID           `json:"cart_id"`
	Items         []*CheckoutItem     `json:"items"`
	Promotions    []*PromotionApplied `json:"promotions,omitempty"`
	Subtotal      float64             `json:"subtotal"`
	TotalDiscount float64             `json:"total_discount"`
	Total         float64             `json:"total"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// CheckoutItem represents an item in a checkout
type CheckoutItem struct {
	ID          uuid.UUID `json:"id"`
	CheckoutID  uuid.UUID `json:"checkout_id"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductSKU  string    `json:"product_sku"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Subtotal    float64   `json:"subtotal"`
	Discount    float64   `json:"discount"`
	Total       float64   `json:"total"`
}

// PromotionApplied represents a promotion applied to a checkout
type PromotionApplied struct {
	ID          uuid.UUID `json:"id"`
	CheckoutID  uuid.UUID `json:"checkout_id"`
	PromotionID uuid.UUID `json:"promotion_id"`
	Description string    `json:"description"`
	Discount    float64   `json:"discount"`
}

// NewCheckout creates a new checkout with the given parameters
func NewCheckout(cartID uuid.UUID, items []*CheckoutItem, subtotal, totalDiscount, total float64) *Checkout {
	return &Checkout{
		ID:            uuid.New(),
		CartID:        cartID,
		Items:         items,
		Subtotal:      subtotal,
		TotalDiscount: totalDiscount,
		Total:         total,
		CreatedAt:     time.Now(),
	}
}

// AddItem adds an item to the checkout
func (c *Checkout) AddItem(productID uuid.UUID, sku, name string, price float64, quantity int, discount float64) {
	subtotal := price * float64(quantity)
	total := subtotal - discount

	item := &CheckoutItem{
		ID:          uuid.New(),
		CheckoutID:  c.ID,
		ProductID:   productID,
		ProductSKU:  sku,
		ProductName: name,
		Quantity:    quantity,
		UnitPrice:   price,
		Subtotal:    subtotal,
		Discount:    discount,
		Total:       total,
	}

	c.Items = append(c.Items, item)
	c.Subtotal += subtotal
	c.TotalDiscount += discount
	c.Total += total
}

// CalculateTotal recalculates the checkout totals
func (c *Checkout) CalculateTotal() {
	subtotal := 0.0
	totalDiscount := 0.0

	for _, item := range c.Items {
		subtotal += item.Subtotal
		totalDiscount += item.Discount
	}

	c.Subtotal = subtotal
	c.TotalDiscount = totalDiscount
	c.Total = subtotal - totalDiscount
}
