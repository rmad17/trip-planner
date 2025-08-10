package main

import (
	"log"
	"triplanner/admin"
	"triplanner/core"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"

	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/postgres"
)

func init() {
	core.LoadEnvs()
}

func main() {
	// Initialize database connection for GoAdmin
	cfg := db.Config{
		Host:       core.GetEnv("DB_HOST", "localhost"),
		Port:       core.GetEnv("DB_PORT", "5432"),
		User:       core.GetEnv("DB_USER", "postgres"),
		Pwd:        core.GetEnv("DB_PASSWORD", ""),
		Name:       core.GetEnv("DB_NAME", "triplanner"),
		MaxIdleCon: 50,
		MaxOpenCon: 150,
		Driver:     "postgres",
	}

	conn := db.GetConnectionByDriver("postgres")
	conn.InitDB(map[string]db.Config{
		"default": cfg,
	})
	conn.SetParams(parameter.Base{})

	// Create admin user seed
	if err := admin.CreateAdminUserSeed(conn); err != nil {
		log.Fatalf("Failed to create admin user seed: %v", err)
	}

	log.Println("Seed completed successfully!")
}