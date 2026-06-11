package routes

import (
	"context"
	"encoding/json"
)

// Provider is the swap surface for routing implementations.
type Provider interface {
	GetRoute(ctx context.Context, q RouteQuery) (*Route, error)
	GetProviderName() string
}

// LngLat is a (longitude, latitude) coordinate pair.
type LngLat struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

// RouteQuery describes a multi-stop route lookup.
type RouteQuery struct {
	Stops []LngLat
	Mode  string // drive | walk | transit | cycle
}

// Route is the canonical route response shape.
type Route struct {
	DistanceKm      float64         `json:"distance_km"`
	DurationMin     int             `json:"duration_min"`
	PolylineGeoJSON json.RawMessage `json:"polyline_geojson" swaggertype:"object"`
	Steps           []RouteStep     `json:"steps,omitempty"`
	Provider        string          `json:"provider"`
}

// RouteStep is an optional turn-by-turn step.
type RouteStep struct {
	Instruction string  `json:"instruction"`
	DistanceKm  float64 `json:"distance_km"`
	DurationMin int     `json:"duration_min"`
}
