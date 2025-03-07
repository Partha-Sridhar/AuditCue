#!/bin/bash

# Exit on error
set -e

echo "🚀 Setting up AuditCue Integration Service directory structure..."

# Create root directory
mkdir -p auditcue-integration
cd auditcue-integration

# Create main directories with placeholders
echo "📁 Creating directory structure and placeholder files..."

# cmd directory
mkdir -p cmd/server
cat > cmd/server/main.go << 'EOF'
package main

// Main entry point for the application
func main() {
	// TODO: Implement server initialization
}
EOF

# config directory
mkdir -p config
cat > config/config.go << 'EOF'
package config

// Config holds all application configuration
type Config struct {
	// TODO: Add configuration fields
}

// Load returns application configuration
func Load() *Config {
	// TODO: Implement configuration loading
	return &Config{}
}
EOF

# internal directory and subdirectories
mkdir -p internal/auth
mkdir -p internal/database
mkdir -p internal/gmail
mkdir -p internal/middleware
mkdir -p internal/slack

# auth module
cat > internal/auth/handlers.go << 'EOF'
package auth

// Handler manages authentication endpoints
type Handler struct {
	// TODO: Implement authentication handlers
}
EOF

cat > internal/auth/manager.go << 'EOF'
package auth

// CredentialsManager handles user credentials
type CredentialsManager struct {
	// TODO: Implement credentials management
}
EOF

cat > internal/auth/state.go << 'EOF'
package auth

// OAuthStateStore manages OAuth state tokens
type OAuthStateStore struct {
	// TODO: Implement OAuth state management
}
EOF

# database module
cat > internal/database/models.go << 'EOF'
package database

// Integration represents the mapping between Slack teams and users
type Integration struct {
	// TODO: Define integration model fields
}
EOF

cat > internal/database/store.go << 'EOF'
package database

// IntegrationStore manages integrations in the database
type IntegrationStore struct {
	// TODO: Implement database operations
}
EOF

# gmail module
cat > internal/gmail/service.go << 'EOF'
package gmail

// Service handles Gmail API operations
type Service struct {
	// TODO: Implement Gmail service
}
EOF

# middleware module
cat > internal/middleware/middleware.go << 'EOF'
package middleware

import "net/http"

// Recovery middleware to prevent server crashes
func Recovery(next http.Handler) http.Handler {
	// TODO: Implement recovery middleware
	return nil
}

// Logging middleware for request logging
func Logging(next http.Handler) http.Handler {
	// TODO: Implement logging middleware
	return nil
}

// CORS middleware for handling cross-origin requests
func CORS(next http.Handler) http.Handler {
	// TODO: Implement CORS middleware
	return nil
}
EOF

# slack module
cat > internal/slack/handler.go << 'EOF'
package slack

// Handler processes incoming Slack events
type Handler struct {
	// TODO: Implement Slack event handler
}
EOF

cat > internal/slack/models.go << 'EOF'
package slack

// SlackEvent represents an incoming Slack event
type SlackEvent struct {
	// TODO: Define Slack event model
}
EOF

cat > internal/slack/service.go << 'EOF'
package slack

// Service provides Slack integration functionality
type Service struct {
	// TODO: Implement Slack service
}
EOF

# pkg directory
mkdir -p pkg/integration
mkdir -p pkg/utils

cat > pkg/integration/integration.go << 'EOF'
package integration

// Integration represents an external service integration
type Integration interface {
	// TODO: Define integration interface
}
EOF

cat > pkg/utils/http.go << 'EOF'
package utils

import "net/http"

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) error {
	// TODO: Implement JSON response helper
	return nil
}
EOF

# scripts directory
mkdir -p scripts
cat > scripts/get_token.py << 'EOF'
#!/usr/bin/env python3
"""
Script to get OAuth tokens for Gmail.
"""

def main():
    # TODO: Implement OAuth token retrieval
    pass

if __name__ == "__main__":
    main()
EOF
chmod +x scripts/get_token.py

cat > scripts/test.py << 'EOF'
#!/usr/bin/env python3
"""
Script to test Gmail integration.
"""

def main():
    # TODO: Implement Gmail test
    pass

if __name__ == "__main__":
    main()
EOF
chmod +x scripts/test.py

# Root files
cat > .env.example << 'EOF'
# Server configuration
PORT=8080

# Database connection
DATABASE_URL=postgres://postgres:password@localhost/auditcue?sslmode=disable

# OAuth configuration
OAUTH_REDIRECT_URL=https://your-app-domain.com/oauth/callback
OAUTH_SUCCESS_URL=https://your-app-domain.com/success

# Credentials
CREDENTIALS_FILE_PATH=user_credentials.json
EOF

cat > .gitignore << 'EOF'
# Binaries
/bin
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out

# Dependency directories
vendor/

# Environment files
.env

# Credentials
user_credentials.json
token.json
credentials.json

# IDE files
.idea/
.vscode/
*.sublime-project
*.sublime-workspace

# OS specific
.DS_Store
EOF

cat > go.mod << 'EOF'
module github.com/auditcue/integration

go 1.19

// TODO: Add dependencies
EOF

cat > README.md << 'EOF'
# AuditCue Integration Service

Integration service that connects Slack and Gmail.

## Features

- Listens for Slack messages
- Sends email notifications via Gmail
- Uses OAuth2 for Gmail authentication

## Setup

TODO: Add setup instructions

## Configuration

See `.env.example` for available configuration options.

## Development

TODO: Add development instructions
EOF

echo "✅ Directory structure created successfully!"
echo "📂 Created in: $(pwd)"