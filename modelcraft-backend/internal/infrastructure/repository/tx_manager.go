package repository

import (
	"context"
	"database/sql"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/pkg/bizerrors"
)

// TxManager manages database transactions, passing a dbgen.Querier explicitly to the callback.
// This is "Option B" (explicit Querier passing) as decided in design.
type TxManager interface {
	// WithTx begins a transaction, passes a Querier bound to that transaction to fn,
	// commits on success, and rolls back on error or panic.
	WithTx(ctx context.Context, fn func(ctx context.Context, q dbgen.Querier) error) error
}

// SqlTxManager is the standard *sql.DB-based TxManager implementation.
type SqlTxManager struct {
	db *sql.DB
}

// NewSqlTxManager creates a SqlTxManager from a *sql.DB.
func NewSqlTxManager(db *sql.DB) TxManager {
	return &SqlTxManager{db: db}
}

// WithTx implements TxManager.
func (m *SqlTxManager) WithTx(ctx context.Context, fn func(ctx context.Context, q dbgen.Querier) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return bizerrors.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	q := dbgen.New(tx)
	if err := fn(ctx, q); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return bizerrors.Errorf("tx rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return bizerrors.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
