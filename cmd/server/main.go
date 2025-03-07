package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/auditcue/integration/config"
	"github.com/auditcue/integration/internal/auth"
	"github.com/auditcue/integration/internal/database"
	"github.com/auditcue/integration/internal/gmail"
	"github.com/auditcue/integration/internal/middleware"
	"github.com/auditcue/integration/internal/slack"
)

func main() {
	log.Println("Starting AuditCue Integration Service...")

	// Load configuration
	cfg := config.Load()
	log.Println("Configuration loaded")

	// Initialize credential manager
	log.Println("Initializing credential manager...")
	credManager, err := auth.NewCredentialsManager(cfg.Credentials.FilePath)
	if err != nil {
		log.Fatalf("Failed to initialize credentials manager: %v", err)
	}

	// Initialize PostgreSQL-based integration store
	log.Println("Initializing PostgreSQL database...")
	integrationStore, err := database.NewIntegrationStore(cfg.Database.URL)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL: %v", err)
	}

	// Ensure database connection is closed on shutdown
	defer integrationStore.Close()

	// Initialize services
	gmailService := gmail.NewService(credManager)
	slackService := slack.NewService(credManager)

	// Create auth handler with integration registration function
	authHandler := auth.NewHandler(
		credManager,
		cfg.OAuth.RedirectURL,
		cfg.OAuth.SuccessURL,
		integrationStore.RegisterIntegration,
	)

	// Create Slack handler with necessary dependencies
	slackHandler := slack.NewHandler(
		slackService,
		gmailService,
		integrationStore.GetUserIDForTeam,
	)

	// API endpoints
	mux := http.NewServeMux()

	// Slack webhook endpoint
	mux.HandleFunc("/api/slack/events", slackHandler.HandleWebhook)
	mux.HandleFunc("/slack/events", slackHandler.HandleWebhook) // Add fallback URL without /api prefix

	// Auth endpoints
	mux.HandleFunc("/api/auth/credentials", authHandler.SaveCredentials)
	mux.HandleFunc("/oauth/callback", authHandler.GmailOAuthCallback)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		if err := integrationStore.DB.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(fmt.Sprintf("Database connection error: %v", err)))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK - Service is healthy"))
	})

	// Debug endpoint to check stored integrations
	mux.HandleFunc("/api/debug/integrations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		allIntegrations := integrationStore.GetAllIntegrations()
		w.Write([]byte(fmt.Sprintf("Found %d integrations: %v",
			len(allIntegrations), allIntegrations)))
	})

	// Add database connection test endpoint
	mux.HandleFunc("/api/db/test", func(w http.ResponseWriter, r *http.Request) {
		if err := integrationStore.DB.Ping(); err != nil {
			http.Error(w, fmt.Sprintf("Database connection error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("Database connection successful"))
	})

	// Chain all middleware
	handler := middleware.Recovery(
		middleware.Logging(
			middleware.CORS(mux),
		),
	)

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine to allow for graceful shutdown
	go func() {
		log.Printf("Server ready! Listening on port %s...", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-quit
	log.Println("Shutdown signal received...")

	// Create a deadline context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	log.Println("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}
