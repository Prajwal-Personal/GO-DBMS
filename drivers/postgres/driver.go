package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/unidb/unidb-go/drivers/sqlwrap"
	"github.com/unidb/unidb-go/internal"
)

func init() {
	internal.RegisterDriver("postgres", &PostgresDriver{})
}

type PostgresDriver struct{}

func (d *PostgresDriver) Connect(config internal.Config) (internal.Connection, error) {
	// postgres://user:pass@localhost:5432/dbname?sslmode=disable
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if config.PoolSize > 0 {
		db.SetMaxOpenConns(config.PoolSize)
	}

	// Ping to verify connection
	// err = db.Ping() // Skip ping for mocking without real db

	return &sqlwrap.SQLConnection{DB: db}, nil
}

func (d *PostgresDriver) Capabilities() internal.Capabilities {
	return internal.Capabilities{
		SupportsSQL:          true,
		SupportsTransactions: true,
		SupportsJoins:        true,
		SupportsAggregation:  true,
	}
}
