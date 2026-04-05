package beatport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var (
	beatportClient *clients.BeatportClient
	dynamoClient   *dynamodb.Client
	tracksTable    = "cr8_tracks"
)

// getBeatportClient returns a singleton Beatport client
func getBeatportClient() (*clients.BeatportClient, error) {
	if beatportClient != nil {
		return beatportClient, nil
	}

	var err error
	beatportClient, err = clients.NewBeatportClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Beatport client: %w", err)
	}

	return beatportClient, nil
}

// getDynamoClient returns a singleton DynamoDB client
func getDynamoClient(ctx context.Context) (*dynamodb.Client, error) {
	if dynamoClient != nil {
		return dynamoClient, nil
	}

	profile := config.GetEnv("AWS_PROFILE", "cr8")

	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithSharedConfigProfile(profile),
		awsconfig.WithRegion("us-east-1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	return dynamoClient, nil
}

// handleSearch handles the beatport_search tool
func handleSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(request, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(request, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	client, err := getBeatportClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	tracks, err := client.SearchTracks(ctx, query, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("search failed: %w", err)), nil
	}

	if len(tracks) == 0 {
		return tools.TextResult(fmt.Sprintf("No results found for: %s", query)), nil
	}

	// Format results
	results := make([]map[string]interface{}, len(tracks))
	for i, track := range tracks {
		result := map[string]interface{}{
			"id":       track.ID,
			"name":     track.Name,
			"mix_name": track.MixName,
			"bpm":      track.BPM,
			"url":      track.URL,
		}

		// Add artists
		if len(track.Artists) > 0 {
			artists := make([]string, len(track.Artists))
			for j, a := range track.Artists {
				artists[j] = a.Name
			}
			result["artists"] = artists
		}

		if track.Key != nil {
			result["key"] = track.Key.Name
			if track.Key.Camelot != nil {
				result["camelot_key"] = track.Key.Camelot.Key
			}
		}

		if track.Genre != nil {
			result["genre"] = track.Genre.Name
		}

		if track.SubGenre != nil {
			result["sub_genre"] = track.SubGenre.Name
		}

		if track.Label != nil {
			result["label"] = track.Label.Name
		}

		results[i] = result
	}

	return tools.JSONResult(map[string]interface{}{
		"query":   query,
		"count":   len(results),
		"results": results,
	}), nil
}

// handleGetTrack handles the beatport_track tool
func handleGetTrack(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	trackID := tools.GetIntParam(request, "track_id", 0)
	if trackID == 0 {
		return tools.ErrorResult(fmt.Errorf("track_id parameter is required")), nil
	}

	client, err := getBeatportClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	track, err := client.GetTrack(ctx, trackID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get track: %w", err)), nil
	}

	// Convert to enrichment data format for consistent output
	data := clients.TrackToEnrichmentData(track)
	data["name"] = track.Name
	data["mix_name"] = track.MixName

	if len(track.Artists) > 0 {
		artists := make([]string, len(track.Artists))
		for i, a := range track.Artists {
			artists[i] = a.Name
		}
		data["artists"] = artists
	}

	if track.LengthMS > 0 {
		data["duration_ms"] = track.LengthMS
		data["duration"] = fmt.Sprintf("%d:%02d", track.LengthMS/60000, (track.LengthMS/1000)%60)
	}

	return tools.JSONResult(data), nil
}

// handleCharts handles the beatport_charts tool
func handleCharts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	genre := tools.GetStringParam(request, "genre")

	client, err := getBeatportClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	charts, err := client.GetCharts(ctx, genre)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get charts: %w", err)), nil
	}

	results := make([]map[string]interface{}, len(charts))
	for i, chart := range charts {
		result := map[string]interface{}{
			"id":   chart.ID,
			"name": chart.Name,
			"url":  chart.URL,
		}
		if chart.Genre != nil {
			result["genre"] = chart.Genre.Name
		}
		if chart.Description != "" {
			result["description"] = chart.Description
		}
		results[i] = result
	}

	return tools.JSONResult(map[string]interface{}{
		"genre":  genre,
		"count":  len(results),
		"charts": results,
	}), nil
}

// handleChartTracks handles the beatport_chart_tracks tool
func handleChartTracks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	chartID := tools.GetIntParam(request, "chart_id", 0)
	if chartID == 0 {
		return tools.ErrorResult(fmt.Errorf("chart_id parameter is required")), nil
	}

	client, err := getBeatportClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	tracks, err := client.GetChartTracks(ctx, chartID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get chart tracks: %w", err)), nil
	}

	results := make([]map[string]interface{}, len(tracks))
	for i, track := range tracks {
		result := map[string]interface{}{
			"position": i + 1,
			"id":       track.ID,
			"name":     track.Name,
			"mix_name": track.MixName,
			"bpm":      track.BPM,
		}

		if len(track.Artists) > 0 {
			artists := make([]string, len(track.Artists))
			for j, a := range track.Artists {
				artists[j] = a.Name
			}
			result["artists"] = artists
		}

		if track.Key != nil {
			result["key"] = track.Key.Name
		}

		if track.Label != nil {
			result["label"] = track.Label.Name
		}

		results[i] = result
	}

	return tools.JSONResult(map[string]interface{}{
		"chart_id": chartID,
		"count":    len(results),
		"tracks":   results,
	}), nil
}

// handleGenres handles the beatport_genres tool
func handleGenres(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getBeatportClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	genres, err := client.GetGenres(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get genres: %w", err)), nil
	}

	results := make([]map[string]interface{}, len(genres))
	for i, genre := range genres {
		results[i] = map[string]interface{}{
			"id":   genre.ID,
			"name": genre.Name,
			"slug": genre.Slug,
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"count":  len(results),
		"genres": results,
	}), nil
}

// handleEnrich handles the beatport_enrich tool
func handleEnrich(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	trackID, errResult := tools.RequireStringParam(request, "track_id")
	if errResult != nil {
		return errResult, nil
	}

	dryRun := tools.GetBoolParam(request, "dry_run", true)

	// Get track from DynamoDB
	dynamo, err := getDynamoClient(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := dynamo.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tracksTable),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: trackID},
		},
	})
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get track: %w", err)), nil
	}

	if result.Item == nil {
		return tools.ErrorResult(fmt.Errorf("track not found: %s", trackID)), nil
	}

	// Extract artist and title
	var cr8Track struct {
		ID     string `dynamodbav:"id"`
		Artist string `dynamodbav:"artist"`
		Title  string `dynamodbav:"title"`
	}
	if err := attributevalue.UnmarshalMap(result.Item, &cr8Track); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to unmarshal track: %w", err)), nil
	}

	// Search Beatport
	client, err := getBeatportClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	query := fmt.Sprintf("%s %s", cr8Track.Artist, cr8Track.Title)
	tracks, err := client.SearchTracks(ctx, query, 5)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("beatport search failed: %w", err)), nil
	}

	if len(tracks) == 0 {
		return tools.JSONResult(map[string]interface{}{
			"track_id": trackID,
			"status":   "no_match",
			"query":    query,
		}), nil
	}

	// Use best match
	enrichment := clients.TrackToEnrichmentData(&tracks[0])

	if !dryRun {
		// Update DynamoDB
		if err := updateTrackMetadata(ctx, dynamo, trackID, enrichment); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to update track: %w", err)), nil
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"track_id":   trackID,
		"dry_run":    dryRun,
		"status":     "enriched",
		"query":      query,
		"match":      tracks[0].Name,
		"match_id":   tracks[0].ID,
		"enrichment": enrichment,
	}), nil
}

// handleBatchEnrich handles the beatport_batch_enrich tool
func handleBatchEnrich(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(request, "limit", 10)
	dryRun := tools.GetBoolParam(request, "dry_run", true)
	filter := tools.GetStringParam(request, "filter")
	if filter == "" {
		filter = "missing_any"
	}

	dynamo, err := getDynamoClient(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Build filter expression based on filter type
	var filterExpr string
	switch filter {
	case "missing_bpm":
		filterExpr = "attribute_not_exists(bpm)"
	case "missing_key":
		filterExpr = "attribute_not_exists(musical_key)"
	default: // missing_any
		filterExpr = "attribute_not_exists(bpm) OR attribute_not_exists(musical_key)"
	}

	// Scan for tracks needing enrichment
	scanResult, err := dynamo.Scan(ctx, &dynamodb.ScanInput{
		TableName:            aws.String(tracksTable),
		FilterExpression:     aws.String(filterExpr),
		Limit:                aws.Int32(int32(limit)),
		ProjectionExpression: aws.String("id, artist, title"),
	})
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to scan tracks: %w", err)), nil
	}

	if len(scanResult.Items) == 0 {
		return tools.JSONResult(map[string]interface{}{
			"status":  "complete",
			"message": "No tracks found needing enrichment",
			"filter":  filter,
		}), nil
	}

	client, err := getBeatportClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	results := []map[string]interface{}{}
	enriched := 0
	noMatch := 0
	errors := 0

	for _, item := range scanResult.Items {
		var track struct {
			ID     string `dynamodbav:"id"`
			Artist string `dynamodbav:"artist"`
			Title  string `dynamodbav:"title"`
		}
		if err := attributevalue.UnmarshalMap(item, &track); err != nil {
			errors++
			continue
		}

		query := fmt.Sprintf("%s %s", track.Artist, track.Title)
		beatportTracks, err := client.SearchTracks(ctx, query, 3)
		if err != nil {
			results = append(results, map[string]interface{}{
				"track_id": track.ID,
				"status":   "error",
				"error":    err.Error(),
			})
			errors++
			continue
		}

		if len(beatportTracks) == 0 {
			results = append(results, map[string]interface{}{
				"track_id": track.ID,
				"status":   "no_match",
				"query":    query,
			})
			noMatch++
			continue
		}

		enrichment := clients.TrackToEnrichmentData(&beatportTracks[0])

		if !dryRun {
			if err := updateTrackMetadata(ctx, dynamo, track.ID, enrichment); err != nil {
				results = append(results, map[string]interface{}{
					"track_id": track.ID,
					"status":   "error",
					"error":    err.Error(),
				})
				errors++
				continue
			}
		}

		results = append(results, map[string]interface{}{
			"track_id": track.ID,
			"status":   "enriched",
			"match":    beatportTracks[0].Name,
			"bpm":      enrichment["bpm"],
			"key":      enrichment["key"],
		})
		enriched++

		// Rate limit: 1 request per second
		time.Sleep(time.Second)
	}

	return tools.JSONResult(map[string]interface{}{
		"dry_run":   dryRun,
		"filter":    filter,
		"processed": len(results),
		"enriched":  enriched,
		"no_match":  noMatch,
		"errors":    errors,
		"results":   results,
	}), nil
}

// handleAuthStatus handles the beatport_auth_status tool
func handleAuthStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getBeatportClient()
	if err != nil {
		return tools.JSONResult(map[string]interface{}{
			"authenticated": false,
			"error":         err.Error(),
		}), nil
	}

	accessToken, refreshToken, expiry := client.GetTokens()

	status := map[string]interface{}{
		"authenticated":     client.IsAuthenticated(),
		"has_access_token":  accessToken != "",
		"has_refresh_token": refreshToken != "",
	}

	if !expiry.IsZero() {
		status["token_expires"] = expiry.Format(time.RFC3339)
		status["token_valid_for"] = time.Until(expiry).String()
	}

	// Check if credentials are configured
	status["has_credentials"] = config.Get().BeatportUsername != ""

	return tools.JSONResult(status), nil
}

// updateTrackMetadata updates a track in DynamoDB with enrichment data
func updateTrackMetadata(ctx context.Context, dynamo *dynamodb.Client, trackID string, data map[string]interface{}) error {
	// Build update expression
	updateParts := []string{}
	exprValues := map[string]types.AttributeValue{}
	exprNames := map[string]string{}

	fieldMap := map[string]string{
		"bpm":            "bpm",
		"key":            "musical_key", // 'key' is reserved
		"camelot_key":    "camelot_key",
		"genre":          "genre",
		"sub_genre":      "sub_genre",
		"label":          "label",
		"mix_name":       "mix_name",
		"remixer":        "remixer",
		"isrc":           "isrc",
		"release_date":   "release_date",
		"artwork_url":    "artwork_url",
		"beatport_id":    "beatport_id",
		"beatport_url":   "beatport_url",
		"catalog_number": "catalog_number",
	}

	for dataKey, attrName := range fieldMap {
		if val, ok := data[dataKey]; ok && val != nil && val != "" && val != 0 {
			placeholder := ":" + attrName
			namePlaceholder := "#" + attrName

			updateParts = append(updateParts, fmt.Sprintf("%s = %s", namePlaceholder, placeholder))
			exprNames[namePlaceholder] = attrName

			switch v := val.(type) {
			case string:
				exprValues[placeholder] = &types.AttributeValueMemberS{Value: v}
			case int:
				exprValues[placeholder] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", v)}
			case float64:
				exprValues[placeholder] = &types.AttributeValueMemberN{Value: fmt.Sprintf("%g", v)}
			default:
				// Try JSON encoding for complex types
				jsonBytes, _ := json.Marshal(v)
				exprValues[placeholder] = &types.AttributeValueMemberS{Value: string(jsonBytes)}
			}
		}
	}

	if len(updateParts) == 0 {
		return nil // Nothing to update
	}

	// Add updated_at timestamp
	updateParts = append(updateParts, "#updated_at = :updated_at")
	exprNames["#updated_at"] = "updated_at"
	exprValues[":updated_at"] = &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)}

	// Add enrichment source
	updateParts = append(updateParts, "#enrichment_source = :enrichment_source")
	exprNames["#enrichment_source"] = "enrichment_source"
	exprValues[":enrichment_source"] = &types.AttributeValueMemberS{Value: "beatport"}

	updateExpr := "SET " + joinStrings(updateParts, ", ")

	_, err := dynamo.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tracksTable),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: trackID},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprNames,
		ExpressionAttributeValues: exprValues,
	})

	return err
}

// joinStrings joins strings with a separator
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

// handleGetPlaylist handles the beatport_playlist tool
func handleGetPlaylist(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playlistID := tools.GetIntParam(request, "playlist_id", 0)
	if playlistID == 0 {
		return tools.ErrorResult(fmt.Errorf("playlist_id parameter is required")), nil
	}

	includeTracks := tools.GetBoolParam(request, "include_tracks", false)

	client, err := getBeatportClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	playlist, err := client.GetPlaylist(ctx, playlistID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get playlist: %w", err)), nil
	}

	result := map[string]interface{}{
		"id":          playlist.ID,
		"name":        playlist.Name,
		"track_count": playlist.TrackCount,
		"genres":      playlist.Genres,
		"bpm_range":   playlist.BPMRange,
		"created":     playlist.CreatedDate,
		"updated":     playlist.UpdatedDate,
	}

	if includeTracks {
		tracks, err := client.GetAllPlaylistTracks(ctx, playlistID)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to get tracks: %w", err)), nil
		}

		trackList := make([]map[string]interface{}, len(tracks))
		for i, track := range tracks {
			trackList[i] = map[string]interface{}{
				"position": i + 1,
				"id":       track.ID,
				"name":     track.Name,
				"mix_name": track.MixName,
				"bpm":      track.BPM,
			}
			if len(track.Artists) > 0 {
				artists := make([]string, len(track.Artists))
				for j, a := range track.Artists {
					artists[j] = a.Name
				}
				trackList[i]["artists"] = artists
			}
			if track.Key != nil {
				trackList[i]["key"] = track.Key.Name
			}
		}
		result["tracks"] = trackList
	}

	return tools.JSONResult(result), nil
}

// handleDownloadTrack handles the beatport_download tool
func handleDownloadTrack(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	trackID := tools.GetIntParam(request, "track_id", 0)
	if trackID == 0 {
		return tools.ErrorResult(fmt.Errorf("track_id parameter is required")), nil
	}

	quality := tools.GetStringParam(request, "quality")
	if quality == "" {
		quality = clients.QualityFLAC
	}

	format := tools.GetStringParam(request, "format")
	if format == "" {
		format = "aiff"
	}

	uploadS3 := tools.GetBoolParam(request, "upload_s3", true)

	client, err := getBeatportClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Get track metadata first
	track, err := client.GetTrack(ctx, trackID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get track info: %w", err)), nil
	}

	// Build filename
	artistName := "Unknown"
	if len(track.Artists) > 0 {
		artistName = track.Artists[0].Name
	}
	baseName := fmt.Sprintf("%s - %s", artistName, track.Name)
	if track.MixName != "" && track.MixName != "Original Mix" {
		baseName = fmt.Sprintf("%s (%s)", baseName, track.MixName)
	}

	// Create temp directory
	tempDir := os.TempDir()
	flacPath := fmt.Sprintf("%s/%d.flac", tempDir, trackID)

	// Download track
	err = client.DownloadTrackToFile(ctx, trackID, quality, flacPath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("download failed: %w", err)), nil
	}
	defer os.Remove(flacPath)

	// Convert if needed
	outputPath := flacPath
	outputExt := "flac"
	if format != "flac" && (format == "aiff" || format == "wav") {
		outputExt = format
		outputPath = fmt.Sprintf("%s/%d.%s", tempDir, trackID, format)
		err = convertAudio(ctx, flacPath, outputPath, format)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("conversion failed: %w", err)), nil
		}
		defer os.Remove(outputPath)
	}

	result := map[string]interface{}{
		"track_id":   trackID,
		"name":       track.Name,
		"artists":    artistName,
		"format":     outputExt,
		"quality":    quality,
		"local_path": outputPath,
	}

	// Upload to S3 if requested
	if uploadS3 {
		s3Key := fmt.Sprintf("beatport/%s.%s", sanitizeFilename(baseName), outputExt)
		err = uploadToS3(ctx, outputPath, s3Key)
		if err != nil {
			result["s3_error"] = err.Error()
		} else {
			result["s3_path"] = s3Key
			result["uploaded"] = true

			// Record in DynamoDB
			enrichment := clients.TrackToEnrichmentData(track)
			enrichment["storage_path"] = s3Key
			enrichment["service"] = "beatport"
			enrichment["source_url"] = track.URL

			dynamo, err := getDynamoClient(ctx)
			if err == nil {
				trackDBID := fmt.Sprintf("beatport-%d", trackID)
				err = createTrackRecord(ctx, dynamo, trackDBID, track, s3Key)
				if err != nil {
					result["db_error"] = err.Error()
				} else {
					result["db_id"] = trackDBID
				}
			}
		}
	}

	return tools.JSONResult(result), nil
}

// handleSyncPlaylist handles the beatport_sync_playlist tool
func handleSyncPlaylist(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playlistID := tools.GetIntParam(request, "playlist_id", 0)
	if playlistID == 0 {
		return tools.ErrorResult(fmt.Errorf("playlist_id parameter is required")), nil
	}

	quality := tools.GetStringParam(request, "quality")
	if quality == "" {
		quality = clients.QualityFLAC
	}

	format := tools.GetStringParam(request, "format")
	if format == "" {
		format = "aiff"
	}

	dryRun := tools.GetBoolParam(request, "dry_run", true)
	limit := tools.GetIntParam(request, "limit", 0)

	client, err := getBeatportClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Get playlist info
	playlist, err := client.GetPlaylist(ctx, playlistID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get playlist: %w", err)), nil
	}

	// Get tracks
	tracks, err := client.GetAllPlaylistTracks(ctx, playlistID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get playlist tracks: %w", err)), nil
	}

	// Apply limit
	if limit > 0 && limit < len(tracks) {
		tracks = tracks[:limit]
	}

	// Check which tracks already exist in DB
	dynamo, err := getDynamoClient(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	existingTracks := make(map[int]bool)
	for _, track := range tracks {
		trackDBID := fmt.Sprintf("beatport-%d", track.ID)
		result, _ := dynamo.GetItem(ctx, &dynamodb.GetItemInput{
			TableName: aws.String(tracksTable),
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: trackDBID},
			},
			ProjectionExpression: aws.String("id"),
		})
		if result != nil && result.Item != nil {
			existingTracks[track.ID] = true
		}
	}

	newTracks := []clients.BeatportTrack{}
	for _, track := range tracks {
		if !existingTracks[track.ID] {
			newTracks = append(newTracks, track)
		}
	}

	// Build response
	response := map[string]interface{}{
		"playlist_id":     playlistID,
		"playlist_name":   playlist.Name,
		"total_tracks":    len(tracks),
		"existing_tracks": len(existingTracks),
		"new_tracks":      len(newTracks),
		"quality":         quality,
		"format":          format,
		"dry_run":         dryRun,
	}

	// Preview new tracks
	preview := make([]map[string]interface{}, 0)
	for i, track := range newTracks {
		if i >= 10 {
			preview = append(preview, map[string]interface{}{"note": fmt.Sprintf("... and %d more", len(newTracks)-10)})
			break
		}
		artistName := "Unknown"
		if len(track.Artists) > 0 {
			artistName = track.Artists[0].Name
		}
		preview = append(preview, map[string]interface{}{
			"id":     track.ID,
			"artist": artistName,
			"name":   track.Name,
			"bpm":    track.BPM,
		})
	}
	response["preview"] = preview

	if dryRun {
		response["message"] = "Dry run - no tracks downloaded. Set dry_run=false to sync."
		return tools.JSONResult(response), nil
	}

	// Actually sync tracks
	synced := 0
	failed := 0
	results := []map[string]interface{}{}

	tempDir := os.TempDir()

	for _, track := range newTracks {
		artistName := "Unknown"
		if len(track.Artists) > 0 {
			artistName = track.Artists[0].Name
		}
		baseName := fmt.Sprintf("%s - %s", artistName, track.Name)
		if track.MixName != "" && track.MixName != "Original Mix" {
			baseName = fmt.Sprintf("%s (%s)", baseName, track.MixName)
		}

		flacPath := fmt.Sprintf("%s/%d.flac", tempDir, track.ID)

		// Download
		err := client.DownloadTrackToFile(ctx, track.ID, quality, flacPath)
		if err != nil {
			results = append(results, map[string]interface{}{
				"id":     track.ID,
				"status": "failed",
				"error":  err.Error(),
			})
			failed++
			continue
		}

		// Convert if needed
		outputPath := flacPath
		outputExt := "flac"
		if format != "flac" {
			outputExt = format
			outputPath = fmt.Sprintf("%s/%d.%s", tempDir, track.ID, format)
			if err := convertAudio(ctx, flacPath, outputPath, format); err != nil {
				os.Remove(flacPath)
				results = append(results, map[string]interface{}{
					"id":     track.ID,
					"status": "failed",
					"error":  fmt.Sprintf("conversion: %v", err),
				})
				failed++
				continue
			}
			os.Remove(flacPath)
		}

		// Upload to S3
		s3Key := fmt.Sprintf("beatport/%s.%s", sanitizeFilename(baseName), outputExt)
		if err := uploadToS3(ctx, outputPath, s3Key); err != nil {
			os.Remove(outputPath)
			results = append(results, map[string]interface{}{
				"id":     track.ID,
				"status": "failed",
				"error":  fmt.Sprintf("s3 upload: %v", err),
			})
			failed++
			continue
		}
		os.Remove(outputPath)

		// Record in DynamoDB
		trackDBID := fmt.Sprintf("beatport-%d", track.ID)
		if err := createTrackRecord(ctx, dynamo, trackDBID, &track, s3Key); err != nil {
			results = append(results, map[string]interface{}{
				"id":     track.ID,
				"status": "partial",
				"s3":     s3Key,
				"error":  fmt.Sprintf("db: %v", err),
			})
		} else {
			results = append(results, map[string]interface{}{
				"id":     track.ID,
				"status": "synced",
				"s3":     s3Key,
				"db_id":  trackDBID,
			})
		}
		synced++

		// Rate limit
		time.Sleep(2 * time.Second)
	}

	response["synced"] = synced
	response["failed"] = failed
	response["results"] = results

	return tools.JSONResult(response), nil
}

// convertAudio converts audio from FLAC to another format using ffmpeg
func convertAudio(ctx context.Context, inputPath, outputPath, format string) error {
	var args []string
	switch format {
	case "aiff":
		args = []string{"-i", inputPath, "-c:a", "pcm_s16be", "-y", outputPath}
	case "wav":
		args = []string{"-i", inputPath, "-c:a", "pcm_s16le", "-y", outputPath}
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v - %s", err, string(output))
	}
	return nil
}

// uploadToS3 uploads a file to S3
func uploadToS3(ctx context.Context, localPath, s3Key string) error {
	profile := config.GetEnv("AWS_PROFILE", "cr8")

	bucket := config.GetEnv("CR8_S3_BUCKET", "cr8-music-storage")

	cmd := exec.CommandContext(ctx, "aws", "s3", "cp", localPath,
		fmt.Sprintf("s3://%s/%s", bucket, s3Key),
		"--profile", profile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("aws s3 cp error: %v - %s", err, string(output))
	}
	return nil
}

// createTrackRecord creates a track record in DynamoDB
func createTrackRecord(ctx context.Context, dynamo *dynamodb.Client, trackID string, track *clients.BeatportTrack, s3Key string) error {
	artistName := "Unknown"
	if len(track.Artists) > 0 {
		artistName = track.Artists[0].Name
	}

	item := map[string]types.AttributeValue{
		"id":           &types.AttributeValueMemberS{Value: trackID},
		"title":        &types.AttributeValueMemberS{Value: track.Name},
		"artist":       &types.AttributeValueMemberS{Value: artistName},
		"service":      &types.AttributeValueMemberS{Value: "beatport"},
		"source_url":   &types.AttributeValueMemberS{Value: track.URL},
		"storage_path": &types.AttributeValueMemberS{Value: s3Key},
		"beatport_id":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", track.ID)},
		"bpm":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%g", track.BPM)},
		"created_at":   &types.AttributeValueMemberS{Value: time.Now().UTC().Format(time.RFC3339)},
	}

	if track.Key != nil {
		item["musical_key"] = &types.AttributeValueMemberS{Value: track.Key.Name}
		if track.Key.Camelot != nil {
			item["camelot_key"] = &types.AttributeValueMemberS{Value: track.Key.Camelot.Key}
		}
	}

	if track.Genre != nil {
		item["genre"] = &types.AttributeValueMemberS{Value: track.Genre.Name}
	}

	if track.SubGenre != nil {
		item["sub_genre"] = &types.AttributeValueMemberS{Value: track.SubGenre.Name}
	}

	if track.Label != nil {
		item["label"] = &types.AttributeValueMemberS{Value: track.Label.Name}
	}

	if track.MixName != "" {
		item["mix_name"] = &types.AttributeValueMemberS{Value: track.MixName}
	}

	if track.ISRC != "" {
		item["isrc"] = &types.AttributeValueMemberS{Value: track.ISRC}
	}

	_, err := dynamo.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tracksTable),
		Item:      item,
	})
	return err
}

// sanitizeFilename removes/replaces characters not safe for filenames
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "'",
		"<", "",
		">", "",
		"|", "-",
	)
	return replacer.Replace(name)
}

// handleSyncLikes handles the beatport_sync_likes tool
// Syncs likes and follows for all tracks/artists in a playlist using Playwright automation
func handleSyncLikes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playlistID, errResult := tools.RequireStringParam(request, "playlist_id")
	if errResult != nil {
		return errResult, nil
	}

	mode := tools.GetStringParam(request, "mode")
	if mode == "" {
		mode = "both"
	}

	// Validate mode
	validModes := map[string]bool{"both": true, "tracks": true, "artists": true, "status": true}
	if !validModes[mode] {
		return tools.ErrorResult(fmt.Errorf("invalid mode: %s (valid: both, tracks, artists, status)", mode)), nil
	}

	// Path to automation script
	scriptPath := "tools/beatport-automation/sync-likes.sh"

	// For status check, run synchronously
	if mode == "status" {
		cmd := exec.CommandContext(ctx, "node", "index.js", "status", "--playlist="+playlistID)
		cmd.Dir = "tools/beatport-automation"
		cmd.Env = append(os.Environ(), "PLAYWRIGHT_BROWSERS_PATH=./node_modules/playwright-core/.local-browsers")

		output, err := cmd.CombinedOutput()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("status check failed: %w", err)), nil
		}

		return tools.JSONResult(map[string]interface{}{
			"playlist_id": playlistID,
			"mode":        "status",
			"output":      string(output),
		}), nil
	}

	// Convert mode to script flag
	modeFlag := ""
	switch mode {
	case "tracks":
		modeFlag = "--tracks-only"
	case "artists":
		modeFlag = "--artists-only"
	}

	// Run sync in background
	logFile := fmt.Sprintf("/tmp/beatport-likes-%s.log", playlistID)
	cmdArgs := []string{scriptPath, playlistID}
	if modeFlag != "" {
		cmdArgs = append(cmdArgs, modeFlag)
	}

	cmd := exec.Command("bash", cmdArgs...)
	bpCfg := config.Get()
	cmd.Env = append(os.Environ(),
		"PLAYWRIGHT_BROWSERS_PATH=./node_modules/playwright-core/.local-browsers",
		fmt.Sprintf("BEATPORT_USERNAME=%s", bpCfg.BeatportUsername),
		fmt.Sprintf("BEATPORT_PASSWORD=%s", bpCfg.BeatportPassword),
	)

	// Redirect output to log file
	logHandle, err := os.Create(logFile)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create log file: %w", err)), nil
	}
	cmd.Stdout = logHandle
	cmd.Stderr = logHandle

	// Start background process
	if err := cmd.Start(); err != nil {
		logHandle.Close()
		return tools.ErrorResult(fmt.Errorf("failed to start sync: %w", err)), nil
	}

	// Detach from parent process
	go func() {
		cmd.Wait()
		logHandle.Close()
	}()

	return tools.JSONResult(map[string]interface{}{
		"status":      "started",
		"playlist_id": playlistID,
		"mode":        mode,
		"pid":         cmd.Process.Pid,
		"log_file":    logFile,
		"message":     fmt.Sprintf("Sync started in background. Monitor with: tail -f %s", logFile),
	}), nil
}
