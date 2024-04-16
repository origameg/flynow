package pricing

import (
	"fmt"
	"strconv"
	"time"
)

// Flight model to be shared outside the Amadeus-base price search package
type FlightForPurchase struct {
	FlightNumber string
	Origin       string
	Destination  string
	Departure    time.Time
	Arrival      time.Time
	Price        float32
	Currency     string
}

// Converts the Amadeus JSON model for a flight offer into the shared data model
// (Given how common this kind of conversion is, I expect there is a standard way
// to do it, but I didn't find an example right away. And it was a good exercise to
// practice basic type conversions.)
func convert(offer *flightOffer) FlightForPurchase {

	singleFlight := offer.Itineraries[0].Segments[0]

	var flight FlightForPurchase

	flight.FlightNumber = singleFlight.Airline + singleFlight.Number
	flight.Origin = singleFlight.Departure.Airport
	flight.Destination = singleFlight.Arrival.Airport

	if dep, err := time.Parse("2006-01-02T15:04:05", singleFlight.Departure.Time); err == nil {
		flight.Departure = dep
	} else {
		logWarning(fmt.Sprintf("Unexpected value for Departure time: %s", singleFlight.Departure.Time))
	}
	if arr, err := time.Parse("2006-01-02T15:04:05", singleFlight.Arrival.Time); err == nil {
		flight.Arrival = arr
	} else {
		logWarning(fmt.Sprintf("Unexpected value for Arrival time: %s", singleFlight.Arrival.Time))
	}

	if p, err := strconv.ParseFloat(offer.Price.Total, 64); err == nil {
		flight.Price = float32(p)
	} else {
		logWarning(fmt.Sprintf("Unexpected value for TotalPrice: %s", offer.Price.Total))
	}
	flight.Currency = offer.Price.Currency

	return flight
}

// Stringer implementation, providing basic information
func (flight FlightForPurchase) String() string {
	return fmt.Sprintf("%s : %s-%s %v -- %v", flight.FlightNumber, flight.Origin, flight.Destination, flight.Departure.Format("15:04"), flight.GetFormattedPrice())
}

// Returns the full flight information in a user-friendly multi-line representation
func (flight FlightForPurchase) GetMultilineString() string {
	s1 := fmt.Sprintf("%s : %s - %s\n", flight.FlightNumber, flight.Origin, flight.Destination)
	s2 := fmt.Sprintf("Departing %v\nArriving %v\n", flight.Departure.Format("2006-01-02 15:04"), flight.Arrival.Format("2006-01-02 15:04"))
	s3 := flight.GetFormattedPrice()
	return s1 + s2 + s3
}

// Provides price information in a currency-specific display format
func (flight FlightForPurchase) GetFormattedPrice() string {

	switch c := flight.Currency; c {
	case "NOK":
		return fmt.Sprintf("%d NOK", int(flight.Price))
	case "EUR":
		return fmt.Sprintf("€%.2f", flight.Price)
	case "USD":
		return fmt.Sprintf("$%.2f", flight.Price)
	default:
		return fmt.Sprintf("%.2f %s", flight.Price, c)
	}
}

// Enable sorting by price
type ByPrice []FlightForPurchase

func (a ByPrice) Len() int           { return len(a) }
func (a ByPrice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPrice) Less(i, j int) bool { return a[i].Price < a[j].Price }

// Enable sorting by destination
type ByDest []FlightForPurchase

func (a ByDest) Len() int           { return len(a) }
func (a ByDest) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDest) Less(i, j int) bool { return a[i].Destination < a[j].Destination }

// Enable sorting by departure time
type ByTime []FlightForPurchase

func (a ByTime) Len() int           { return len(a) }
func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime) Less(i, j int) bool { return a[i].Departure.Before(a[j].Departure) }
