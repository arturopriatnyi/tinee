package redis

import (
	"context"
	"encoding/json"

	"tinee/internal/service"
)

// LinkCache is the link cache.
type LinkCache struct {
	db *DB
}

// NewLinkCache creates and returns a new LinkCache instance.
func NewLinkCache(db *DB) *LinkCache {
	return &LinkCache{db: db}
}

// Set saves a service.Link to Redis with alias key.
func (c *LinkCache) Set(ctx context.Context, alias string, l service.Link) error {
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}

	_, err = c.db.client.Set(ctx, alias, s, 0).Result()

	return err
}

// Get gets service.Link by alias from Redis.
func (c *LinkCache) Get(ctx context.Context, alias string) (l service.Link, err error) {
	s, err := c.db.client.Get(ctx, alias).Result()
	if err != nil {
		return service.Link{}, err
	}

	return l, json.Unmarshal([]byte(s), &l)
}
