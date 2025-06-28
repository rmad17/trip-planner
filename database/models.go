package database

import (
	"triplanner/accounts"
	"triplanner/core"
	"triplanner/trips"
)

// AutoMigrateAll runs GORM AutoMigrate for all models
func AutoMigrateAll() error {
	models := []interface{}{
		&accounts.User{},
		&trips.TripPlan{},
	}

	return core.DB.AutoMigrate(models...)
}

// GetAllModels returns all models for Atlas
func GetAllModels() []interface{} {
	var models []interface{}
	models = append(models, accounts.GetModels()...)
	models = append(models, trips.GetModels()...)
	return models
}
