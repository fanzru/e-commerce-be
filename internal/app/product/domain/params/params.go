package params

import (
	"github.com/google/uuid"
)

// CreateProductParams defines the parameters for creating a product
type CreateProductParams struct {
	SKU       string  `json:"sku" binding:"required"`
	Name      string  `json:"name" binding:"required"`
	Price     float64 `json:"price" binding:"required,gt=0"`
	Inventory int     `json:"inventory" binding:"required,gte=0"`
}

// UpdateProductParams defines the parameters for updating a product
type UpdateProductParams struct {
	Name      string  `json:"name" binding:"omitempty"`
	Price     float64 `json:"price" binding:"omitempty,gt=0"`
	Inventory int     `json:"inventory" binding:"omitempty,gte=0"`
}

// ProductResponse defines the response structure for a product
type ProductResponse struct {
	ID        uuid.UUID `json:"id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Inventory int       `json:"inventory"`
}

// ProductListResponse defines the response structure for a list of products
type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int               `json:"total"`
}

// ProductQueryParams defines the query parameters for listing products
type ProductQueryParams struct {
	Page  int    `form:"page" binding:"omitempty,min=1"`
	Limit int    `form:"limit" binding:"omitempty,min=1,max=100"`
	SKU   string `form:"sku" binding:"omitempty"`
	Name  string `form:"name" binding:"omitempty"`
}
