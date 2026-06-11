package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// GoogleRoutesProvider uses Google Routes API v2.
type GoogleRoutesProvider struct {
	APIKey     string
	HTTPClient *http.Client
}

// NewGoogleRoutesProvider returns a provider configured from env.
func NewGoogleRoutesProvider() *GoogleRoutesProvider {
	key := os.Getenv("GOOGLE_MAPS_API_KEY")
	if key == "" {
		// Re-use existing GOOGLE_API_KEY if scoped to Routes API.
		key = os.Getenv("GOOGLE_API_KEY")
	}
	return &GoogleRoutesProvider{
		APIKey: key,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProviderName implements Provider.
func (p *GoogleRoutesProvider) GetProviderName() string { return "google" }

type googleLatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type googleLocation struct {
	LatLng googleLatLng `json:"latLng"`
}

type googleWaypoint struct {
	Location googleLocation `json:"location"`
}

type googleRoutesRequest struct {
	Origin                   googleWaypoint   `json:"origin"`
	Destination              googleWaypoint   `json:"destination"`
	Intermediates            []googleWaypoint `json:"intermediates,omitempty"`
	TravelMode               string           `json:"travelMode"`
	RoutingPreference        string           `json:"routingPreference,omitempty"`
	PolylineEncoding         string           `json:"polylineEncoding"`
	ComputeAlternativeRoutes bool             `json:"computeAlternativeRoutes"`
}

type googleRoutesResponse struct {
	Routes []struct {
		DistanceMeters int    `json:"distanceMeters"`
		Duration       string `json:"duration"`
		Polyline       struct {
			GeoJsonLinestring json.RawMessage `json:"geoJsonLinestring"`
		} `json:"polyline"`
	} `json:"routes"`
}

// GetRoute implements Provider.
func (p *GoogleRoutesProvider) GetRoute(ctx context.Context, q RouteQuery) (*Route, error) {
	if p.APIKey == "" {
		return nil, fmt.Errorf("GOOGLE_MAPS_API_KEY (or GOOGLE_API_KEY) not configured")
	}
	if len(q.Stops) < 2 {
		return nil, fmt.Errorf("routes: at least 2 stops required")
	}

	travelMode := googleTravelMode(q.Mode)

	body := googleRoutesRequest{
		Origin:           waypoint(q.Stops[0]),
		Destination:      waypoint(q.Stops[len(q.Stops)-1]),
		TravelMode:       travelMode,
		PolylineEncoding: "GEO_JSON_LINESTRING",
	}
	if travelMode == "DRIVE" {
		body.RoutingPreference = "TRAFFIC_AWARE"
	}
	if len(q.Stops) > 2 {
		intermediates := make([]googleWaypoint, 0, len(q.Stops)-2)
		for _, s := range q.Stops[1 : len(q.Stops)-1] {
			intermediates = append(intermediates, waypoint(s))
		}
		body.Intermediates = intermediates
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("routes: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://routes.googleapis.com/directions/v2:computeRoutes", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("routes: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Goog-Api-Key", p.APIKey)
	req.Header.Set("X-Goog-FieldMask", "routes.distanceMeters,routes.duration,routes.polyline.geoJsonLinestring")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("routes: do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("routes: read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("routes: google api error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var out googleRoutesResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return nil, fmt.Errorf("routes: parse response: %w", err)
	}
	if len(out.Routes) == 0 {
		return nil, fmt.Errorf("routes: no route returned")
	}
	r := out.Routes[0]

	durMin, err := parseGoogleDurationMin(r.Duration)
	if err != nil {
		return nil, fmt.Errorf("routes: parse duration %q: %w", r.Duration, err)
	}

	polyline := r.Polyline.GeoJsonLinestring
	if len(polyline) == 0 {
		polyline = json.RawMessage(`null`)
	}

	return &Route{
		DistanceKm:      float64(r.DistanceMeters) / 1000.0,
		DurationMin:     durMin,
		PolylineGeoJSON: polyline,
		Provider:        p.GetProviderName(),
	}, nil
}

func waypoint(s LngLat) googleWaypoint {
	return googleWaypoint{
		Location: googleLocation{
			LatLng: googleLatLng{Latitude: s.Lat, Longitude: s.Lng},
		},
	}
}

func googleTravelMode(mode string) string {
	switch mode {
	case "walk":
		return "WALK"
	case "transit":
		return "TRANSIT"
	case "cycle":
		return "BICYCLE"
	case "drive":
		return "DRIVE"
	default:
		return "DRIVE"
	}
}

// parseGoogleDurationMin parses a Google "duration" value like "375s" into minutes.
func parseGoogleDurationMin(d string) (int, error) {
	if d == "" {
		return 0, nil
	}
	if d[len(d)-1] == 's' {
		d = d[:len(d)-1]
	}
	dur, err := time.ParseDuration(d + "s")
	if err != nil {
		return 0, err
	}
	return int(dur.Minutes() + 0.5), nil
}
