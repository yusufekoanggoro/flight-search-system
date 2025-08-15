package delivery

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisclient/stream"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/domain"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/usecase"
)

type RedisHandler struct {
	usecase             usecase.Usecase
	redisStreamConsumer stream.ConsumerGroup
	validate            *validator.Validate
}

func NewRedisHandler(usecase usecase.Usecase, redisStreamConsumer stream.ConsumerGroup) *RedisHandler {
	return &RedisHandler{
		usecase:             usecase,
		redisStreamConsumer: redisStreamConsumer,
		validate:            validator.New(),
	}
}

func (h *RedisHandler) RegisterSubscribers() {
	h.redisStreamConsumer.AddMessageHandler("flight.search.requested", h.processSearch)
}

func (h *RedisHandler) processSearch(key, value []byte) error {
	var temp map[string]string
	if err := json.Unmarshal([]byte(value), &temp); err != nil {
		log.Println(err)
	}

	var req domain.FlightSearchRequest
	if err := json.Unmarshal([]byte(temp["data"]), &req); err != nil {
		log.Println(err)
		return err
	}

	if err := h.validate.Struct(&req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if _, err := h.usecase.ProcessSearch(context.Background(), &req); err != nil {
		return fmt.Errorf("failed to submit flight search: %w", err)
	}

	return nil
}
