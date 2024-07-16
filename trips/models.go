package trips

import (
	"time"
	"triplanner/accounts"
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
	UserID     uuid.UUID
	User       accounts.User
}
