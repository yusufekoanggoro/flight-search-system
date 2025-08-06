package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/yusufekoanggoro/flight-search-system/main-service/dto"
	"github.com/yusufekoanggoro/flight-search-system/main-service/repository"
)

type Usecase interface {
	SubmitSearch(ctx context.Context, req dto.FlightSearchRequest) (*dto.FlightSearchResponse, error)
	StreamSearchResults(
		ctx context.Context,
		stream string,
		group string,
		consumerID string,
		expectedID string,
	) (<-chan redis.XMessage, error)
}

type usecase struct {
	redisRepo repository.RedisRepository
}

func NewFlightUsecase(redisRepo repository.RedisRepository) Usecase {
	return &usecase{redisRepo: redisRepo}
}

func (u *usecase) SubmitSearch(ctx context.Context, req dto.FlightSearchRequest) (*dto.FlightSearchResponse, error) {
	searchID := uuid.NewString()

	payload := map[string]interface{}{
		"search_id":  searchID,
		"from":       req.From,
		"to":         req.To,
		"date":       req.Date,
		"passengers": req.Passengers,
	}

	if err := u.redisRepo.Publish(ctx, "flight.search.requested", payload); err != nil {
		return nil, err
	}

	return &dto.FlightSearchResponse{
		SearchID: searchID,
		Status:   "processing",
	}, nil
}

func (u *usecase) StreamSearchResults(
	ctx context.Context,
	stream string,
	group string,
	consumerID string,
	expectedID string,
) (<-chan redis.XMessage, error) {
	return u.redisRepo.SubscribeToSearchResults(ctx, stream, group, consumerID, expectedID)
}
