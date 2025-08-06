package redisstream

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisstream/stream"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/dto"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/usecase"
)

type ConsumerHandler struct {
	rootCtx  context.Context
	uc       usecase.Usecase
	validate *validator.Validate
	consumer stream.Consumer
}

func NewConsumerHandler(rootCtx context.Context, uc usecase.Usecase, consumer stream.Consumer) *ConsumerHandler {
	return &ConsumerHandler{
		rootCtx:  rootCtx,
		uc:       uc,
		consumer: consumer,
		validate: validator.New(),
	}
}

func (h *ConsumerHandler) RegisterSubscribers(ctx context.Context) {
	streamName := "flight.search.requested"
	group := "flight.search.group"
	consumerID := "provider.service.1"
	h.consumer.SubscribeGroup(ctx, streamName, group, consumerID, h.consumeFlightSearchRequested, "*")
}

func (h *ConsumerHandler) consumeFlightSearchRequested(msg redis.XMessage, rawData string) (bool, error) {
	var payload dto.FlightSearchRequest

	// Unmarshal pesan JSON
	if err := json.Unmarshal([]byte(rawData), &payload); err != nil {
		log.Printf("[FlightSearch] Unmarshal error: %v", err)
		return false, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Validasi
	if err := h.validate.Struct(payload); err != nil {
		log.Printf("[FlightSearch] Validation failed: %v", err)
		return true, nil // Tetap ACK untuk hindari retry terus
	}

	if err := h.uc.ProcessFlightSearchRequest(h.rootCtx, payload); err != nil {
		log.Printf("[FlightSearch] Usecase error: %v", err)
		return false, err // Tidak ACK agar bisa retry
	}

	log.Printf("[FlightSearch] Successfully processed search_id=%s", payload.SearchID)
	return true, nil
}
