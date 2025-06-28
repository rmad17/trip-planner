package main

import (
	"log"
	"triplanner/core"
	"triplanner/database"
)

func init() {
	core.LoadEnvs()
	core.ConnectDB()
}

func main() {
	// Use the centralized migration function
	err := database.AutoMigrateAll()
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Migration completed successfully!")
}
