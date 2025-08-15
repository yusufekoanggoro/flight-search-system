package delivery

import (
	"encoding/json"
	"log"

	"github.com/yusufekoanggoro/flight-search-system/main-service/domain"
	"github.com/yusufekoanggoro/flight-search-system/main-service/shared/sse"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisclient/stream"
)

type RedisHandler struct {
	redisStreamConsumer stream.ConsumerGroup
	hub                 sse.Hub
}

func NewRedisHandler(redisStreamConsumer stream.ConsumerGroup, hub sse.Hub) *RedisHandler {
	return &RedisHandler{
		redisStreamConsumer: redisStreamConsumer,
		hub:                 hub,
	}
}

func (h *RedisHandler) RegisterSubscribers() {
	h.redisStreamConsumer.AddMessageHandler("flight.search.results", h.searchResult)
}

func (h *RedisHandler) searchResult(key, value []byte) error {
	var temp map[string]string
	if err := json.Unmarshal([]byte(value), &temp); err != nil {
		log.Println(err)
	}

	var resp domain.FlightSearchResult
	if err := json.Unmarshal([]byte(temp["data"]), &resp); err != nil {
		log.Println(err)
	} else {
		data := sse.SSEMessage{
			Event:    "search result",
			SearchID: resp.SearchID,
			Data:     []byte("asd"),
		}
		h.hub.SendToOne(&data)
	}
	return nil
}
