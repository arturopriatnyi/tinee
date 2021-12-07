// Package redis is responsible for caching to Redis.
package redis

import (
	"context"

	"github.com/go-redis/redis/v8"

	"urx/internal/config"
)

// DB represents Redis database.
type DB struct {
	cfg    config.Redis
	client *redis.Client
}

// Open connects to the database and returns DB instance.
func Open(ctx context.Context, cfg config.Redis) (*DB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
	})
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return &DB{cfg: cfg, client: client}, nil
}

// Close closes database connection.
func (db *DB) Close() error {
	return db.client.Close()
}
