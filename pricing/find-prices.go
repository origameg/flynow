package pricing

import (
	"encoding/json"
	"errors"
	"flynow/config"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Given a departure airport code, and a list of possible destination airports,
// search for flight options, and identify the cheapest flight to each one.
// Note: Although the Amadeus API does include an open-ended flight search, it does
// not appear to be supported for OSL
func FindPrices(origin string, destinations []string, currencyCode string) (results []FlightForPurchase, err error) {

	// Get Authorization token for Amadeus API
	token, err := getAmadeusToken()
	if err != nil {
		return results, fmt.Errorf("creating HTTP request: %w", err)
	}
	// TODO: If this method is called more than once (i.e. as part of a service), cache the token based on its expiry

	// Create channels to store flight options and error information
	flights := make(chan FlightForPurchase)
	errs := make(chan error)

	// Perform a flight search for each destination
	wg := new(sync.WaitGroup)
	for _, destCode := range destinations {
		wg.Add(1)
		go func(destCode string) {
			defer wg.Done()
			getCheapestFlight(origin, destCode, currencyCode, token, flights, errs)
		}(destCode)
	}

	// Wait until all searches have completed
	go func(wg *sync.WaitGroup, flights chan FlightForPurchase, errs chan error) {
		wg.Wait()
		close(flights)
		close(errs)
	}(wg, flights, errs)

	// Combine any errors (hopefully none🤞)
	if len(errs) > 0 {
		err = errors.New("finding flight prices: ")
		for nextErr := range errs {
			err = fmt.Errorf("%w; %w", err, nextErr)
		}
		return nil, err
	}

	// Create a slice with all the results
	results = make([]FlightForPurchase, 0, len(flights))
	for f := range flights {
		// Skip any unitialized struct values - Is there a best practice for this?
		if f.FlightNumber != "" {
			results = append(results, f)
		}
	}

	return results, nil
}

// Performs a REST call to the Amadeus token endpoint to get a Bearer token for use in all
// subsequent Amadeus API requests. Note that a valid client ID and client secret must be provided
// but have not been checked into source code.
func getAmadeusToken() (string, error) {

	const tokenEndpoint = "https://test.api.amadeus.com/v1/security/oauth2/token"

	// TODO: Read these from secure storage
	clientId, clientSecret := config.GetAmadeusCredentials()

	// Construct the POST request body as URL-encoded form data
	bodyData := url.Values{}
	bodyData.Set("grant_type", "client_credentials")
	bodyData.Set("client_id", clientId)
	bodyData.Set("client_secret", clientSecret)
	body := bodyData.Encode()

	request, err := http.NewRequest(http.MethodPost, tokenEndpoint, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating HTTP request: %w", err)
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	var httpClient = &http.Client{}

	// Call the token API
	response, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("requesting Amadeus token: %w", err)
	}

	if response.StatusCode != 200 {
		return "", fmt.Errorf("unexpected response code (%d) getting token", response.StatusCode)
	}

	// Read and parse the response data
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	var jsonToken amadeusToken
	json.Unmarshal(responseBody, &jsonToken)

	return jsonToken.AccessToken, nil
}

// Performs a REST call to the Amadeus flight search API to retrieve flight offers for direct flights on the given route,
// departing today. The results are then evaluated to identify the cheapest option.
func getCheapestFlight(originCode string, destCode string, currencyCode string, token string, flights chan FlightForPurchase, errs chan error) {

	const searchEndpoint = "https://test.api.amadeus.com/v2/shopping/flight-offers"
	today := time.Now()

	// Construct the API request
	request, err := http.NewRequest(http.MethodGet, searchEndpoint, nil)
	if err != nil {
		errs <- fmt.Errorf("creating HTTP request: %w", err)
	}

	// Set the authorization header
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Set the query parameters
	query := request.URL.Query()
	query.Add("originLocationCode", originCode)
	query.Add("destinationLocationCode", destCode)
	query.Add("departureDate", today.Format("2006-01-02"))
	query.Add("adults", fmt.Sprint(1))
	query.Add("travelClass", "ECONOMY")
	query.Add("nonStop", "true")
	query.Add("currencyCode", currencyCode)
	request.URL.RawQuery = query.Encode()

	httpClient := &http.Client{}

	var response *http.Response
	backoffSeconds := int32(1)
	for response == nil || response.StatusCode == 429 {

		// Call the API
		response, err = httpClient.Do(request)
		if err != nil {
			errs <- fmt.Errorf("getting Amadeus flights: %w", err)
			return
		}

		// Perform an exponential backoff & retry on 429
		if response.StatusCode == 429 {
			backoff := ((backoffSeconds * 1000) + rand.Int31n(500))
			duration := time.Duration(backoff) * time.Millisecond
			time.Sleep(duration)
			backoffSeconds = backoffSeconds * 2
		}
	}

	if response.StatusCode != 200 {
		errs <- fmt.Errorf("unexpected response code (%d) searching for flights to %s", response.StatusCode, destCode)
		return
	}

	// Read and parse the response
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		errs <- fmt.Errorf("reading response body: %w", err)
		return
	}

	var flightResults flightSearchResponse
	json.Unmarshal(responseBody, &flightResults)

	// Find the cheapest option (if any) that actually matches the input criteria
	if found, flight := evaluateFlights(&flightResults, originCode, destCode, currencyCode); found {
		flights <- flight
	}
}

// Given the parsed JSON response from the flight search, identify the cheapest flight offer that
// matches the given input parameters. Although the flight search *should* return only direct flights
// between the origin and destination, there is some room for discrepancy. For example, Amadeus may
// return flights from TRF, even though the IATA code "OSL" specifically designates Gardermoen
func evaluateFlights(response *flightSearchResponse, originCode string, destCode string, currencyCode string) (found bool, result FlightForPurchase) {

	// Confirm the count is correct
	if response.Metadata.Count != len(response.Flights) {
		logWarning(fmt.Sprintf("Response contained %d flight offers, but  Count was %d", len(response.Flights), response.Metadata.Count))
	}

	// Return false if there are no results
	if len(response.Flights) == 0 {
		return false, result
	}

	var cheapestFlight *flightOffer = nil

	// Loop over the offers to find the cheapest one
	for i := 0; i < response.Metadata.Count; i++ {
		offer := response.Flights[i]

		// Verify that it's a single-leg flight
		if len(offer.Itineraries) > 1 || len(offer.Itineraries[0].Segments) > 1 {
			logWarning("Offer contained more than one flight")
			continue
		}
		flightInfo := offer.Itineraries[0].Segments[0]

		// Verify the airport codes (i.e. NOT TORP!!! 😜)
		if flightInfo.Departure.Airport != originCode || flightInfo.Arrival.Airport != destCode {
			logWarning(fmt.Sprintf("Offer contained incorrect flight: %s - %s", flightInfo.Departure.Airport, flightInfo.Arrival.Airport))
			continue
		}

		// Check the price
		if cheapestFlight == nil || offer.Price.Total < cheapestFlight.Price.Total {
			if offer.Price.Currency != currencyCode {
				logWarning(fmt.Sprintf("Offer was in incorrect currency: %s", offer.Price.Currency))
				continue
			} else {
				cheapestFlight = &offer
			}
		}
	}

	if cheapestFlight != nil {
		result = convert(cheapestFlight)
		return true, result
	}

	return false, result
}

// Placeholder function for future logging enhancements
func logWarning(message string) {
	// TODO: Introduce a more robust logging infrastructure
	//fmt.Printf("WARNING: %s\n", message)
}
