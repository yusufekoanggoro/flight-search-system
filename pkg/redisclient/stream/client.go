package stream

import (
	"context"
	"encoding/json"
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

func (c *client) Add(ctx context.Context, stream string, values interface{}) error {
	data, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{
			"data": string(data),
		},
	}

	id, err := c.client.XAdd(ctx, args).Result()
	if err != nil {
		log.Printf("failed to add to stream %s: %v", stream, err)
		return fmt.Errorf("failed to add to stream %s: %w", stream, err)
	}
	log.Printf("add to stream %s with id %s", stream, id)
	return nil
}
