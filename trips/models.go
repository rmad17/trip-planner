package trips

import (
	"time"
	"triplanner/core"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TripPlan struct {
	core.BaseModel
	PlaceName  string
	PlaceID    string
	StartDate  *time.Time
	EndDate    *time.Time
	MinDays    *int8
	TravelMode *string
	Notes      *string
	Hotels     pq.StringArray `gorm:"type:text[]"`
	Tags       pq.StringArray `gorm:"type:text[]"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null"`
}

// Add method to get models for Atlas
func GetModels() []interface{} {
	return []interface{}{
		&TripPlan{},
	}
}
