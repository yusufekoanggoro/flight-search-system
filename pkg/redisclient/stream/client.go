package stream

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type client struct {
	client redis.UniversalClient
}

func NewClient(rdb redis.UniversalClient) Client {
	return &client{
		client: rdb,
	}
}

func (c *client) Add(ctx context.Context, stream string, values map[string]interface{}) error {
	args := &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}
	id, err := c.client.XAdd(ctx, args).Result()
	if err != nil {
		log.Printf("failed to add to stream %s: %v", stream, err)
		return fmt.Errorf("failed to add to stream %s: %w", stream, err)
	}
	log.Printf("add to stream %s with id %s", stream, id)
	return nil
}
