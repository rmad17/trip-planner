package hotels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"triplanner/core"
)

// AmadeusHotelProvider wraps the shared Amadeus client for hotel offers.
type AmadeusHotelProvider struct {
	Client *core.AmadeusClient
}

// NewAmadeusHotelProvider builds the provider with the shared Amadeus client.
func NewAmadeusHotelProvider() *AmadeusHotelProvider {
	return &AmadeusHotelProvider{Client: core.DefaultAmadeusClient()}
}

// GetProviderName implements Provider.
func (p *AmadeusHotelProvider) GetProviderName() string { return "amadeus" }

// amadeusCityListResponse maps the /v1/reference-data/locations/cities response.
type amadeusCityListResponse struct {
	Data []struct {
		IataCode string `json:"iataCode"`
		Name     string `json:"name"`
	} `json:"data"`
}

// amadeusHotelsByCityResponse is the v1 hotels-by-city listing payload.
type amadeusHotelsByCityResponse struct {
	Data []struct {
		HotelID  string  `json:"hotelId"`
		Name     string  `json:"name"`
		IataCode string  `json:"iataCode"`
		Latitude float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"data"`
}

// amadeusHotelOffersResponse is the v3/shopping/hotel-offers payload.
type amadeusHotelOffersResponse struct {
	Data []struct {
		Type  string `json:"type"`
		Hotel struct {
			HotelID   string  `json:"hotelId"`
			Name      string  `json:"name"`
			CityCode  string  `json:"cityCode"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Address   struct {
				Lines      []string `json:"lines"`
				CityName   string   `json:"cityName"`
				CountryCode string  `json:"countryCode"`
			} `json:"address"`
			Rating string `json:"rating"`
		} `json:"hotel"`
		Available bool `json:"available"`
		Offers    []struct {
			ID    string `json:"id"`
			Price struct {
				Currency string `json:"currency"`
				Total    string `json:"total"`
				Base     string `json:"base"`
				Variations struct {
					Average struct {
						Base string `json:"base"`
					} `json:"average"`
				} `json:"variations"`
			} `json:"price"`
			CheckInDate  string `json:"checkInDate"`
			CheckOutDate string `json:"checkOutDate"`
			RateCode     string `json:"rateCode"`
		} `json:"offers"`
	} `json:"data"`
}

// SearchHotels implements Provider.
func (p *AmadeusHotelProvider) SearchHotels(ctx context.Context, q HotelSearchQuery) ([]HotelOffer, error) {
	if !p.Client.Configured() {
		return nil, fmt.Errorf("amadeus: API credentials not configured")
	}

	cityCode, err := p.resolveCityCode(ctx, q.City)
	if err != nil {
		return nil, err
	}

	// Step 1: list hotel IDs in the city.
	hotelListV := url.Values{}
	hotelListV.Set("cityCode", cityCode)
	var hotelList amadeusHotelsByCityResponse
	if err := p.Client.DoJSON(ctx, "/v1/reference-data/locations/hotels/by-city", hotelListV, &hotelList); err != nil {
		return nil, fmt.Errorf("amadeus hotels: list by city: %w", err)
	}
	if len(hotelList.Data) == 0 {
		return nil, nil
	}

	// Cap to first 20 hotel IDs to stay under the URL length limit and rate budget.
	max := 20
	if len(hotelList.Data) < max {
		max = len(hotelList.Data)
	}
	ids := make([]string, 0, max)
	for _, h := range hotelList.Data[:max] {
		ids = append(ids, h.HotelID)
	}

	// Step 2: get offers for those hotel IDs.
	offersV := url.Values{}
	offersV.Set("hotelIds", strings.Join(ids, ","))
	if q.CheckIn != "" {
		offersV.Set("checkInDate", q.CheckIn)
	}
	if q.CheckOut != "" {
		offersV.Set("checkOutDate", q.CheckOut)
	}
	if q.Guests > 0 {
		offersV.Set("adults", strconv.Itoa(q.Guests))
	}
	if q.Currency != "" {
		offersV.Set("currency", q.Currency)
	}
	offersV.Set("bestRateOnly", "true")

	var offers amadeusHotelOffersResponse
	if err := p.Client.DoJSON(ctx, "/v3/shopping/hotel-offers", offersV, &offers); err != nil {
		return nil, fmt.Errorf("amadeus hotels: offers: %w", err)
	}

	out := make([]HotelOffer, 0, len(offers.Data))
	for _, item := range offers.Data {
		if !item.Available || len(item.Offers) == 0 {
			continue
		}
		first := item.Offers[0]
		total, _ := strconv.ParseFloat(first.Price.Total, 64)
		perNight := total
		if v, err := strconv.ParseFloat(first.Price.Variations.Average.Base, 64); err == nil && v > 0 {
			perNight = v
		}
		if q.MaxPricePerNight > 0 && perNight > q.MaxPricePerNight {
			continue
		}

		var rating float32
		if r, err := strconv.ParseFloat(item.Hotel.Rating, 32); err == nil {
			rating = float32(r)
		}

		raw, _ := json.Marshal(item)

		out = append(out, HotelOffer{
			Provider:     p.GetProviderName(),
			OfferID:      first.ID,
			Name:         item.Hotel.Name,
			Address:      strings.Join(item.Hotel.Address.Lines, ", "),
			CheckInDate:  first.CheckInDate,
			CheckOutDate: first.CheckOutDate,
			CostPerNight: perNight,
			TotalCost:    total,
			Currency:     first.Price.Currency,
			Rating:       rating,
			Lat:          item.Hotel.Latitude,
			Lng:          item.Hotel.Longitude,
			Raw:          raw,
		})
	}
	return out, nil
}

// resolveCityCode tries to resolve a free-text city to an Amadeus IATA city code.
// If the input already looks like a 3-letter IATA code, it is returned as-is.
func (p *AmadeusHotelProvider) resolveCityCode(ctx context.Context, city string) (string, error) {
	c := strings.TrimSpace(city)
	if c == "" {
		return "", fmt.Errorf("hotels: city is required")
	}
	if len(c) == 3 && strings.ToUpper(c) == c {
		return c, nil
	}
	v := url.Values{}
	v.Set("keyword", c)
	v.Set("max", "1")
	var resp amadeusCityListResponse
	if err := p.Client.DoJSON(ctx, "/v1/reference-data/locations/cities", v, &resp); err != nil {
		return "", fmt.Errorf("amadeus hotels: resolve city %q: %w", c, err)
	}
	if len(resp.Data) == 0 || resp.Data[0].IataCode == "" {
		return "", fmt.Errorf("amadeus hotels: no city code found for %q", c)
	}
	return resp.Data[0].IataCode, nil
}
