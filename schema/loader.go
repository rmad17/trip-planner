package schema

import (
	"triplanner/accounts"
	"triplanner/trips"
)

// GetAllModels returns all GORM models for Atlas
func GetAllModels() []interface{} {
	return []interface{}{
		&accounts.User{},
		&trips.TripPlan{},
		&trips.TripHop{},
		&trips.Stay{},
		// Add other models as you create them
	}
}

// You can also create a function to get models by module
func GetAccountsModels() []interface{} {
	return []interface{}{
		&accounts.User{},
	}
}

func GetTripsModels() []interface{} {
	return []interface{}{
		&trips.TripPlan{},
		&trips.TripHop{},
		&trips.Stay{},
	}
}
