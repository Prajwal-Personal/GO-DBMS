package internal

import "context"

// Result represents a unified database result
type Result interface {
	Columns() []string
	Next() bool
	Scan(dest ...any) error
	Close() error
}

// ExecResult represents the result of a non-read query
type ExecResult interface {
	RowsAffected() int64
	LastInsertId() (int64, error)
}

// Tx represents a database transaction
type Tx interface {
	Query(ctx context.Context, query string, args ...any) (Result, error)
	Exec(ctx context.Context, query string, args ...any) (ExecResult, error)

	Commit() error
	Rollback() error
}

// Connection represents a single database connection
type Connection interface {
	Query(ctx context.Context, query string, args ...any) (Result, error)
	Exec(ctx context.Context, query string, args ...any) (ExecResult, error)

	BeginTx(ctx context.Context) (Tx, error)

	Close() error
}

// Capabilities defines what a driver supports
type Capabilities struct {
	SupportsSQL          bool
	SupportsTransactions bool
	SupportsJoins        bool
	SupportsAggregation  bool
}

// Config represents connection configuration
type Config struct {
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	PoolSize int
}

// Driver is the internal interface all DB drivers MUST implement
type Driver interface {
	Connect(config Config) (Connection, error)
	Capabilities() Capabilities
}
