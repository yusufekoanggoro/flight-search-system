package domain

type FlightSearchRequest struct {
	From       string `json:"from" validate:"required"`
	To         string `json:"to" validate:"required"`
	Date       string `json:"date" validate:"required"`
	Passengers int    `json:"passengers" validate:"required,min=1"`
}

type FlightSearchResponse struct {
	SearchID string `json:"search_id"`
	Status   string `json:"status"`
}
