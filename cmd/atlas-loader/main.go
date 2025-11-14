package main

import (
	"fmt"
	"log"
	"os"

	"triplanner/database"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	// Get all models from database package
	models := database.GetAllModels()

	// Generate schema for PostgreSQL
	stmts, err := gormschema.New("postgres").Load(models...)
	if err != nil {
		log.Fatalf("failed to load gorm schema: %v", err)
	}

	// Output schema statements as strings (not bytes)
	if _, err := fmt.Fprint(os.Stdout, stmts); err != nil {
		log.Fatalf("failed to write schema: %v", err)
	}
}
