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

func (c *client) Ack(ctx context.Context, stream, group, messageID string) error {
	n, err := c.client.XAck(ctx, stream, group, messageID).Result()
	if err != nil {
		log.Printf("failed to ack message %s in stream %s: %v", messageID, stream, err)
		return fmt.Errorf("failed to ack message %s: %w", messageID, err)
	}

	if n == 0 {
		log.Printf("warning: message %s was not acknowledged (maybe already acked or doesn't exist)", messageID)
	} else {
		log.Printf("acknowledged message %s in stream %s", messageID, stream)
	}

	return nil
}
