package main

import (
	"flynow/pricing"
	"flynow/schedule"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {

	// TODO: Pass in these values (e.g. as part of a REST call)
	const origin = "OSL"
	const currency = "NOK"
	const orderBy = "price"

	fmt.Printf("Searching for potential destinations from %s...\n", origin)

	// Get a list of destination airports, based on real-time scheduled flights
	scheduleClient := schedule.GetClient()
	destinations, err := scheduleClient.GetScheduledDestinations(origin)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("\nSearching for flights departing today to:")
	fmt.Println(strings.Join(destinations, ","))

	// Perform a series of flight searches to find the cheapest option for each of the possible destinations
	flightOptions, err := pricing.FindPrices(origin, destinations, currency)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Sort by the desired field
	switch orderBy {
	case "price":
		sort.Sort(pricing.ByPrice(flightOptions))
	case "time":
		sort.Sort(pricing.ByTime(flightOptions))
	case "dest":
		sort.Sort(pricing.ByDest(flightOptions))
	}

	// Output the results
	fmt.Println("\nFound the following flights")
	printResults(flightOptions)
}

// Print a formatted table, showing the resulting flight options
func printResults(flightOptions []pricing.FlightForPurchase) {
	fmt.Printf("%s\t%s\t%s\t\t%s\t\t%s\n", "From", "To", "Departing", "Arriving", "Price")
	fmt.Println("______________________________________________________________________")

	for _, f := range flightOptions {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", f.Origin, f.Destination, f.Departure.Format("2006-01-02 15:04"), f.Arrival.Format("2006-01-02 15:04"), f.GetFormattedPrice())
	}

	fmt.Println("______________________________________________________________________")
}
