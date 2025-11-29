package trips

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CreateTripRequest represents the request body for creating a new trip
type CreateTripRequest struct {
	Name        *string        `json:"place_name" binding:"required" example:"Trip to Paris" swaggertype:"string" description:"Name of the trip destination"`
	StartDate   *time.Time     `json:"start_date" example:"2024-06-01T00:00:00Z" description:"Start date of the trip"`
	EndDate     *time.Time     `json:"end_date" example:"2024-06-10T00:00:00Z" description:"End date of the trip"`
	MinDays     *int16         `json:"min_days" example:"7" description:"Minimum number of days for the trip"`
	TravelModes pq.StringArray `json:"travel_modes,omitempty" example:"flight,train" swaggertype:"array,string" description:"Modes of travel (flight, car, train, bus, etc.) - optional"`
	Notes       *string        `json:"notes" example:"Romantic getaway" description:"Additional notes about the trip"`
	Hotels      pq.StringArray `json:"hotels,omitempty" example:"Hotel de Paris,Le Bristol" swaggertype:"array,string" description:"List of preferred hotels - optional"`
	Tags        pq.StringArray `json:"tags" example:"romantic,europe,culture" swaggertype:"array,string" description:"Tags associated with the trip"`
	UserID      uuid.UUID      `json:"-" swaggerignore:"true"` // Hidden from Swagger docs as it's set internally
}
