package domain

type FlightSearchResult struct {
	SearchID     string        `json:"search_id"`
	Status       string        `json:"status"`
	Results      []interface{} `json:"results,omitempty"`       // bisa kosong atau array data
	TotalResults *int          `json:"total_results,omitempty"` // bisa kosong atau angka
}
