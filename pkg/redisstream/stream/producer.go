package stream

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Producer interface {
	SendMessage(ctx context.Context, stream string, values map[string]interface{}) error
}

type producer struct {
	client *redis.Client
}

func NewProducer(client *redis.Client) Producer {
	return &producer{
		client: client,
	}
}

func (p *producer) SendMessage(ctx context.Context, stream string, values map[string]interface{}) error {
	args := &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}
	id, err := p.client.XAdd(ctx, args).Result()
	if err != nil {
		return fmt.Errorf("failed to produce to stream %s: %w", stream, err)
	}
	fmt.Printf("Produced to stream %s with ID %s\n", stream, id)
	return nil
}
