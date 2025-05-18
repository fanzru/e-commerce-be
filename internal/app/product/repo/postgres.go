package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/fanzru/e-commerce-be/internal/app/product/domain/entity"
	domainErrors "github.com/fanzru/e-commerce-be/internal/app/product/domain/errs"
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
			return nil, domainErrors.ErrProductNotFound
		}
		return nil, fmt.Errorf("error querying product by ID: %w", err)
	}

	return &product, nil
}

// List retrieves a list of products with pagination and filtering
func (r *ProductPostgresRepository) List(ctx context.Context, page, limit int, sku, name string) ([]*entity.Product, int, error) {
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
	}

	if name != "" {
		whereClause += fmt.Sprintf(" AND name ILIKE $%d", argPos)
		args = append(args, "%"+name+"%")
		argPos++
	}

	// Count total matches first
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM products %s`, whereClause)
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
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
			return nil, 0, fmt.Errorf("error scanning product row: %w", err)
		}
		products = append(products, &product)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating product rows: %w", err)
	}

	return products, total, nil
}

// Create creates a new product
func (r *ProductPostgresRepository) Create(ctx context.Context, product *entity.Product) error {
	// Check if SKU already exists
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE sku = $1 AND deleted_at IS NULL)", product.SKU).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking SKU existence: %w", err)
	}

	if exists {
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
			return domainErrors.ErrProductSKUAlreadyExists
		}
		return fmt.Errorf("error creating product: %w", err)
	}

	return nil
}

// Update updates an existing product
func (r *ProductPostgresRepository) Update(ctx context.Context, product *entity.Product) error {
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
		return fmt.Errorf("error updating product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrProductNotFound
	}

	return nil
}

// Delete deletes a product by its ID (soft delete)
func (r *ProductPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE products
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domainErrors.ErrProductNotFound
	}

	return nil
}
