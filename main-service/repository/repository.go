package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/yusufekoanggoro/flight-search-system/main-service/dto"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisstream/stream"
)

type RedisRepository interface {
	Publish(ctx context.Context, stream string, data map[string]interface{}) error
	SubscribeToSearchResults(ctx context.Context, stream, group, consumerID, expectedID string) (<-chan redis.XMessage, error)
}

type redisRepo struct {
	producer stream.Producer
	consumer stream.Consumer
}

func NewRedisRepository(producer stream.Producer, consumer stream.Consumer) RedisRepository {
	return &redisRepo{producer: producer, consumer: consumer}
}

func (r *redisRepo) Publish(ctx context.Context, stream string, data map[string]interface{}) error {
	err := r.producer.SendMessage(ctx, stream, data)
	if err != nil {
		return err
	}
	return nil
}

func (r *redisRepo) SubscribeToSearchResults(
	ctx context.Context,
	stream, group, consumerID, expectedID string,
) (<-chan redis.XMessage, error) {
	messageChan := make(chan redis.XMessage, 10)
	go func() {
		defer close(messageChan)
		log.Printf("[SSE] Start consumer for search_id=%s (consumerID=%s)", expectedID, consumerID)

		err := r.consumer.SubscribeGroup(ctx, stream, group, consumerID, func(msg redis.XMessage, rawData string) (bool, error) {

			var result dto.FlightSearchResult
			if err := json.Unmarshal([]byte(rawData), &result); err != nil {
				log.Println("unmarshal error:", err)
				return false, fmt.Errorf("failed to unmarshal: %w", err)
			}

			if result.SearchID != expectedID {
				log.Printf("Message %s skipped (search ID mismatch)", msg.ID)
				return false, nil
			}

			select {
			case messageChan <- msg:
				return true, nil
			case <-ctx.Done():
				log.Printf("[SSE] Context done before sending message %s", msg.ID)
				return false, nil // tidak ACK
			}
		}, ">")
		if err != nil {
			log.Printf("[SSE] Error from consumer: %v", err)
		}
	}()
	return messageChan, nil
}
