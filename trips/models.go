package main

import (
	"time"
	"triplanner/core"
	"triplanner/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Trip struct {
	core.Base
	UserID      uuid.UUID
	User        models.User
	place_name  string
	place_id    string
	start_date  *time.Time
	end_date    *time.Time
	min_days    *int8
	travel_mode *string
	notes       *string
	hotels      pq.StringArray `gorm:"type:text[]"`
	tags        pq.StringArray `gorm:"type:text[]"`
}
