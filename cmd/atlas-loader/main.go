package main

import (
	"fmt"
	"log"
	"os"

	"triplanner/accounts"
	"triplanner/trips"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	// Collect all your models
	models := []interface{}{
		&accounts.User{},
		&trips.TripPlan{},
		&trips.TripHop{},
		&trips.Stay{},
	}

	// Generate schema for PostgreSQL
	stmts, err := gormschema.New("postgres").Load(models...)
	if err != nil {
		log.Fatalf("failed to load gorm schema: %v", err)
	}

	// Output schema statements as strings (not bytes)
	fmt.Fprint(os.Stdout, stmts) // Use Print instead of Println to avoid extra newlines
}
