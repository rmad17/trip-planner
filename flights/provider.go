package flights

import (
	"context"
	"encoding/json"
)

// Provider is the swap surface for flight-search implementations.
type Provider interface {
	SearchFlights(ctx context.Context, q FlightSearchQuery) ([]FlightOffer, error)
	GetProviderName() string
}

// FlightSearchQuery describes a one-way or round-trip flight search.
type FlightSearchQuery struct {
	From     string // IATA code (e.g. DEL)
	To       string // IATA code
	Depart   string // YYYY-MM-DD
	Return   string // YYYY-MM-DD, "" for one-way
	Pax      int
	Cabin    string // economy | premium_economy | business | first
	Currency string
}

// FlightOffer is the canonical flight offer shape.
type FlightOffer struct {
	Provider string          `json:"provider"`
	OfferID  string          `json:"offer_id"`
	Price    float64         `json:"price"`
	Currency string          `json:"currency"`
	Cabin    string          `json:"cabin"`
	Segments []FlightSegment `json:"segments"`
	DeepLink string          `json:"deep_link,omitempty"`
	Raw      json.RawMessage `json:"-"`
}

// FlightSegment is one leg of a flight itinerary.
type FlightSegment struct {
	From         string `json:"from"`   // IATA
	To           string `json:"to"`     // IATA
	Depart       string `json:"depart"` // RFC3339
	Arrive       string `json:"arrive"` // RFC3339
	Carrier      string `json:"carrier"`
	FlightNumber string `json:"flight_number"`
	DurationMin  int    `json:"duration_min"`
}
