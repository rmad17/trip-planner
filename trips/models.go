package trips

import (
	"time"
	"triplanner/core"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TripPlan struct {
	core.BaseModel
	Name       *string
	StartDate  *time.Time
	EndDate    *time.Time
	MinDays    *int8
	TravelMode *string
	Notes      *string
	Hotels     pq.StringArray `gorm:"type:text[]"`
	Tags       pq.StringArray `gorm:"type:text[]"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null"`
}

type TripHop struct {
	core.BaseModel
	Name        *string
	MapSource   *string
	PlaceID     *string
	StartDate   *time.Time
	EndDate     *time.Time
	Notes       *string
	POIs        pq.StringArray `gorm:"type:text[]"`
	PreviousHop uuid.UUID
	NextHop     uuid.UUID
	TripPlan    uuid.UUID `gorm:"type:uuid;not null"`
}

type Stay struct {
	core.BaseModel
	GoogleLocation *string
	MapboxLocation *string
	StayType       *string
	StayNotes      *string
	StartDate      *time.Time
	EndDate        *time.Time
	IsPrepaid      *bool
	PaymentMode    *string
	TripHop        uuid.UUID `gorm:"type:uuid;not null"`
}

// Add method to get models for Atlas
func GetModels() []interface{} {
	return []interface{}{
		&TripPlan{},
		&TripHop{},
		&Stay{},
	}
}
