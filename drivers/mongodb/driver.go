package mongodb

import (
	"context"
	"errors"
	"fmt"

	"github.com/unidb/unidb-go/internal"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	internal.RegisterDriver("mongodb", &MongoDriver{})
}

type MongoDriver struct{}

func (d *MongoDriver) Connect(config internal.Config) (internal.Connection, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d", config.Username, config.Password, config.Host, config.Port)
	if config.Username == "" && config.Password == "" {
		uri = fmt.Sprintf("mongodb://%s:%d", config.Host, config.Port)
	}

	clientOptions := options.Client().ApplyURI(uri)
	if config.PoolSize > 0 {
		clientOptions.SetMaxPoolSize(uint64(config.PoolSize))
	}

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	db := client.Database(config.Database)

	return &MongoConnection{client: client, db: db}, nil
}

func (d *MongoDriver) Capabilities() internal.Capabilities {
	return internal.Capabilities{
		SupportsSQL:          false, // Translates via middleware
		SupportsTransactions: true,  // Depending on mongo setup
		SupportsJoins:        false, // Very limited
		SupportsAggregation:  true,
	}
}

type MongoConnection struct {
	client *mongo.Client
	db     *mongo.Database
}

func (c *MongoConnection) Query(ctx context.Context, query string, args ...any) (internal.Result, error) {
	// For MVP, just return an error since NoSQL translation happens at a higher level
	// or requires specialized logic
	return nil, errors.New("mongodb query requires AST translation")
}

func (c *MongoConnection) Exec(ctx context.Context, query string, args ...any) (internal.ExecResult, error) {
	return nil, errors.New("mongodb exec requires AST translation")
}

func (c *MongoConnection) BeginTx(ctx context.Context) (internal.Tx, error) {
	return nil, errors.New("mongodb transactions not fully implemented in MVP")
}

func (c *MongoConnection) Close() error {
	return c.client.Disconnect(context.Background())
}
