package flights

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"triplanner/core"
)

// AmadeusFlightProvider wraps the shared Amadeus client for flight offers.
type AmadeusFlightProvider struct {
	Client *core.AmadeusClient
}

// NewAmadeusFlightProvider builds the provider with the shared Amadeus client.
func NewAmadeusFlightProvider() *AmadeusFlightProvider {
	return &AmadeusFlightProvider{Client: core.DefaultAmadeusClient()}
}

// GetProviderName implements Provider.
func (p *AmadeusFlightProvider) GetProviderName() string { return "amadeus" }

type amadeusFlightOffersResponse struct {
	Data []struct {
		Type        string `json:"type"`
		ID          string `json:"id"`
		Source      string `json:"source"`
		OneWay      bool   `json:"oneWay"`
		Itineraries []struct {
			Duration string `json:"duration"`
			Segments []struct {
				Departure struct {
					IataCode string `json:"iataCode"`
					Terminal string `json:"terminal"`
					At       string `json:"at"`
				} `json:"departure"`
				Arrival struct {
					IataCode string `json:"iataCode"`
					Terminal string `json:"terminal"`
					At       string `json:"at"`
				} `json:"arrival"`
				CarrierCode string `json:"carrierCode"`
				Number      string `json:"number"`
				Duration    string `json:"duration"`
			} `json:"segments"`
		} `json:"itineraries"`
		Price struct {
			Currency   string `json:"currency"`
			Total      string `json:"total"`
			GrandTotal string `json:"grandTotal"`
		} `json:"price"`
		TravelerPricings []struct {
			FareDetailsBySegment []struct {
				Cabin string `json:"cabin"`
			} `json:"fareDetailsBySegment"`
		} `json:"travelerPricings"`
	} `json:"data"`
}

// SearchFlights implements Provider.
func (p *AmadeusFlightProvider) SearchFlights(ctx context.Context, q FlightSearchQuery) ([]FlightOffer, error) {
	if !p.Client.Configured() {
		return nil, fmt.Errorf("amadeus: API credentials not configured")
	}
	if q.From == "" || q.To == "" {
		return nil, fmt.Errorf("flights: from and to are required")
	}
	if q.Depart == "" {
		return nil, fmt.Errorf("flights: depart date is required")
	}

	pax := q.Pax
	if pax <= 0 {
		pax = 1
	}

	v := url.Values{}
	v.Set("originLocationCode", strings.ToUpper(q.From))
	v.Set("destinationLocationCode", strings.ToUpper(q.To))
	v.Set("departureDate", q.Depart)
	if q.Return != "" {
		v.Set("returnDate", q.Return)
	}
	v.Set("adults", strconv.Itoa(pax))
	v.Set("max", "10")
	if q.Currency != "" {
		v.Set("currencyCode", strings.ToUpper(q.Currency))
	}
	if cabin := normalizeCabin(q.Cabin); cabin != "" {
		v.Set("travelClass", cabin)
	}

	var resp amadeusFlightOffersResponse
	if err := p.Client.DoJSON(ctx, "/v2/shopping/flight-offers", v, &resp); err != nil {
		return nil, fmt.Errorf("amadeus flights: %w", err)
	}

	out := make([]FlightOffer, 0, len(resp.Data))
	for _, item := range resp.Data {
		price, _ := strconv.ParseFloat(item.Price.GrandTotal, 64)
		if price == 0 {
			price, _ = strconv.ParseFloat(item.Price.Total, 64)
		}

		segments := make([]FlightSegment, 0)
		for _, itin := range item.Itineraries {
			for _, s := range itin.Segments {
				segments = append(segments, FlightSegment{
					From:         s.Departure.IataCode,
					To:           s.Arrival.IataCode,
					Depart:       s.Departure.At,
					Arrive:       s.Arrival.At,
					Carrier:      s.CarrierCode,
					FlightNumber: s.CarrierCode + s.Number,
					DurationMin:  parseISO8601Min(s.Duration),
				})
			}
		}

		cabin := q.Cabin
		if len(item.TravelerPricings) > 0 && len(item.TravelerPricings[0].FareDetailsBySegment) > 0 {
			cabin = strings.ToLower(item.TravelerPricings[0].FareDetailsBySegment[0].Cabin)
		}

		raw, _ := json.Marshal(item)

		out = append(out, FlightOffer{
			Provider: p.GetProviderName(),
			OfferID:  item.ID,
			Price:    price,
			Currency: item.Price.Currency,
			Cabin:    cabin,
			Segments: segments,
			Raw:      raw,
		})
	}
	return out, nil
}

func normalizeCabin(c string) string {
	switch strings.ToLower(strings.TrimSpace(c)) {
	case "", "economy":
		return "ECONOMY"
	case "premium_economy", "premium economy":
		return "PREMIUM_ECONOMY"
	case "business":
		return "BUSINESS"
	case "first":
		return "FIRST"
	}
	return ""
}

// parseISO8601Min parses an ISO-8601 duration like "PT9H30M" into total minutes.
func parseISO8601Min(d string) int {
	if d == "" || !strings.HasPrefix(d, "PT") {
		return 0
	}
	body := d[2:]
	total := 0
	num := strings.Builder{}
	for _, ch := range body {
		if ch >= '0' && ch <= '9' {
			num.WriteRune(ch)
			continue
		}
		n, _ := strconv.Atoi(num.String())
		num.Reset()
		switch ch {
		case 'H':
			total += n * 60
		case 'M':
			total += n
		case 'S':
			// ignore
		}
	}
	// Fallback for plain "PT...S" parsable by time.ParseDuration.
	if total == 0 {
		if dur, err := time.ParseDuration(strings.ToLower(body)); err == nil {
			return int(dur.Minutes() + 0.5)
		}
	}
	return total
}
