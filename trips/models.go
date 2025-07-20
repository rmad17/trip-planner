package trips

import (
	"time"
	"triplanner/core"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// TripPlan represents a trip plan in the system
type TripPlan struct {
	core.BaseModel
	Name       *string        `json:"name" example:"Trip to Paris" description:"Name of the trip"`
	StartDate  *time.Time     `json:"start_date" example:"2024-06-01T00:00:00Z" description:"Start date of the trip"`
	EndDate    *time.Time     `json:"end_date" example:"2024-06-10T00:00:00Z" description:"End date of the trip"`
	MinDays    *int8          `json:"min_days" example:"7" description:"Minimum number of days for the trip"`
	TravelMode *string        `json:"travel_mode" example:"flight" description:"Mode of travel"`
	Notes      *string        `json:"notes" example:"Romantic getaway" description:"Additional notes"`
	Hotels     pq.StringArray `json:"hotels" gorm:"type:text[]" swaggertype:"array,string" example:"Hotel de Paris,Le Bristol" description:"List of hotels"`
	Tags       pq.StringArray `json:"tags" gorm:"type:text[]" swaggertype:"array,string" example:"romantic,europe" description:"Trip tags"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the user who created the trip"`
}

// TripHop represents a hop/leg in a trip itinerary
type TripHop struct {
	core.BaseModel
	Name        *string        `json:"name" example:"Paris Visit" description:"Name of the hop"`
	MapSource   *string        `json:"map_source" example:"google" description:"Map service used (google, mapbox, etc.)"`
	PlaceID     *string        `json:"place_id" example:"ChIJD7fiBh9u5kcRYJSMaMOCCwQ" description:"Place ID from map service"`
	StartDate   *time.Time     `json:"start_date" example:"2024-06-01T00:00:00Z" description:"Start date of the hop"`
	EndDate     *time.Time     `json:"end_date" example:"2024-06-03T00:00:00Z" description:"End date of the hop"`
	Notes       *string        `json:"notes" example:"Visit Eiffel Tower" description:"Notes for this hop"`
	POIs        pq.StringArray `json:"pois" gorm:"type:text[]" swaggertype:"array,string" example:"Eiffel Tower,Louvre Museum" description:"Points of interest"`
	PreviousHop uuid.UUID      `json:"previous_hop" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of previous hop in sequence"`
	NextHop     uuid.UUID      `json:"next_hop" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of next hop in sequence"`
	TripPlan    uuid.UUID      `json:"trip_plan" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the parent trip plan"`
}

// Stay represents accommodation details for a trip hop
type Stay struct {
	core.BaseModel
	GoogleLocation *string    `json:"google_location" example:"ChIJD7fiBh9u5kcRYJSMaMOCCwQ" description:"Google Maps location identifier"`
	MapboxLocation *string    `json:"mapbox_location" example:"paris.hotel.123" description:"Mapbox location identifier"`
	StayType       *string    `json:"stay_type" example:"hotel" description:"Type of accommodation (hotel, airbnb, hostel, etc.)"`
	StayNotes      *string    `json:"stay_notes" example:"Near Eiffel Tower" description:"Notes about the accommodation"`
	StartDate      *time.Time `json:"start_date" example:"2024-06-01T00:00:00Z" description:"Check-in date"`
	EndDate        *time.Time `json:"end_date" example:"2024-06-03T00:00:00Z" description:"Check-out date"`
	IsPrepaid      *bool      `json:"is_prepaid" example:"true" description:"Whether the stay is prepaid"`
	PaymentMode    *string    `json:"payment_mode" example:"credit_card" description:"Payment method used"`
	TripHop        uuid.UUID  `json:"trip_hop" gorm:"type:uuid;not null" example:"123e4567-e89b-12d3-a456-426614174000" description:"ID of the associated trip hop"`
}

// Add method to get models for Atlas
func GetModels() []interface{} {
	return []interface{}{
		&TripPlan{},
		&TripHop{},
		&Stay{},
	}
}
