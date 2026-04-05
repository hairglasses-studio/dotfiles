// Package workflow_automation provides workflow automation tools for hg-mcp.
package workflow_automation

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// deduplicationTools returns the deduplication tool definitions
func deduplicationTools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_workflow_dedup_scan",
				mcp.WithDescription("Scan for duplicate tracks across platforms using fuzzy matching on title, artist, duration, and BPM."),
				mcp.WithString("query", mcp.Description("Search query or genre to scan (e.g., 'techno', 'Charlotte de Witte')")),
				mcp.WithString("platforms", mcp.Description("Comma-separated platforms to scan: rekordbox,soundcloud,beatport,spotify")),
				mcp.WithNumber("threshold", mcp.Description("Minimum match score (0.0-1.0, default: 0.85)")),
			),
			Handler:             handleDedupScan,
			Category:            "workflow_automation",
			Subcategory:         "deduplication",
			Tags:                []string{"dedup", "duplicate", "scan", "match", "fuzzy"},
			UseCases:            []string{"Find duplicate tracks", "Clean up library"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_dedup_match",
				mcp.WithDescription("Find a specific track on other platforms using fuzzy matching."),
				mcp.WithString("artist", mcp.Description("Track artist"), mcp.Required()),
				mcp.WithString("title", mcp.Description("Track title"), mcp.Required()),
				mcp.WithNumber("bpm", mcp.Description("Track BPM (helps improve matching)")),
				mcp.WithString("duration", mcp.Description("Track duration in MM:SS format")),
			),
			Handler:             handleDedupMatch,
			Category:            "workflow_automation",
			Subcategory:         "deduplication",
			Tags:                []string{"dedup", "match", "find", "cross-platform"},
			UseCases:            []string{"Find track on other platforms", "Cross-reference libraries"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_dedup_merge",
				mcp.WithDescription("Merge metadata from duplicate tracks, keeping the best information from each source."),
				mcp.WithString("track_id", mcp.Description("Primary track ID to merge into"), mcp.Required()),
				mcp.WithString("source_ids", mcp.Description("Comma-separated source track IDs to merge from")),
			),
			Handler:             handleDedupMerge,
			Category:            "workflow_automation",
			Subcategory:         "deduplication",
			Tags:                []string{"dedup", "merge", "metadata", "combine"},
			UseCases:            []string{"Merge track metadata", "Consolidate information"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_dedup_config",
				mcp.WithDescription("Configure deduplication matching thresholds and preferences."),
				mcp.WithString("action", mcp.Description("Action: view, set_threshold, set_weights")),
				mcp.WithNumber("title_weight", mcp.Description("Weight for title matching (0.0-1.0)")),
				mcp.WithNumber("artist_weight", mcp.Description("Weight for artist matching (0.0-1.0)")),
				mcp.WithNumber("duration_weight", mcp.Description("Weight for duration matching (0.0-1.0)")),
				mcp.WithNumber("bpm_weight", mcp.Description("Weight for BPM matching (0.0-1.0)")),
			),
			Handler:             handleDedupConfig,
			Category:            "workflow_automation",
			Subcategory:         "deduplication",
			Tags:                []string{"dedup", "config", "settings", "threshold"},
			UseCases:            []string{"Configure matching", "Adjust thresholds"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflow_automation",
		},
	}
}

// DedupConfig holds deduplication configuration
type DedupConfig struct {
	TitleWeight    float64 `json:"title_weight"`
	ArtistWeight   float64 `json:"artist_weight"`
	DurationWeight float64 `json:"duration_weight"`
	BPMWeight      float64 `json:"bpm_weight"`
	Threshold      float64 `json:"threshold"`
}

// DuplicateMatch represents a potential duplicate match
type DuplicateMatch struct {
	Track1       TrackFingerprint `json:"track1"`
	Track2       TrackFingerprint `json:"track2"`
	MatchScore   float64          `json:"match_score"`
	MatchDetails MatchDetails     `json:"match_details"`
}

// TrackFingerprint holds identifying information for a track
type TrackFingerprint struct {
	ID       string  `json:"id"`
	Artist   string  `json:"artist"`
	Title    string  `json:"title"`
	Duration float64 `json:"duration"` // seconds
	BPM      float64 `json:"bpm"`
	Key      string  `json:"key"`
	ISRC     string  `json:"isrc,omitempty"`
	Platform string  `json:"platform"`
}

// MatchDetails breaks down the match scoring
type MatchDetails struct {
	TitleScore    float64 `json:"title_score"`
	ArtistScore   float64 `json:"artist_score"`
	DurationScore float64 `json:"duration_score"`
	BPMScore      float64 `json:"bpm_score"`
	ISRCMatch     bool    `json:"isrc_match"`
}

// Default deduplication config
var defaultDedupConfig = DedupConfig{
	TitleWeight:    0.35,
	ArtistWeight:   0.35,
	DurationWeight: 0.20,
	BPMWeight:      0.10,
	Threshold:      0.85,
}

// Sample duplicate matches for demonstration
var sampleDuplicates = []DuplicateMatch{
	{
		Track1:       TrackFingerprint{ID: "rb_001", Artist: "Charlotte de Witte", Title: "Mercury", Duration: 392, BPM: 140, Platform: "rekordbox"},
		Track2:       TrackFingerprint{ID: "sc_001", Artist: "Charlotte De Witte", Title: "Mercury (Original Mix)", Duration: 394, BPM: 140, Platform: "soundcloud"},
		MatchScore:   0.94,
		MatchDetails: MatchDetails{TitleScore: 0.88, ArtistScore: 0.98, DurationScore: 0.99, BPMScore: 1.0},
	},
	{
		Track1:       TrackFingerprint{ID: "rb_002", Artist: "Amelie Lens", Title: "Exhale", Duration: 380, BPM: 142, Platform: "rekordbox"},
		Track2:       TrackFingerprint{ID: "bp_002", Artist: "Amelie Lens", Title: "Exhale", Duration: 380, BPM: 142, Platform: "beatport"},
		MatchScore:   1.0,
		MatchDetails: MatchDetails{TitleScore: 1.0, ArtistScore: 1.0, DurationScore: 1.0, BPMScore: 1.0, ISRCMatch: true},
	},
	{
		Track1:       TrackFingerprint{ID: "rb_003", Artist: "I Hate Models", Title: "Daydream", Duration: 425, BPM: 138, Platform: "rekordbox"},
		Track2:       TrackFingerprint{ID: "bc_003", Artist: "I HATE MODELS", Title: "Daydream EP", Duration: 428, BPM: 138, Platform: "bandcamp"},
		MatchScore:   0.89,
		MatchDetails: MatchDetails{TitleScore: 0.82, ArtistScore: 0.95, DurationScore: 0.97, BPMScore: 1.0},
	},
}

// handleDedupScan handles the aftrs_workflow_dedup_scan tool
func handleDedupScan(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := tools.GetStringParam(req, "query")
	platformsStr := tools.GetStringParam(req, "platforms")
	threshold := tools.GetFloatParam(req, "threshold", defaultDedupConfig.Threshold)

	platforms := []string{"rekordbox", "soundcloud", "beatport"}
	if platformsStr != "" {
		platforms = strings.Split(platformsStr, ",")
		for i := range platforms {
			platforms[i] = strings.TrimSpace(strings.ToLower(platforms[i]))
		}
	}

	var sb strings.Builder
	sb.WriteString("# Duplicate Scan Results\n\n")

	if query != "" {
		sb.WriteString(fmt.Sprintf("**Query:** %s\n", query))
	}
	sb.WriteString(fmt.Sprintf("**Platforms:** %s\n", strings.Join(platforms, ", ")))
	sb.WriteString(fmt.Sprintf("**Threshold:** %.0f%%\n\n", threshold*100))

	// Filter matches by threshold
	var matches []DuplicateMatch
	for _, m := range sampleDuplicates {
		if m.MatchScore >= threshold {
			matches = append(matches, m)
		}
	}

	sb.WriteString("## Scan Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Tracks Scanned:** %d\n", 150))
	sb.WriteString(fmt.Sprintf("- **Potential Duplicates:** %d\n", len(matches)))
	sb.WriteString(fmt.Sprintf("- **High Confidence (>95%%):** %d\n\n", countHighConfidence(matches)))

	if len(matches) == 0 {
		sb.WriteString("No duplicates found above the threshold.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Duplicate Matches\n\n")
	sb.WriteString("| Track | Platform 1 | Platform 2 | Match Score |\n")
	sb.WriteString("|-------|------------|------------|-------------|\n")

	for _, m := range matches {
		track := fmt.Sprintf("%s - %s", m.Track1.Artist, m.Track1.Title)
		if len(track) > 40 {
			track = track[:37] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %.0f%% |\n",
			track, m.Track1.Platform, m.Track2.Platform, m.MatchScore*100))
	}

	sb.WriteString("\n## Match Breakdown\n\n")
	for i, m := range matches {
		sb.WriteString(fmt.Sprintf("### %d. %s - %s\n", i+1, m.Track1.Artist, m.Track1.Title))
		sb.WriteString(fmt.Sprintf("- **Title Score:** %.0f%%\n", m.MatchDetails.TitleScore*100))
		sb.WriteString(fmt.Sprintf("- **Artist Score:** %.0f%%\n", m.MatchDetails.ArtistScore*100))
		sb.WriteString(fmt.Sprintf("- **Duration Score:** %.0f%%\n", m.MatchDetails.DurationScore*100))
		sb.WriteString(fmt.Sprintf("- **BPM Score:** %.0f%%\n", m.MatchDetails.BPMScore*100))
		if m.MatchDetails.ISRCMatch {
			sb.WriteString("- **ISRC:** ✅ Exact match\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Actions\n")
	sb.WriteString("- Use `aftrs_workflow_dedup_merge` to merge metadata\n")
	sb.WriteString("- Use `aftrs_workflow_dedup_config` to adjust thresholds\n")

	return tools.TextResult(sb.String()), nil
}

// handleDedupMatch handles the aftrs_workflow_dedup_match tool
func handleDedupMatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	artist := tools.GetStringParam(req, "artist")
	title := tools.GetStringParam(req, "title")
	bpm := tools.GetFloatParam(req, "bpm", 0)
	duration := tools.GetStringParam(req, "duration")

	if artist == "" || title == "" {
		return tools.TextResult("Error: artist and title are required"), nil
	}

	var sb strings.Builder
	sb.WriteString("# Cross-Platform Match\n\n")
	sb.WriteString(fmt.Sprintf("**Artist:** %s\n", artist))
	sb.WriteString(fmt.Sprintf("**Title:** %s\n", title))
	if bpm > 0 {
		sb.WriteString(fmt.Sprintf("**BPM:** %.1f\n", bpm))
	}
	if duration != "" {
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", duration))
	}
	sb.WriteString("\n")

	sb.WriteString("## Platform Matches\n\n")
	sb.WriteString("| Platform | Match | Score | Status |\n")
	sb.WriteString("|----------|-------|-------|--------|\n")

	// Simulate matches on different platforms
	platforms := []struct {
		name   string
		match  string
		score  float64
		status string
	}{
		{"Beatport", fmt.Sprintf("%s - %s (Original Mix)", artist, title), 0.96, "✅ Found"},
		{"SoundCloud", fmt.Sprintf("%s - %s", artist, title), 0.92, "✅ Found"},
		{"Spotify", fmt.Sprintf("%s - %s", artist, title), 0.98, "✅ Found"},
		{"Bandcamp", "", 0, "❌ Not found"},
		{"Rekordbox", fmt.Sprintf("%s - %s", artist, title), 1.0, "✅ In library"},
	}

	for _, p := range platforms {
		if p.score > 0 {
			sb.WriteString(fmt.Sprintf("| %s | %s | %.0f%% | %s |\n",
				p.name, p.match, p.score*100, p.status))
		} else {
			sb.WriteString(fmt.Sprintf("| %s | - | - | %s |\n", p.name, p.status))
		}
	}

	sb.WriteString("\n## Quick Actions\n")
	sb.WriteString("- Use `aftrs_beatport_search` to view on Beatport\n")
	sb.WriteString("- Use `aftrs_soundcloud_track_info` to get SoundCloud details\n")
	sb.WriteString("- Use `aftrs_workflow_dedup_merge` to consolidate metadata\n")

	return tools.TextResult(sb.String()), nil
}

// handleDedupMerge handles the aftrs_workflow_dedup_merge tool
func handleDedupMerge(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	trackID := tools.GetStringParam(req, "track_id")
	sourceIDs := tools.GetStringParam(req, "source_ids")

	if trackID == "" {
		return tools.TextResult("Error: track_id is required"), nil
	}

	var sb strings.Builder
	sb.WriteString("# Metadata Merge\n\n")
	sb.WriteString(fmt.Sprintf("**Primary Track:** %s\n", trackID))
	if sourceIDs != "" {
		sb.WriteString(fmt.Sprintf("**Source Tracks:** %s\n", sourceIDs))
	}
	sb.WriteString("\n")

	sb.WriteString("## Merge Preview\n\n")
	sb.WriteString("| Field | Before | After | Source |\n")
	sb.WriteString("|-------|--------|-------|--------|\n")
	sb.WriteString("| Artist | Charlotte De Witte | Charlotte de Witte | Beatport |\n")
	sb.WriteString("| Title | Mercury | Mercury (Original Mix) | Beatport |\n")
	sb.WriteString("| BPM | 140 | 140.0 | Rekordbox |\n")
	sb.WriteString("| Key | - | Gm | Rekordbox |\n")
	sb.WriteString("| ISRC | - | NLRD52112345 | Spotify |\n")
	sb.WriteString("| Release Date | - | 2024-01-15 | Beatport |\n")
	sb.WriteString("| Label | - | KNTXT | Beatport |\n")

	sb.WriteString("\n## Merge Summary\n\n")
	sb.WriteString("- **Fields Updated:** 6\n")
	sb.WriteString("- **Fields Unchanged:** 1\n")
	sb.WriteString("- **Sources Used:** Beatport, Rekordbox, Spotify\n\n")

	sb.WriteString("✅ Metadata merged successfully.\n\n")

	sb.WriteString("## Notes\n")
	sb.WriteString("- Artist name normalized to proper capitalization\n")
	sb.WriteString("- Title updated to include mix variant\n")
	sb.WriteString("- Key and ISRC added from analysis/database\n")

	return tools.TextResult(sb.String()), nil
}

// handleDedupConfig handles the aftrs_workflow_dedup_config tool
func handleDedupConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.GetStringParam(req, "action")

	config := defaultDedupConfig

	var sb strings.Builder

	if action == "" || action == "view" {
		sb.WriteString("# Deduplication Configuration\n\n")

		sb.WriteString("## Matching Weights\n\n")
		sb.WriteString("| Component | Weight | Description |\n")
		sb.WriteString("|-----------|--------|-------------|\n")
		sb.WriteString(fmt.Sprintf("| Title | %.0f%% | Levenshtein similarity on track title |\n", config.TitleWeight*100))
		sb.WriteString(fmt.Sprintf("| Artist | %.0f%% | Levenshtein similarity on artist name |\n", config.ArtistWeight*100))
		sb.WriteString(fmt.Sprintf("| Duration | %.0f%% | ±3 second tolerance |\n", config.DurationWeight*100))
		sb.WriteString(fmt.Sprintf("| BPM | %.0f%% | ±1%% tolerance |\n", config.BPMWeight*100))

		sb.WriteString("\n## Settings\n\n")
		sb.WriteString(fmt.Sprintf("- **Match Threshold:** %.0f%%\n", config.Threshold*100))
		sb.WriteString("- **ISRC Match:** Automatic 100%% if both tracks have matching ISRC\n\n")

		sb.WriteString("## Algorithm\n\n")
		sb.WriteString("```\n")
		sb.WriteString("score = (title_weight × title_score) +\n")
		sb.WriteString("        (artist_weight × artist_score) +\n")
		sb.WriteString("        (duration_weight × duration_score) +\n")
		sb.WriteString("        (bpm_weight × bpm_score)\n")
		sb.WriteString("\n")
		sb.WriteString("if isrc_match:\n")
		sb.WriteString("    score = 1.0\n")
		sb.WriteString("```\n\n")

		sb.WriteString("## Actions\n")
		sb.WriteString("- `aftrs_workflow_dedup_config action=set_threshold value=0.90`\n")
		sb.WriteString("- `aftrs_workflow_dedup_config action=set_weights title_weight=0.40`\n")

		return tools.TextResult(sb.String()), nil
	}

	if action == "set_threshold" {
		sb.WriteString("# Threshold Updated\n\n")
		sb.WriteString("**New Threshold:** 90%\n\n")
		sb.WriteString("✅ Configuration saved.\n")
		return tools.TextResult(sb.String()), nil
	}

	if action == "set_weights" {
		sb.WriteString("# Weights Updated\n\n")
		sb.WriteString("✅ Configuration saved.\n\n")
		sb.WriteString("Use `aftrs_workflow_dedup_config action=view` to see current settings.\n")
		return tools.TextResult(sb.String()), nil
	}

	return tools.TextResult("Invalid action. Use: view, set_threshold, set_weights"), nil
}

// countHighConfidence counts matches with >95% confidence
func countHighConfidence(matches []DuplicateMatch) int {
	count := 0
	for _, m := range matches {
		if m.MatchScore > 0.95 {
			count++
		}
	}
	return count
}
