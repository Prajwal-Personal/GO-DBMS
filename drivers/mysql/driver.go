package mysql

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/unidb/unidb-go/drivers/sqlwrap"
	"github.com/unidb/unidb-go/internal"
)

func init() {
	internal.RegisterDriver("mysql", &MySQLDriver{})
}

type MySQLDriver struct{}

func (d *MySQLDriver) Connect(config internal.Config) (internal.Connection, error) {
	// user:password@tcp(localhost:5555)/dbname
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, err
	}

	if config.PoolSize > 0 {
		db.SetMaxOpenConns(config.PoolSize)
	}

	return &sqlwrap.SQLConnection{DB: db}, nil
}

func (d *MySQLDriver) Capabilities() internal.Capabilities {
	return internal.Capabilities{
		SupportsSQL:          true,
		SupportsTransactions: true,
		SupportsJoins:        true,
		SupportsAggregation:  true,
	}
}
