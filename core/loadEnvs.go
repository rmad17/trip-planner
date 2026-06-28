package core

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func LoadEnvs() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatal("Error loading .env file")
	}
	if err != nil {
		log.Println("Warning: no .env file found, using environment variables")
	}
}
