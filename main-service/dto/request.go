package dto

type FlightSearchRequest struct {
	From string `json:"from" validate:"required"`
	To   string `json:"to" validate:"required"`
	// Date       string `json:"date" validate:"required,date"`
	Date       string `json:"date" validate:"required"`
	Passengers int    `json:"passengers" validate:"required,min=1"`
}
