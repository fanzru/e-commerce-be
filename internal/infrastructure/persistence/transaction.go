package persistence

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/fanzru/e-commerce-be/internal/infrastructure/middleware"
	"github.com/google/uuid"
)

// TxKey is the context key for transaction
type TxKey string

const (
	txKey TxKey = "tx"
)

// TransactionManager manages database transactions
type TransactionManager struct {
	db *sql.DB
}

var (
	m    *TransactionManager
	once sync.Once
)

// ProvideTransactionManager returns a singleton TransactionManager instance
func ProvideTransactionManager(db *sql.DB) *TransactionManager {
	once.Do(func() {
		m = NewTransactionManager(db)
	})
	return m
}

// NewTransactionManager creates a new TransactionManager
func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{
		db: db,
	}
}

// RunInTransaction executes the provided function within a transaction
func (m *TransactionManager) RunInTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	logger := middleware.Logger.With(
		"method", "TransactionManager.RunInTransaction",
		"transaction_id", uuid.New().String(),
	)
	logger.Debug("Starting transaction")
	startTime := time.Now()

	// Check if there's already a transaction in the context
	tx := TxFromContext(ctx)
	isOuterTx := false

	if tx == nil {
		// Start a new transaction
		var err error
		tx, err = m.db.BeginTx(ctx, nil)
		if err != nil {
			logger.Error("Failed to begin transaction", "error", err.Error())
			return err
		}
		isOuterTx = true
	}

	// Add panic recovery to rollback transaction
	defer func() {
		if r := recover(); r != nil {
			if isOuterTx && tx != nil {
				err := tx.Rollback()
				if err != nil {
					logger.Error("Failed to rollback transaction after panic", "error", err.Error(), "panic", r)
				}
			}
			// Re-panic to preserve the original panic
			panic(r)
		}
	}()

	// Store transaction in context
	ctx = ContextWithTx(ctx, tx)

	// Execute the callback function
	err := f(ctx)

	// Only commit/rollback if this is the outer transaction
	if !isOuterTx {
		return err
	}

	// Handle the result
	if err != nil {
		rollbackErr := tx.Rollback()
		if rollbackErr != nil {
			logger.Error("Failed to rollback transaction", "error", rollbackErr.Error(), "original_error", err.Error())
			return rollbackErr
		}
		logger.Warn("Transaction rolled back", "error", err.Error())
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit transaction", "error", err.Error())
		return err
	}

	duration := time.Since(startTime)
	logger.Info("Transaction committed successfully",
		"duration_ms", duration.Milliseconds())
	return nil
}

// ContextWithTx adds a transaction to context
func ContextWithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// TxFromContext retrieves transaction from context
func TxFromContext(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value(txKey).(*sql.Tx); ok {
		return tx
	}
	return nil
}

// GetQueryable returns either the transaction from context or the database
func (m *TransactionManager) GetQueryable(ctx context.Context) interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
} {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return m.db
}
