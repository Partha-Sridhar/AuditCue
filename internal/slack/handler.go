package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/auditcue/integration/internal/gmail"
	"github.com/auditcue/integration/pkg/utils"
)

// Handler processes incoming Slack events
type Handler struct {
	SlackService     *Service
	GmailService     *gmail.Service
	GetUserIDForTeam func(string) (string, bool)
}

// NewHandler creates a new Slack webhook handler
func NewHandler(slackService *Service, gmailService *gmail.Service, getUserIDForTeam func(string) (string, bool)) *Handler {
	return &Handler{
		SlackService:     slackService,
		GmailService:     gmailService,
		GetUserIDForTeam: getUserIDForTeam,
	}
}

// HandleWebhook processes Slack webhooks and manages events
func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 SlackMessageListener received request from: %s, method: %s", r.RemoteAddr, r.Method)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error reading request body: %v", err)
		utils.RespondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Log the raw request body for debugging
	log.Printf("🔍 FULL SLACK EVENT DETAILS:")
	log.Printf("Full Raw Body: %s", string(body))

	// First, check if this is a challenge request by trying to parse it as such
	var challenge SlackChallenge
	if err := json.Unmarshal(body, &challenge); err == nil && challenge.Type == "url_verification" {
		log.Printf("✅ Received URL verification challenge: %s", challenge.Challenge)
		// Ensure proper JSON response format for Slack
		utils.RespondJSON(w, http.StatusOK, map[string]string{"challenge": challenge.Challenge})
		return
	}

	// If not a challenge, try parsing as a normal event
	var event SlackEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("❌ Error parsing JSON: %v", err)
		log.Printf("📄 Raw body content: %s", string(body))
		utils.RespondError(w, http.StatusBadRequest, "Invalid event JSON")
		return
	}

	// Log detailed event information
	log.Printf("🔍 FULL SLACK EVENT DETAILS:")
	log.Printf("Event Type: %s", event.Type)
	log.Printf("Team ID: %s", event.TeamID)
	log.Printf("Channel ID: %s", event.Event.Channel)
	log.Printf("User ID: %s", event.Event.User)
	log.Printf("Message Text: %s", event.Event.Text)
	log.Printf("Event Subtype: %s", event.Event.Subtype)
	log.Printf("Bot ID: %s", event.Event.BotID)

	// For actual event notifications
	if event.Type == "event_callback" {
		log.Printf("📨 Received event_callback with Event Type: %s", event.Event.Type)

		// Check if we have a team ID
		if event.TeamID == "" {
			log.Printf("❌ No team ID found in event")
			utils.RespondError(w, http.StatusBadRequest, "No team ID in event")
			return
		}

		// Find userID associated with this Slack team
		userID, exists := h.GetUserIDForTeam(event.TeamID)
		if !exists {
			log.Printf("❌ No integration found for team ID: %s", event.TeamID)
			utils.RespondError(w, http.StatusNotFound, "Integration not found")
			return
		}

		log.Printf("✅ Found integration: TeamID=%s mapped to UserID=%s", event.TeamID, userID)

		// Process incoming Slack messages
		if event.Event.Type == "message" {
			// Check if this is a bot message or a message edit, which we should ignore
			if event.Event.BotID != "" || event.Event.Subtype == "message_changed" || event.Event.Subtype == "bot_message" {
				log.Printf("🤖 Ignoring bot message or message edit: subtype=%s, bot_id=%s",
					event.Event.Subtype, event.Event.BotID)
				w.WriteHeader(http.StatusOK)
				return
			}

			log.Printf("💬 Processing message event from user %s: %s", event.Event.User, event.Event.Text)
			channelID := event.Event.Channel

			// Create timeout context for processing
			ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
			defer cancel()

			// Get sender's Slack user info to exclude them from recipients
			senderSlackID := event.Event.User
			senderEmail, err := h.SlackService.getUserEmail(ctx, userID, senderSlackID)
			if err != nil {
				log.Printf("⚠️ Could not get sender email: %v", err)
				// Continue with processing even if we couldn't get the sender email
			} else {
				log.Printf("👤 Sender email: %s", senderEmail)
			}

			log.Printf("🔍 Fetching users for channel: %s", channelID)

			// Get all channel member emails
			emails, err := h.SlackService.GetChannelMembers(ctx, userID, channelID)
			if err != nil {
				log.Printf("❌ Error fetching users: %s", err)
				utils.RespondError(w, http.StatusInternalServerError, "Error fetching users")
				return
			}

			log.Printf("✅ Found %d emails for channel %s: %v", len(emails), channelID, emails)

			// Exclude the sender from the recipients list
			var recipients []string
			for _, email := range emails {
				if email != senderEmail && email != "" {
					recipients = append(recipients, email)
				}
			}

			if len(recipients) == 0 {
				log.Printf("⚠️ No recipients left after excluding sender")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("No recipients to send email to"))
				return
			}

			log.Printf("📧 Recipients after excluding sender: %v", recipients)

			subject := "New Slack Message in Channel"
			message := fmt.Sprintf("New Slack message received in channel from user %s:\n\n%s",
				event.Event.User, event.Event.Text)

			recipientsStr := strings.Join(recipients, ",")
			log.Printf("📧 Attempting to send email to: %s", recipientsStr)

			// Respond to Slack immediately
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Message received and processing"))

			// Process email in background
			go func() {
				err := h.GmailService.SendEmail(userID, recipientsStr, subject, message)
				if err != nil {
					log.Printf("❌ Error sending email: %s", err)
				} else {
					log.Println("✅ Email successfully sent to all recipients!")
				}
			}()
			return
		} else {
			log.Printf("⚠️ Ignoring non-message event: %s", event.Event.Type)
		}
	} else {
		log.Printf("⚠️ Ignoring event with type: %s (not an event_callback)", event.Type)
	}

	// Default response for unhandled events
	log.Printf("⚠️ Event received but not specifically handled")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Event received"))
}
