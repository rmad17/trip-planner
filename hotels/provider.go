package hotels

import (
	"context"
	"encoding/json"
)

// Provider is the swap surface for hotel-search implementations.
type Provider interface {
	SearchHotels(ctx context.Context, q HotelSearchQuery) ([]HotelOffer, error)
	GetProviderName() string
}

// HotelSearchQuery describes a hotel availability search.
type HotelSearchQuery struct {
	City             string  // Free-text city or IATA-like city code
	CheckIn          string  // YYYY-MM-DD
	CheckOut         string  // YYYY-MM-DD
	Guests           int
	Currency         string
	MaxPricePerNight float64 // 0 = no cap
}

// HotelOffer mirrors the existing Stay model field names so it can be spread
// into a "create stay" payload on the FE.
type HotelOffer struct {
	Provider     string          `json:"provider"`
	OfferID      string          `json:"offer_id"`
	Name         string          `json:"name"`
	Address      string          `json:"address"`
	CheckInDate  string          `json:"check_in_date"`
	CheckOutDate string          `json:"check_out_date"`
	CostPerNight float64         `json:"cost_per_night"`
	TotalCost    float64         `json:"total_cost"`
	Currency     string          `json:"currency"`
	Rating       float32         `json:"rating,omitempty"`
	Lat          float64         `json:"lat,omitempty"`
	Lng          float64         `json:"lng,omitempty"`
	DeepLink     string          `json:"deep_link,omitempty"`
	Raw          json.RawMessage `json:"-"`
}
