package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yusufekoanggoro/flight-search-system/pkg/redisclient/stream"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/domain"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/service"
)

type Usecase interface {
	ProcessSearch(ctx context.Context, req *domain.FlightSearchRequest) ([]byte, error)
}

type usecase struct {
	redisStream   stream.Client
	flightService service.FlightService
}

func NewFlightUsecase(redisStream stream.Client, flightService service.FlightService) Usecase {
	return &usecase{
		redisStream:   redisStream,
		flightService: flightService,
	}
}

func (u *usecase) ProcessSearch(ctx context.Context, req *domain.FlightSearchRequest) ([]byte, error) {
	results, err := u.flightService.SearchFlights(ctx, req)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to search flights: %w", err)
	}

	response := domain.FlightSearchResponse{
		SearchID: req.SearchID,
		Status:   "completed",
		Results:  results,
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to marshal response: %w", err)
	}
	fmt.Println(response.SearchID)
	u.redisStream.Add(ctx, "flight.search.results", response)

	return bytes, nil
}
