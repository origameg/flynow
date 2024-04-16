package schedule

// Relevant partial models for the AviationStack REST API response

// Top-level response model for real-time flight searches
type flightsResponse struct {
	Page    pagination   `json:"pagination"`
	Flights []flightInfo `json:"data"`
}

// Pagination info for flight list
type pagination struct {
	PageLimit int `json:"limit"`
	Offset    int `json:"offset"`
	Count     int `json:"count"`
	Total     int `json:"total"`
}

// Information for a single returned flight
type flightInfo struct {
	Date      string       `json:"flight_date"`
	Status    string       `json:"flight_status"`
	Departure flightTime   `json:"departure"`
	Arrival   flightTime   `json:"arrival"`
	Airline   airline      `json:"airline"`
	Number    flightNumber `json:"flight"`
}

// Arrival or Departure time and location
type flightTime struct {
	Airport string `json:"iata"`
	Time    string `json:"scheduled"`
}

// Airline information
type airline struct {
	Code string `json:"iata"`
}

// Flight number
type flightNumber struct {
	Number string `json:"iata"`
}
