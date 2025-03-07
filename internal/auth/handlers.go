package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/auditcue/integration/pkg/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// SaveUserCredentialsRequest represents the request to save user credentials
type SaveUserCredentialsRequest struct {
	UserID        string `json:"user_id"`
	GmailClientID string `json:"gmail_client_id"`
	GmailSecret   string `json:"gmail_secret"`
	SlackBotToken string `json:"slack_bot_token"`
	SlackTeamID   string `json:"slack_team_id"`
}

// Handler manages authentication endpoints
type Handler struct {
	CredManager  *CredentialsManager
	StateStore   *OAuthStateStore
	RedirectURL  string
	SuccessURL   string
	RegisterTeam func(string, string) error
}

// NewHandler creates a new auth handler
func NewHandler(credManager *CredentialsManager, redirectURL, successURL string, registerTeam func(string, string) error) *Handler {
	return &Handler{
		CredManager:  credManager,
		StateStore:   NewOAuthStateStore(),
		RedirectURL:  redirectURL,
		SuccessURL:   successURL,
		RegisterTeam: registerTeam,
	}
}

// SaveCredentials handles saving user credentials and generating OAuth URL
func (h *Handler) SaveCredentials(w http.ResponseWriter, r *http.Request) {
	log.Printf("SaveCredentials called from: %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		utils.RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SaveUserCredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	log.Printf("Received credentials save request for user: %s, team: %s", req.UserID, req.SlackTeamID)

	// Save slack token immediately
	userCreds := UserCredentials{
		UserID:        req.UserID,
		GmailClientID: req.GmailClientID,
		GmailSecret:   req.GmailSecret,
		SlackBotToken: req.SlackBotToken,
	}

	if err := h.CredManager.SaveCredentials(userCreds); err != nil {
		log.Printf("Error saving credentials: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Error saving credentials")
		return
	}

	log.Printf("Credentials saved successfully for user: %s", req.UserID)

	// Register the integration mapping
	if req.SlackTeamID != "" && h.RegisterTeam != nil {
		log.Printf("Registering integration for team: %s", req.SlackTeamID)
		if err := h.RegisterTeam(req.SlackTeamID, req.UserID); err != nil {
			log.Printf("Error registering integration: %v", err)
			utils.RespondError(w, http.StatusInternalServerError, "Error registering integration")
			return
		}
		log.Printf("Integration registered successfully: %s -> %s", req.SlackTeamID, req.UserID)
	}

	// Generate Gmail OAuth URL
	config := &oauth2.Config{
		ClientID:     req.GmailClientID,
		ClientSecret: req.GmailSecret,
		RedirectURL:  h.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/gmail.send"},
		Endpoint:     google.Endpoint,
	}

	// Generate secure random state
	state, err := h.StateStore.GenerateState(req.UserID)
	if err != nil {
		log.Printf("Error generating OAuth state: %v", err)
		utils.RespondError(w, http.StatusInternalServerError, "Error generating OAuth state")
		return
	}

	// Generate auth URL
	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	// Return the auth URL to the frontend
	utils.RespondJSON(w, http.StatusOK, map[string]string{
		"auth_url": authURL,
	})

	log.Printf("Auth URL generated successfully for user: %s", req.UserID)
}

// GmailOAuthCallback handles the Gmail OAuth callback
func (h *Handler) GmailOAuthCallback(w http.ResponseWriter, r *http.Request) {
	log.Printf("GmailOAuthCallback called from: %s", r.RemoteAddr)

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	log.Printf("Received OAuth callback with state: %s", state)

	// Validate state and get userID
	userID, valid := h.StateStore.ValidateState(state)
	if !valid {
		log.Printf("Invalid state parameter: %s", state)
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	log.Printf("State validation successful for user: %s", userID)

	// Get the user's stored credentials
	userCreds, err := h.CredManager.GetCredentials(userID)
	if err != nil {
		log.Printf("User not found: %s", userID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Create OAuth config
	config := &oauth2.Config{
		ClientID:     userCreds.GmailClientID,
		ClientSecret: userCreds.GmailSecret,
		RedirectURL:  h.RedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/gmail.send"},
		Endpoint:     google.Endpoint,
	}

	// Define a timeout duration
	timeout := 10 * time.Second

	// Exchange code for token with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		http.Error(w, "Failed to authenticate with Google. Please try again.", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully exchanged code for token")

	// Store the refresh token
	if token.RefreshToken != "" {
		log.Printf("Received refresh token, saving for user: %s", userID)
		userCreds.GmailRefreshToken = token.RefreshToken
		if err := h.CredManager.SaveCredentials(userCreds); err != nil {
			log.Printf("Error saving Gmail refresh token: %v", err)
			http.Error(w, "Error saving token", http.StatusInternalServerError)
			return
		}
		log.Printf("Refresh token saved successfully")
	} else {
		log.Printf("Warning: No refresh token received from Google")
	}

	// Redirect to frontend with success message
	http.Redirect(w, r, h.SuccessURL, http.StatusTemporaryRedirect)
	log.Printf("OAuth flow completed successfully for user: %s", userID)
}
