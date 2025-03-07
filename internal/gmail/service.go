package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/auditcue/integration/internal/auth"
	"google.golang.org/api/gmail/v1"
)

// Service handles Gmail API operations
type Service struct {
	CredManager *auth.CredentialsManager
}

// NewService creates a new Gmail service
func NewService(credManager *auth.CredentialsManager) *Service {
	return &Service{
		CredManager: credManager,
	}
}

// SendEmail sends an email using Gmail API with retries
func (s *Service) SendEmail(userID string, to, subject, messageText string) error {
	log.Printf("📨 Attempting to send email from userID: %s", userID)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get Gmail service with retry
	var srv *gmail.Service
	var err error
	var backoff time.Duration = 1 * time.Second

	// Try up to 3 times to get Gmail service
	for attempt := 1; attempt <= 3; attempt++ {
		srv, err = s.CredManager.GetGmailService(ctx, userID)
		if err == nil {
			break
		}
		log.Printf("❌ Error retrieving Gmail service (attempt %d/3): %v", attempt, err)

		if attempt < 3 {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}
	}

	if err != nil {
		return fmt.Errorf("unable to retrieve Gmail service after retries: %v", err)
	}
	log.Println("✅ Successfully retrieved Gmail service")

	// Get sender email
	profile, err := srv.Users.GetProfile("me").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("unable to get user profile: %v", err)
	}
	fromEmail := profile.EmailAddress
	log.Printf("📧 Retrieved sender email: %s", fromEmail)

	// Construct email message with proper MIME headers
	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"+
			"%s",
		fromEmail, to, subject, messageText,
	))

	// Proper Base64 encoding (URL-safe)
	encodedMsg := base64.URLEncoding.EncodeToString(msg)
	encodedMsg = strings.ReplaceAll(encodedMsg, "+", "-") // Gmail requires URL-safe base64
	encodedMsg = strings.ReplaceAll(encodedMsg, "/", "_") // Gmail requires URL-safe base64
	encodedMsg = strings.TrimRight(encodedMsg, "=")       // Remove padding

	// Create Gmail API message
	gmailMessage := &gmail.Message{
		Raw: encodedMsg,
	}

	// Send email with retry
	backoff = 1 * time.Second
	var response *gmail.Message

	for attempt := 1; attempt <= 3; attempt++ {
		log.Printf("📤 Sending email (attempt %d/3)...", attempt)
		response, err = srv.Users.Messages.Send("me", gmailMessage).Context(ctx).Do()
		if err == nil {
			break
		}
		log.Printf("❌ Error sending email (attempt %d/3): %v", attempt, err)

		if attempt < 3 {
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
		}
	}

	if err != nil {
		return fmt.Errorf("unable to send email after retries: %v", err)
	}

	log.Printf("✅ Email sent successfully! Message ID: %s", response.Id)
	return nil
}
