package api

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/unidb/unidb-go/internal"

	// Import drivers to run their init() functions
	_ "github.com/unidb/unidb-go/drivers/mongodb"
	_ "github.com/unidb/unidb-go/drivers/mysql"
	_ "github.com/unidb/unidb-go/drivers/postgres"
)

// ParseConnectionString parses a URI into internal.Config
// e.g. postgres://user:pass@localhost:5432/dbname
func ParseConnectionString(connStr string) (internal.Config, error) {
	u, err := url.Parse(connStr)
	if err != nil {
		return internal.Config{}, err
	}

	cfg := internal.Config{
		Driver: u.Scheme,
		Host:   u.Hostname(),
	}

	if port := u.Port(); port != "" {
		p, _ := strconv.Atoi(port)
		cfg.Port = p
	} else {
		// Defaults
		switch cfg.Driver {
		case "postgres":
			cfg.Port = 5432
		case "mysql":
			cfg.Port = 3306
		case "mongodb":
			cfg.Port = 27017
		}
	}

	cfg.Database = strings.TrimPrefix(u.Path, "/")

	if u.User != nil {
		cfg.Username = u.User.Username()
		cfg.Password, _ = u.User.Password()
	}

	return cfg, nil
}

// Connect initializes a new database connection
func Connect(connStr string, opts ...Option) (*DB, error) {
	cfg, err := ParseConnectionString(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	db := &DB{
		config: cfg,
	}

	// Apply options
	for _, opt := range opts {
		opt(db)
	}

	// Get driver
	driver, err := internal.GetDriver(db.config.Driver)
	if err != nil {
		return nil, err
	}

	// Connect
	conn, err := driver.Connect(db.config)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	db.conn = conn
	return db, nil
}
