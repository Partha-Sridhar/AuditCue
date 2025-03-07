package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// OAuthState represents a stored OAuth state with metadata
type OAuthState struct {
	UserID string
	Expiry time.Time
}

// OAuthStateStore securely manages OAuth state tokens
type OAuthStateStore struct {
	states map[string]OAuthState
	mu     sync.RWMutex
}

// NewOAuthStateStore creates a new state store
func NewOAuthStateStore() *OAuthStateStore {
	return &OAuthStateStore{
		states: make(map[string]OAuthState),
	}
}

// GenerateState creates and stores a secure random state
func (s *OAuthStateStore) GenerateState(userID string) (string, error) {
	// Generate cryptographically secure random state
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)

	// Store state with expiration (15 minutes)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[state] = OAuthState{
		UserID: userID,
		Expiry: time.Now().Add(15 * time.Minute),
	}

	// Clean up expired states
	s.cleanupExpiredStates()

	return state, nil
}

// ValidateState checks if a state is valid and returns the associated userID
func (s *OAuthStateStore) ValidateState(state string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stateData, exists := s.states[state]
	if !exists {
		return "", false
	}

	// Check expiration
	if time.Now().After(stateData.Expiry) {
		delete(s.states, state)
		return "", false
	}

	// Remove the state after use (one-time use)
	delete(s.states, state)
	return stateData.UserID, true
}

// cleanupExpiredStates removes any expired states
func (s *OAuthStateStore) cleanupExpiredStates() {
	now := time.Now()
	for state, data := range s.states {
		if now.After(data.Expiry) {
			delete(s.states, state)
		}
	}
}
