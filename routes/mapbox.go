package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"triplanner/core"
)

// MapboxRouteProvider uses the Mapbox Directions API as a fallback impl.
type MapboxRouteProvider struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewMapboxRouteProvider returns a provider configured from env (uses MAPBOX_TOKEN).
func NewMapboxRouteProvider() *MapboxRouteProvider {
	return &MapboxRouteProvider{
		APIKey: os.Getenv(core.SEARCH_API_KEY),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderName implements Provider.
func (p *MapboxRouteProvider) GetProviderName() string { return "mapbox" }

type mapboxDirectionsResponse struct {
	Routes []struct {
		Distance float64         `json:"distance"`
		Duration float64         `json:"duration"`
		Geometry json.RawMessage `json:"geometry"`
	} `json:"routes"`
	Code string `json:"code"`
}

// GetRoute implements Provider.
func (p *MapboxRouteProvider) GetRoute(ctx context.Context, q RouteQuery) (*Route, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("MAPBOX_TOKEN not configured")
	}
	if len(q.Stops) < 2 {
		return nil, fmt.Errorf("routes: at least 2 stops required")
	}

	profile := mapboxProfile(q.Mode)

	parts := make([]string, 0, len(q.Stops))
	for _, s := range q.Stops {
		parts = append(parts, fmt.Sprintf("%f,%f", s.Lng, s.Lat))
	}
	coords := strings.Join(parts, ";")

	v := url.Values{}
	v.Set("access_token", p.APIKey)
	v.Set("geometries", "geojson")
	v.Set("overview", "full")

	endpoint := fmt.Sprintf("https://api.mapbox.com/directions/v5/%s/%s?%s", profile, coords, v.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("routes: build request: %w", err)
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("routes: do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("routes: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("routes: mapbox api error (status %d): %s", resp.StatusCode, string(body))
	}

	var out mapboxDirectionsResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("routes: parse response: %w", err)
	}
	if len(out.Routes) == 0 {
		return nil, fmt.Errorf("routes: mapbox returned no routes")
	}

	r := out.Routes[0]
	geom := r.Geometry
	if len(geom) == 0 {
		geom = json.RawMessage(`null`)
	}

	return &Route{
		DistanceKm:      r.Distance / 1000.0,
		DurationMin:     int(r.Duration/60 + 0.5),
		PolylineGeoJSON: geom,
		Provider:        p.GetProviderName(),
	}, nil
}

func mapboxProfile(mode string) string {
	switch mode {
	case "walk":
		return "mapbox/walking"
	case "cycle":
		return "mapbox/cycling"
	case "transit", "drive":
		return "mapbox/driving"
	default:
		return "mapbox/driving"
	}
}
