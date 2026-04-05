// +build ignore

package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Playlist struct {
	ID               int        `json:"id"`
	Service          string     `json:"service"`
	Username         string     `json:"username"`
	URL              string     `json:"url"`
	GDrivePath       string     `json:"gdrive_path"`
	Name             *string    `json:"name"`
	Description      *string    `json:"description"`
	TrackCount       int        `json:"track_count"`
	LastSyncedAt     *time.Time `json:"last_synced_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	IndexedAt        *time.Time `json:"indexed_at"`
	IndexStatus      string     `json:"index_status"`
	TracksDiscovered int        `json:"tracks_discovered"`
	TracksQueued     int        `json:"tracks_queued"`
	TracksCompleted  int        `json:"tracks_completed"`
	TracksFailed     int        `json:"tracks_failed"`
}

type DynamoPlaylist struct {
	ID               string  `dynamodbav:"id"`
	Service          string  `dynamodbav:"service"`
	Username         string  `dynamodbav:"username"`
	URL              string  `dynamodbav:"url"`
	GDrivePath       string  `dynamodbav:"gdrive_path,omitempty"`
	Name             string  `dynamodbav:"name,omitempty"`
	Description      string  `dynamodbav:"description,omitempty"`
	TrackCount       int     `dynamodbav:"track_count"`
	LastSyncedAt     string  `dynamodbav:"last_synced_at,omitempty"`
	CreatedAt        string  `dynamodbav:"created_at"`
	UpdatedAt        string  `dynamodbav:"updated_at"`
	IndexedAt        string  `dynamodbav:"indexed_at,omitempty"`
	IndexStatus      string  `dynamodbav:"index_status"`
	TracksDiscovered int     `dynamodbav:"tracks_discovered"`
	TracksQueued     int     `dynamodbav:"tracks_queued"`
	TracksCompleted  int     `dynamodbav:"tracks_completed"`
	TracksFailed     int     `dynamodbav:"tracks_failed"`
	UserID           string  `dynamodbav:"user_id"`
	Status           string  `dynamodbav:"status"`
}

func main() {
	ctx := context.Background()

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithSharedConfigProfile("cr8"),
	)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	// Read playlists.json
	data, err := os.ReadFile("exports/playlists.json")
	if err != nil {
		log.Fatalf("Failed to read playlists.json: %v", err)
	}

	var playlists []Playlist
	if err := json.Unmarshal(data, &playlists); err != nil {
		log.Fatalf("Failed to parse playlists.json: %v", err)
	}

	log.Printf("Migrating %d playlists to DynamoDB...", len(playlists))

	for _, p := range playlists {
		dp := DynamoPlaylist{
			ID:               strconv.Itoa(p.ID),
			Service:          p.Service,
			Username:         p.Username,
			URL:              p.URL,
			GDrivePath:       p.GDrivePath,
			TrackCount:       p.TrackCount,
			CreatedAt:        p.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        p.UpdatedAt.Format(time.RFC3339),
			IndexStatus:      p.IndexStatus,
			TracksDiscovered: p.TracksDiscovered,
			TracksQueued:     p.TracksQueued,
			TracksCompleted:  p.TracksCompleted,
			TracksFailed:     p.TracksFailed,
			UserID:           p.Username,
			Status:           p.IndexStatus,
		}

		if p.Name != nil {
			dp.Name = *p.Name
		}
		if p.Description != nil {
			dp.Description = *p.Description
		}
		if p.LastSyncedAt != nil {
			dp.LastSyncedAt = p.LastSyncedAt.Format(time.RFC3339)
		}
		if p.IndexedAt != nil {
			dp.IndexedAt = p.IndexedAt.Format(time.RFC3339)
		}

		item, err := attributevalue.MarshalMap(dp)
		if err != nil {
			log.Printf("Failed to marshal playlist %d: %v", p.ID, err)
			continue
		}

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("cr8_playlists"),
			Item:      item,
		})
		if err != nil {
			log.Printf("Failed to put playlist %d: %v", p.ID, err)
			continue
		}

		log.Printf("Migrated playlist %d: %s", p.ID, p.URL)
	}

	log.Printf("Migration complete!")
}
