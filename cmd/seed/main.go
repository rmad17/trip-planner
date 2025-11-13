package main

import (
	"log"
	"os"
	"triplanner/admin"
	"triplanner/core"

	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/GoAdminGroup/go-admin/modules/db"

	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/postgres"
)

func init() {
	core.LoadEnvs()
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Initialize database connection for GoAdmin
	cfg := config.Database{
		Host:         getEnv("DB_HOST", "localhost"),
		Port:         getEnv("DB_PORT", "5432"),
		User:         getEnv("DB_USER", "postgres"),
		Pwd:          getEnv("DB_PASSWORD", ""),
		Name:         getEnv("DB_NAME", "triplanner"),
		MaxIdleConns: 50,
		MaxOpenConns: 150,
		Driver:       "postgres",
	}

	conn := db.GetConnectionByDriver("postgres")
	conn.InitDB(map[string]config.Database{
		"default": cfg,
	})

	// Create admin user seed
	if err := admin.CreateAdminUserSeed(conn); err != nil {
		log.Fatalf("Failed to create admin user seed: %v", err)
	}

	log.Println("Seed completed successfully!")
}