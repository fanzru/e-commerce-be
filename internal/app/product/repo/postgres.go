package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/product/domain/entity"
	domainErrors "github.com/fanzru/e-commerce-be/internal/app/product/domain/errs"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/google/uuid"
)

// ProductPostgresRepository implements ProductRepository using PostgreSQL
type ProductPostgresRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sql.DB) ProductRepository {
	return &ProductPostgresRepository{
		db: db,
	}
}

// GetByID retrieves a product by its ID
func (r *ProductPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	logger := middleware.Logger.With(
		"method", "ProductRepository.GetByID",
		"product_id", id.String(),
	)
	logger.Debug("Fetching product by ID")
	startTime := time.Now()

	query := `
		SELECT id, sku, name, price, inventory
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`

	var product entity.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.Price,
		&product.Inventory,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Product not found", "error", "ErrProductNotFound")
			return nil, domainErrors.ErrProductNotFound
		}
		logger.Error("Failed to query product by ID", "error", err.Error())
		return nil, fmt.Errorf("error querying product by ID: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved product",
		"sku", product.SKU,
		"name", product.Name,
		"price", product.Price,
		"inventory", product.Inventory,
		"duration_ms", duration.Milliseconds())

	return &product, nil
}

// List retrieves a list of products with pagination and filtering
func (r *ProductPostgresRepository) List(ctx context.Context, page, limit int, sku, name string) ([]*entity.Product, int, error) {
	logger := middleware.Logger.With(
		"method", "ProductRepository.List",
		"page", page,
		"limit", limit,
	)
	logger.Debug("Listing products with filters")
	startTime := time.Now()

	offset := (page - 1) * limit

	// Base query for filtering
	whereClause := "WHERE deleted_at IS NULL"
	args := []interface{}{}
	argPos := 1

	// Add filters if provided
	if sku != "" {
		whereClause += fmt.Sprintf(" AND sku ILIKE $%d", argPos)
		args = append(args, "%"+sku+"%")
		argPos++
		logger = logger.With("filter_sku", sku)
	}

	if name != "" {
		whereClause += fmt.Sprintf(" AND name ILIKE $%d", argPos)
		args = append(args, "%"+name+"%")
		argPos++
		logger = logger.With("filter_name", name)
	}

	// Count total matches first
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM products %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		logger.Error("Failed to count products", "error", err.Error())
		return nil, 0, fmt.Errorf("error counting products: %w", err)
	}

	// Now fetch the actual data with pagination
	query := fmt.Sprintf(`
		SELECT id, sku, name, price, inventory
		FROM products
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argPos, argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Error("Failed to query products", "error", err.Error())
		return nil, 0, fmt.Errorf("error querying products: %w", err)
	}
	defer rows.Close()

	products := []*entity.Product{}
	for rows.Next() {
		var product entity.Product
		err := rows.Scan(
			&product.ID,
			&product.SKU,
			&product.Name,
			&product.Price,
			&product.Inventory,
		)
		if err != nil {
			logger.Error("Failed to scan product row", "error", err.Error())
			return nil, 0, fmt.Errorf("error scanning product row: %w", err)
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Failed to iterate product rows", "error", err.Error())
		return nil, 0, fmt.Errorf("error iterating product rows: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully listed products",
		"total_count", total,
		"returned_count", len(products),
		"duration_ms", duration.Milliseconds())

	return products, total, nil
}

// Create creates a new product
func (r *ProductPostgresRepository) Create(ctx context.Context, product *entity.Product) error {
	logger := middleware.Logger.With(
		"method", "ProductRepository.Create",
		"sku", product.SKU,
		"name", product.Name,
	)
	logger.Debug("Creating new product")
	startTime := time.Now()

	// Check if SKU already exists
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE sku = $1 AND deleted_at IS NULL)", product.SKU).Scan(&exists)
	if err != nil {
		logger.Error("Failed to check SKU existence", "error", err.Error())
		return fmt.Errorf("error checking SKU existence: %w", err)
	}

	if exists {
		logger.Warn("Product SKU already exists", "error", "ErrProductSKUAlreadyExists")
		return domainErrors.ErrProductSKUAlreadyExists
	}

	// Generate a new UUID if not provided
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	query := `
		INSERT INTO products (id, sku, name, price, inventory)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = r.db.ExecContext(ctx, query,
		product.ID,
		product.SKU,
		product.Name,
		product.Price,
		product.Inventory,
	)

	if err != nil {
		// Check for unique constraint violation on SKU
		if strings.Contains(err.Error(), "unique constraint") && strings.Contains(err.Error(), "sku") {
			logger.Warn("Product SKU already exists (constraint violation)", "error", "ErrProductSKUAlreadyExists")
			return domainErrors.ErrProductSKUAlreadyExists
		}
		logger.Error("Failed to create product", "error", err.Error())
		return fmt.Errorf("error creating product: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully created product",
		"product_id", product.ID.String(),
		"price", product.Price,
		"inventory", product.Inventory,
		"duration_ms", duration.Milliseconds())

	return nil
}

// Update updates an existing product
func (r *ProductPostgresRepository) Update(ctx context.Context, product *entity.Product) error {
	logger := middleware.Logger.With(
		"method", "ProductRepository.Update",
		"product_id", product.ID.String(),
	)
	logger.Debug("Updating product")
	startTime := time.Now()

	query := `
		UPDATE products
		SET name = $1, price = $2, inventory = $3, updated_at = NOW()
		WHERE id = $4 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Price,
		product.Inventory,
		product.ID,
	)

	if err != nil {
		logger.Error("Failed to update product", "error", err.Error())
		return fmt.Errorf("error updating product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err.Error())
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("Product not found", "error", "ErrProductNotFound")
		return domainErrors.ErrProductNotFound
	}

	duration := time.Since(startTime)
	logger.Info("Successfully updated product",
		"name", product.Name,
		"price", product.Price,
		"inventory", product.Inventory,
		"duration_ms", duration.Milliseconds())

	return nil
}

// Delete deletes a product by its ID (soft delete)
func (r *ProductPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	logger := middleware.Logger.With(
		"method", "ProductRepository.Delete",
		"product_id", id.String(),
	)
	logger.Debug("Deleting product")
	startTime := time.Now()

	query := `
		UPDATE products
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Error("Failed to delete product", "error", err.Error())
		return fmt.Errorf("error deleting product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err.Error())
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Warn("Product not found", "error", "ErrProductNotFound")
		return domainErrors.ErrProductNotFound
	}

	duration := time.Since(startTime)
	logger.Info("Successfully deleted product",
		"duration_ms", duration.Milliseconds())

	return nil
}
