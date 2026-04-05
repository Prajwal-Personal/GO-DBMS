package sqlwrap

import (
	"context"
	"database/sql"
	"github.com/unidb/unidb-go/internal"
)

// SQLConnection wraps *sql.DB
type SQLConnection struct {
	DB *sql.DB
}

func (c *SQLConnection) Query(ctx context.Context, query string, args ...any) (internal.Result, error) {
	rows, err := c.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLResult{Rows: rows}, nil
}

func (c *SQLConnection) Exec(ctx context.Context, query string, args ...any) (internal.ExecResult, error) {
	res, err := c.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLExecResult{Result: res}, nil
}

func (c *SQLConnection) BeginTx(ctx context.Context) (internal.Tx, error) {
	tx, err := c.DB.BeginTx(ctx, nil) // use default options for now
	if err != nil {
		return nil, err
	}
	return &SQLTx{Tx: tx}, nil
}

func (c *SQLConnection) Close() error {
	return c.DB.Close()
}

// SQLResult wraps *sql.Rows
type SQLResult struct {
	Rows *sql.Rows
}

func (r *SQLResult) Columns() []string {
	cols, _ := r.Rows.Columns()
	return cols
}

func (r *SQLResult) Next() bool {
	return r.Rows.Next()
}

func (r *SQLResult) Scan(dest ...any) error {
	return r.Rows.Scan(dest...)
}

func (r *SQLResult) Close() error {
	return r.Rows.Close()
}

// SQLExecResult wraps sql.Result
type SQLExecResult struct {
	Result sql.Result
}

func (r *SQLExecResult) RowsAffected() int64 {
	v, _ := r.Result.RowsAffected()
	return v
}

func (r *SQLExecResult) LastInsertId() (int64, error) {
	return r.Result.LastInsertId()
}

// SQLTx wraps *sql.Tx
type SQLTx struct {
	Tx *sql.Tx
}

func (t *SQLTx) Query(ctx context.Context, query string, args ...any) (internal.Result, error) {
	rows, err := t.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLResult{Rows: rows}, nil
}

func (t *SQLTx) Exec(ctx context.Context, query string, args ...any) (internal.ExecResult, error) {
	res, err := t.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &SQLExecResult{Result: res}, nil
}

func (t *SQLTx) Commit() error {
	return t.Tx.Commit()
}

func (t *SQLTx) Rollback() error {
	return t.Tx.Rollback()
}
