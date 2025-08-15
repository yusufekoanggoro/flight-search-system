package redisclient

import (
	"context"
	"fmt"
	"strings"

	"log"

	"github.com/redis/go-redis/v9"
)

type client struct {
	client redis.UniversalClient
}

func New(ctx context.Context, config *Config) (Client, error) {
	log.Printf("connecting to redis at %s (db=%d)...", config.Addr, config.DB)

	rdb := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    strings.Split(config.Addr, ","),
		Password: config.Password,
		DB:       config.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Printf("successfully connected to redis at %s (db=%d)", config.Addr, config.DB)
	return &client{client: rdb}, nil
}

func (c *client) Client() redis.UniversalClient {
	return c.client
}

func (c *client) Close() error {
	log.Println("closing Redis client...")

	err := c.client.Close()
	if err != nil {
		log.Printf("failed to close Redis client: %v", err)
		return fmt.Errorf("redis close error: %w", err)
	}

	log.Println("redis client closed successfully")
	return nil
}
