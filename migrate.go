package main

import (
	"triplanner/core"
	"triplanner/models"
)

func init() {
	core.LoadEnvs()
	core.ConnectDB()
}

func main() {
	core.DB.AutoMigrate(&models.User{})
}
