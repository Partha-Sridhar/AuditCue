package database

import (
	"time"
)

// Integration represents the mapping between Slack teams and users
type Integration struct {
	ID          int64     `json:"id"`
	SlackTeamID string    `json:"slack_team_id"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
