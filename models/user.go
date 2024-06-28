package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	Username  string    `json:"username" gorm:"unique"`
	Password  string    `json:"password"`
	Email     *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
