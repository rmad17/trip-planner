package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
)

// MockRouteProvider returns deterministic routes computed from great-circle
// distance between the supplied stops. Used when ROUTES_PROVIDER=mock.
type MockRouteProvider struct{}

// NewMockRouteProvider returns a new mock provider.
func NewMockRouteProvider() *MockRouteProvider { return &MockRouteProvider{} }

// GetProviderName implements Provider.
func (p *MockRouteProvider) GetProviderName() string { return "mock" }

// GetRoute implements Provider with a great-circle approximation.
func (p *MockRouteProvider) GetRoute(_ context.Context, q RouteQuery) (*Route, error) {
	if len(q.Stops) < 2 {
		return nil, fmt.Errorf("routes: at least 2 stops required")
	}

	totalKm := 0.0
	for i := 1; i < len(q.Stops); i++ {
		totalKm += haversineKm(q.Stops[i-1], q.Stops[i])
	}

	// 60 km/h average for "drive" / fallback; 5 km/h for walk; 15 km/h for cycle.
	speedKmh := 60.0
	switch q.Mode {
	case "walk":
		speedKmh = 5.0
	case "cycle":
		speedKmh = 15.0
	case "transit":
		speedKmh = 40.0
	}
	durMin := int(totalKm/speedKmh*60 + 0.5)

	coords := make([][]float64, 0, len(q.Stops))
	for _, s := range q.Stops {
		coords = append(coords, []float64{s.Lng, s.Lat})
	}
	geom := map[string]interface{}{
		"type":        "LineString",
		"coordinates": coords,
	}
	geomBytes, _ := json.Marshal(geom)

	return &Route{
		DistanceKm:      totalKm,
		DurationMin:     durMin,
		PolylineGeoJSON: geomBytes,
		Provider:        p.GetProviderName(),
	}, nil
}

func haversineKm(a, b LngLat) float64 {
	const R = 6371.0
	lat1 := a.Lat * math.Pi / 180
	lat2 := b.Lat * math.Pi / 180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180

	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
	return R * c
}
