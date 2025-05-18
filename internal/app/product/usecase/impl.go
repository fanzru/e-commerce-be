package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/product/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/product/repo"
	"github.com/google/uuid"
)

// productUseCase implements the ProductUseCase interface
type productUseCase struct {
	repo repo.ProductRepository
}

// NewProductUseCase creates a new instance of productUseCase
func NewProductUseCase(repo repo.ProductRepository) ProductUseCase {
	return &productUseCase{
		repo: repo,
	}
}

// List returns a list of products with pagination and filtering
func (u *productUseCase) List(ctx context.Context, page, limit int, sku, name string) ([]*entity.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	return u.repo.List(ctx, page, limit, sku, name)
}

// GetByID returns a product by its ID
func (u *productUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}

	return u.repo.GetByID(ctx, id)
}

// Create creates a new product
func (u *productUseCase) Create(ctx context.Context, sku, name string, price float64, inventory int) (*entity.Product, error) {
	if sku == "" {
		return nil, errors.New("SKU is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	if price <= 0 {
		return nil, errors.New("price must be greater than zero")
	}
	if inventory < 0 {
		return nil, errors.New("inventory cannot be negative")
	}

	product := entity.NewProduct(sku, name, price, inventory)
	err := u.repo.Create(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

// Update updates an existing product
func (u *productUseCase) Update(ctx context.Context, id uuid.UUID, name string, price float64, inventory int) (*entity.Product, error) {
	if id == uuid.Nil {
		return nil, errors.New("invalid product ID")
	}

	product, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if name != "" {
		product.Name = name
	}
	if price > 0 {
		product.Price = price
	}
	if inventory >= 0 {
		product.Inventory = inventory
	}

	product.UpdatedAt = time.Now()

	err = u.repo.Update(ctx, product)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

// Delete deletes a product
func (u *productUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return errors.New("invalid product ID")
	}

	return u.repo.Delete(ctx, id)
}
