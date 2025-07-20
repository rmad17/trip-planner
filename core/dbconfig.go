package core

import (
	"log/slog"
	"os"

	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	DSN         string
	Environment string
}

var DB *gorm.DB

// GetDatabaseConfig returns database configuration based on environment
func GetDatabaseConfig() DatabaseConfig {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	var dsn string
	switch env {
	case "test":
		dsn = os.Getenv("TEST_DB_URL")
		if dsn == "" {
			slog.Error("TEST_DB_URL environment variable not set")
		}
	default:
		dsn = os.Getenv("DB_URL")
		if dsn == "" {
			slog.Error("DB_URL environment variable not set")
		}
	}

	return DatabaseConfig{
		DSN:         dsn,
		Environment: env,
	}
}

// ConnectDB connects to the database using the appropriate configuration
func ConnectDB() {
	config := GetDatabaseConfig()
	
	var err error
	DB, err = gorm.Open(postgres.Open(config.DSN), &gorm.Config{})

	if err != nil {
		msg := "Failed to connect to DB: " + err.Error()
		slog.Error(msg)
		return
	}

	// Only set up logging for non-test environments
	if config.Environment != "test" {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		log := zap.New(zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(colorable.NewColorableStdout()),
			zapcore.DebugLevel,
		))

		log.Info("Connected to database ...", zap.String("environment", config.Environment))
	}
}

// ConnectTestDB connects to the test database specifically
func ConnectTestDB() {
	// Temporarily set environment to test
	originalEnv := os.Getenv("APP_ENV")
	os.Setenv("APP_ENV", "test")
	
	// Connect to test database
	ConnectDB()
	
	// Restore original environment
	if originalEnv == "" {
		os.Unsetenv("APP_ENV")
	} else {
		os.Setenv("APP_ENV", originalEnv)
	}
}

// GetDB returns the current database instance
func GetDB() *gorm.DB {
	return DB
}