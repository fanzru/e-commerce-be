package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/google/uuid"
)

// ExampleRepository demonstrates how to use the TransactionManager
type ExampleRepository struct {
	db        *sql.DB
	txManager *TransactionManager
}

// NewExampleRepository creates a new ExampleRepository
func NewExampleRepository(db *sql.DB, txManager *TransactionManager) *ExampleRepository {
	return &ExampleRepository{
		db:        db,
		txManager: txManager,
	}
}

// CreateOrderWithItems creates an order with items in a transaction
func (r *ExampleRepository) CreateOrderWithItems(ctx context.Context, orderID uuid.UUID, items []uuid.UUID) error {
	logger := middleware.Logger.With(
		"method", "ExampleRepository.CreateOrderWithItems",
		"order_id", orderID.String(),
	)
	logger.Info("Creating order with items")
	startTime := time.Now()

	// Use transaction manager to ensure all operations are atomic
	err := r.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		// Get the queryable (either tx from context or db)
		queryable := r.txManager.GetQueryable(ctx)

		// 1. Insert order
		orderQuery := `
			INSERT INTO orders (id, created_at)
			VALUES ($1, $2)
		`
		_, err := queryable.ExecContext(ctx, orderQuery, orderID, time.Now())
		if err != nil {
			logger.Error("Failed to insert order", "error", err.Error())
			return fmt.Errorf("failed to insert order: %w", err)
		}

		// 2. Insert order items
		for i, itemID := range items {
			itemQuery := `
				INSERT INTO order_items (id, order_id, item_number, item_id)
				VALUES ($1, $2, $3, $4)
			`
			_, err := queryable.ExecContext(ctx, itemQuery, uuid.New(), orderID, i+1, itemID)
			if err != nil {
				logger.Error("Failed to insert order item", "error", err.Error())
				return fmt.Errorf("failed to insert order item: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	duration := time.Since(startTime)
	logger.Info("Successfully created order with items",
		"item_count", len(items),
		"duration_ms", duration.Milliseconds())

	return nil
}

// Example of how to use transactions across multiple repositories
func ExampleUsageAcrossRepositories(
	ctx context.Context,
	txManager *TransactionManager,
	orderRepo *ExampleRepository,
	inventoryRepo *ExampleRepository,
) error {
	// Execute everything in a single transaction
	return txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		orderID := uuid.New()
		items := []uuid.UUID{uuid.New(), uuid.New()}

		// Create order with items
		if err := orderRepo.CreateOrderWithItems(ctx, orderID, items); err != nil {
			return err
		}

		// Update inventory (using the same transaction from context)
		for _, itemID := range items {
			if err := inventoryRepo.DecrementInventory(ctx, itemID, 1); err != nil {
				return err
			}
		}

		return nil
	})
}

// DecrementInventory is a sample method for the example
func (r *ExampleRepository) DecrementInventory(ctx context.Context, itemID uuid.UUID, quantity int) error {
	queryable := r.txManager.GetQueryable(ctx)

	query := `
		UPDATE inventory
		SET quantity = quantity - $1
		WHERE item_id = $2
	`
	_, err := queryable.ExecContext(ctx, query, quantity, itemID)
	return err
}
