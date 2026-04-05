// Package mixcloud provides MCP tools for Mixcloud DJ mix discovery and downloads.
package mixcloud

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewMixcloudClient)

// Module implements the Mixcloud tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "mixcloud"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Mixcloud DJ mix discovery, search, and download tools"
}

// Tools returns the Mixcloud tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_mixcloud_status",
				mcp.WithDescription("Get Mixcloud API connection status and download tool availability"),
			),
			Handler:             handleMixcloudStatus,
			Category:            "mixcloud",
			Subcategory:         "status",
			Tags:                []string{"mixcloud", "dj", "mixes", "status"},
			UseCases:            []string{"Check Mixcloud connectivity", "Verify download tools"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mixcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_mixcloud_health",
				mcp.WithDescription("Check Mixcloud integration health with recommendations"),
			),
			Handler:             handleMixcloudHealth,
			Category:            "mixcloud",
			Subcategory:         "status",
			Tags:                []string{"mixcloud", "health", "diagnostics"},
			UseCases:            []string{"Diagnose Mixcloud issues", "Get setup recommendations"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mixcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_mixcloud_search",
				mcp.WithDescription("Search Mixcloud for mixes, DJs, or tags"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithString("type", mcp.Description("Search type: cloudcast/mix, user, tag (default: cloudcast)"), mcp.Enum("cloudcast", "mix", "user", "tag")),
			),
			Handler:             handleMixcloudSearch,
			Category:            "mixcloud",
			Subcategory:         "search",
			Tags:                []string{"mixcloud", "search", "dj", "mixes", "discovery"},
			UseCases:            []string{"Find DJ mixes", "Search for DJs", "Discover new mixes"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mixcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_mixcloud_user",
				mcp.WithDescription("Get Mixcloud user/DJ profile details"),
				mcp.WithString("username", mcp.Required(), mcp.Description("Mixcloud username")),
			),
			Handler:             handleMixcloudUser,
			Category:            "mixcloud",
			Subcategory:         "users",
			Tags:                []string{"mixcloud", "user", "dj", "profile"},
			UseCases:            []string{"Get DJ info", "View user profile", "Check follow stats"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mixcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_mixcloud_shows",
				mcp.WithDescription("Get a user's uploaded mixes/shows"),
				mcp.WithString("username", mcp.Required(), mcp.Description("Mixcloud username")),
				mcp.WithNumber("limit", mcp.Description("Maximum mixes to return (default: 20)")),
			),
			Handler:             handleMixcloudShows,
			Category:            "mixcloud",
			Subcategory:         "shows",
			Tags:                []string{"mixcloud", "shows", "mixes", "cloudcasts"},
			UseCases:            []string{"Browse DJ's mixes", "Get user uploads", "View discography"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mixcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_mixcloud_favorites",
				mcp.WithDescription("Get a user's favorite mixes"),
				mcp.WithString("username", mcp.Required(), mcp.Description("Mixcloud username")),
				mcp.WithNumber("limit", mcp.Description("Maximum mixes to return (default: 20)")),
			),
			Handler:             handleMixcloudFavorites,
			Category:            "mixcloud",
			Subcategory:         "favorites",
			Tags:                []string{"mixcloud", "favorites", "likes", "mixes"},
			UseCases:            []string{"Get liked mixes", "Browse favorites", "Discover via favorites"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mixcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_mixcloud_mix",
				mcp.WithDescription("Get detailed info about a specific mix including tracklist"),
				mcp.WithString("key", mcp.Required(), mcp.Description("Mix key (e.g., /username/mix-name/ or full URL)")),
			),
			Handler:             handleMixcloudMix,
			Category:            "mixcloud",
			Subcategory:         "shows",
			Tags:                []string{"mixcloud", "mix", "tracklist", "details"},
			UseCases:            []string{"Get mix details", "View tracklist", "Check play count"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mixcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_mixcloud_tags",
				mcp.WithDescription("Get popular Mixcloud genre tags for discovery"),
			),
			Handler:             handleMixcloudTags,
			Category:            "mixcloud",
			Subcategory:         "discovery",
			Tags:                []string{"mixcloud", "tags", "genres", "discovery"},
			UseCases:            []string{"Browse genres", "Discover by tag", "Find genre mixes"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mixcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_mixcloud_discover",
				mcp.WithDescription("Discover mixes by genre/tag"),
				mcp.WithString("tag", mcp.Required(), mcp.Description("Genre tag (e.g., house, techno, deep-house)")),
				mcp.WithNumber("limit", mcp.Description("Maximum mixes to return (default: 20)")),
			),
			Handler:             handleMixcloudDiscover,
			Category:            "mixcloud",
			Subcategory:         "discovery",
			Tags:                []string{"mixcloud", "discover", "genre", "browse"},
			UseCases:            []string{"Browse genre mixes", "Discover new music", "Find trending mixes"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mixcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_mixcloud_download",
				mcp.WithDescription("Download a mix from Mixcloud (requires yt-dlp)"),
				mcp.WithString("url", mcp.Required(), mcp.Description("Mixcloud mix URL")),
				mcp.WithString("output_dir", mcp.Description("Output directory for downloaded file (optional)")),
			),
			Handler:             handleMixcloudDownload,
			Category:            "mixcloud",
			Subcategory:         "download",
			Tags:                []string{"mixcloud", "download", "music", "archive"},
			UseCases:            []string{"Download DJ mixes", "Build mix archive", "Offline listening"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "mixcloud",
			IsWrite:             true,
		},
	}
}

func handleMixcloudStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Mixcloud client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleMixcloudHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Mixcloud client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}

func handleMixcloudSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Mixcloud client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	searchType := tools.GetStringParam(req, "type")
	if searchType == "" || searchType == "mix" {
		searchType = "cloudcast"
	}

	result, err := client.Search(ctx, query, searchType)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("search failed: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

func handleMixcloudUser(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Mixcloud client: %w", err)), nil
	}

	username, errResult := tools.RequireStringParam(req, "username")
	if errResult != nil {
		return errResult, nil
	}

	user, err := client.GetUser(ctx, username)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get user: %w", err)), nil
	}

	return tools.JSONResult(user), nil
}

func handleMixcloudShows(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Mixcloud client: %w", err)), nil
	}

	username, errResult := tools.RequireStringParam(req, "username")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	cloudcasts, err := client.GetUserCloudcasts(ctx, username, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get shows: %w", err)), nil
	}

	result := map[string]interface{}{
		"username": username,
		"shows":    cloudcasts,
		"count":    len(cloudcasts),
	}

	return tools.JSONResult(result), nil
}

func handleMixcloudFavorites(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Mixcloud client: %w", err)), nil
	}

	username, errResult := tools.RequireStringParam(req, "username")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	favorites, err := client.GetUserFavorites(ctx, username, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get favorites: %w", err)), nil
	}

	result := map[string]interface{}{
		"username":  username,
		"favorites": favorites,
		"count":     len(favorites),
	}

	return tools.JSONResult(result), nil
}

func handleMixcloudMix(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Mixcloud client: %w", err)), nil
	}

	key, errResult := tools.RequireStringParam(req, "key")
	if errResult != nil {
		return errResult, nil
	}

	// Handle full URL input
	if strings.Contains(key, "mixcloud.com") {
		// Extract path from URL
		key = strings.TrimPrefix(key, "https://www.mixcloud.com")
		key = strings.TrimPrefix(key, "https://mixcloud.com")
		key = strings.TrimPrefix(key, "http://www.mixcloud.com")
		key = strings.TrimPrefix(key, "http://mixcloud.com")
	}

	cloudcast, err := client.GetCloudcast(ctx, key)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get mix: %w", err)), nil
	}

	return tools.JSONResult(cloudcast), nil
}

func handleMixcloudTags(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Mixcloud client: %w", err)), nil
	}

	tags, err := client.GetPopularTags(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get tags: %w", err)), nil
	}

	result := map[string]interface{}{
		"tags":  tags,
		"count": len(tags),
		"note":  "Use aftrs_mixcloud_discover to browse mixes for a specific tag",
	}

	return tools.JSONResult(result), nil
}

func handleMixcloudDiscover(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Mixcloud client: %w", err)), nil
	}

	tag, errResult := tools.RequireStringParam(req, "tag")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 20)

	cloudcasts, err := client.GetTagCloudcasts(ctx, tag, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to discover mixes: %w", err)), nil
	}

	result := map[string]interface{}{
		"tag":   tag,
		"mixes": cloudcasts,
		"count": len(cloudcasts),
	}

	return tools.JSONResult(result), nil
}

func handleMixcloudDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Mixcloud client: %w", err)), nil
	}

	mixURL, errResult := tools.RequireStringParam(req, "url")
	if errResult != nil {
		return errResult, nil
	}

	outputDir := tools.GetStringParam(req, "output_dir")

	result, err := client.Download(ctx, mixURL, outputDir)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("download failed: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}
