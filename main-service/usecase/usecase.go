package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/yusufekoanggoro/flight-search-system/main-service/domain"
	"github.com/yusufekoanggoro/flight-search-system/pkg/redisclient/stream"
)

type Usecase interface {
	SubmitSearch(ctx context.Context, req *domain.FlightSearchRequest) (*domain.FlightSearchResponse, error)
}

type usecase struct {
	redisStream stream.Client
}

func NewFlightUsecase(redisStream stream.Client) Usecase {
	return &usecase{
		redisStream: redisStream,
	}
}

func (u *usecase) SubmitSearch(ctx context.Context, req *domain.FlightSearchRequest) (*domain.FlightSearchResponse, error) {
	searchID := uuid.NewString()

	payload := map[string]interface{}{
		"search_id":  searchID,
		"from":       req.From,
		"to":         req.To,
		"date":       req.Date,
		"passengers": req.Passengers,
	}

	u.redisStream.Add(ctx, "flight.search.requested", payload)

	return &domain.FlightSearchResponse{
		SearchID: searchID,
		Status:   "processing",
	}, nil
}
