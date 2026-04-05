// Package analytics provides MCP performance analytics tools.
package analytics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for analytics
type Module struct{}

var getClient = tools.LazyClient(clients.NewAnalyticsClient)

func (m *Module) Name() string {
	return "analytics"
}

func (m *Module) Description() string {
	return "Performance analytics and session tracking"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_session_start",
				mcp.WithDescription("Start a new performance session for tracking. Logs tracks, transitions, and BPM throughout the set."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Session name (e.g., 'Friday Night Set')")),
				mcp.WithString("venue", mcp.Description("Venue name")),
				mcp.WithString("type", mcp.Description("Session type: dj, live, hybrid (default: dj)")),
			),
			Handler:             handleSessionStart,
			Category:            "analytics",
			Subcategory:         "session",
			Tags:                []string{"session", "start", "tracking", "performance"},
			UseCases:            []string{"Start DJ set tracking", "Begin live performance logging"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "analytics",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_session_end",
				mcp.WithDescription("End the current performance session. Calculates final metrics and saves session data."),
			),
			Handler:             handleSessionEnd,
			Category:            "analytics",
			Subcategory:         "session",
			Tags:                []string{"session", "end", "stop", "save"},
			UseCases:            []string{"End DJ set", "Complete performance logging"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "analytics",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_session_status",
				mcp.WithDescription("Get current session status with live metrics. Shows tracks played, BPM history, and performance stats."),
				mcp.WithString("session_id", mcp.Description("Session ID (default: current session)")),
			),
			Handler:             handleSessionStatus,
			Category:            "analytics",
			Subcategory:         "session",
			Tags:                []string{"session", "status", "metrics", "current"},
			UseCases:            []string{"Check set progress", "View live metrics"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "analytics",
		},
		{
			Tool: mcp.NewTool("aftrs_track_log",
				mcp.WithDescription("Log a track play in the current session. Automatically calculates transitions."),
				mcp.WithString("title", mcp.Required(), mcp.Description("Track title")),
				mcp.WithString("artist", mcp.Required(), mcp.Description("Track artist")),
				mcp.WithNumber("bpm", mcp.Required(), mcp.Description("Track BPM")),
				mcp.WithString("key", mcp.Description("Track key (e.g., '5A', 'Dm')")),
				mcp.WithString("source", mcp.Description("Source: rekordbox, serato, traktor, ableton")),
				mcp.WithNumber("energy", mcp.Description("Energy level 1-10")),
			),
			Handler:             handleTrackLog,
			Category:            "analytics",
			Subcategory:         "session",
			Tags:                []string{"track", "log", "play", "record"},
			UseCases:            []string{"Log track in set", "Record play history"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "analytics",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_analytics_report",
				mcp.WithDescription("Generate analytics report across all sessions. Shows most played tracks, artists, venue stats, and trends."),
				mcp.WithString("period", mcp.Description("Time period: all, week, month, year (default: all)")),
			),
			Handler:             handleAnalyticsReport,
			Category:            "analytics",
			Subcategory:         "reporting",
			Tags:                []string{"analytics", "report", "stats", "trends"},
			UseCases:            []string{"View play history", "Analyze performance trends"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "analytics",
		},
		{
			Tool: mcp.NewTool("aftrs_sessions_list",
				mcp.WithDescription("List all recorded performance sessions."),
				mcp.WithNumber("limit", mcp.Description("Maximum sessions to show (default: 10)")),
			),
			Handler:             handleSessionsList,
			Category:            "analytics",
			Subcategory:         "session",
			Tags:                []string{"sessions", "list", "history"},
			UseCases:            []string{"View past sets", "Find session to analyze"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "analytics",
		},
	}
}

func handleSessionStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}
	venue := tools.GetStringParam(req, "venue")
	sessionType := tools.OptionalStringParam(req, "type", "dj")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	session, err := client.StartSession(ctx, name, venue, sessionType)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Session Started\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", session.ID[:16]))
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", session.Name))
	if venue != "" {
		sb.WriteString(fmt.Sprintf("**Venue:** %s\n", venue))
	}
	sb.WriteString(fmt.Sprintf("**Type:** %s\n", sessionType))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n\n", session.StartTime.Format("15:04:05")))
	sb.WriteString("Use `aftrs_track_log` to record tracks as you play them.\n")

	return tools.TextResult(sb.String()), nil
}

func handleSessionEnd(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	session, err := client.EndSession(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Session Ended\n\n")
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", session.Name))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", session.Duration.Round(time.Minute)))
	sb.WriteString("\n## Summary\n\n")

	if session.Metrics != nil {
		sb.WriteString(fmt.Sprintf("- **Tracks Played:** %d\n", session.Metrics.TotalTracks))
		sb.WriteString(fmt.Sprintf("- **Unique Artists:** %d\n", session.Metrics.UniqueArtists))
		sb.WriteString(fmt.Sprintf("- **Average BPM:** %.1f\n", session.Metrics.AverageBPM))
		sb.WriteString(fmt.Sprintf("- **BPM Range:** %.1f - %.1f\n", session.Metrics.MinBPM, session.Metrics.MaxBPM))
		sb.WriteString(fmt.Sprintf("- **Transitions:** %d\n", session.Metrics.TransitionCount))
		if session.Metrics.AverageTrackTime > 0 {
			sb.WriteString(fmt.Sprintf("- **Avg Track Time:** %s\n", session.Metrics.AverageTrackTime.Round(time.Second)))
		}
	}

	sb.WriteString(fmt.Sprintf("\nSession saved to: `%s`\n", session.ID[:16]))

	return tools.TextResult(sb.String()), nil
}

func handleSessionStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := tools.GetStringParam(req, "session_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var session *clients.PerformanceSession
	if sessionID != "" {
		session, err = client.GetSession(ctx, sessionID)
	} else {
		session, err = client.GetCurrentSession(ctx)
	}
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Session: %s\n\n", session.Name))

	// Duration
	var duration time.Duration
	if session.EndTime.IsZero() {
		duration = time.Since(session.StartTime)
		sb.WriteString(fmt.Sprintf("**Status:** 🔴 Live (%s)\n", duration.Round(time.Second)))
	} else {
		duration = session.Duration
		sb.WriteString(fmt.Sprintf("**Status:** Ended (%s)\n", duration.Round(time.Minute)))
	}

	if session.Venue != "" {
		sb.WriteString(fmt.Sprintf("**Venue:** %s\n", session.Venue))
	}
	sb.WriteString("\n")

	// Metrics
	if session.Metrics != nil {
		sb.WriteString("## Metrics\n\n")
		sb.WriteString(fmt.Sprintf("- Tracks: **%d**\n", session.Metrics.TotalTracks))
		sb.WriteString(fmt.Sprintf("- Artists: **%d**\n", session.Metrics.UniqueArtists))
		sb.WriteString(fmt.Sprintf("- BPM: **%.1f** avg (%.0f-%.0f)\n",
			session.Metrics.AverageBPM, session.Metrics.MinBPM, session.Metrics.MaxBPM))
		sb.WriteString(fmt.Sprintf("- Transitions: **%d**\n", session.Metrics.TransitionCount))
		sb.WriteString("\n")
	}

	// Recent tracks
	if len(session.Tracks) > 0 {
		sb.WriteString("## Recent Tracks\n\n")
		start := len(session.Tracks) - 5
		if start < 0 {
			start = 0
		}
		for i := start; i < len(session.Tracks); i++ {
			track := session.Tracks[i]
			key := ""
			if track.Key != "" {
				key = fmt.Sprintf(" [%s]", track.Key)
			}
			sb.WriteString(fmt.Sprintf("%d. **%s** - %s (%.1f BPM%s)\n",
				i+1, track.Title, track.Artist, track.BPM, key))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleTrackLog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}
	artist, errResult := tools.RequireStringParam(req, "artist")
	if errResult != nil {
		return errResult, nil
	}
	bpm := tools.GetFloatParam(req, "bpm", 0)
	key := tools.GetStringParam(req, "key")
	source := tools.GetStringParam(req, "source")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	track, err := client.LogTrack(ctx, title, artist, bpm, key, source)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Get current session for context
	session, _ := client.GetCurrentSession(ctx)

	var sb strings.Builder
	sb.WriteString("# Track Logged\n\n")
	sb.WriteString(fmt.Sprintf("**#%d:** %s - %s\n", len(session.Tracks), track.Title, track.Artist))
	sb.WriteString(fmt.Sprintf("**BPM:** %.1f", track.BPM))
	if key != "" {
		sb.WriteString(fmt.Sprintf(" | **Key:** %s", key))
	}
	sb.WriteString("\n")

	// Show transition info if not first track
	if len(session.Tracks) > 1 {
		prevTrack := session.Tracks[len(session.Tracks)-2]
		bpmDiff := track.BPM - prevTrack.BPM
		sb.WriteString(fmt.Sprintf("\n**Transition:** %+.1f BPM from previous\n", bpmDiff))
	}

	return tools.TextResult(sb.String()), nil
}

func handleAnalyticsReport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	analytics, err := client.GetAnalytics(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Performance Analytics\n\n")

	sb.WriteString("## Overview\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Sessions:** %d\n", analytics.TotalSessions))
	sb.WriteString(fmt.Sprintf("- **Total Playtime:** %s\n", analytics.TotalPlaytime.Round(time.Minute)))
	sb.WriteString(fmt.Sprintf("- **Total Tracks:** %d\n", analytics.TotalTracks))
	sb.WriteString(fmt.Sprintf("- **Average BPM:** %.1f\n", analytics.AverageBPM))
	sb.WriteString("\n")

	if analytics.MostPlayedTrack != "" {
		sb.WriteString("## Top Tracks\n\n")
		sb.WriteString(fmt.Sprintf("**Most Played:** %s (%d plays)\n\n",
			analytics.MostPlayedTrack, analytics.TrackFrequency[analytics.MostPlayedTrack]))

		// Top 5 tracks
		type trackCount struct {
			title string
			count int
		}
		var tracks []trackCount
		for t, c := range analytics.TrackFrequency {
			tracks = append(tracks, trackCount{t, c})
		}
		if len(tracks) > 5 {
			tracks = tracks[:5]
		}
		for _, t := range tracks {
			sb.WriteString(fmt.Sprintf("- %s (%d)\n", t.title, t.count))
		}
		sb.WriteString("\n")
	}

	if analytics.MostPlayedArtist != "" {
		sb.WriteString("## Top Artists\n\n")
		sb.WriteString(fmt.Sprintf("**Most Played:** %s (%d tracks)\n",
			analytics.MostPlayedArtist, analytics.ArtistFrequency[analytics.MostPlayedArtist]))
	}

	if len(analytics.VenueStats) > 0 {
		sb.WriteString("\n## Venues\n\n")
		for name, stats := range analytics.VenueStats {
			sb.WriteString(fmt.Sprintf("- **%s:** %d sessions, %s total\n",
				name, stats.SessionCount, stats.TotalDuration.Round(time.Minute)))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleSessionsList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(req, "limit", 10)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	sessions := client.ListSessions(ctx)

	var sb strings.Builder
	sb.WriteString("# Performance Sessions\n\n")

	if len(sessions) == 0 {
		sb.WriteString("No sessions recorded yet.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**Total:** %d sessions\n\n", len(sessions)))
	sb.WriteString("| Date | Name | Duration | Tracks | Avg BPM |\n")
	sb.WriteString("|------|------|----------|--------|----------|\n")

	count := 0
	for _, s := range sessions {
		if count >= limit {
			break
		}

		duration := s.Duration
		if duration == 0 && !s.EndTime.IsZero() {
			duration = s.EndTime.Sub(s.StartTime)
		}

		avgBPM := "-"
		if s.Metrics != nil && s.Metrics.AverageBPM > 0 {
			avgBPM = fmt.Sprintf("%.1f", s.Metrics.AverageBPM)
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s |\n",
			s.StartTime.Format("01-02 15:04"),
			s.Name,
			duration.Round(time.Minute),
			len(s.Tracks),
			avgBPM))
		count++
	}

	return tools.TextResult(sb.String()), nil
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
