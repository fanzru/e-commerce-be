package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fanzru/e-commerce-be/internal/app/checkout/domain/entity"
	domainErrors "github.com/fanzru/e-commerce-be/internal/app/checkout/domain/errs"
	"github.com/google/uuid"
)

// CheckoutPostgresRepository implements CheckoutRepository using PostgreSQL
type CheckoutPostgresRepository struct {
	db *sql.DB
}

// NewCheckoutRepository creates a new checkout repository
func NewCheckoutRepository(db *sql.DB) CheckoutRepository {
	return &CheckoutPostgresRepository{
		db: db,
	}
}

// GetByID retrieves a checkout by its ID
func (r *CheckoutPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Checkout, error) {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the checkout
	checkoutQuery := `
		SELECT id, cart_id, subtotal, total_discount, total, created_at, updated_at
		FROM checkouts
		WHERE id = $1
	`

	var checkout entity.Checkout
	err = tx.QueryRowContext(ctx, checkoutQuery, id).Scan(
		&checkout.ID,
		&checkout.CartID,
		&checkout.Subtotal,
		&checkout.TotalDiscount,
		&checkout.Total,
		&checkout.CreatedAt,
		&checkout.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrCheckoutNotFound
		}
		return nil, fmt.Errorf("error querying checkout: %w", err)
	}

	// Get checkout items
	itemsQuery := `
		SELECT id, checkout_id, product_id, product_sku, product_name, quantity, unit_price, subtotal, discount, total
		FROM checkout_items
		WHERE checkout_id = $1
		ORDER BY id
	`

	itemRows, err := tx.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("error querying checkout items: %w", err)
	}
	defer itemRows.Close()

	checkout.Items = []*entity.CheckoutItem{}
	for itemRows.Next() {
		var item entity.CheckoutItem
		err := itemRows.Scan(
			&item.ID,
			&item.CheckoutID,
			&item.ProductID,
			&item.ProductSKU,
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
			&item.Subtotal,
			&item.Discount,
			&item.Total,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning checkout item: %w", err)
		}
		checkout.Items = append(checkout.Items, &item)
	}

	if err = itemRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating checkout items: %w", err)
	}

	// Get applied promotions
	promotionsQuery := `
		SELECT id, checkout_id, promotion_id, description, discount
		FROM checkout_promotions
		WHERE checkout_id = $1
		ORDER BY id
	`

	promotionRows, err := tx.QueryContext(ctx, promotionsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("error querying checkout promotions: %w", err)
	}
	defer promotionRows.Close()

	checkout.Promotions = []*entity.PromotionApplied{}
	for promotionRows.Next() {
		var promotion entity.PromotionApplied
		err := promotionRows.Scan(
			&promotion.ID,
			&promotion.CheckoutID,
			&promotion.PromotionID,
			&promotion.Description,
			&promotion.Discount,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning checkout promotion: %w", err)
		}
		checkout.Promotions = append(checkout.Promotions, &promotion)
	}

	if err = promotionRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating checkout promotions: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	return &checkout, nil
}

// GetByCartID retrieves a checkout by its cart ID
func (r *CheckoutPostgresRepository) GetByCartID(ctx context.Context, cartID uuid.UUID) (*entity.Checkout, error) {
	query := `
		SELECT id FROM checkouts
		WHERE cart_id = $1
	`

	var id uuid.UUID
	err := r.db.QueryRowContext(ctx, query, cartID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainErrors.ErrCheckoutNotFound
		}
		return nil, fmt.Errorf("error querying checkout by cart ID: %w", err)
	}

	return r.GetByID(ctx, id)
}

// Create creates a new checkout
func (r *CheckoutPostgresRepository) Create(ctx context.Context, checkout *entity.Checkout) error {
	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if a checkout already exists for this cart
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM checkouts WHERE cart_id = $1)", checkout.CartID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking if checkout exists: %w", err)
	}

	if exists {
		return domainErrors.ErrCartAlreadyCheckedOut
	}

	// Insert checkout
	if checkout.ID == uuid.Nil {
		checkout.ID = uuid.New()
	}

	checkoutQuery := `
		INSERT INTO checkouts (id, cart_id, subtotal, total_discount, total)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err = tx.QueryRowContext(ctx, checkoutQuery,
		checkout.ID,
		checkout.CartID,
		checkout.Subtotal,
		checkout.TotalDiscount,
		checkout.Total,
	).Scan(
		&checkout.CreatedAt,
		&checkout.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("error inserting checkout: %w", err)
	}

	// Insert checkout items
	for _, item := range checkout.Items {
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}
		item.CheckoutID = checkout.ID

		itemQuery := `
			INSERT INTO checkout_items (id, checkout_id, product_id, product_sku, product_name, quantity, unit_price, subtotal, discount, total)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`

		_, err = tx.ExecContext(ctx, itemQuery,
			item.ID,
			item.CheckoutID,
			item.ProductID,
			item.ProductSKU,
			item.ProductName,
			item.Quantity,
			item.UnitPrice,
			item.Subtotal,
			item.Discount,
			item.Total,
		)

		if err != nil {
			return fmt.Errorf("error inserting checkout item: %w", err)
		}
	}

	// Insert applied promotions
	if checkout.Promotions != nil {
		for _, promotion := range checkout.Promotions {
			if promotion.ID == uuid.Nil {
				promotion.ID = uuid.New()
			}
			promotion.CheckoutID = checkout.ID

			promotionQuery := `
				INSERT INTO checkout_promotions (id, checkout_id, promotion_id, description, discount)
				VALUES ($1, $2, $3, $4, $5)
			`

			_, err = tx.ExecContext(ctx, promotionQuery,
				promotion.ID,
				promotion.CheckoutID,
				promotion.PromotionID,
				promotion.Description,
				promotion.Discount,
			)

			if err != nil {
				return fmt.Errorf("error inserting checkout promotion: %w", err)
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

// ListCheckouts retrieves a list of checkouts with pagination
func (r *CheckoutPostgresRepository) ListCheckouts(ctx context.Context, page, limit int) ([]*entity.Checkout, int, error) {
	// Calculate offset
	offset := (page - 1) * limit

	// Count total checkouts
	var total int
	countQuery := `SELECT COUNT(*) FROM checkouts`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting checkouts: %w", err)
	}

	// Get checkouts with pagination
	query := `
		SELECT id, cart_id, subtotal, total_discount, total, created_at, updated_at
		FROM checkouts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying checkouts: %w", err)
	}
	defer rows.Close()

	checkouts := make([]*entity.Checkout, 0)
	for rows.Next() {
		var checkout entity.Checkout
		err := rows.Scan(
			&checkout.ID,
			&checkout.CartID,
			&checkout.Subtotal,
			&checkout.TotalDiscount,
			&checkout.Total,
			&checkout.CreatedAt,
			&checkout.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning checkout: %w", err)
		}

		// For efficiency, we're not loading items and promotions here
		// They will be loaded when getting a specific checkout
		checkouts = append(checkouts, &checkout)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating checkout rows: %w", err)
	}

	return checkouts, total, nil
}
