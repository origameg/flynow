package pricing

// Relevant partial models for the Amadeus REST API responses

// Response model for requests to get API Bearer token
type amadeusToken struct {
	AccessToken string `json:"access_token"`
	Expiry      int    `json:"expires_in"`
}

// Top-level response model for flight searches
type flightSearchResponse struct {
	Metadata offersMetadata `json:"meta"`
	Flights  []flightOffer  `json:"data"`
}

// General information about flight offers returned
type offersMetadata struct {
	Count int `json:"count"`
}

// A single offer within the flight search response
type flightOffer struct {
	SeatsAvailable int         `json:"numberOfBookableSeats"`
	Price          price       `json:"price"`
	Itineraries    []itinerary `json:"itineraries"`
}

// Price of the flight offer
type price struct {
	Total    string `json:"grandTotal"`
	Currency string `json:"currency"`
}

// Full multi-segment flight information
type itinerary struct {
	Segments []flight `json:"segments"`
}

// Single flight
type flight struct {
	Departure flightTime `json:"departure"`
	Arrival   flightTime `json:"arrival"`
	Airline   string     `json:"carrierCode"`
	Number    string     `json:"number"`
}

// Arrival or Departure time and location
type flightTime struct {
	Airport string `json:"iataCode"`
	Time    string `json:"at"`
}
