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
	place_name  string
	place_id    string
	start_date  *time.Time
	end_date    *time.Time
	min_days    *int8
	travel_mode *string
	notes       *string
	hotels      pq.StringArray `gorm:"type:text[]"`
	tags        pq.StringArray `gorm:"type:text[]"`
	UserID      uuid.UUID
	User        accounts.User
}
