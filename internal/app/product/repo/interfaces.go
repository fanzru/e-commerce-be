package repo

import (
	"context"

	"github.com/fanzru/e-commerce-be/internal/app/product/domain/entity"
	"github.com/google/uuid"
)

// ProductRepository defines the interface for product repository
type ProductRepository interface {
	// GetByID retrieves a product by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)

	// List retrieves a list of products with pagination and filtering
	List(ctx context.Context, page, limit int, sku, name string) ([]*entity.Product, int, error)

	// Create creates a new product
	Create(ctx context.Context, product *entity.Product) error

	// Update updates an existing product
	Update(ctx context.Context, product *entity.Product) error

	// Delete deletes a product by its ID
	Delete(ctx context.Context, id uuid.UUID) error
}
