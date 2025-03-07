package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server struct {
		Port         string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		IdleTimeout  time.Duration
	}
	Database struct {
		URL             string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
	}
	OAuth struct {
		RedirectURL string
		SuccessURL  string
	}
	Credentials struct {
		FilePath string
	}
}

// Load returns application configuration loaded from environment variables
func Load() *Config {
	cfg := &Config{}

	// Server config
	cfg.Server.Port = getEnv("PORT", "8080")
	cfg.Server.ReadTimeout = getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second)
	cfg.Server.WriteTimeout = getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second)
	cfg.Server.IdleTimeout = getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second)

	// Database config
	cfg.Database.URL = getEnv("DATABASE_URL", "postgres://postgres:password@localhost/auditcue?sslmode=disable")
	cfg.Database.MaxOpenConns = getIntEnv("DB_MAX_OPEN_CONNS", 25)
	cfg.Database.MaxIdleConns = getIntEnv("DB_MAX_IDLE_CONNS", 5)
	cfg.Database.ConnMaxLifetime = getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute)

	// OAuth config
	cfg.OAuth.RedirectURL = getEnv("OAUTH_REDIRECT_URL", "https://auditcue-integration-production.up.railway.app/oauth/callback")
	cfg.OAuth.SuccessURL = getEnv("OAUTH_SUCCESS_URL", "https://auditcue-integration-production.up.railway.app/success")

	// Credentials config
	cfg.Credentials.FilePath = getEnv("CREDENTIALS_FILE_PATH", "user_credentials.json")

	return cfg
}

// Helper functions to read environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
