package domain

type FlightSearchRequest struct {
	SearchID string `json:"search_id"`
	From     string `json:"from"`
	To       string `json:"to"`
	Date     string `json:"date"`
}
