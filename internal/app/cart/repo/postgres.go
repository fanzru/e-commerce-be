package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/cart/domain/entity"
	domainErrors "github.com/fanzru/e-commerce-be/internal/app/cart/domain/errs"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/google/uuid"
)

// CartPostgresRepository implements CartRepository using PostgreSQL
type CartPostgresRepository struct {
	db *sql.DB
}

// NewCartRepository creates a new cart repository
func NewCartRepository(db *sql.DB) CartRepository {
	return &CartPostgresRepository{
		db: db,
	}
}

// GetByID retrieves a cart by its ID with all items
func (r *CartPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Cart, error) {
	logger := middleware.Logger.With(
		"method", "CartRepository.GetByID",
		"cart_id", id.String(),
	)
	logger.Debug("Fetching cart by ID")
	startTime := time.Now()

	// First, get the cart
	cartQuery := `
		SELECT id, user_id, created_at, updated_at
		FROM carts
		WHERE id = $1 AND deleted_at IS NULL
	`

	var cart entity.Cart
	err := r.db.QueryRowContext(ctx, cartQuery, id).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Cart not found", "error", "ErrCartNotFound")
			return nil, domainErrors.ErrCartNotFound
		}
		logger.Error("Failed to query cart by ID", "error", err.Error())
		return nil, fmt.Errorf("error querying cart by ID: %w", err)
	}

	logger.Debug("Cart found, fetching cart items", "user_id", cart.UserID)

	// Then, get all cart items with product details
	itemsQuery := `
		SELECT 
			ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
			p.sku, p.name, p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = $1 AND ci.deleted_at IS NULL
		ORDER BY ci.created_at
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		logger.Error("Failed to query cart items", "error", err.Error())
		return nil, fmt.Errorf("error querying cart items: %w", err)
	}
	defer rows.Close()

	cart.Items = []*entity.CartItem{}
	itemCount := 0
	for rows.Next() {
		var item entity.CartItem
		err := rows.Scan(
			&item.ID,
			&item.CartID,
			&item.ProductID,
			&item.Quantity,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.ProductSKU,
			&item.ProductName,
			&item.UnitPrice,
		)
		if err != nil {
			logger.Error("Failed to scan cart item", "error", err.Error())
			return nil, fmt.Errorf("error scanning cart item row: %w", err)
		}
		cart.Items = append(cart.Items, &item)
		itemCount++
	}

	if err = rows.Err(); err != nil {
		logger.Error("Failed to iterate cart items", "error", err.Error())
		return nil, fmt.Errorf("error iterating cart item rows: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved cart with items",
		"user_id", cart.UserID,
		"item_count", itemCount,
		"duration_ms", duration.Milliseconds())

	return &cart, nil
}

// Create creates a new empty cart
func (r *CartPostgresRepository) Create(ctx context.Context, cart *entity.Cart) error {
	// Generate a new UUID if not provided
	if cart.ID == uuid.Nil {
		cart.ID = uuid.New()
	}

	query := `
		INSERT INTO carts (id, user_id)
		VALUES ($1, $2)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, cart.ID, cart.UserID).Scan(
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("error creating cart: %w", err)
	}

	return nil
}

// Delete deletes a cart by its ID (soft delete)
func (r *CartPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// First, soft delete all cart items
	itemsQuery := `
		UPDATE cart_items
		SET deleted_at = NOW()
		WHERE cart_id = $1 AND deleted_at IS NULL
	`

	_, err = tx.ExecContext(ctx, itemsQuery, id)
	if err != nil {
		return fmt.Errorf("error deleting cart items: %w", err)
	}

	// Then, soft delete the cart
	cartQuery := `
		UPDATE carts
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`

	var cartID uuid.UUID
	err = tx.QueryRowContext(ctx, cartQuery, id).Scan(&cartID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainErrors.ErrCartNotFound
		}
		return fmt.Errorf("error deleting cart: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// AddItem adds an item to a cart or updates its quantity if already exists
func (r *CartPostgresRepository) AddItem(ctx context.Context, item *entity.CartItem) error {
	logger := middleware.Logger.With(
		"method", "CartRepository.AddItem",
		"cart_id", item.CartID.String(),
		"product_id", item.ProductID.String(),
		"quantity", item.Quantity,
	)
	logger.InfoContext(ctx, "Adding item to cart")
	startTime := time.Now()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if cart exists
	var cartExists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM carts WHERE id = $1 AND deleted_at IS NULL)", item.CartID).Scan(&cartExists)
	if err != nil {
		return fmt.Errorf("error checking cart existence: %w", err)
	}

	if !cartExists {
		return domainErrors.ErrCartNotFound
	}

	// First, get product details - we need this for both new items and updates
	productQuery := `
		SELECT sku, name, price
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`
	err = tx.QueryRowContext(ctx, productQuery, item.ProductID).Scan(
		&item.ProductSKU,
		&item.ProductName,
		&item.UnitPrice,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainErrors.ErrProductNotFound
		}
		return fmt.Errorf("error getting product details: %w", err)
	}

	// Check if the item already exists
	var existingItemID uuid.UUID
	var existingQuantity int
	err = tx.QueryRowContext(ctx,
		"SELECT id, quantity FROM cart_items WHERE cart_id = $1 AND product_id = $2 AND deleted_at IS NULL",
		item.CartID, item.ProductID).Scan(&existingItemID, &existingQuantity)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error checking existing cart item: %w", err)
	}

	// If item exists, update its quantity
	if err == nil {
		updateQuery := `
			UPDATE cart_items
			SET quantity = $1, updated_at = NOW()
			WHERE id = $2
			RETURNING created_at, updated_at
		`

		newQuantity := existingQuantity + item.Quantity
		err = tx.QueryRowContext(ctx, updateQuery, newQuantity, existingItemID).Scan(
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("error updating cart item quantity: %w", err)
		}

		item.ID = existingItemID
		item.Quantity = newQuantity
	} else {
		// Otherwise, insert a new item
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}

		insertQuery := `
			INSERT INTO cart_items (id, cart_id, product_id, quantity, product_sku, product_name, unit_price)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING created_at, updated_at
		`

		err = tx.QueryRowContext(ctx, insertQuery,
			item.ID,
			item.CartID,
			item.ProductID,
			item.Quantity,
			item.ProductSKU,
			item.ProductName,
			item.UnitPrice).Scan(
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("error inserting cart item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully added item to cart",
		"item_id", item.ID.String(),
		"product_name", item.ProductName,
		"unit_price", item.UnitPrice,
		"total_price", item.UnitPrice*float64(item.Quantity),
		"duration_ms", duration.Milliseconds())

	return nil
}

// UpdateItem updates a cart item's quantity
func (r *CartPostgresRepository) UpdateItem(ctx context.Context, itemID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return domainErrors.ErrInvalidQuantity
	}

	query := `
		UPDATE cart_items
		SET quantity = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
		RETURNING id
	`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, quantity, itemID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainErrors.ErrCartItemNotFound
		}
		return fmt.Errorf("error updating cart item: %w", err)
	}

	return nil
}

// DeleteItem removes an item from a cart
func (r *CartPostgresRepository) DeleteItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	query := `
		UPDATE cart_items
		SET deleted_at = NOW()
		WHERE id = $1 AND cart_id = $2 AND deleted_at IS NULL
		RETURNING id
	`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, itemID, cartID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainErrors.ErrCartItemNotFound
		}
		return fmt.Errorf("error deleting cart item: %w", err)
	}

	return nil
}

// GetItem gets a specific item from a cart
func (r *CartPostgresRepository) GetItem(ctx context.Context, cartID, itemID uuid.UUID) (*entity.CartItem, error) {
	query := `
		SELECT 
			ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
			p.sku, p.name, p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.id = $1 AND ci.cart_id = $2 AND ci.deleted_at IS NULL
	`

	var item entity.CartItem
	err := r.db.QueryRowContext(ctx, query, itemID, cartID).Scan(
		&item.ID,
		&item.CartID,
		&item.ProductID,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.ProductSKU,
		&item.ProductName,
		&item.UnitPrice,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrCartItemNotFound
		}
		return nil, fmt.Errorf("error getting cart item: %w", err)
	}

	return &item, nil
}

// GetByUserID retrieves a cart by user ID with all items
func (r *CartPostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Cart, error) {
	logger := middleware.Logger.With(
		"method", "CartRepository.GetByUserID",
		"user_id", userID.String(),
	)
	logger.Debug("Fetching cart by user ID")
	startTime := time.Now()

	// First, get the cart
	cartQuery := `
		SELECT id, user_id, created_at, updated_at
		FROM carts
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`

	var cart entity.Cart
	err := r.db.QueryRowContext(ctx, cartQuery, userID).Scan(
		&cart.ID,
		&cart.UserID,
		&cart.CreatedAt,
		&cart.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("No cart found for user", "error", "ErrCartNotFound")
			return nil, domainErrors.ErrCartNotFound
		}
		logger.Error("Failed to query cart by user ID", "error", err.Error())
		return nil, fmt.Errorf("error querying cart by user ID: %w", err)
	}

	logger.Debug("Cart found, fetching cart items", "cart_id", cart.ID)

	// Then, get all cart items with product details
	itemsQuery := `
		SELECT 
			ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
			p.sku, p.name, p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = $1 AND ci.deleted_at IS NULL
		ORDER BY ci.created_at
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, cart.ID)
	if err != nil {
		logger.Error("Failed to query cart items", "error", err.Error())
		return nil, fmt.Errorf("error querying cart items: %w", err)
	}
	defer rows.Close()

	cart.Items = []*entity.CartItem{}
	itemCount := 0
	totalValue := 0.0

	for rows.Next() {
		var item entity.CartItem
		err := rows.Scan(
			&item.ID,
			&item.CartID,
			&item.ProductID,
			&item.Quantity,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.ProductSKU,
			&item.ProductName,
			&item.UnitPrice,
		)
		if err != nil {
			logger.Error("Failed to scan cart item", "error", err.Error())
			return nil, fmt.Errorf("error scanning cart item row: %w", err)
		}
		cart.Items = append(cart.Items, &item)
		itemCount++
		totalValue += float64(item.Quantity) * item.UnitPrice
	}

	if err = rows.Err(); err != nil {
		logger.Error("Failed to iterate cart items", "error", err.Error())
		return nil, fmt.Errorf("error iterating cart item rows: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved user cart with items",
		"cart_id", cart.ID,
		"item_count", itemCount,
		"total_value", totalValue,
		"duration_ms", duration.Milliseconds())

	return &cart, nil
}
