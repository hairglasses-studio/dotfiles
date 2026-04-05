// Package youtube_live provides MCP tools for YouTube Live streaming integration.
package youtube_live

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewYouTubeLiveClient)

// Module implements the ToolModule interface for YouTube Live.
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string        { return "youtube_live" }
func (m *Module) Description() string { return "YouTube Live streaming integration" }

func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool:        mcp.NewTool("aftrs_ytlive_status", mcp.WithDescription("Get YouTube Live broadcast status and metrics.")),
			Handler:     handleStatus,
			Category:    "youtube_live",
			Subcategory: "status",
			Tags:        []string{"youtube", "streaming", "status"},
			UseCases:    []string{"Check if broadcasting", "Get viewer count"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_ytlive_broadcast", mcp.WithDescription("Get broadcast details."),
				mcp.WithString("action", mcp.Description("Action: list, get")),
				mcp.WithString("broadcast_id", mcp.Description("Broadcast ID for get")),
				mcp.WithString("status", mcp.Description("Filter: active, all, completed, upcoming"))),
			Handler:     handleBroadcast,
			Category:    "youtube_live",
			Subcategory: "broadcast",
			Tags:        []string{"youtube", "broadcast", "live"},
			UseCases:    []string{"List broadcasts", "Get broadcast details"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_ytlive_chat", mcp.WithDescription("Get live chat messages."),
				mcp.WithString("video_id", mcp.Required(), mcp.Description("Video/stream ID")),
				mcp.WithString("page_token", mcp.Description("Pagination token"))),
			Handler:     handleChat,
			Category:    "youtube_live",
			Subcategory: "chat",
			Tags:        []string{"youtube", "chat", "live"},
			UseCases:    []string{"Read live chat", "Monitor chat"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_ytlive_video", mcp.WithDescription("Get video/stream details with live stats."),
				mcp.WithString("video_id", mcp.Required(), mcp.Description("Video ID"))),
			Handler:     handleVideo,
			Category:    "youtube_live",
			Subcategory: "video",
			Tags:        []string{"youtube", "video", "stats"},
			UseCases:    []string{"Get video details", "Check live stats"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_ytlive_search", mcp.WithDescription("Search for live streams."),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithNumber("max_results", mcp.Description("Max results (default 25)"))),
			Handler:     handleSearch,
			Category:    "youtube_live",
			Subcategory: "search",
			Tags:        []string{"youtube", "search", "live"},
			UseCases:    []string{"Find live streams", "Search content"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_ytlive_channel", mcp.WithDescription("Get current live stream for a channel."),
				mcp.WithString("channel_id", mcp.Description("Channel ID (uses configured default if omitted)"))),
			Handler:     handleChannel,
			Category:    "youtube_live",
			Subcategory: "channel",
			Tags:        []string{"youtube", "channel", "live"},
			UseCases:    []string{"Check if channel is live", "Get channel stream"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_ytlive_metrics", mcp.WithDescription("Get live stream metrics and statistics."),
				mcp.WithString("video_id", mcp.Required(), mcp.Description("Video ID"))),
			Handler:     handleMetrics,
			Category:    "youtube_live",
			Subcategory: "metrics",
			Tags:        []string{"youtube", "metrics", "analytics"},
			UseCases:    []string{"Get viewer stats", "Monitor performance"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool:        mcp.NewTool("aftrs_ytlive_health", mcp.WithDescription("Check YouTube API health.")),
			Handler:     handleHealth,
			Category:    "youtube_live",
			Subcategory: "health",
			Tags:        []string{"youtube", "health", "status"},
			UseCases:    []string{"Check API connection", "Diagnose issues"},
			Complexity:  tools.ComplexitySimple,
		},
	}

	// Apply circuit breaker to all tools — separate from "google" due to strict daily quotas
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "youtube"
	}

	return allTools
}

func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.JSONResult(status), nil
}

func handleBroadcast(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	action := tools.OptionalStringParam(req, "action", "list")

	switch action {
	case "list":
		status := tools.OptionalStringParam(req, "status", "active")
		broadcasts, err := client.GetBroadcasts(ctx, status)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(broadcasts), nil

	case "get":
		broadcastID, errResult := tools.RequireStringParam(req, "broadcast_id")
		if errResult != nil {
			return errResult, nil
		}
		broadcast, err := client.GetBroadcast(ctx, broadcastID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.JSONResult(broadcast), nil
	}

	return tools.ErrorResult(fmt.Errorf("unknown action: %s", action)), nil
}

func handleChat(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	videoID := tools.GetStringParam(req, "video_id")
	pageToken := tools.GetStringParam(req, "page_token")

	video, err := client.GetVideo(ctx, videoID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get video: %v", err)), nil
	}

	if video.LiveDetails == nil || video.LiveDetails.ActiveLiveChatID == "" {
		return tools.ErrorResult(fmt.Errorf("video does not have an active live chat")), nil
	}

	messages, nextToken, err := client.GetLiveChatMessages(ctx, video.LiveDetails.ActiveLiveChatID, pageToken)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"messages":        messages,
		"next_page_token": nextToken,
	}), nil
}

func handleVideo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	videoID := tools.GetStringParam(req, "video_id")
	video, err := client.GetVideo(ctx, videoID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(video), nil
}

func handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	query := tools.GetStringParam(req, "query")
	maxResults := tools.GetIntParam(req, "max_results", 25)

	videos, err := client.SearchLiveStreams(ctx, query, maxResults)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(videos), nil
}

func handleChannel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	channelID := tools.GetStringParam(req, "channel_id")
	video, err := client.GetChannelLiveStream(ctx, channelID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if video == nil {
		return tools.TextResult("Channel is not currently live"), nil
	}

	return tools.JSONResult(video), nil
}

func handleMetrics(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	videoID := tools.GetStringParam(req, "video_id")
	video, err := client.GetVideo(ctx, videoID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	metrics := map[string]interface{}{
		"video_id": video.ID,
		"title":    video.Title,
	}

	if video.Statistics != nil {
		metrics["view_count"] = video.Statistics.ViewCount
		metrics["like_count"] = video.Statistics.LikeCount
		metrics["comment_count"] = video.Statistics.CommentCount
	}

	if video.LiveDetails != nil {
		metrics["concurrent_viewers"] = video.LiveDetails.ConcurrentViewers
		metrics["started_at"] = video.LiveDetails.ActualStartTime
		metrics["is_live"] = video.LiveDetails.ActualEndTime == ""
	}

	return tools.JSONResult(metrics), nil
}

func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(health), nil
}
