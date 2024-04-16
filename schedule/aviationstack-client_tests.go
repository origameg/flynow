package schedule

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestFindUniqueDestinations(t *testing.T) {

	// Arrange
	const expected = 22
	fakeResponse, err := getFakeResponseData()
	if err != nil {
		t.Skip("Unable to parse test data: %w", err)
	}

	// Act
	actual := findUniqueDestinations("OSL", fakeResponse.Flights)

	// Assert
	if len(actual) != expected {
		t.Errorf("Found %d results; Expected %d", len(actual), expected)
	}
}

func getFakeResponseData() (fakeFlights flightsResponse, err error) {

	data, err := os.ReadFile("./schedule/sample-scheduled-flights.json")
	if err != nil {
		return fakeFlights, fmt.Errorf("reading json file: %w", err)
	}

	var flightData flightsResponse
	err = json.Unmarshal(data, &flightData)
	if err != nil {
		return fakeFlights, fmt.Errorf("parsing test data: %w", err)
	}

	return flightData, nil
}
