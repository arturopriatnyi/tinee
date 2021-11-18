// Package mongodb is responsible for data persistence to MongoDB.
package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"urx/internal/config"
)

// DB represents MongoDB database.
type DB struct {
	cfg    config.MongoDB
	client *mongo.Client
}

// Open connects to the database and returns its instance.
func Open(ctx context.Context, cfg config.MongoDB) (*DB, error) {
	clientOptions := options.Client().ApplyURI(cfg.URL).SetAuth(
		options.Credential{
			Username: cfg.Username,
			Password: cfg.Password,
		},
	)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	return &DB{cfg: cfg, client: client}, client.Ping(ctx, nil)
}

// Collection returns collection by name.
func (db *DB) Collection(name string) *mongo.Collection {
	return db.client.Database(db.cfg.DbName).Collection(name)
}

// Close closes database connection.
func (db *DB) Close(ctx context.Context) error {
	return db.client.Disconnect(ctx)
}
