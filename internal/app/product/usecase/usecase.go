package usecase

import (
	"context"

	"github.com/fanzru/e-commerce-be/internal/app/product/domain/entity"
	"github.com/google/uuid"
)

// ProductUseCase defines the interface for product use cases
type ProductUseCase interface {
	// List returns a list of products with pagination and filtering
	List(ctx context.Context, page, limit int, sku, name string) ([]*entity.Product, int, error)

	// GetByID returns a product by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)

	// Create creates a new product
	Create(ctx context.Context, sku, name string, price float64, inventory int) (*entity.Product, error)

	// Update updates an existing product
	Update(ctx context.Context, id uuid.UUID, name string, price float64, inventory int) (*entity.Product, error)

	// Delete deletes a product
	Delete(ctx context.Context, id uuid.UUID) error
}
