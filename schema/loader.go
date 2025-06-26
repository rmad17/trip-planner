package schema

import (
	"triplanner/accounts"
	"triplanner/trips"
)

// GetAllModels returns all GORM models from different modules
func GetAllModels() []interface{} {
	var models []interface{}
	// Add models from each module
	models = append(models, accounts.GetModels()...)
	models = append(models, trips.GetModels()...)

	return models
}
