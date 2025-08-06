package repository

import (
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisstream/stream"
)

type RedisRepository interface {
	// Publish(ctx context.Context, stream string, data map[string]interface{}) error
	// SubscribeToSearchResults(ctx context.Context, stream, group, consumerID, expectedID string) (<-chan redis.XMessage, error)
}

type redisRepo struct {
	producer stream.Producer
	consumer stream.Consumer
}

func NewRedisRepository(producer stream.Producer, consumer stream.Consumer) RedisRepository {
	return &redisRepo{producer: producer, consumer: consumer}
}
