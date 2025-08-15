package service

import (
	"context"
	"time"

	"github.com/yusufekoanggoro/flight-search-system/provider-service/domain"
)

type FlightService interface {
	SearchFlights(ctx context.Context, req *domain.FlightSearchRequest) ([]*domain.FlightSearchResult, error)
}

type flightService struct{}

func NewFlightService() FlightService {
	return &flightService{}
}

// SearchFlights → Simulasi API eksternal
func (s *flightService) SearchFlights(ctx context.Context, req *domain.FlightSearchRequest) ([]*domain.FlightSearchResult, error) {
	// Simulasi delay API call
	time.Sleep(500 * time.Millisecond)

	// Dummy data
	results := []*domain.FlightSearchResult{
		{Airline: "AirAsia", From: req.From, To: req.To, Date: req.Date, PriceUSD: 150.0},
		{Airline: "Garuda", From: req.From, To: req.To, Date: req.Date, PriceUSD: 300.0},
	}

	return results, nil
}
