package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// AWSStorage implements sync state and playlist storage using DynamoDB.
type AWSStorage struct {
	client    *dynamodb.Client
	profile   string
	region    string
	tableName string
	mu        sync.RWMutex
}

// AWSOption configures AWSStorage.
type AWSOption func(*AWSStorage)

// WithProfile sets the AWS profile.
func WithProfile(profile string) AWSOption {
	return func(s *AWSStorage) { s.profile = profile }
}

// WithRegion sets the AWS region.
func WithRegion(region string) AWSOption {
	return func(s *AWSStorage) { s.region = region }
}

// WithTable sets the DynamoDB table name.
func WithTable(table string) AWSOption {
	return func(s *AWSStorage) { s.tableName = table }
}

// NewAWSStorage creates a new DynamoDB-backed storage.
func NewAWSStorage(ctx context.Context, opts ...AWSOption) (*AWSStorage, error) {
	s := &AWSStorage{
		profile:   "cr8",
		region:    "us-east-1",
		tableName: "cr8_sync_state",
	}
	for _, opt := range opts {
		opt(s)
	}

	configOpts := []func(*config.LoadOptions) error{config.WithRegion(s.region)}
	if s.profile != "" {
		configOpts = append(configOpts, config.WithSharedConfigProfile(s.profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, configOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	s.client = dynamodb.NewFromConfig(cfg)
	return s, nil
}

// SyncStateRecord represents a sync state record in DynamoDB.
type SyncStateRecord struct {
	ID          string    `dynamodbav:"id"`
	Service     string    `dynamodbav:"service"`
	User        string    `dynamodbav:"user"`
	Playlist    string    `dynamodbav:"playlist"`
	LastSync    time.Time `dynamodbav:"last_sync"`
	TrackCount  int       `dynamodbav:"track_count"`
	Synced      int       `dynamodbav:"synced"`
	Failed      int       `dynamodbav:"failed"`
	LastTrackID string    `dynamodbav:"last_track_id,omitempty"`
	Errors      []string  `dynamodbav:"errors,omitempty"`
	UpdatedAt   time.Time `dynamodbav:"updated_at"`
}

// PlaylistRecord represents a playlist in DynamoDB.
type PlaylistRecord struct {
	ID               string     `dynamodbav:"id"`
	Service          string     `dynamodbav:"service"`
	Username         string     `dynamodbav:"username"`
	URL              string     `dynamodbav:"url"`
	GDrivePath       string     `dynamodbav:"gdrive_path,omitempty"`
	Name             string     `dynamodbav:"name,omitempty"`
	Description      string     `dynamodbav:"description,omitempty"`
	TrackCount       int        `dynamodbav:"track_count"`
	LastSyncedAt     *time.Time `dynamodbav:"last_synced_at,omitempty"`
	CreatedAt        time.Time  `dynamodbav:"created_at"`
	UpdatedAt        time.Time  `dynamodbav:"updated_at"`
	IndexedAt        *time.Time `dynamodbav:"indexed_at,omitempty"`
	IndexStatus      string     `dynamodbav:"index_status"`
	TracksDiscovered int        `dynamodbav:"tracks_discovered"`
	TracksQueued     int        `dynamodbav:"tracks_queued"`
	TracksCompleted  int        `dynamodbav:"tracks_completed"`
	TracksFailed     int        `dynamodbav:"tracks_failed"`
	UserID           string     `dynamodbav:"user_id,omitempty"`
}

// GetSyncState retrieves sync state for a service/user/playlist.
func (s *AWSStorage) GetSyncState(ctx context.Context, service, user, playlist string) (*SyncStateRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id := fmt.Sprintf("%s#%s#%s", service, user, playlist)
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("cr8_sync_state"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get sync state: %w", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var record SyncStateRecord
	if err := attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sync state: %w", err)
	}
	return &record, nil
}

// UpdateSyncState updates or creates sync state.
func (s *AWSStorage) UpdateSyncState(ctx context.Context, record *SyncStateRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	record.ID = fmt.Sprintf("%s#%s#%s", record.Service, record.User, record.Playlist)
	record.UpdatedAt = time.Now()

	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return fmt.Errorf("failed to marshal sync state: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("cr8_sync_state"),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put sync state: %w", err)
	}
	return nil
}

// ListSyncStates lists all sync states for a service and user.
func (s *AWSStorage) ListSyncStates(ctx context.Context, service, user string) ([]*SyncStateRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prefix := fmt.Sprintf("%s#%s#", service, user)
	result, err := s.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String("cr8_sync_state"),
		FilterExpression: aws.String("begins_with(id, :prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":prefix": &types.AttributeValueMemberS{Value: prefix},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan sync states: %w", err)
	}

	var records []*SyncStateRecord
	for _, item := range result.Items {
		var record SyncStateRecord
		if err := attributevalue.UnmarshalMap(item, &record); err != nil {
			continue
		}
		records = append(records, &record)
	}
	return records, nil
}

// GetPlaylist retrieves a playlist by ID.
func (s *AWSStorage) GetPlaylist(ctx context.Context, id string) (*PlaylistRecord, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("cr8_playlists"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist: %w", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var record PlaylistRecord
	if err := attributevalue.UnmarshalMap(result.Item, &record); err != nil {
		return nil, fmt.Errorf("failed to unmarshal playlist: %w", err)
	}
	return &record, nil
}

// UpdatePlaylist creates or updates a playlist.
func (s *AWSStorage) UpdatePlaylist(ctx context.Context, record *PlaylistRecord) error {
	record.UpdatedAt = time.Now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}

	item, err := attributevalue.MarshalMap(record)
	if err != nil {
		return fmt.Errorf("failed to marshal playlist: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("cr8_playlists"),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put playlist: %w", err)
	}
	return nil
}

// ListPlaylists lists all playlists, optionally filtered by user.
func (s *AWSStorage) ListPlaylists(ctx context.Context, userID string) ([]*PlaylistRecord, error) {
	var input *dynamodb.ScanInput
	if userID != "" {
		input = &dynamodb.ScanInput{
			TableName:        aws.String("cr8_playlists"),
			FilterExpression: aws.String("user_id = :uid OR username = :uid"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":uid": &types.AttributeValueMemberS{Value: userID},
			},
		}
	} else {
		input = &dynamodb.ScanInput{TableName: aws.String("cr8_playlists")}
	}

	result, err := s.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan playlists: %w", err)
	}

	var records []*PlaylistRecord
	for _, item := range result.Items {
		var record PlaylistRecord
		if err := attributevalue.UnmarshalMap(item, &record); err != nil {
			continue
		}
		records = append(records, &record)
	}
	return records, nil
}

// QueueItem represents an item in the sync queue.
type QueueItem struct {
	ID          string     `dynamodbav:"id"`
	PlaylistID  string     `dynamodbav:"playlist_id"`
	TrackURL    string     `dynamodbav:"track_url"`
	Status      string     `dynamodbav:"status"` // pending, processing, completed, failed
	Priority    int        `dynamodbav:"priority"`
	RetryCount  int        `dynamodbav:"retry_count"`
	Error       string     `dynamodbav:"error,omitempty"`
	CreatedAt   time.Time  `dynamodbav:"created_at"`
	UpdatedAt   time.Time  `dynamodbav:"updated_at"`
	CompletedAt *time.Time `dynamodbav:"completed_at,omitempty"`
}

// EnqueueTrack adds a track to the sync queue.
func (s *AWSStorage) EnqueueTrack(ctx context.Context, item *QueueItem) error {
	item.UpdatedAt = time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}
	if item.Status == "" {
		item.Status = "pending"
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal queue item: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("cr8_sync_queue"),
		Item:      av,
	})
	return err
}

// DequeueTrack gets the next pending track from the queue.
func (s *AWSStorage) DequeueTrack(ctx context.Context) (*QueueItem, error) {
	result, err := s.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String("cr8_sync_queue"),
		IndexName:              aws.String("status-index"),
		KeyConditionExpression: aws.String("#status = :pending"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pending": &types.AttributeValueMemberS{Value: "pending"},
		},
		Limit:            aws.Int32(1),
		ScanIndexForward: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query queue: %w", err)
	}
	if len(result.Items) == 0 {
		return nil, nil
	}

	var item QueueItem
	if err := attributevalue.UnmarshalMap(result.Items[0], &item); err != nil {
		return nil, fmt.Errorf("failed to unmarshal queue item: %w", err)
	}

	// Mark as processing
	item.Status = "processing"
	item.UpdatedAt = time.Now()
	if err := s.updateQueueItem(ctx, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

// CompleteTrack marks a track as completed.
func (s *AWSStorage) CompleteTrack(ctx context.Context, id string, err error) error {
	now := time.Now()
	status := "completed"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
	}

	_, updateErr := s.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String("cr8_sync_queue"),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET #status = :status, updated_at = :now, completed_at = :now, #error = :err"),
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
			"#error":  "error",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: status},
			":now":    &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			":err":    &types.AttributeValueMemberS{Value: errMsg},
		},
	})
	return updateErr
}

func (s *AWSStorage) updateQueueItem(ctx context.Context, item *QueueItem) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}
	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("cr8_sync_queue"),
		Item:      av,
	})
	return err
}

// GetQueueStats returns queue statistics.
func (s *AWSStorage) GetQueueStats(ctx context.Context) (map[string]int, error) {
	stats := map[string]int{"pending": 0, "processing": 0, "completed": 0, "failed": 0}

	for status := range stats {
		result, err := s.client.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String("cr8_sync_queue"),
			IndexName:              aws.String("status-index"),
			KeyConditionExpression: aws.String("#status = :status"),
			ExpressionAttributeNames: map[string]string{
				"#status": "status",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":status": &types.AttributeValueMemberS{Value: status},
			},
			Select: types.SelectCount,
		})
		if err != nil {
			continue
		}
		stats[status] = int(result.Count)
	}
	return stats, nil
}

// RecordSyncHistory saves a sync operation to history.
func (s *AWSStorage) RecordSyncHistory(ctx context.Context, playlistID, status string, details map[string]interface{}) error {
	id := fmt.Sprintf("%s#%d", playlistID, time.Now().UnixNano())
	detailsJSON, _ := json.Marshal(details)

	_, err := s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("cr8_sync_history"),
		Item: map[string]types.AttributeValue{
			"id":          &types.AttributeValueMemberS{Value: id},
			"playlist_id": &types.AttributeValueMemberS{Value: playlistID},
			"status":      &types.AttributeValueMemberS{Value: status},
			"details":     &types.AttributeValueMemberS{Value: string(detailsJSON)},
			"created_at":  &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
		},
	})
	return err
}

// Health checks DynamoDB connectivity.
func (s *AWSStorage) Health(ctx context.Context) error {
	_, err := s.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String("cr8_playlists"),
	})
	return err
}
