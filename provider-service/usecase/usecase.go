package usecase

import (
	"context"
	"log"

	"github.com/yusufekoanggoro/flight-search-system/provider-service/dto"
	"github.com/yusufekoanggoro/flight-search-system/provider-service/repository"
)

type Usecase interface {
	ProcessFlightSearchRequest(ctx context.Context, payload dto.FlightSearchRequest) error
}

type usecase struct {
	redisRepo repository.RedisRepository
}

func NewUsecase(redisRepo repository.RedisRepository) Usecase {
	return &usecase{redisRepo: redisRepo}
}

func (u *usecase) ProcessFlightSearchRequest(ctx context.Context, payload dto.FlightSearchRequest) error {
	log.Printf("[Usecase] Processing flight search for search_id=%s, origin=%s, destination=%s, date=%s",
		payload.SearchID, payload.From, payload.To, payload.Date)

	// TODO: Tambahkan logika pencarian penerbangan (call provider, simpan ke DB, publish stream, dll)

	// Simulasi pemrosesan
	// log.Printf("[Usecase] Done processing search_id=%s", payload.SearchID)
	return nil
}
