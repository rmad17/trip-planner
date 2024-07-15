package main

import (
	"triplanner/accounts"
	"triplanner/core"
)

func init() {
	core.LoadEnvs()
	core.ConnectDB()
}

func main() {
	core.DB.AutoMigrate(&accounts.User{})
}
