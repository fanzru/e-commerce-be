package entity

import (
	"time"

	"github.com/google/uuid"
)

// CartItem represents an item in a cart
type CartItem struct {
	ID          uuid.UUID  `json:"id"`
	CartID      uuid.UUID  `json:"cart_id"`
	ProductID   uuid.UUID  `json:"product_id"`
	ProductSKU  string     `json:"product_sku"`
	ProductName string     `json:"product_name"`
	UnitPrice   float64    `json:"unit_price"`
	Quantity    int        `json:"quantity"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// Cart represents a shopping cart
type Cart struct {
	ID        uuid.UUID   `json:"id"`
	UserID    uuid.UUID   `json:"user_id"`
	Items     []*CartItem `json:"items"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	DeletedAt *time.Time  `json:"deleted_at,omitempty"`
}

// NewCart creates a new empty cart
func NewCart() *Cart {
	now := time.Now()
	return &Cart{
		ID:        uuid.New(),
		Items:     []*CartItem{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewCartWithUser creates a new empty cart associated with a user
func NewCartWithUser(userID uuid.UUID) *Cart {
	cart := NewCart()
	cart.UserID = userID
	return cart
}

// AddItem adds a product to the cart or updates its quantity if it already exists
func (c *Cart) AddItem(productID uuid.UUID, sku, name string, price float64, quantity int) *CartItem {
	// Check if the item already exists in the cart
	for i, item := range c.Items {
		if item.ProductID == productID {
			// Update existing item
			c.Items[i].Quantity += quantity
			c.Items[i].UpdatedAt = time.Now()
			c.UpdatedAt = time.Now()
			return c.Items[i]
		}
	}

	// Create new item
	item := &CartItem{
		ID:          uuid.New(),
		CartID:      c.ID,
		ProductID:   productID,
		ProductSKU:  sku,
		ProductName: name,
		UnitPrice:   price,
		Quantity:    quantity,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	c.Items = append(c.Items, item)
	c.UpdatedAt = time.Now()
	return item
}

// UpdateItemQuantity updates the quantity of an item in the cart
func (c *Cart) UpdateItemQuantity(productID uuid.UUID, quantity int) bool {
	for i, item := range c.Items {
		if item.ProductID == productID {
			if quantity <= 0 {
				// Remove item if quantity is 0 or negative
				return c.RemoveItem(productID)
			}
			c.Items[i].Quantity = quantity
			c.Items[i].UpdatedAt = time.Now()
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// RemoveItem removes an item from the cart
func (c *Cart) RemoveItem(productID uuid.UUID) bool {
	for i, item := range c.Items {
		if item.ProductID == productID {
			// Remove item by replacing it with the last element and truncating
			c.Items[i] = c.Items[len(c.Items)-1]
			c.Items = c.Items[:len(c.Items)-1]
			c.UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// GetItem gets an item from the cart by product ID
func (c *Cart) GetItem(productID uuid.UUID) *CartItem {
	for i, item := range c.Items {
		if item.ProductID == productID {
			return c.Items[i]
		}
	}
	return nil
}

// Clear removes all items from the cart
func (c *Cart) Clear() {
	c.Items = []*CartItem{}
	c.UpdatedAt = time.Now()
}

// IsEmpty checks if the cart is empty
func (c *Cart) IsEmpty() bool {
	return len(c.Items) == 0
}

// Count returns the number of items in the cart
func (c *Cart) Count() int {
	return len(c.Items)
}

// TotalItems returns the total quantity of all items in the cart
func (c *Cart) TotalItems() int {
	total := 0
	for _, item := range c.Items {
		total += item.Quantity
	}
	return total
}

// Subtotal calculates the subtotal of all items in the cart (before promotions)
func (c *Cart) Subtotal() float64 {
	subtotal := 0.0
	for _, item := range c.Items {
		subtotal += item.UnitPrice * float64(item.Quantity)
	}
	return subtotal
}
