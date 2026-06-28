package trips

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CreateTripRequest represents the request body for creating a new trip
type CreateTripRequest struct {
	Name        *string        `json:"place_name" binding:"required"`
	StartDate   *time.Time     `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	MinDays     *int16         `json:"min_days"`
	TravelMode  string         `json:"travel_mode,omitempty"`  // Frontend sends comma-separated string
	TravelModes pq.StringArray `json:"travel_modes,omitempty"` // Or native array
	Notes       *string        `json:"notes"`
	Hotels      pq.StringArray `json:"hotels,omitempty"`
	Tags        pq.StringArray `json:"tags"`
	UserID      uuid.UUID      `json:"-" swaggerignore:"true"`
}

// ParsedTravelModes returns travel modes from whichever field the client sent.
func (r *CreateTripRequest) ParsedTravelModes() pq.StringArray {
	return parseTravelModes(r.TravelMode, r.TravelModes)
}

// UpdateTripRequest is the update payload for trip plans.
// Accepts travel_mode as a comma-separated string to match the frontend.
type UpdateTripRequest struct {
	Name        *string        `json:"name"`
	Description *string        `json:"description"`
	StartDate   *time.Time     `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	TravelMode  string         `json:"travel_mode"`  // comma-separated from frontend
	TravelModes pq.StringArray `json:"travel_modes"` // or native array
	Notes       *string        `json:"notes"`
	Budget      *float64       `json:"budget"`
	Currency    string         `json:"currency"`
	Status      string         `json:"status"`
	TripType    *string        `json:"trip_type"`
	Tags        pq.StringArray `json:"tags"`
	IsPublic    *bool          `json:"is_public"`
	Timezone    *string        `json:"timezone"`
}

// ParsedTravelModes returns travel modes from whichever field the client sent.
func (r *UpdateTripRequest) ParsedTravelModes() pq.StringArray {
	return parseTravelModes(r.TravelMode, r.TravelModes)
}

// parseTravelModes returns travel modes from a comma-separated string or a native array.
// The array takes precedence when both are present.
func parseTravelModes(single string, multi pq.StringArray) pq.StringArray {
	if len(multi) > 0 {
		return multi
	}
	if single == "" {
		return nil
	}
	parts := strings.Split(single, ",")
	modes := make(pq.StringArray, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			modes = append(modes, p)
		}
	}
	return modes
}
