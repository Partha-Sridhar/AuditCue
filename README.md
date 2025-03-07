# AuditCue Integration Service

A modular integration service that connects Slack and Gmail, allowing Slack messages to trigger email notifications.

## Features

- Listens for new messages in Slack channels
- Sends email notifications to channel members via Gmail
- Secure OAuth 2.0 authentication for Gmail
- Credentials management with database integration
- Extensible architecture for adding new integrations

## Prerequisites

- Go 1.19 or higher
- PostgreSQL database
- Slack App configured with Bot token
- Gmail API credentials

## Setup

1. Clone the repository:

   ```bash
   git clone https://github.com/auditcue/integration.git
   cd integration
   ```

2. Create a `.env` file based on `.env.example`:

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. Set up the PostgreSQL database:

   ```sql
   CREATE DATABASE auditcue;
   CREATE TABLE integrations (
     id SERIAL PRIMARY KEY,
     slack_team_id TEXT UNIQUE NOT NULL,
     user_id TEXT NOT NULL,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
     updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
   );
   ```

4. Get Gmail API credentials:

   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project
   - Enable the Gmail API
   - Create OAuth credentials
   - Download the credentials JSON file as `credentials.json`

5. Run the OAuth token generator:

   ```bash
   python3 scripts/get_token.py
   ```

6. Build and run the service:
   ```bash
   go build -o bin/server ./cmd/server
   ./bin/server
   ```

## Configuration

See `.env.example` for available configuration options. Key settings include:

- `PORT`: HTTP server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string
- `OAUTH_REDIRECT_URL`: Callback URL for OAuth flow
- `OAUTH_SUCCESS_URL`: URL to redirect after successful authentication
- `CREDENTIALS_FILE_PATH`: Path to store user credentials

## API Endpoints

- `/api/slack/events`: Webhook endpoint for Slack events
- `/api/auth/credentials`: Save user credentials
- `/oauth/callback`: OAuth callback endpoint
- `/health`: Health check endpoint
- `/api/debug/integrations`: Debug endpoint to view integrations

## Usage

1. Configure your Slack App:

   - Create a new Slack App at [api.slack.com](https://api.slack.com/apps)
   - Add Bot Token Scopes: `channels:history`, `users:read`, `users:read.email`
   - Set the Event Subscription URL to `https://your-domain.com/api/slack/events`
   - Subscribe to `message.channels` bot events

2. Set up the integration:
   - Send a POST request to `/api/auth/credentials` with user credentials
   - Complete the Gmail OAuth flow
   - Test by sending a message in a Slack channel

## Development

To test Gmail functionality without setting up the full service:

```bash
# Test sending an email
python3 scripts/test.py
```

## Architecture

The service follows a modular architecture:

- `cmd/server`: Application entry point
- `config`: Configuration management
- `internal/auth`: Authentication and credentials
- `internal/database`: Database integration
- `internal/gmail`: Gmail service integration
- `internal/middleware`: HTTP middleware
- `internal/slack`: Slack service integration
- `pkg/integration`: Core integration interfaces
- `pkg/utils`: Utility functions

## License

[MIT License](LICENSE)
