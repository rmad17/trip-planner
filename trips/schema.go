package trips

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CreateTripRequest struct {
	PlaceName  string         `json:"place_name" binding:"required"`
	PlaceID    string         `json:"place_id"`
	StartDate  *time.Time     `json:"start_date"`
	EndDate    *time.Time     `json:"end_date"`
	MinDays    *int16         `json:"min_days"`
	TravelMode *string        `json:"travel_mode"`
	Notes      *string        `json:"notes"`
	Hotels     pq.StringArray `gorm:"type:text[]"`
	Tags       pq.StringArray `gorm:"type:text[]"`
	UserID     uuid.UUID
}
