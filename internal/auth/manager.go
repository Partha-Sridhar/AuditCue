package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// UserCredentials represents the credentials for a specific user
type UserCredentials struct {
	UserID            string `json:"user_id"`
	GmailClientID     string `json:"gmail_client_id"`
	GmailSecret       string `json:"gmail_secret"`
	GmailRefreshToken string `json:"gmail_refresh_token"`
	SlackBotToken     string `json:"slack_bot_token"`
}

// CredentialsManager handles storing and retrieving user credentials
type CredentialsManager struct {
	credentialsFile string
	credentials     map[string]UserCredentials
	mu              sync.RWMutex
}

// NewCredentialsManager creates a new credentials manager
func NewCredentialsManager(filename string) (*CredentialsManager, error) {
	cm := &CredentialsManager{
		credentialsFile: filename,
		credentials:     make(map[string]UserCredentials),
	}

	// Try to load existing credentials
	if _, err := os.Stat(filename); err == nil {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("error reading credentials file: %v", err)
		}

		if err := json.Unmarshal(data, &cm.credentials); err != nil {
			return nil, fmt.Errorf("error parsing credentials file: %v", err)
		}

		log.Printf("Loaded %d user credentials from %s", len(cm.credentials), filename)
	} else {
		log.Printf("No existing credentials file found at %s, creating new", filename)
	}

	return cm, nil
}

// SaveCredentials adds or updates user credentials
func (cm *CredentialsManager) SaveCredentials(creds UserCredentials) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.credentials[creds.UserID] = creds

	// Save to file
	data, err := json.MarshalIndent(cm.credentials, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing credentials: %v", err)
	}

	if err := os.WriteFile(cm.credentialsFile, data, 0600); err != nil {
		return fmt.Errorf("error writing credentials file: %v", err)
	}

	log.Printf("Saved credentials for user %s", creds.UserID)

	return nil
}

// GetCredentials retrieves credentials for a specific user
func (cm *CredentialsManager) GetCredentials(userID string) (UserCredentials, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	creds, ok := cm.credentials[userID]
	if !ok {
		return UserCredentials{}, fmt.Errorf("no credentials found for user %s", userID)
	}

	return creds, nil
}

// GetGmailService creates a Gmail service for a specific user
func (cm *CredentialsManager) GetGmailService(ctx context.Context, userID string) (*gmail.Service, error) {
	creds, err := cm.GetCredentials(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %v", err)
	}

	// Create OAuth config
	config := &oauth2.Config{
		ClientID:     creds.GmailClientID,
		ClientSecret: creds.GmailSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/gmail.send"},
	}

	// Use stored refresh token
	token := &oauth2.Token{
		RefreshToken: creds.GmailRefreshToken,
	}

	client := config.Client(ctx, token)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to create Gmail service: %v", err)
	}

	return srv, nil
}

// GetSlackToken returns the Slack bot token for a user
func (cm *CredentialsManager) GetSlackToken(userID string) (string, error) {
	creds, err := cm.GetCredentials(userID)
	if err != nil {
		return "", err
	}

	return creds.SlackBotToken, nil
}
