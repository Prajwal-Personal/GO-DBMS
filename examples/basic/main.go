package main

import (
	"context"
	"fmt"
	"log"

	"github.com/unidb/unidb-go/api"
	"github.com/unidb/unidb-go/metrics"
	"github.com/unidb/unidb-go/security"
	
	// register drivers
	_ "github.com/unidb/unidb-go/drivers/postgres"
	_ "github.com/unidb/unidb-go/drivers/mysql"
)

func main() {
	// Connect to postgres
	db, err := api.Connect("postgres://user:pass@localhost:5432/mydb", api.WithMaxPool(20))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer db.Close()

	// Apply Middlewares
	secEngine := &security.SecurityEngine{}
	db.Use(security.SecurityMiddleware(secEngine))
	db.Use(metrics.MetricsMiddleware())

	ctx := context.Background()

	// This is a normal query, should succeed.
	res, err := db.Query(ctx, "SELECT id, name FROM users WHERE id = 1")
	if err != nil {
		log.Printf("Query error: %v", err)
	} else {
		fmt.Printf("Query succeeded. Columns: %v\n", res.Columns())
	}

	// This is an injection attempt, should be blocked by security engine.
	_, err = db.Query(ctx, "SELECT * FROM users WHERE id = 1 OR 1=1")
	if err != nil {
		log.Printf("Security engine successfully blocked query: %v", err)
	}
}
