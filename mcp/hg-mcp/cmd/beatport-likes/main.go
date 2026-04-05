// beatport-likes - Extract track/artist IDs from playlist for Playwright automation
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: beatport-likes <playlist-id>")
		fmt.Println("Outputs track and artist IDs for use with Playwright automation")
		os.Exit(1)
	}

	playlistID := os.Args[1]

	username := os.Getenv("BEATPORT_USERNAME")
	password := os.Getenv("BEATPORT_PASSWORD")
	if username == "" || password == "" {
		fmt.Fprintln(os.Stderr, "BEATPORT_USERNAME and BEATPORT_PASSWORD required")
		os.Exit(1)
	}

	ctx := context.Background()
	client, err := clients.NewBeatportClientWithCredentials(username, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
		os.Exit(1)
	}

	if err := client.Authenticate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Auth failed: %v\n", err)
		os.Exit(1)
	}

	// Parse playlist ID
	var pid int
	if _, err := fmt.Sscanf(playlistID, "%d", &pid); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid playlist ID: %s\n", playlistID)
		os.Exit(1)
	}

	// Get all playlist tracks
	allTracks, err := client.GetAllPlaylistTracks(ctx, pid)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching tracks: %v\n", err)
		os.Exit(1)
	}

	// Extract unique track IDs
	var trackIDs []string
	for _, t := range allTracks {
		trackIDs = append(trackIDs, fmt.Sprintf("%d", t.ID))
	}

	// Extract unique artist IDs
	artistSet := make(map[int]bool)
	var artistIDs []string
	for _, t := range allTracks {
		for _, a := range t.Artists {
			if !artistSet[a.ID] {
				artistSet[a.ID] = true
				artistIDs = append(artistIDs, fmt.Sprintf("%d", a.ID))
			}
		}
	}

	fmt.Printf("TRACKS=%s\n", strings.Join(trackIDs, ","))
	fmt.Printf("ARTISTS=%s\n", strings.Join(artistIDs, ","))
	fmt.Printf("# %d tracks, %d unique artists\n", len(trackIDs), len(artistIDs))
}
