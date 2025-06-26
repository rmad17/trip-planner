package main

import (
	"fmt"
	"log"

	"triplanner/accounts"
	"triplanner/trips"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	// Collect all your models
	models := []interface{}{
		&accounts.User{},
		&trips.TripPlan{},
	}

	// Generate schema for PostgreSQL
	stmts, err := gormschema.New("postgres").Load(models...)
	if err != nil {
		log.Fatalf("failed to load gorm schema: %v", err)
	}

	// Output schema statements
	for _, stmt := range stmts {
		fmt.Println(stmt)
	}
}
