package hotels

import (
	"context"
	"fmt"
	"time"
)

// MockHotelProvider returns fixture data deterministic on the search query.
type MockHotelProvider struct{}

// NewMockHotelProvider returns a new mock provider.
func NewMockHotelProvider() *MockHotelProvider { return &MockHotelProvider{} }

// GetProviderName implements Provider.
func (p *MockHotelProvider) GetProviderName() string { return "mock" }

// SearchHotels implements Provider with three fixture offers.
func (p *MockHotelProvider) SearchHotels(_ context.Context, q HotelSearchQuery) ([]HotelOffer, error) {
	currency := q.Currency
	if currency == "" {
		currency = "USD"
	}
	nights := nightsBetween(q.CheckIn, q.CheckOut)
	if nights == 0 {
		nights = 1
	}

	prototypes := []struct {
		name         string
		address      string
		rating       float32
		costPerNight float64
		lat, lng     float64
	}{
		{"Mock Grand Hotel", "1 Mock Plaza", 4.5, 180, 48.8566, 2.3522},
		{"Mock Boutique Inn", "12 Boutique Street", 4.0, 120, 48.8606, 2.3376},
		{"Mock Budget Stay", "44 Budget Lane", 3.0, 70, 48.8530, 2.3499},
	}

	out := make([]HotelOffer, 0, len(prototypes))
	for i, h := range prototypes {
		if q.MaxPricePerNight > 0 && h.costPerNight > q.MaxPricePerNight {
			continue
		}
		out = append(out, HotelOffer{
			Provider:     p.GetProviderName(),
			OfferID:      fmt.Sprintf("mock-hotel-%d", i+1),
			Name:         h.name + " - " + q.City,
			Address:      h.address + ", " + q.City,
			CheckInDate:  q.CheckIn,
			CheckOutDate: q.CheckOut,
			CostPerNight: h.costPerNight,
			TotalCost:    h.costPerNight * float64(nights),
			Currency:     currency,
			Rating:       h.rating,
			Lat:          h.lat,
			Lng:          h.lng,
		})
	}
	return out, nil
}

func nightsBetween(checkIn, checkOut string) int {
	if checkIn == "" || checkOut == "" {
		return 0
	}
	in, err1 := time.Parse("2006-01-02", checkIn)
	out, err2 := time.Parse("2006-01-02", checkOut)
	if err1 != nil || err2 != nil {
		return 0
	}
	d := int(out.Sub(in).Hours() / 24)
	if d < 0 {
		return 0
	}
	return d
}
