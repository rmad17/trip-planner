package accounts

import "triplanner/core"

type User struct {
	core.Base
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
	Email    *string
}
