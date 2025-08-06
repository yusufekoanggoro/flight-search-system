package redisstream

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Client() *redis.Client
	Shutdown() error
}

type Config struct {
	Addr     string
	Password string
	DB       int
}

type redisClient struct {
	client *redis.Client
}

func NewRedisClient(ctx context.Context, cfg Config) (RedisClient, error) {
	rdc := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdc.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis at %s: %w", cfg.Addr, err)
	}

	log.Printf("Successfully connected to Redis at %s (DB: %d)", cfg.Addr, cfg.DB)

	return &redisClient{
		client: rdc,
	}, nil
}

func (rs *redisClient) Client() *redis.Client {
	return rs.client
}

func (rs *redisClient) Shutdown() error {
	err := rs.client.Close()
	if err != nil {
		log.Printf("Redis shutdown error: %v", err)
		return err
	}

	log.Println("Redis client closed successfully")
	return nil
}
