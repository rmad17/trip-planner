package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"triplanner/accounts"
	"triplanner/core"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	email := flag.String("email", "", "User email address (required)")
	password := flag.String("password", "", "Password (required when creating a new user)")
	username := flag.String("username", "", "Username (defaults to email for new users)")
	flag.Parse()

	if *email == "" {
		fmt.Fprintln(os.Stderr, "Usage: createadmin -email=user@example.com [-password=secret] [-username=admin]")
		os.Exit(1)
	}

	core.LoadEnvs()
	core.ConnectDB()

	var user accounts.User
	result := core.DB.Where("email = ?", *email).First(&user)

	if result.Error != nil {
		// User not found — create new admin account
		if *password == "" {
			fmt.Fprintln(os.Stderr, "Error: -password is required when creating a new admin user")
			os.Exit(1)
		}

		uname := *username
		if uname == "" {
			uname = *email
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		emailVal := *email
		user = accounts.User{
			Username:      uname,
			Email:         &emailVal,
			Password:      string(hash),
			EmailVerified: true,
			IsAdmin:       true,
		}

		if err := core.DB.Create(&user).Error; err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}

		fmt.Printf("Admin user created: %s (id: %s)\n", *email, user.ID)
	} else {
		// User exists — promote to admin
		if err := core.DB.Model(&user).Update("is_admin", true).Error; err != nil {
			log.Fatalf("Failed to promote user to admin: %v", err)
		}
		fmt.Printf("User %s (id: %s) promoted to admin\n", *email, user.ID)
	}
}
