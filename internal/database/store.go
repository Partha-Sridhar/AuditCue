package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// IntegrationStore manages mappings between Slack teams and users
type IntegrationStore struct {
	DB *sql.DB
	mu sync.Mutex
}

// NewIntegrationStore creates a new integration store
func NewIntegrationStore(connStr string) (*IntegrationStore, error) {
	// Log the connection string (with password redacted)
	redactedConnStr := redactConnectionString(connStr)
	log.Printf("Connecting to PostgreSQL with: %s", redactedConnStr)

	// Retry connection a few times
	var db *sql.DB
	var err error
	for retries := 0; retries < 3; retries++ {
		// Open database connection
		db, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("Failed to open PostgreSQL connection (attempt %d): %v", retries+1, err)
			time.Sleep(time.Second * 2)
			continue
		}

		// Test the connection
		err = db.Ping()
		if err == nil {
			break // Connection successful
		}

		log.Printf("Failed to ping PostgreSQL (attempt %d): %v", retries+1, err)
		db.Close()
		time.Sleep(time.Second * 2)
	}

	if err != nil {
		log.Printf("All connection attempts failed: %v", err)
		return nil, err
	}

	log.Printf("Successfully connected to PostgreSQL database")

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create store
	store := &IntegrationStore{
		DB: db,
	}

	// Ensure tables exist
	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %v", err)
	}

	// Count integrations for logging
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM integrations").Scan(&count)
	if err != nil {
		log.Printf("Failed to count integrations: %v", err)
	} else {
		log.Printf("PostgreSQL initialized with %d existing integrations", count)
	}

	return store, nil
}

// initSchema creates tables if they don't exist
func (s *IntegrationStore) initSchema() error {
	_, err := s.DB.Exec(`
		CREATE TABLE IF NOT EXISTS integrations (
			id SERIAL PRIMARY KEY,
			slack_team_id TEXT UNIQUE NOT NULL,
			user_id TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`)
	return err
}

// RegisterIntegration adds a new integration mapping to the database
func (s *IntegrationStore) RegisterIntegration(slackTeamID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Registering integration for team %s to user %s", slackTeamID, userID)

	// Insert or update the integration mapping
	_, err := s.DB.Exec(`
		INSERT INTO integrations (slack_team_id, user_id, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (slack_team_id) DO UPDATE SET 
			user_id = $2,
			updated_at = NOW()
	`, slackTeamID, userID)

	if err != nil {
		log.Printf("Error saving integration to PostgreSQL: %v", err)
		return err
	}

	log.Printf("Successfully saved integration mapping to PostgreSQL")
	return nil
}

// GetUserIDForTeam returns the user ID associated with a Slack team
func (s *IntegrationStore) GetUserIDForTeam(slackTeamID string) (string, bool) {
	var userID string
	err := s.DB.QueryRow(
		"SELECT user_id FROM integrations WHERE slack_team_id = $1",
		slackTeamID,
	).Scan(&userID)

	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("Database error looking up team %s: %v", slackTeamID, err)
		}
		return "", false
	}

	return userID, true
}

// GetAllIntegrations returns all current integration mappings
func (s *IntegrationStore) GetAllIntegrations() map[string]string {
	result := make(map[string]string)

	rows, err := s.DB.Query("SELECT slack_team_id, user_id FROM integrations")
	if err != nil {
		log.Printf("Error querying integrations: %v", err)
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var teamID, userID string
		if err := rows.Scan(&teamID, &userID); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		result[teamID] = userID
	}

	return result
}

// Close closes the database connection
func (s *IntegrationStore) Close() {
	if s.DB != nil {
		s.DB.Close()
		log.Printf("PostgreSQL connection closed")
	}
}

// Helper function to redact the password in connection strings
func redactConnectionString(connStr string) string {
	redacted := connStr
	if idx := strings.Index(connStr, "://"); idx > 0 {
		if atIdx := strings.Index(connStr[idx:], "@"); atIdx > 0 {
			userPass := connStr[idx+3 : idx+atIdx]
			if passIdx := strings.Index(userPass, ":"); passIdx > 0 {
				redacted = connStr[:idx+3+passIdx+1] + "********" + connStr[idx+atIdx:]
			}
		}
	}
	return redacted
}
