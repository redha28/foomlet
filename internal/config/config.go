package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Global config instance
var AppConfig *Config

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// Initialize loads config values from .env and sets up the global config
func Initialize() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	AppConfig = &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "ewallet"),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "yourAccessTokenSecret123"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "yourRefreshTokenSecret456"),
			AccessExpiry:  getDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
			RefreshExpiry: getDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		},
	}

	return nil
}

// GetConfig returns the global config instance
func GetConfig() *Config {
	if AppConfig == nil {
		// If not initialized, load with default values
		AppConfig = &Config{
			// ...default values...
		}
	}
	return AppConfig
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
