package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
	"triplanner/core"

	_ "github.com/lib/pq"
)

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func main() {
	// Load environment variables
	core.LoadEnvs()

	// Connect to PostgreSQL using DB_URL
	psqlInfo := getEnv("DB_URL", "postgres://postgres:postgres@localhost:5432/trip?sslmode=disable")

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Create GoAdmin tables
	if err := createGoAdminTables(db); err != nil {
		log.Fatalf("Failed to create GoAdmin tables: %v", err)
	}

	// Create admin user
	if err := createAdminUser(db); err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Println("GoAdmin setup completed successfully!")
	log.Println("Admin credentials:")
	log.Println("Username: admin")
	log.Println("Password: admin")
	log.Println("Admin URL: http://localhost:8080/admin")
}

func createGoAdminTables(db *sql.DB) error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS goadmin_users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			password VARCHAR(100) NOT NULL,
			name VARCHAR(100) NOT NULL,
			avatar VARCHAR(255) DEFAULT '',
			remember_token VARCHAR(100) DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS goadmin_roles (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			slug VARCHAR(50) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS goadmin_permissions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			slug VARCHAR(50) NOT NULL UNIQUE,
			http_method VARCHAR(255) DEFAULT '',
			http_path TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS goadmin_role_users (
			role_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (role_id, user_id)
		);`,

		`CREATE TABLE IF NOT EXISTS goadmin_role_permissions (
			role_id INTEGER NOT NULL,
			permission_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (role_id, permission_id)
		);`,

		`CREATE TABLE IF NOT EXISTS goadmin_user_permissions (
			user_id INTEGER NOT NULL,
			permission_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (user_id, permission_id)
		);`,

		`CREATE TABLE IF NOT EXISTS goadmin_menu (
			id SERIAL PRIMARY KEY,
			parent_id INTEGER DEFAULT 0,
			type INTEGER DEFAULT 0,
			title VARCHAR(50) NOT NULL,
			icon VARCHAR(50) NOT NULL DEFAULT '',
			uri VARCHAR(255) DEFAULT '',
			header VARCHAR(150) DEFAULT '',
			plugin_name VARCHAR(150) DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS goadmin_role_menu (
			role_id INTEGER NOT NULL,
			menu_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (role_id, menu_id)
		);`,

		`CREATE TABLE IF NOT EXISTS goadmin_operation_log (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			path VARCHAR(255) NOT NULL,
			method VARCHAR(10) NOT NULL,
			ip VARCHAR(15) NOT NULL,
			input TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS goadmin_session (
			id VARCHAR(191) PRIMARY KEY,
			values TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,
	}

	for _, query := range tables {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	log.Println("GoAdmin tables created successfully")
	return nil
}

func createAdminUser(db *sql.DB) error {
	// Check if admin user already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM goadmin_users WHERE username = $1", "admin").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check admin user: %v", err)
	}

	if count > 0 {
		log.Println("Admin user already exists, skipping creation")
		return nil
	}

	now := time.Now()

	// Create password hash (MD5 for GoAdmin compatibility)
	hasher := md5.New()
	hasher.Write([]byte("admin"))
	passwordHash := fmt.Sprintf("%x", hasher.Sum(nil))

	// Insert admin user
	_, err = db.Exec(`
		INSERT INTO goadmin_users (username, password, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, "admin", passwordHash, "Administrator", now, now)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	// Create administrator role
	_, err = db.Exec(`
		INSERT INTO goadmin_roles (name, slug, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (slug) DO NOTHING
	`, "Administrator", "administrator", now, now)
	if err != nil {
		return fmt.Errorf("failed to create admin role: %v", err)
	}

	// Get role ID
	var roleID int
	err = db.QueryRow("SELECT id FROM goadmin_roles WHERE slug = $1", "administrator").Scan(&roleID)
	if err != nil {
		return fmt.Errorf("failed to get role ID: %v", err)
	}

	// Get user ID
	var userID int
	err = db.QueryRow("SELECT id FROM goadmin_users WHERE username = $1", "admin").Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to get user ID: %v", err)
	}

	// Assign role to user
	_, err = db.Exec(`
		INSERT INTO goadmin_role_users (role_id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
	`, roleID, userID, now, now)
	if err != nil {
		return fmt.Errorf("failed to assign role to admin user: %v", err)
	}

	// Create default menu items
	menuItems := []struct {
		ID       int
		ParentID int
		Type     int
		Title    string
		Icon     string
		URI      string
	}{
		{1, 0, 1, "Admin", "fa-tasks", ""},
		{2, 1, 1, "Users", "fa-users", "/info/goadmin_users"},
		{3, 1, 1, "Roles", "fa-user", "/info/goadmin_roles"},
		{4, 1, 1, "Permission", "fa-ban", "/info/goadmin_permissions"},
		{5, 1, 1, "Menu", "fa-bars", "/menu"},
		{6, 1, 1, "Operation log", "fa-history", "/info/goadmin_operation_log"},
		{7, 0, 1, "Trip Management", "fa-plane", ""},
		{8, 7, 1, "Trip Plans", "fa-map", "/info/trip_plans"},
		{9, 7, 1, "Users", "fa-users", "/info/users"},
		{10, 7, 1, "Travellers", "fa-users", "/info/travellers"},
		{11, 7, 1, "Expenses", "fa-money", "/info/expenses"},
		{12, 7, 1, "Documents", "fa-file", "/info/documents"},
	}

	for _, item := range menuItems {
		_, err := db.Exec(`
			INSERT INTO goadmin_menu (id, parent_id, type, title, icon, uri, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title,
				icon = EXCLUDED.icon,
				uri = EXCLUDED.uri,
				updated_at = EXCLUDED.updated_at
		`, item.ID, item.ParentID, item.Type, item.Title, item.Icon, item.URI, now, now)
		if err != nil {
			return fmt.Errorf("failed to create menu item: %v", err)
		}
	}

	// Assign all menu items to admin role
	_, err = db.Exec(`
		INSERT INTO goadmin_role_menu (role_id, menu_id, created_at, updated_at)
		SELECT $1, id, $2, $3 FROM goadmin_menu
		ON CONFLICT DO NOTHING
	`, roleID, now, now)
	if err != nil {
		return fmt.Errorf("failed to assign menu to admin role: %v", err)
	}

	log.Println("Admin user created successfully")
	return nil
}