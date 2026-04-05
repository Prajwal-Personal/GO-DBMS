package api

import (
	"context"

	"github.com/unidb/unidb-go/internal"
)

// DB represents the unified database instance.
type DB struct {
	conn       internal.Connection
	config     internal.Config
	middleware []Middleware
}

// finalHandler is the base execution handler that directly calls the driver.
func (db *DB) finalHandler() Handler {
	return func(ctx context.Context, query string, args ...any) (internal.Result, error) {
		return db.conn.Query(ctx, query, args...)
	}
}

// Query executes a query that returns rows, typically a SELECT.
func (db *DB) Query(ctx context.Context, query string, args ...any) (internal.Result, error) {
	handler := db.finalHandler()

	// Apply middleware in reverse order so the first added is the first executed
	for i := len(db.middleware) - 1; i >= 0; i-- {
		handler = db.middleware[i](handler)
	}

	return handler(ctx, query, args...)
}

// Exec executes a query without returning any rows, typically an INSERT, UPDATE, or DELETE.
func (db *DB) Exec(ctx context.Context, query string, args ...any) (internal.ExecResult, error) {
	// For Exec, we can build a similar middleware pipeline, or reuse handler with some casting.
	// For MVP, we will directly call the conn.Exec. 
	// A proper implementation would have a unified ExecHandler as well.
	return db.conn.Exec(ctx, query, args...)
}

// Transaction options
type TxOptions struct {
	Isolation string
	ReadOnly  bool
}

// BeginTx starts a database transaction.
func (db *DB) BeginTx(ctx context.Context, opts *TxOptions) (internal.Tx, error) {
	// Options ignored for MVP
	return db.conn.BeginTx(ctx)
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}
