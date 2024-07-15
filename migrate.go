package main

import (
	"triplanner/accounts"
	"triplanner/core"
	"triplanner/trips"
)

func init() {
	core.LoadEnvs()
	core.ConnectDB()
}

func main() {
	core.DB.AutoMigrate(&accounts.User{}, &trips.TripPlan{})
}
