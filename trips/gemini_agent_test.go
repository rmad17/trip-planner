package trips

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestEnrichTripWithMockProviders verifies acceptance criterion #5: when all
// three providers are set to "mock", the enrichment step produces non-empty
// suggested_hotels, suggested_flights (intercity hops), and route_to_next
// without any external network call.
func TestEnrichTripWithMockProviders(t *testing.T) {
	t.Setenv("HOTELS_PROVIDER", "mock")
	t.Setenv("FLIGHTS_PROVIDER", "mock")
	t.Setenv("ROUTES_PROVIDER", "mock")

	// Sanity-check env didn't leak from a prior test.
	if got := os.Getenv("HOTELS_PROVIDER"); got != "mock" {
		t.Fatalf("HOTELS_PROVIDER = %q, want mock", got)
	}

	plan := &TripGenerationResponse{
		TripName:  "Mock test",
		TotalDays: 5,
		Hops: []GeneratedHop{
			{
				Name:           "Delhi",
				City:           "Delhi",
				Country:        "India",
				StartDate:      "2026-06-01",
				EndDate:        "2026-06-03",
				Transportation: "flight",
				HopOrder:       1,
			},
			{
				Name:           "Paris",
				City:           "Paris",
				Country:        "France",
				StartDate:      "2026-06-03",
				EndDate:        "2026-06-08",
				Transportation: "flight",
				HopOrder:       2,
			},
		},
	}

	req := TripGenerationRequest{
		Source:       "Mumbai",
		Destinations: []string{"Delhi", "Paris"},
		StartDate:    time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC),
		NumTravelers: 2,
		Currency:     CurrencyEUR,
	}

	enrichTripWithProviders(context.Background(), plan, req)

	// Every hop should have suggested hotels.
	for i, hop := range plan.Hops {
		if len(hop.SuggestedHotels) == 0 {
			t.Errorf("hop %d (%s): expected non-empty SuggestedHotels", i, hop.City)
		}
	}

	// Both hops use flight transport, so both should have suggested flights.
	for i, hop := range plan.Hops {
		if len(hop.SuggestedFlights) == 0 {
			t.Errorf("hop %d (%s): expected non-empty SuggestedFlights for flight transport", i, hop.City)
		}
	}

	// The first hop should have a route-to-next pointing to the second.
	if plan.Hops[0].RouteToNext == nil {
		t.Errorf("hop 0: expected non-nil RouteToNext to Paris")
	} else {
		r := plan.Hops[0].RouteToNext
		if r.DistanceKm <= 0 {
			t.Errorf("hop 0 RouteToNext: expected DistanceKm > 0, got %v", r.DistanceKm)
		}
		if len(r.PolylineGeoJSON) == 0 {
			t.Errorf("hop 0 RouteToNext: expected non-empty polyline")
		}
		if r.Provider == "" {
			t.Errorf("hop 0 RouteToNext: expected provider name to be set")
		}
	}

	// The last hop should have RouteToNext nil (no next hop).
	if plan.Hops[len(plan.Hops)-1].RouteToNext != nil {
		t.Errorf("last hop: expected RouteToNext nil")
	}
}
