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

// GetByUserID retrieves all cart items for a user
func (r *CartPostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Cart, error) {
	logger := middleware.Logger.With(
		"method", "CartRepository.GetByUserID",
		"user_id", userID.String(),
	)
	logger.Debug("Fetching cart items for user")
	startTime := time.Now()

	// Create a new cart with the user ID
	cart := entity.NewCartWithUser(userID)

	// Get all cart items with product details
	itemsQuery := `
		SELECT 
			ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
			p.sku, p.name, p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1 AND ci.deleted_at IS NULL
		ORDER BY ci.created_at
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, userID)
	if err != nil {
		logger.Error("Failed to query cart items", "error", err.Error())
		return nil, fmt.Errorf("error querying cart items: %w", err)
	}
	defer rows.Close()

	cart.Items = []*entity.CartItem{}
	itemCount := 0

	for rows.Next() {
		var item entity.CartItem
		var productSKU string
		var productName string
		var unitPrice float64

		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ProductID,
			&item.Quantity,
			&item.CreatedAt,
			&item.UpdatedAt,
			&productSKU,
			&productName,
			&unitPrice,
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
	logger.Info("Successfully retrieved cart items for user",
		"item_count", itemCount,
		"duration_ms", duration.Milliseconds())

	return cart, nil
}

// AddItem adds an item to a user's cart or updates its quantity if already exists
func (r *CartPostgresRepository) AddItem(ctx context.Context, item *entity.CartItem) error {
	logger := middleware.Logger.With(
		"method", "CartRepository.AddItem",
		"user_id", item.UserID.String(),
		"product_id", item.ProductID.String(),
	)
	logger.Debug("Adding item to user's cart")
	startTime := time.Now()

	// Check if this product already exists in the user's cart
	existingItem, err := r.GetItemByProductID(ctx, item.UserID, item.ProductID)
	if err != nil && !errors.Is(err, domainErrors.ErrItemNotFound) {
		logger.Error("Failed to check for existing item", "error", err.Error())
		return fmt.Errorf("error checking for existing item: %w", err)
	}

	if existingItem != nil {
		// Update quantity if item exists
		logger.Debug("Item already exists, updating quantity",
			"item_id", existingItem.ID.String(),
			"old_quantity", existingItem.Quantity,
			"new_quantity", existingItem.Quantity+item.Quantity)

		return r.UpdateItem(ctx, existingItem.ID, existingItem.Quantity+item.Quantity)
	}

	// Otherwise, insert a new item
	query := `
		INSERT INTO cart_items (
			id, user_id, product_id, quantity
		) VALUES (
			$1, $2, $3, $4
		)
	`

	_, err = r.db.ExecContext(ctx, query,
		item.ID,
		item.UserID,
		item.ProductID,
		item.Quantity,
	)

	if err != nil {
		logger.Error("Failed to insert cart item", "error", err.Error())
		return fmt.Errorf("error inserting cart item: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully added item to cart",
		"item_id", item.ID.String(),
		"duration_ms", duration.Milliseconds())

	return nil
}

// UpdateItem updates a cart item's quantity
func (r *CartPostgresRepository) UpdateItem(ctx context.Context, itemID uuid.UUID, quantity int) error {
	logger := middleware.Logger.With(
		"method", "CartRepository.UpdateItem",
		"item_id", itemID.String(),
		"quantity", quantity,
	)
	logger.Debug("Updating cart item quantity")
	startTime := time.Now()

	if quantity <= 0 {
		logger.Debug("Quantity is zero or negative, deleting item instead")
		// Get the item to find the user ID
		query := `
			SELECT user_id FROM cart_items 
			WHERE id = $1 AND deleted_at IS NULL
		`
		var userID uuid.UUID
		err := r.db.QueryRowContext(ctx, query, itemID).Scan(&userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domainErrors.ErrItemNotFound
			}
			return fmt.Errorf("error getting item user ID: %w", err)
		}

		return r.DeleteItem(ctx, userID, itemID)
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
			logger.Warn("Cart item not found", "error", "ErrItemNotFound")
			return domainErrors.ErrItemNotFound
		}
		logger.Error("Failed to update cart item", "error", err.Error())
		return fmt.Errorf("error updating cart item: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully updated cart item quantity",
		"duration_ms", duration.Milliseconds())

	return nil
}

// DeleteItem removes an item from a user's cart
func (r *CartPostgresRepository) DeleteItem(ctx context.Context, userID, itemID uuid.UUID) error {
	logger := middleware.Logger.With(
		"method", "CartRepository.DeleteItem",
		"user_id", userID.String(),
		"item_id", itemID.String(),
	)
	logger.Debug("Deleting cart item")
	startTime := time.Now()

	query := `
		DELETE FROM cart_items 
		WHERE id = $1 AND user_id = $2
		RETURNING id
	`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, itemID, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Cart item not found", "error", "ErrItemNotFound")
			return domainErrors.ErrItemNotFound
		}
		logger.Error("Failed to delete cart item", "error", err.Error())
		return fmt.Errorf("error deleting cart item: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully deleted cart item",
		"duration_ms", duration.Milliseconds())

	return nil
}

// GetItem gets a specific item from a user's cart
func (r *CartPostgresRepository) GetItem(ctx context.Context, userID, itemID uuid.UUID) (*entity.CartItem, error) {
	logger := middleware.Logger.With(
		"method", "CartRepository.GetItem",
		"user_id", userID.String(),
		"item_id", itemID.String(),
	)
	logger.Debug("Getting cart item")
	startTime := time.Now()

	query := `
		SELECT 
			ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
			p.sku, p.name, p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.id = $1 AND ci.user_id = $2 AND ci.deleted_at IS NULL
	`

	var item entity.CartItem
	var productSKU string
	var productName string
	var unitPrice float64

	err := r.db.QueryRowContext(ctx, query, itemID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.ProductID,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
		&productSKU,
		&productName,
		&unitPrice,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Cart item not found", "error", "ErrItemNotFound")
			return nil, domainErrors.ErrItemNotFound
		}
		logger.Error("Failed to get cart item", "error", err.Error())
		return nil, fmt.Errorf("error getting cart item: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved cart item",
		"duration_ms", duration.Milliseconds())

	return &item, nil
}

// GetItemByProductID gets a specific item by product ID from a user's cart
func (r *CartPostgresRepository) GetItemByProductID(ctx context.Context, userID, productID uuid.UUID) (*entity.CartItem, error) {
	logger := middleware.Logger.With(
		"method", "CartRepository.GetItemByProductID",
		"user_id", userID.String(),
		"product_id", productID.String(),
	)
	logger.Debug("Getting cart item by product ID")
	startTime := time.Now()

	query := `
		SELECT 
			ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
			p.sku, p.name, p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.product_id = $1 AND ci.user_id = $2 AND ci.deleted_at IS NULL
	`

	var item entity.CartItem
	var productSKU string
	var productName string
	var unitPrice float64

	err := r.db.QueryRowContext(ctx, query, productID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.ProductID,
		&item.Quantity,
		&item.CreatedAt,
		&item.UpdatedAt,
		&productSKU,
		&productName,
		&unitPrice,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Cart item not found", "error", "ErrItemNotFound")
			return nil, domainErrors.ErrItemNotFound
		}
		logger.Error("Failed to get cart item by product ID", "error", err.Error())
		return nil, fmt.Errorf("error getting cart item by product ID: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved cart item by product ID",
		"duration_ms", duration.Milliseconds())

	return &item, nil
}

// ClearUserCart removes all items from a user's cart
func (r *CartPostgresRepository) ClearUserCart(ctx context.Context, userID uuid.UUID) error {
	logger := middleware.Logger.With(
		"method", "CartRepository.ClearUserCart",
		"user_id", userID.String(),
	)
	logger.Debug("Clearing user cart")
	startTime := time.Now()

	query := `
		UPDATE cart_items 
		SET deleted_at = NOW() 
		WHERE user_id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		logger.Error("Failed to clear user cart", "error", err.Error())
		return fmt.Errorf("error clearing user cart: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", "error", err.Error())
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully cleared user cart",
		"items_removed", rowsAffected,
		"duration_ms", duration.Milliseconds())

	return nil
}

// GetCartInfo retrieves cart with product details for display
func (r *CartPostgresRepository) GetCartInfo(ctx context.Context, userID uuid.UUID) (*entity.CartInfo, error) {
	logger := middleware.Logger.With(
		"method", "CartRepository.GetCartInfo",
		"user_id", userID.String(),
	)
	logger.Debug("Fetching cart info with product details")
	startTime := time.Now()

	// Create a new cart info with the user ID
	cartInfo := &entity.CartInfo{
		UserID:    userID,
		Items:     []*entity.CartItemInfo{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Subtotal:  0,
	}

	// Get all cart items with product details
	itemsQuery := `
		SELECT 
			ci.id, ci.user_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
			p.sku, p.name, p.price
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.user_id = $1 AND ci.deleted_at IS NULL
		ORDER BY ci.created_at
	`

	rows, err := r.db.QueryContext(ctx, itemsQuery, userID)
	if err != nil {
		logger.Error("Failed to query cart items", "error", err.Error())
		return nil, fmt.Errorf("error querying cart items: %w", err)
	}
	defer rows.Close()

	itemCount := 0
	var totalSubtotal float64 = 0

	for rows.Next() {
		var item entity.CartItemInfo
		err := rows.Scan(
			&item.ID,
			&item.UserID,
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

		// Calculate subtotal for the item
		item.Subtotal = item.UnitPrice * float64(item.Quantity)
		totalSubtotal += item.Subtotal

		cartInfo.Items = append(cartInfo.Items, &item)
		itemCount++
	}

	if err = rows.Err(); err != nil {
		logger.Error("Failed to iterate cart items", "error", err.Error())
		return nil, fmt.Errorf("error iterating cart item rows: %w", err)
	}

	cartInfo.Subtotal = totalSubtotal

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved cart info with product details",
		"item_count", itemCount,
		"subtotal", totalSubtotal,
		"duration_ms", duration.Milliseconds())

	return cartInfo, nil
}
