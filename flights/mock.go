package flights

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MockFlightProvider returns deterministic fixture flights.
type MockFlightProvider struct{}

// NewMockFlightProvider returns a new mock provider.
func NewMockFlightProvider() *MockFlightProvider { return &MockFlightProvider{} }

// GetProviderName implements Provider.
func (p *MockFlightProvider) GetProviderName() string { return "mock" }

// SearchFlights implements Provider with three fixture offers (different prices/cabins).
func (p *MockFlightProvider) SearchFlights(_ context.Context, q FlightSearchQuery) ([]FlightOffer, error) {
	if q.From == "" || q.To == "" {
		return nil, fmt.Errorf("flights: from and to are required")
	}
	pax := q.Pax
	if pax <= 0 {
		pax = 1
	}
	currency := q.Currency
	if currency == "" {
		currency = "USD"
	}
	cabin := q.Cabin
	if cabin == "" {
		cabin = "economy"
	}

	from := strings.ToUpper(q.From)
	to := strings.ToUpper(q.To)

	depart, err := time.Parse("2006-01-02", q.Depart)
	if err != nil {
		depart = time.Now().Add(48 * time.Hour)
	}
	leg1Depart := depart.Add(8 * time.Hour)
	leg1Arrive := leg1Depart.Add(7 * time.Hour)

	prototypes := []struct {
		offerID string
		carrier string
		flight  string
		price   float64
	}{
		{"mock-flight-1", "AI", "AI201", 480},
		{"mock-flight-2", "AF", "AF445", 540},
		{"mock-flight-3", "LH", "LH763", 620},
	}

	out := make([]FlightOffer, 0, len(prototypes))
	for _, pt := range prototypes {
		segs := []FlightSegment{
			{
				From:         from,
				To:           to,
				Depart:       leg1Depart.Format(time.RFC3339),
				Arrive:       leg1Arrive.Format(time.RFC3339),
				Carrier:      pt.carrier,
				FlightNumber: pt.flight,
				DurationMin:  int(leg1Arrive.Sub(leg1Depart).Minutes()),
			},
		}
		if q.Return != "" {
			ret, err := time.Parse("2006-01-02", q.Return)
			if err == nil {
				retDepart := ret.Add(10 * time.Hour)
				retArrive := retDepart.Add(7 * time.Hour)
				segs = append(segs, FlightSegment{
					From:         to,
					To:           from,
					Depart:       retDepart.Format(time.RFC3339),
					Arrive:       retArrive.Format(time.RFC3339),
					Carrier:      pt.carrier,
					FlightNumber: pt.flight + "R",
					DurationMin:  int(retArrive.Sub(retDepart).Minutes()),
				})
			}
		}

		out = append(out, FlightOffer{
			Provider: p.GetProviderName(),
			OfferID:  pt.offerID,
			Price:    pt.price * float64(pax),
			Currency: currency,
			Cabin:    cabin,
			Segments: segs,
		})
	}
	return out, nil
}
