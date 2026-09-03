package database

import (
	"context"
	"fmt"

	"github.com/masterfabric-go/masterfabric/internal/shared/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// NewMongoClient creates a new MongoDB client from config.
// The URI is read from config (env MONGODB_URI), never hardcoded.
func NewMongoClient(ctx context.Context, cfg config.MongoDBConfig) (*mongo.Client, error) {
	opts := options.Client().ApplyURI(cfg.URI)

	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return client, nil
}

// GetDatabase returns the configured MongoDB database handle.
func GetDatabase(client *mongo.Client, cfg config.MongoDBConfig) *mongo.Database {
	return client.Database(cfg.Database)
}
