package main

import (
	"fmt"
	"log"

	"triplanner/database"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	// Get all models from centralized registry
	models := database.GetAllModels()

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
