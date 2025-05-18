package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	ServerPort int
	Database   DatabaseConfig
	JWT        JWTConfig
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey       string
	ExpirationHours int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host                   string
	Port                   int
	User                   string
	Password               string
	Name                   string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxLifetimeMinutes int
}

// Load loads configuration from environment variables with defaults
func Load() (*Config, error) {
	// Try to load from .env file first
	err := godotenv.Load()
	if err != nil {
		// If .env doesn't exist, try sample-env
		err = godotenv.Load("sample-env")
		if err != nil {
			log.Println("Warning: No .env or sample-env file found. Using environment variables or defaults.")
		}
	}

	// Server configuration
	serverPort := getEnvInt("SERVER_PORT", 8080)

	// JWT configuration
	jwtSecretKey := getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production")
	jwtExpirationHours := getEnvInt("JWT_EXPIRATION_HOURS", 24)

	// Database configuration
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnvInt("DB_PORT", 5432)
	dbUser := getEnv("DB_USER", "postgres") // Default to user's username
	dbPassword := getEnv("DB_PASSWORD", "") // Default to empty password
	dbName := getEnv("DB_NAME", "ecommerce")
	dbMaxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	dbMaxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 25)
	dbConnMaxLifetimeMinutes := getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)

	return &Config{
		ServerPort: serverPort,
		Database: DatabaseConfig{
			Host:                   dbHost,
			Port:                   dbPort,
			User:                   dbUser,
			Password:               dbPassword,
			Name:                   dbName,
			MaxOpenConns:           dbMaxOpenConns,
			MaxIdleConns:           dbMaxIdleConns,
			ConnMaxLifetimeMinutes: dbConnMaxLifetimeMinutes,
		},
		JWT: JWTConfig{
			SecretKey:       jwtSecretKey,
			ExpirationHours: jwtExpirationHours,
		},
	}, nil
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// getEnvInt gets an environment variable as an integer or returns the default value
func getEnvInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}

	return intValue
}
