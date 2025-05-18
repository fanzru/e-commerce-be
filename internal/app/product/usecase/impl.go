package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/product/domain/entity"
	"github.com/fanzru/e-commerce-be/internal/app/product/domain/errs"
	"github.com/fanzru/e-commerce-be/internal/app/product/repo"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/google/uuid"
)

// productUseCase implements the ProductUseCase interface
type productUseCase struct {
	productRepo repo.ProductRepository
}

// NewProductUseCase creates a new instance of productUseCase
func NewProductUseCase(productRepo repo.ProductRepository) ProductUseCase {
	return &productUseCase{
		productRepo: productRepo,
	}
}

// List returns a list of products with pagination and filtering
func (u *productUseCase) List(ctx context.Context, page, limit int, sku, name string) ([]*entity.Product, int, error) {
	logger := middleware.Logger.With(
		"method", "ProductUseCase.List",
		"page", page,
		"limit", limit,
	)
	if sku != "" {
		logger = logger.With("sku", sku)
	}
	if name != "" {
		logger = logger.With("name", name)
	}
	logger.Info("Listing products with filters")
	startTime := time.Now()

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	products, total, err := u.productRepo.List(ctx, page, limit, sku, name)
	if err != nil {
		logger.Error("Failed to list products", "error", err.Error())
		return nil, 0, fmt.Errorf("error listing products: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully listed products",
		"total", total,
		"returned", len(products),
		"duration_ms", duration.Milliseconds())

	return products, total, nil
}

// GetByID returns a product by its ID
func (u *productUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	logger := middleware.Logger.With(
		"method", "ProductUseCase.GetByID",
		"product_id", id.String(),
	)
	logger.Info("Getting product by ID")
	startTime := time.Now()

	product, err := u.productRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get product", "error", err.Error())
		return nil, fmt.Errorf("error getting product: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved product",
		"sku", product.SKU,
		"name", product.Name,
		"price", product.Price,
		"inventory", product.Inventory,
		"duration_ms", duration.Milliseconds())

	return product, nil
}

// Create creates a new product
func (u *productUseCase) Create(ctx context.Context, sku, name string, price float64, inventory int) (*entity.Product, error) {
	logger := middleware.Logger.With(
		"method", "ProductUseCase.Create",
		"sku", sku,
		"name", name,
		"price", price,
		"inventory", inventory,
	)
	logger.Info("Creating new product")
	startTime := time.Now()

	// Validate input
	if sku == "" {
		logger.Warn("Invalid input: Empty SKU", "error", "ErrInvalidInput")
		return nil, errs.ErrInvalidInput
	}
	if name == "" {
		logger.Warn("Invalid input: Empty name", "error", "ErrInvalidInput")
		return nil, errs.ErrInvalidInput
	}
	if price <= 0 {
		logger.Warn("Invalid input: Price must be positive", "error", "ErrInvalidInput")
		return nil, errs.ErrInvalidInput
	}
	if inventory < 0 {
		logger.Warn("Invalid input: Inventory must be non-negative", "error", "ErrInvalidInput")
		return nil, errs.ErrInvalidInput
	}

	// Create product entity
	product := &entity.Product{
		ID:        uuid.New(),
		SKU:       sku,
		Name:      name,
		Price:     price,
		Inventory: inventory,
	}

	// Save to repository
	err := u.productRepo.Create(ctx, product)
	if err != nil {
		logger.Error("Failed to create product", "error", err.Error())
		return nil, fmt.Errorf("error creating product: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully created product",
		"product_id", product.ID.String(),
		"duration_ms", duration.Milliseconds())

	return product, nil
}

// Update updates an existing product
func (u *productUseCase) Update(ctx context.Context, id uuid.UUID, name string, price float64, inventory int) (*entity.Product, error) {
	logger := middleware.Logger.With(
		"method", "ProductUseCase.Update",
		"product_id", id.String(),
		"name", name,
		"price", price,
		"inventory", inventory,
	)
	logger.Info("Updating product")
	startTime := time.Now()

	// Validate input
	if name == "" {
		logger.Warn("Invalid input: Empty name", "error", "ErrInvalidInput")
		return nil, errs.ErrInvalidInput
	}
	if price <= 0 {
		logger.Warn("Invalid input: Price must be positive", "error", "ErrInvalidInput")
		return nil, errs.ErrInvalidInput
	}
	if inventory < 0 {
		logger.Warn("Invalid input: Inventory must be non-negative", "error", "ErrInvalidInput")
		return nil, errs.ErrInvalidInput
	}

	// Get existing product
	product, err := u.productRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get product for update", "error", err.Error())
		return nil, fmt.Errorf("error getting product for update: %w", err)
	}

	// Update fields
	product.Name = name
	product.Price = price
	product.Inventory = inventory

	// Save to repository
	err = u.productRepo.Update(ctx, product)
	if err != nil {
		logger.Error("Failed to update product", "error", err.Error())
		return nil, fmt.Errorf("error updating product: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully updated product",
		"duration_ms", duration.Milliseconds())

	return product, nil
}

// Delete deletes a product
func (u *productUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	logger := middleware.Logger.With(
		"method", "ProductUseCase.Delete",
		"product_id", id.String(),
	)
	logger.Info("Deleting product")
	startTime := time.Now()

	err := u.productRepo.Delete(ctx, id)
	if err != nil {
		logger.Error("Failed to delete product", "error", err.Error())
		return fmt.Errorf("error deleting product: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully deleted product",
		"duration_ms", duration.Milliseconds())

	return nil
}
