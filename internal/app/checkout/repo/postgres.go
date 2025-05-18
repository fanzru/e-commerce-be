package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/fanzru/e-commerce-be/internal/app/checkout/domain/entity"
	domainErrors "github.com/fanzru/e-commerce-be/internal/app/checkout/domain/errs"
	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
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
	logger := middleware.Logger.With(
		"method", "CheckoutRepository.GetByID",
		"checkout_id", id.String(),
	)
	logger.Debug("Fetching checkout by ID")
	startTime := time.Now()

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin transaction", "error", err.Error())
		return nil, fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the checkout
	checkoutQuery := `
		SELECT id, user_id, subtotal, total_discount, total, 
		       payment_status, payment_method, payment_reference, notes, status, 
		       created_at, updated_at, completed_at
		FROM checkouts
		WHERE id = $1
	`

	var checkout entity.Checkout
	var userID sql.NullString
	var paymentMethod, paymentReference, notes sql.NullString
	var completedAt sql.NullTime

	err = tx.QueryRowContext(ctx, checkoutQuery, id).Scan(
		&checkout.ID,
		&userID,
		&checkout.Subtotal,
		&checkout.TotalDiscount,
		&checkout.Total,
		&checkout.PaymentStatus,
		&paymentMethod,
		&paymentReference,
		&notes,
		&checkout.Status,
		&checkout.CreatedAt,
		&checkout.UpdatedAt,
		&completedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Checkout not found", "error", "ErrCheckoutNotFound")
			return nil, domainErrors.ErrCheckoutNotFound
		}
		logger.Error("Failed to query checkout", "error", err.Error())
		return nil, fmt.Errorf("error querying checkout: %w", err)
	}

	// Handle nullable fields
	if userID.Valid {
		userUUID, err := uuid.Parse(userID.String)
		if err == nil {
			checkout.UserID = &userUUID
		}
	}
	if paymentMethod.Valid {
		checkout.PaymentMethod = &paymentMethod.String
	}
	if paymentReference.Valid {
		checkout.PaymentReference = &paymentReference.String
	}
	if notes.Valid {
		checkout.Notes = &notes.String
	}
	if completedAt.Valid {
		checkout.CompletedAt = &completedAt.Time
	}

	logger.Debug("Checkout found, fetching checkout items")

	// Get checkout items
	itemsQuery := `
		SELECT id, checkout_id, product_id, product_sku, product_name, quantity, unit_price, subtotal, discount, total
		FROM checkout_items
		WHERE checkout_id = $1
		ORDER BY id
	`

	itemRows, err := tx.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		logger.Error("Failed to query checkout items", "error", err.Error())
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
			logger.Error("Failed to scan checkout item", "error", err.Error())
			return nil, fmt.Errorf("error scanning checkout item: %w", err)
		}
		checkout.Items = append(checkout.Items, &item)
	}

	if err = itemRows.Err(); err != nil {
		logger.Error("Failed to iterate checkout items", "error", err.Error())
		return nil, fmt.Errorf("error iterating checkout items: %w", err)
	}

	// Get applied promotions
	promotionsQuery := `
		SELECT id, checkout_id, promotion_id, description, discount
		FROM promotion_applied
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

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", "error", err.Error())
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved checkout with items",
		"user_id", checkout.UserID,
		"payment_status", checkout.PaymentStatus,
		"status", checkout.Status,
		"item_count", len(checkout.Items),
		"subtotal", checkout.Subtotal,
		"total_discount", checkout.TotalDiscount,
		"total", checkout.Total,
		"duration_ms", duration.Milliseconds())

	return &checkout, nil
}

// Create creates a new checkout
func (r *CheckoutPostgresRepository) Create(ctx context.Context, checkout *entity.Checkout) error {
	logger := middleware.Logger.With(
		"method", "CheckoutRepository.Create",
	)
	if checkout.UserID != nil {
		logger = logger.With("user_id", checkout.UserID.String())
	}
	logger.Debug("Creating new checkout")
	startTime := time.Now()

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin transaction", "error", err.Error())
		return fmt.Errorf("error beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert checkout
	if checkout.ID == uuid.Nil {
		checkout.ID = uuid.New()
	}

	// Set default values if not provided
	if checkout.PaymentStatus == "" {
		checkout.PaymentStatus = entity.PaymentStatusPending
	}
	if checkout.Status == "" {
		checkout.Status = entity.OrderStatusCreated
	}

	checkoutQuery := `
		INSERT INTO checkouts (
			id, user_id, subtotal, total_discount, total, 
			payment_status, payment_method, payment_reference, notes, status, completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`

	err = tx.QueryRowContext(ctx, checkoutQuery,
		checkout.ID,
		checkout.UserID,
		checkout.Subtotal,
		checkout.TotalDiscount,
		checkout.Total,
		checkout.PaymentStatus,
		checkout.PaymentMethod,
		checkout.PaymentReference,
		checkout.Notes,
		checkout.Status,
		checkout.CompletedAt,
	).Scan(
		&checkout.CreatedAt,
		&checkout.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to insert checkout", "error", err.Error())
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
			logger.Error("Failed to insert checkout item", "error", err.Error())
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
				INSERT INTO promotion_applied (id, checkout_id, promotion_id, description, discount)
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

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", "error", err.Error())
		return fmt.Errorf("error committing transaction: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully created checkout",
		"checkout_id", checkout.ID.String(),
		"user_id", checkout.UserID,
		"payment_status", checkout.PaymentStatus,
		"status", checkout.Status,
		"item_count", len(checkout.Items),
		"subtotal", checkout.Subtotal,
		"total_discount", checkout.TotalDiscount,
		"total", checkout.Total,
		"duration_ms", duration.Milliseconds())

	return nil
}

// List retrieves a list of checkouts with pagination
func (r *CheckoutPostgresRepository) List(ctx context.Context, page, limit int) ([]*entity.Checkout, int, error) {
	logger := middleware.Logger.With(
		"method", "CheckoutRepository.List",
		"page", page,
		"limit", limit,
	)
	logger.Debug("Listing checkouts")
	startTime := time.Now()

	offset := (page - 1) * limit

	// Count total matches first
	countQuery := `SELECT COUNT(*) FROM checkouts`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		logger.Error("Failed to count checkouts", "error", err.Error())
		return nil, 0, fmt.Errorf("error counting checkouts: %w", err)
	}

	// Now fetch the checkouts with pagination
	query := `
		SELECT id, user_id, subtotal, total_discount, total, 
		       payment_status, payment_method, payment_reference, notes, status, 
		       created_at, updated_at, completed_at
		FROM checkouts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		logger.Error("Failed to query checkouts", "error", err.Error())
		return nil, 0, fmt.Errorf("error querying checkouts: %w", err)
	}
	defer rows.Close()

	checkouts := []*entity.Checkout{}
	for rows.Next() {
		var checkout entity.Checkout
		var userID sql.NullString
		var paymentMethod, paymentReference, notes sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&checkout.ID,
			&userID,
			&checkout.Subtotal,
			&checkout.TotalDiscount,
			&checkout.Total,
			&checkout.PaymentStatus,
			&paymentMethod,
			&paymentReference,
			&notes,
			&checkout.Status,
			&checkout.CreatedAt,
			&checkout.UpdatedAt,
			&completedAt,
		)
		if err != nil {
			logger.Error("Failed to scan checkout row", "error", err.Error())
			return nil, 0, fmt.Errorf("error scanning checkout row: %w", err)
		}

		// Handle nullable fields
		if userID.Valid {
			userUUID, err := uuid.Parse(userID.String)
			if err == nil {
				checkout.UserID = &userUUID
			}
		}
		if paymentMethod.Valid {
			checkout.PaymentMethod = &paymentMethod.String
		}
		if paymentReference.Valid {
			checkout.PaymentReference = &paymentReference.String
		}
		if notes.Valid {
			checkout.Notes = &notes.String
		}
		if completedAt.Valid {
			checkout.CompletedAt = &completedAt.Time
		}

		// Set empty items slice, but don't fetch items to reduce load
		checkout.Items = []*entity.CheckoutItem{}
		checkouts = append(checkouts, &checkout)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Failed to iterate checkout rows", "error", err.Error())
		return nil, 0, fmt.Errorf("error iterating checkout rows: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully listed checkouts",
		"total_count", total,
		"returned_count", len(checkouts),
		"duration_ms", duration.Milliseconds())

	return checkouts, total, nil
}

// GetByUserID retrieves a list of checkouts by user ID
func (r *CheckoutPostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.Checkout, int, error) {
	logger := middleware.Logger.With(
		"method", "CheckoutRepository.GetByUserID",
		"user_id", userID.String(),
		"page", page,
		"limit", limit,
	)
	logger.Debug("Fetching checkouts by user ID")
	startTime := time.Now()

	offset := (page - 1) * limit

	// Count total matches for this user
	countQuery := `SELECT COUNT(*) FROM checkouts WHERE user_id = $1`
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		logger.Error("Failed to count user checkouts", "error", err.Error())
		return nil, 0, fmt.Errorf("error counting user checkouts: %w", err)
	}

	// Now fetch the checkouts with pagination
	query := `
		SELECT id, user_id, subtotal, total_discount, total, 
		       payment_status, payment_method, payment_reference, notes, status, 
		       created_at, updated_at, completed_at
		FROM checkouts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to query user checkouts", "error", err.Error())
		return nil, 0, fmt.Errorf("error querying user checkouts: %w", err)
	}
	defer rows.Close()

	checkouts := []*entity.Checkout{}
	for rows.Next() {
		var checkout entity.Checkout
		var userIDNull sql.NullString
		var paymentMethod, paymentReference, notes sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&checkout.ID,
			&userIDNull,
			&checkout.Subtotal,
			&checkout.TotalDiscount,
			&checkout.Total,
			&checkout.PaymentStatus,
			&paymentMethod,
			&paymentReference,
			&notes,
			&checkout.Status,
			&checkout.CreatedAt,
			&checkout.UpdatedAt,
			&completedAt,
		)
		if err != nil {
			logger.Error("Failed to scan checkout row", "error", err.Error())
			return nil, 0, fmt.Errorf("error scanning checkout row: %w", err)
		}

		// Handle nullable fields
		if userIDNull.Valid {
			userUUID, err := uuid.Parse(userIDNull.String)
			if err == nil {
				checkout.UserID = &userUUID
			}
		}
		if paymentMethod.Valid {
			checkout.PaymentMethod = &paymentMethod.String
		}
		if paymentReference.Valid {
			checkout.PaymentReference = &paymentReference.String
		}
		if notes.Valid {
			checkout.Notes = &notes.String
		}
		if completedAt.Valid {
			checkout.CompletedAt = &completedAt.Time
		}

		// Set empty items slice, but don't fetch items to reduce load
		checkout.Items = []*entity.CheckoutItem{}
		checkouts = append(checkouts, &checkout)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Failed to iterate checkout rows", "error", err.Error())
		return nil, 0, fmt.Errorf("error iterating checkout rows: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully retrieved user checkouts",
		"user_id", userID,
		"total_count", total,
		"returned_count", len(checkouts),
		"duration_ms", duration.Milliseconds())

	return checkouts, total, nil
}

// UpdatePaymentStatus updates the payment status of a checkout
func (r *CheckoutPostgresRepository) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status entity.PaymentStatus, paymentMethod, paymentReference string) error {
	logger := middleware.Logger.With(
		"method", "CheckoutRepository.UpdatePaymentStatus",
		"checkout_id", id.String(),
		"payment_status", status,
		"payment_method", paymentMethod,
	)
	logger.Debug("Updating payment status")
	startTime := time.Now()

	query := `
		UPDATE checkouts
		SET payment_status = $1, 
		    payment_method = $2, 
		    payment_reference = $3,
		    status = CASE WHEN $1 = 'PAID' AND status = 'CREATED' THEN 'PROCESSING' ELSE status END,
		    updated_at = NOW()
		WHERE id = $4
		RETURNING id
	`

	var checkoutID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, status, paymentMethod, paymentReference, id).Scan(&checkoutID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Checkout not found", "error", "ErrCheckoutNotFound")
			return domainErrors.ErrCheckoutNotFound
		}
		logger.Error("Failed to update payment status", "error", err.Error())
		return fmt.Errorf("error updating payment status: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully updated payment status",
		"payment_status", status,
		"duration_ms", duration.Milliseconds())

	return nil
}

// UpdateOrderStatus updates the order status of a checkout
func (r *CheckoutPostgresRepository) UpdateOrderStatus(ctx context.Context, id uuid.UUID, status entity.OrderStatus) error {
	logger := middleware.Logger.With(
		"method", "CheckoutRepository.UpdateOrderStatus",
		"checkout_id", id.String(),
		"order_status", status,
	)
	logger.Debug("Updating order status")
	startTime := time.Now()

	var completedAt interface{} = nil
	if status == entity.OrderStatusDelivered {
		completedAt = time.Now()
	}

	query := `
		UPDATE checkouts
		SET status = $1, 
		    completed_at = CASE WHEN $2 IS NOT NULL THEN $2 ELSE completed_at END,
		    updated_at = NOW()
		WHERE id = $3
		RETURNING id
	`

	var checkoutID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, status, completedAt, id).Scan(&checkoutID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Warn("Checkout not found", "error", "ErrCheckoutNotFound")
			return domainErrors.ErrCheckoutNotFound
		}
		logger.Error("Failed to update order status", "error", err.Error())
		return fmt.Errorf("error updating order status: %w", err)
	}

	duration := time.Since(startTime)
	logger.Info("Successfully updated order status",
		"order_status", status,
		"duration_ms", duration.Milliseconds())

	return nil
}
