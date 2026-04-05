// Gmail OAuth helper - generates Gmail-scoped credentials
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
)

func main() {
	// Use existing client ID and secret from Google Drive API credentials
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")

	if clientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID environment variable is required")
	}
	if clientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET environment variable is required")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes: []string{
			gmail.GmailReadonlyScope,
			gmail.GmailLabelsScope,
		},
		RedirectURL: "http://localhost:8085/callback",
	}

	// Start local server for callback
	codeChan := make(chan string)
	errChan := make(chan error)

	server := &http.Server{Addr: ":8085"}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no code in callback")
			fmt.Fprintf(w, "Error: No authorization code received")
			return
		}
		codeChan <- code
		fmt.Fprintf(w, "<html><body><h1>Success!</h1><p>Gmail authorization complete. You can close this window.</p></body></html>")
	})

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Generate auth URL
	authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	fmt.Println("\n=== Gmail OAuth Setup ===")
	fmt.Println("\n1. Open this URL in your browser:")
	fmt.Println(authURL)
	fmt.Println("\n2. Sign in with your Google account and grant access")
	fmt.Println("3. The page will redirect back automatically")

	// Wait for code
	var code string
	select {
	case code = <-codeChan:
		fmt.Println("Authorization code received!")
	case err := <-errChan:
		log.Fatalf("Error: %v", err)
	}

	// Exchange code for token
	ctx := context.Background()
	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Failed to exchange code: %v", err)
	}

	// Save credentials file
	creds := map[string]string{
		"type":          "authorized_user",
		"client_id":     clientID,
		"client_secret": clientSecret,
		"refresh_token": token.RefreshToken,
	}

	homeDir, _ := os.UserHomeDir()
	credPath := filepath.Join(homeDir, ".config", "gcloud", "gmail_credentials.json")
	os.MkdirAll(filepath.Dir(credPath), 0700)

	data, _ := json.MarshalIndent(creds, "", "  ")
	if err := os.WriteFile(credPath, data, 0600); err != nil {
		log.Fatalf("Failed to save credentials: %v", err)
	}

	fmt.Printf("\nCredentials saved to: %s\n", credPath)
	fmt.Println("\nYou can now use the Gmail import tools!")

	server.Shutdown(ctx)
}
