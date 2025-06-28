package accounts

import "triplanner/core"

type User struct {
	core.BaseModel
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
	Email    *string
}

// Add method to get models for Atlas
func GetModels() []interface{} {
	return []interface{}{
		&User{},
	}
}
