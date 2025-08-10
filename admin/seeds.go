package admin

import (
	"crypto/md5"
	"fmt"
	"time"

	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/modules/db/dialect"
)

// CreateAdminUserSeed creates the default admin user for GoAdmin
func CreateAdminUserSeed(conn db.Connection) error {
	// Create GoAdmin tables first if they don't exist
	if err := createGoAdminTables(conn); err != nil {
		return fmt.Errorf("failed to create goadmin tables: %v", err)
	}

	// Check if admin user already exists
	adminExists, err := conn.Query("SELECT COUNT(*) as count FROM goadmin_users WHERE username = ?", "admin")
	if err != nil {
		return fmt.Errorf("failed to check admin user: %v", err)
	}

	if len(adminExists) > 0 && adminExists[0]["count"].(string) != "0" {
		fmt.Println("Admin user already exists, skipping creation")
		return nil
	}

	// Create default admin user
	now := time.Now()
	hasher := md5.New()
	hasher.Write([]byte("admin"))
	passwordHash := fmt.Sprintf("%x", hasher.Sum(nil))

	// Insert admin user
	_, err = conn.Exec(`
		INSERT INTO goadmin_users (id, username, password, name, avatar, created_at, updated_at)
		VALUES (1, 'admin', ?, 'Administrator', '', ?, ?)
	`, passwordHash, now, now)

	if err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	// Create admin role if it doesn't exist
	roleExists, err := conn.Query("SELECT COUNT(*) as count FROM goadmin_roles WHERE slug = ?", "administrator")
	if err != nil {
		return fmt.Errorf("failed to check admin role: %v", err)
	}

	var roleID int64 = 1
	if len(roleExists) == 0 || roleExists[0]["count"].(string) == "0" {
		// Create administrator role
		_, err = conn.Exec(`
			INSERT INTO goadmin_roles (id, name, slug, created_at, updated_at)
			VALUES (1, 'Administrator', 'administrator', ?, ?)
		`, now, now)
		if err != nil {
			return fmt.Errorf("failed to create admin role: %v", err)
		}
	}

	// Assign role to user
	_, err = conn.Exec(`
		INSERT INTO goadmin_role_users (role_id, user_id, created_at, updated_at)
		VALUES (?, 1, ?, ?)
		ON CONFLICT DO NOTHING
	`, roleID, now, now)

	if err != nil {
		return fmt.Errorf("failed to assign role to admin user: %v", err)
	}

	fmt.Println("Admin user created successfully:")
	fmt.Println("Username: admin")
	fmt.Println("Password: admin")
	fmt.Println("Please change the password after first login!")

	return nil
}

// createGoAdminTables creates the necessary GoAdmin tables
func createGoAdminTables(conn db.Connection) error {
	// Get the current dialect
	d := dialect.GetDialectByDriver("postgres")

	tables := []string{
		// Users table
		d.GetName() + ` CREATE TABLE IF NOT EXISTS goadmin_users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			password VARCHAR(100) NOT NULL,
			name VARCHAR(100) NOT NULL,
			avatar VARCHAR(255) DEFAULT '',
			remember_token VARCHAR(100) DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		// Roles table
		d.GetName() + ` CREATE TABLE IF NOT EXISTS goadmin_roles (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			slug VARCHAR(50) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		// Permissions table
		d.GetName() + ` CREATE TABLE IF NOT EXISTS goadmin_permissions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			slug VARCHAR(50) NOT NULL UNIQUE,
			http_method VARCHAR(255) DEFAULT '',
			http_path TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		// Role users junction table
		d.GetName() + ` CREATE TABLE IF NOT EXISTS goadmin_role_users (
			role_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (role_id, user_id)
		);`,

		// Role permissions junction table
		d.GetName() + ` CREATE TABLE IF NOT EXISTS goadmin_role_permissions (
			role_id INTEGER NOT NULL,
			permission_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (role_id, permission_id)
		);`,

		// User permissions junction table
		d.GetName() + ` CREATE TABLE IF NOT EXISTS goadmin_user_permissions (
			user_id INTEGER NOT NULL,
			permission_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (user_id, permission_id)
		);`,

		// Menu table
		d.GetName() + ` CREATE TABLE IF NOT EXISTS goadmin_menu (
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

		// Role menu junction table
		d.GetName() + ` CREATE TABLE IF NOT EXISTS goadmin_role_menu (
			role_id INTEGER NOT NULL,
			menu_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			PRIMARY KEY (role_id, menu_id)
		);`,

		// Operation log table
		d.GetName() + ` CREATE TABLE IF NOT EXISTS goadmin_operation_log (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			path VARCHAR(255) NOT NULL,
			method VARCHAR(10) NOT NULL,
			ip VARCHAR(15) NOT NULL,
			input TEXT DEFAULT '',
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,

		// Session table
		d.GetName() + ` CREATE TABLE IF NOT EXISTS goadmin_session (
			id VARCHAR(191) PRIMARY KEY,
			values TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);`,
	}

	for _, query := range tables {
		// Remove the dialect prefix for execution
		actualQuery := query[len(d.GetName()):]
		if _, err := conn.Exec(actualQuery); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
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
		{1, 0, 1, "Admin", "fa-tasks", "", ""},
		{2, 1, 1, "Users", "fa-users", "/info/users", ""},
		{3, 1, 1, "Roles", "fa-user", "/info/roles", ""},
		{4, 1, 1, "Permission", "fa-ban", "/info/permission", ""},
		{5, 1, 1, "Menu", "fa-bars", "/menu", ""},
		{6, 1, 1, "Operation log", "fa-history", "/info/op", ""},
		{7, 0, 1, "Trip Management", "fa-plane", "", ""},
		{8, 7, 1, "Trip Plans", "fa-map", "/info/trip_plans", ""},
		{9, 7, 1, "Travellers", "fa-users", "/info/travellers", ""},
		{10, 7, 1, "Expenses", "fa-money", "/info/expenses", ""},
		{11, 7, 1, "Documents", "fa-file", "/info/documents", ""},
	}

	for _, item := range menuItems {
		_, err := conn.Exec(`
			INSERT INTO goadmin_menu (id, parent_id, type, title, icon, uri, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
			ON CONFLICT (id) DO NOTHING
		`, item.ID, item.ParentID, item.Type, item.Title, item.Icon, item.URI)
		if err != nil {
			return fmt.Errorf("failed to create menu item: %v", err)
		}
	}

	return nil
}