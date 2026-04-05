// Package music_discovery provides MCP tools for unified music search across all platforms.
package music_discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for music discovery tools
type Module struct{}

func (m *Module) Name() string {
	return "music_discovery"
}

func (m *Module) Description() string {
	return "Unified music discovery tools for cross-platform search, track matching, price comparison, and metadata enrichment across 9+ music platforms"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// Status
		{
			Tool: mcp.NewTool("aftrs_music_status",
				mcp.WithDescription("Get connection status of all music platforms (Beatport, Traxsource, Juno, Boomkat, Spotify, SoundCloud, Bandcamp, Mixcloud, Discogs)"),
			),
			Handler:             handleStatus,
			Category:            "music_discovery",
			Subcategory:         "status",
			Tags:                []string{"music", "discovery", "status", "platforms"},
			UseCases:            []string{"Check platform availability", "Verify API connections"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "music_discovery",
		},

		// Unified Search
		{
			Tool: mcp.NewTool("aftrs_music_discover",
				mcp.WithDescription("Search for tracks across all music platforms simultaneously. Returns unified results with platform availability, BPM, key, genre, and pricing information."),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("Search query (artist, track title, or combination)"),
				),
				mcp.WithString("platforms",
					mcp.Description("Comma-separated list of platforms to search (default: all). Options: beatport, traxsource, juno, boomkat, spotify, soundcloud, bandcamp, mixcloud, discogs"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Maximum results per platform (default: 20)"),
				),
			),
			Handler:             handleDiscover,
			Category:            "music_discovery",
			Subcategory:         "search",
			Tags:                []string{"music", "discovery", "search", "unified", "cross-platform"},
			UseCases:            []string{"Find tracks across all stores", "Compare platform availability", "Unified music search"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "music_discovery",
		},

		// Track Match
		{
			Tool: mcp.NewTool("aftrs_music_track_match",
				mcp.WithDescription("Find the same track across all platforms. Returns platform IDs, URLs, and availability for each matching platform."),
				mcp.WithString("artist",
					mcp.Required(),
					mcp.Description("Artist name"),
				),
				mcp.WithString("title",
					mcp.Required(),
					mcp.Description("Track title"),
				),
				mcp.WithString("platforms",
					mcp.Description("Comma-separated list of platforms to search (default: all)"),
				),
			),
			Handler:             handleTrackMatch,
			Category:            "music_discovery",
			Subcategory:         "matching",
			Tags:                []string{"music", "discovery", "match", "cross-platform", "dedup"},
			UseCases:            []string{"Find track across stores", "Get all platform links", "Cross-reference tracks"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "music_discovery",
		},

		// Price Comparison
		{
			Tool: mcp.NewTool("aftrs_music_price_compare",
				mcp.WithDescription("Compare prices for a track across all music stores (Beatport, Traxsource, Juno, Boomkat, Bandcamp)"),
				mcp.WithString("artist",
					mcp.Required(),
					mcp.Description("Artist name"),
				),
				mcp.WithString("title",
					mcp.Required(),
					mcp.Description("Track title"),
				),
			),
			Handler:             handlePriceCompare,
			Category:            "music_discovery",
			Subcategory:         "pricing",
			Tags:                []string{"music", "discovery", "price", "compare", "store"},
			UseCases:            []string{"Find cheapest price", "Compare store pricing", "Save money on purchases"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "music_discovery",
		},

		// Metadata Enrichment
		{
			Tool: mcp.NewTool("aftrs_music_metadata_enrich",
				mcp.WithDescription("Aggregate metadata for a track from multiple platforms. Returns best available BPM, key, genre, and other metadata by cross-referencing platforms."),
				mcp.WithString("artist",
					mcp.Required(),
					mcp.Description("Artist name"),
				),
				mcp.WithString("title",
					mcp.Required(),
					mcp.Description("Track title"),
				),
			),
			Handler:             handleMetadataEnrich,
			Category:            "music_discovery",
			Subcategory:         "metadata",
			Tags:                []string{"music", "discovery", "metadata", "bpm", "key", "enrich"},
			UseCases:            []string{"Get accurate BPM/key", "Aggregate track metadata", "Enrich local library"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "music_discovery",
		},

		// New Releases
		{
			Tool: mcp.NewTool("aftrs_music_new_releases",
				mcp.WithDescription("Get new releases from all music platforms combined. Returns unified release feed sorted by date."),
				mcp.WithString("platforms",
					mcp.Description("Comma-separated list of platforms (default: beatport, traxsource, juno, boomkat)"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Maximum releases to return (default: 20)"),
				),
			),
			Handler:             handleNewReleases,
			Category:            "music_discovery",
			Subcategory:         "discovery",
			Tags:                []string{"music", "discovery", "releases", "new", "feed"},
			UseCases:            []string{"Browse new releases", "Combined release feed", "Stay up to date"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "music_discovery",
		},

		// Library Status
		{
			Tool: mcp.NewTool("aftrs_music_library_status",
				mcp.WithDescription("Get unified view of music library across all platforms. Shows track counts, playlists, and recent activity."),
			),
			Handler:             handleLibraryStatus,
			Category:            "music_discovery",
			Subcategory:         "library",
			Tags:                []string{"music", "discovery", "library", "unified", "stats"},
			UseCases:            []string{"View library overview", "Check sync status", "Cross-platform stats"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "music_discovery",
		},

		// Platform Info
		{
			Tool: mcp.NewTool("aftrs_music_platforms",
				mcp.WithDescription("List all supported music platforms with their features, capabilities, and metadata availability"),
			),
			Handler:             handlePlatforms,
			Category:            "music_discovery",
			Subcategory:         "info",
			Tags:                []string{"music", "discovery", "platforms", "info", "features"},
			UseCases:            []string{"View platform features", "Check BPM/key availability", "Platform comparison"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "music_discovery",
		},
	}
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Helper function to parse platforms from comma-separated string
func parsePlatforms(platformsStr string) []string {
	if platformsStr == "" {
		return nil // Return nil to use all platforms
	}

	parts := strings.Split(platformsStr, ",")
	platforms := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(strings.ToLower(p))
		if p != "" {
			platforms = append(platforms, p)
		}
	}
	return platforms
}

// handleStatus returns connection status of all music platforms
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMusicDiscoveryClient()

	status, err := client.Status(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

// handleDiscover performs unified search across all platforms
func handleDiscover(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMusicDiscoveryClient()

	// Parse arguments
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	platformsStr := tools.GetStringParam(req, "platforms")
	platforms := parsePlatforms(platformsStr)

	limit := tools.GetIntParam(req, "limit", 20)

	results, err := client.Search(ctx, query, platforms, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("search failed: %w", err)), nil
	}

	return tools.JSONResult(results), nil
}

// handleTrackMatch finds the same track across all platforms
func handleTrackMatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMusicDiscoveryClient()

	artist, errResult := tools.RequireStringParam(req, "artist")
	if errResult != nil {
		return errResult, nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	platformsStr := tools.GetStringParam(req, "platforms")
	platforms := parsePlatforms(platformsStr)

	match, err := client.MatchTrack(ctx, artist, title, platforms)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("track match failed: %w", err)), nil
	}

	return tools.JSONResult(match), nil
}

// handlePriceCompare compares prices across platforms
func handlePriceCompare(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMusicDiscoveryClient()

	artist, errResult := tools.RequireStringParam(req, "artist")
	if errResult != nil {
		return errResult, nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	comparison, err := client.ComparePrices(ctx, artist, title)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("price comparison failed: %w", err)), nil
	}

	return tools.JSONResult(comparison), nil
}

// handleMetadataEnrich aggregates metadata from multiple platforms
func handleMetadataEnrich(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMusicDiscoveryClient()

	artist, errResult := tools.RequireStringParam(req, "artist")
	if errResult != nil {
		return errResult, nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	enriched, err := client.EnrichMetadata(ctx, artist, title)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("metadata enrichment failed: %w", err)), nil
	}

	return tools.JSONResult(enriched), nil
}

// handleNewReleases gets new releases from all platforms
func handleNewReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMusicDiscoveryClient()

	platformsStr := tools.GetStringParam(req, "platforms")
	platforms := parsePlatforms(platformsStr)

	limit := tools.GetIntParam(req, "limit", 20)

	releases, err := client.GetNewReleases(ctx, platforms, limit)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get new releases: %w", err)), nil
	}

	return tools.JSONResult(releases), nil
}

// handleLibraryStatus returns unified library view
func handleLibraryStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMusicDiscoveryClient()

	status, err := client.GetLibraryStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get library status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

// handlePlatforms lists all supported platforms
func handlePlatforms(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMusicDiscoveryClient()

	platforms, err := client.GetPlatforms(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get platforms: %w", err)), nil
	}

	return tools.JSONResult(platforms), nil
}
