package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/auditcue/integration/internal/auth"
)

// Service provides Slack integration functionality
type Service struct {
	CredManager *auth.CredentialsManager
}

// NewService creates a new Slack service
func NewService(credManager *auth.CredentialsManager) *Service {
	return &Service{
		CredManager: credManager,
	}
}

// GetChannelMembers fetches all members from a Slack channel
func (s *Service) GetChannelMembers(ctx context.Context, userID, channelID string) ([]string, error) {
	log.Printf("🔍 DETAILED GetChannelMembers DEBUG:")
	log.Printf("User ID: %s", userID)
	log.Printf("Channel ID: %s", channelID)

	// Get the Slack token for this user
	slackToken, err := s.CredManager.GetSlackToken(userID)
	if err != nil {
		log.Printf("❌ Failed to get Slack token for userID=%s: %v", userID, err)
		return nil, fmt.Errorf("failed to get Slack token: %v", err)
	}
	log.Printf("✅ Got Slack token for userID=%s (token length: %d)", userID, len(slackToken))

	url := "https://slack.com/api/conversations.members?channel=" + channelID
	log.Printf("🌐 Fetching channel members from URL: %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("❌ Failed to create request: %v", err)
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+slackToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ HTTP request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read response body: %v", err)
		return nil, err
	}

	log.Printf("📄 API Response: %s", string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("❌ Failed to parse JSON: %v", err)
		return nil, err
	}

	log.Printf("FULL API RESPONSE PARSED: %+v", result)

	if !result["ok"].(bool) {
		log.Printf("❌ Slack API error: %v", result["error"])
		return nil, fmt.Errorf("failed to fetch members: %v", result["error"])
	}

	userIDs, ok := result["members"].([]interface{})
	if !ok {
		log.Printf("❌ Failed to parse members array")
		return nil, fmt.Errorf("members array not found or invalid format")
	}

	// Limit the number of users to process to avoid overloading
	maxUsers := 50 // Set a reasonable limit
	if len(userIDs) > maxUsers {
		log.Printf("⚠️ Limiting user retrieval to %d users out of %d total", maxUsers, len(userIDs))
		userIDs = userIDs[:maxUsers]
	}

	log.Printf("✅ Processing %d user IDs in channel", len(userIDs))

	// Get all user emails
	return s.getUserEmails(ctx, userID, userIDs)
}

// getUserEmails fetches emails for a list of Slack users
func (s *Service) getUserEmails(ctx context.Context, ownerID string, userIDs []interface{}) ([]string, error) {
	// Create a channel for concurrent processing results
	type emailResult struct {
		userID string
		email  string
		err    error
	}

	resultChan := make(chan emailResult, len(userIDs))
	var emails []string
	var wg sync.WaitGroup

	// Process users concurrently with a limit
	semaphore := make(chan struct{}, 5) // Limit concurrent requests to 5

	for i, id := range userIDs {
		slackUserID := id.(string)
		wg.Add(1)

		// Use goroutine to fetch emails concurrently
		go func(i int, slackUserID string) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			// Create a context with timeout for this specific user
			userCtx, userCancel := context.WithTimeout(ctx, 3*time.Second)
			defer userCancel()

			log.Printf("🔍 Getting email for user %d/%d (ID: %s)", i+1, len(userIDs), slackUserID)
			email, err := s.getUserEmail(userCtx, ownerID, slackUserID)

			// Send result through channel
			resultChan <- emailResult{
				userID: slackUserID,
				email:  email,
				err:    err,
			}
		}(i, slackUserID)
	}

	// Close result channel after all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results from channel
	for result := range resultChan {
		if result.err != nil {
			log.Printf("⚠️ Error getting email for user %s: %v", result.userID, result.err)
			continue
		}

		if result.email != "" {
			log.Printf("✅ Found email for user %s: %s", result.userID, result.email)
			emails = append(emails, result.email)
		} else {
			log.Printf("⚠️ No email found for user %s", result.userID)
		}
	}

	log.Printf("📧 Retrieved %d emails from %d users", len(emails), len(userIDs))
	return emails, nil
}

// getUserEmail retrieves a user's email
func (s *Service) getUserEmail(ctx context.Context, ownerID, slackUserID string) (string, error) {
	log.Printf("🔍 DETAILED getUserEmail DEBUG:")
	log.Printf("Owner ID: %s", ownerID)
	log.Printf("Slack User ID: %s", slackUserID)

	// Get the Slack token for this user
	slackToken, err := s.CredManager.GetSlackToken(ownerID)
	if err != nil {
		log.Printf("❌ CRITICAL: Failed to get Slack token for ownerID %s: %v", ownerID, err)
		return "", fmt.Errorf("failed to get Slack token: %v", err)
	}
	log.Printf("✅ Slack Token Length: %d", len(slackToken))

	url := "https://slack.com/api/users.info?user=" + slackUserID
	log.Printf("🌐 Fetching user info from URL: %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("❌ Failed to create request: %v", err)
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+slackToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ HTTP request failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read response body: %v", err)
		return "", err
	}

	log.Printf("📄 API Response for user %s: %s", slackUserID, string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("❌ Failed to parse JSON: %v", err)
		return "", err
	}

	if !result["ok"].(bool) {
		log.Printf("❌ Slack API error: %v", result["error"])
		return "", fmt.Errorf("failed to fetch user info: %v", result["error"])
	}

	user, ok := result["user"].(map[string]interface{})
	if !ok {
		log.Printf("❌ User object not found or invalid format")
		return "", fmt.Errorf("user object not found")
	}

	profile, ok := user["profile"].(map[string]interface{})
	if !ok {
		log.Printf("❌ Profile object not found or invalid format")
		return "", fmt.Errorf("profile object not found")
	}

	// Log all keys in the profile
	log.Printf("PROFILE KEYS:")
	for key := range profile {
		log.Printf("  - %s: %v", key, profile[key])
	}

	if email, exists := profile["email"].(string); exists {
		log.Printf("✅ Found email for user %s: %s", slackUserID, email)
		return email, nil
	}

	log.Printf("⚠️ No email found in profile for user %s", slackUserID)
	return "", nil
}
