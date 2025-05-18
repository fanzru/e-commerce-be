package entity

import (
	"time"

	"github.com/google/uuid"
)

// PaymentStatus represents the status of a payment
type PaymentStatus string

// OrderStatus represents the status of an order
type OrderStatus string

const (
	// Payment statuses
	PaymentStatusPending  PaymentStatus = "PENDING"
	PaymentStatusPaid     PaymentStatus = "PAID"
	PaymentStatusFailed   PaymentStatus = "FAILED"
	PaymentStatusRefunded PaymentStatus = "REFUNDED"

	// Order statuses
	OrderStatusCreated    OrderStatus = "CREATED"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusShipped    OrderStatus = "SHIPPED"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

// Checkout represents a completed checkout/order
type Checkout struct {
	ID               uuid.UUID           `json:"id"`
	UserID           *uuid.UUID          `json:"user_id,omitempty"`
	Items            []*CheckoutItem     `json:"items"`
	Promotions       []*PromotionApplied `json:"promotions,omitempty"`
	Subtotal         float64             `json:"subtotal"`
	TotalDiscount    float64             `json:"total_discount"`
	Total            float64             `json:"total"`
	PaymentStatus    PaymentStatus       `json:"payment_status"`
	PaymentMethod    *string             `json:"payment_method,omitempty"`
	PaymentReference *string             `json:"payment_reference,omitempty"`
	Notes            *string             `json:"notes,omitempty"`
	Status           OrderStatus         `json:"status"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	CompletedAt      *time.Time          `json:"completed_at,omitempty"`
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
func NewCheckout(userID *uuid.UUID, items []*CheckoutItem, subtotal, totalDiscount, total float64) *Checkout {
	return &Checkout{
		ID:            uuid.New(),
		UserID:        userID,
		Items:         items,
		Subtotal:      subtotal,
		TotalDiscount: totalDiscount,
		Total:         total,
		PaymentStatus: PaymentStatusPending,
		Status:        OrderStatusCreated,
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

// SetPaymentStatus updates the payment status
func (c *Checkout) SetPaymentStatus(status PaymentStatus) {
	c.PaymentStatus = status
	c.UpdatedAt = time.Now()

	// If payment is marked as paid, update order status
	if status == PaymentStatusPaid && c.Status == OrderStatusCreated {
		c.Status = OrderStatusProcessing
	}
}

// SetOrderStatus updates the order status
func (c *Checkout) SetOrderStatus(status OrderStatus) {
	c.Status = status
	c.UpdatedAt = time.Now()

	// Set completed time if order is delivered
	if status == OrderStatusDelivered && c.CompletedAt == nil {
		now := time.Now()
		c.CompletedAt = &now
	}
}

// MarkAsPaid marks the checkout as paid
func (c *Checkout) MarkAsPaid(paymentMethod, paymentReference string) {
	c.PaymentStatus = PaymentStatusPaid
	c.PaymentMethod = &paymentMethod
	c.PaymentReference = &paymentReference
	c.UpdatedAt = time.Now()

	// Update order status to processing
	c.Status = OrderStatusProcessing
}
