package accounts

import (
	"triplanner/core"

	"github.com/google/uuid"
)

// MapProvider represents the map service provider preference
type MapProvider string

const (
	MapProviderGoogle MapProvider = "google"
	MapProviderMapbox MapProvider = "mapbox"
)

// UserPreferences holds user-specific preferences
type UserPreferences struct {
	core.BaseModel
	UserID             uuid.UUID   `json:"user_id" gorm:"type:uuid;not null;unique" description:"ID of the user"`
	MapProvider        MapProvider `json:"map_provider" gorm:"type:varchar(20);default:'google'" example:"google" description:"Preferred map service provider (google, mapbox)"`
	DefaultStorageProv string      `json:"default_storage_provider" gorm:"type:varchar(50);default:'digitalocean'" example:"digitalocean" description:"Default storage provider for document uploads"`
	Language           string      `json:"language" gorm:"type:varchar(10);default:'en'" example:"en" description:"Preferred language code"`
	Timezone           string      `json:"timezone" gorm:"default:'UTC'" example:"America/New_York" description:"User's timezone"`
	Currency           string      `json:"currency" gorm:"type:varchar(3);default:'USD'" example:"USD" description:"Preferred currency for expenses"`
}

type User struct {
	core.BaseModel
	Username      string           `json:"username" gorm:"unique"`
	Password      string           `json:"password"`
	Email         *string          `json:"email" gorm:"unique"`
	// Google OAuth fields
	GoogleID      *string          `json:"google_id" gorm:"unique"`
	Name          *string          `json:"name"`
	FirstName     *string          `json:"first_name"`
	LastName      *string          `json:"last_name"`
	AvatarURL     *string          `json:"avatar_url"`
	Locale        *string          `json:"locale"`
	// OAuth metadata
	Provider      *string          `json:"provider"`
	AccessToken   *string          `json:"-" gorm:"type:text"` // Hidden from JSON
	RefreshToken  *string          `json:"-" gorm:"type:text"` // Hidden from JSON
	ExpiresAt     *int64           `json:"-"`
	Preferences   *UserPreferences `json:"preferences" gorm:"foreignKey:UserID"`
}

// Add method to get models for Atlas
func GetModels() []interface{} {
	return []interface{}{
		&User{},
		&UserPreferences{},
	}
}
