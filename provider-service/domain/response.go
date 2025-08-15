package domain

type FlightSearchResult struct {
	Airline  string  `json:"airline"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Date     string  `json:"date"`
	PriceUSD float64 `json:"price_usd"`
}

type FlightSearchResponse struct {
	SearchID string                `json:"search_id"`
	Status   string                `json:"status"`
	Results  []*FlightSearchResult `json:"results"`
}
