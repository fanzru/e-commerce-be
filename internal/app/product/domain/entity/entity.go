package entity

import (
	"time"

	"github.com/google/uuid"
)

// Product represents a product entity
type Product struct {
	ID        uuid.UUID  `json:"id"`
	SKU       string     `json:"sku"`
	Name      string     `json:"name"`
	Price     float64    `json:"price"`
	Inventory int        `json:"inventory"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewProduct creates a new product with the given parameters
func NewProduct(sku, name string, price float64, inventory int) *Product {
	now := time.Now()
	return &Product{
		ID:        uuid.New(),
		SKU:       sku,
		Name:      name,
		Price:     price,
		Inventory: inventory,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// HasEnoughInventory checks if the product has enough inventory for the given quantity
func (p *Product) HasEnoughInventory(quantity int) bool {
	return p.Inventory >= quantity
}

// ReduceInventory reduces the inventory by the given quantity
func (p *Product) ReduceInventory(quantity int) {
	p.Inventory -= quantity
	p.UpdatedAt = time.Now()
}

// RestoreInventory increases the inventory by the given quantity
func (p *Product) RestoreInventory(quantity int) {
	p.Inventory += quantity
	p.UpdatedAt = time.Now()
}
