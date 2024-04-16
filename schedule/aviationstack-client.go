package schedule

import (
	"encoding/json"
	"errors"
	"flynow/config"
	"fmt"
	"io"
	"net/http"
)

// Interface for the flight schedule client, so that it can be mocked in consuming code

type ScheduleClient interface {
	GetScheduledDestinations(origin string) ([]string, error)
}

func GetClient() ScheduleClient {
	client := aviationStackClient{}
	return &client
}

type aviationStackClient struct{}

// Performs a REST call to the AviationStack flights endpoint to get a list of realtime flights
// scheduled to depart from the given airport. Note: Although this request uses paged results,
// only the first page is used, due to extremely limited number of API calls allowed on the
// free version.
func (client *aviationStackClient) GetScheduledDestinations(origin string) (destinations []string, err error) {

	const flightsEndpoint = "http://api.aviationstack.com/v1/flights"

	// TODO: Read this from secure storage
	apiKey := config.GetAviationStackCredentials()

	request, err := http.NewRequest(http.MethodGet, flightsEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	query := request.URL.Query()
	query.Add("access_key", apiKey)
	query.Add("dep_iata", origin)
	query.Add("flight_status", "scheduled")
	request.URL.RawQuery = query.Encode()

	httpClient := &http.Client{}

	// Call the flights API
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("requesting scheduled flights: %w", err)
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected response code (%d) getting scheduled flights", response.StatusCode)
	}

	// Read and parse the response data
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var scheduledFlights flightsResponse
	json.Unmarshal(responseBody, &scheduledFlights)

	if scheduledFlights.Page.Count > 0 && scheduledFlights.Flights == nil {
		err = errors.New("missing flight info in aviationstack response")
		return nil, err
	}

	// Find the unique destinations based on the realtime flight data
	return findUniqueDestinations(origin, scheduledFlights.Flights), nil
}

func findUniqueDestinations(origin string, scheduledFlights []flightInfo) (destinations []string) {

	destMap := make(map[string]bool)

	// Find flights departing on the given day
	for _, flight := range scheduledFlights {

		// This shouldn't happen, but just in case
		if flight.Departure.Airport != origin {
			continue
		}

		destMap[flight.Arrival.Airport] = true
	}

	destinations = make([]string, 0, len(destMap))
	for dest := range destMap {
		destinations = append(destinations, dest)
	}

	return destinations
}
