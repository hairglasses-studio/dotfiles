// Package discogs provides MCP tools for Discogs music database and collection management.
package discogs

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewDiscogsClient)

// Module implements the Discogs tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "discogs"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Discogs music database, vinyl collection, and marketplace tools"
}

// Tools returns the Discogs tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_discogs_status",
				mcp.WithDescription("Get Discogs API connection status and rate limits"),
			),
			Handler:             handleDiscogsStatus,
			Category:            "discogs",
			Subcategory:         "status",
			Tags:                []string{"discogs", "music", "vinyl", "status"},
			UseCases:            []string{"Check Discogs connectivity", "View rate limits"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_health",
				mcp.WithDescription("Check Discogs integration health with recommendations"),
			),
			Handler:             handleDiscogsHealth,
			Category:            "discogs",
			Subcategory:         "status",
			Tags:                []string{"discogs", "health", "diagnostics"},
			UseCases:            []string{"Diagnose Discogs issues", "Get setup recommendations"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_search",
				mcp.WithDescription("Search Discogs database for releases, artists, or labels"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithString("type", mcp.Description("Search type: release, master, artist, label (default: all)"), mcp.Enum("all", "release", "master", "artist", "label")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:             handleDiscogsSearch,
			Category:            "discogs",
			Subcategory:         "search",
			Tags:                []string{"discogs", "search", "music", "vinyl", "database"},
			UseCases:            []string{"Find releases", "Search artists", "Look up labels"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_release",
				mcp.WithDescription("Get detailed release information including tracklist"),
				mcp.WithNumber("id", mcp.Required(), mcp.Description("Discogs release ID")),
			),
			Handler:             handleDiscogsRelease,
			Category:            "discogs",
			Subcategory:         "database",
			Tags:                []string{"discogs", "release", "album", "vinyl", "tracklist"},
			UseCases:            []string{"Get release details", "View tracklist", "Check pressing info"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_master",
				mcp.WithDescription("Get master release information (canonical release)"),
				mcp.WithNumber("id", mcp.Required(), mcp.Description("Discogs master release ID")),
			),
			Handler:             handleDiscogsMaster,
			Category:            "discogs",
			Subcategory:         "database",
			Tags:                []string{"discogs", "master", "release", "canonical"},
			UseCases:            []string{"Get master release", "Find all pressings", "Canonical album info"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_artist",
				mcp.WithDescription("Get artist profile and discography summary"),
				mcp.WithNumber("id", mcp.Required(), mcp.Description("Discogs artist ID")),
			),
			Handler:             handleDiscogsArtist,
			Category:            "discogs",
			Subcategory:         "database",
			Tags:                []string{"discogs", "artist", "profile", "discography"},
			UseCases:            []string{"Get artist info", "View profile", "Find releases"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_artist_releases",
				mcp.WithDescription("Get an artist's releases"),
				mcp.WithNumber("id", mcp.Required(), mcp.Description("Discogs artist ID")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:             handleDiscogsArtistReleases,
			Category:            "discogs",
			Subcategory:         "database",
			Tags:                []string{"discogs", "artist", "releases", "discography"},
			UseCases:            []string{"Browse artist discography", "Find releases by artist"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_label",
				mcp.WithDescription("Get record label information"),
				mcp.WithNumber("id", mcp.Required(), mcp.Description("Discogs label ID")),
			),
			Handler:             handleDiscogsLabel,
			Category:            "discogs",
			Subcategory:         "database",
			Tags:                []string{"discogs", "label", "record", "catalog"},
			UseCases:            []string{"Get label info", "View label profile", "Find sublabels"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_collection",
				mcp.WithDescription("Get user's vinyl collection (requires DISCOGS_TOKEN)"),
				mcp.WithString("username", mcp.Description("Discogs username (uses authenticated user if omitted)")),
				mcp.WithNumber("folder_id", mcp.Description("Collection folder ID (0 = All, default: 0)")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:             handleDiscogsCollection,
			Category:            "discogs",
			Subcategory:         "collection",
			Tags:                []string{"discogs", "collection", "vinyl", "records"},
			UseCases:            []string{"View vinyl collection", "Browse owned records", "Catalog inventory"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_collection_folders",
				mcp.WithDescription("Get user's collection folders"),
				mcp.WithString("username", mcp.Description("Discogs username (uses authenticated user if omitted)")),
			),
			Handler:             handleDiscogsCollectionFolders,
			Category:            "discogs",
			Subcategory:         "collection",
			Tags:                []string{"discogs", "collection", "folders", "organize"},
			UseCases:            []string{"List collection folders", "View organization"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_wantlist",
				mcp.WithDescription("Get user's wantlist"),
				mcp.WithString("username", mcp.Description("Discogs username (uses authenticated user if omitted)")),
				mcp.WithNumber("page", mcp.Description("Page number (default: 1)")),
			),
			Handler:             handleDiscogsWantlist,
			Category:            "discogs",
			Subcategory:         "wantlist",
			Tags:                []string{"discogs", "wantlist", "wants", "wishlist"},
			UseCases:            []string{"View wanted records", "Track releases to buy"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
		{
			Tool: mcp.NewTool("aftrs_discogs_marketplace",
				mcp.WithDescription("Search marketplace listings for a release"),
				mcp.WithNumber("release_id", mcp.Required(), mcp.Description("Discogs release ID to search marketplace for")),
				mcp.WithString("currency", mcp.Description("Currency for prices (default: USD)"), mcp.Enum("USD", "EUR", "GBP", "AUD", "CAD", "JPY")),
			),
			Handler:             handleDiscogsMarketplace,
			Category:            "discogs",
			Subcategory:         "marketplace",
			Tags:                []string{"discogs", "marketplace", "buy", "sell", "prices"},
			UseCases:            []string{"Find copies for sale", "Check market prices", "Buy vinyl"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discogs",
		},
	}
}

func handleDiscogsStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleDiscogsHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}

func handleDiscogsSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	searchType := tools.GetStringParam(req, "type")
	if searchType == "all" {
		searchType = ""
	}

	page := tools.GetIntParam(req, "page", 1)

	result, err := client.Search(ctx, query, searchType, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("search failed: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

func handleDiscogsRelease(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	id, errResult := tools.RequireIntParam(req, "id")
	if errResult != nil {
		return errResult, nil
	}

	release, err := client.GetRelease(ctx, id)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get release: %w", err)), nil
	}

	return tools.JSONResult(release), nil
}

func handleDiscogsMaster(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	id, errResult := tools.RequireIntParam(req, "id")
	if errResult != nil {
		return errResult, nil
	}

	master, err := client.GetMaster(ctx, id)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get master: %w", err)), nil
	}

	return tools.JSONResult(master), nil
}

func handleDiscogsArtist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	id, errResult := tools.RequireIntParam(req, "id")
	if errResult != nil {
		return errResult, nil
	}

	artist, err := client.GetArtist(ctx, id)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get artist: %w", err)), nil
	}

	return tools.JSONResult(artist), nil
}

func handleDiscogsArtistReleases(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	id, errResult := tools.RequireIntParam(req, "id")
	if errResult != nil {
		return errResult, nil
	}

	page := tools.GetIntParam(req, "page", 1)

	result, err := client.GetArtistReleases(ctx, id, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get releases: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

func handleDiscogsLabel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	id, errResult := tools.RequireIntParam(req, "id")
	if errResult != nil {
		return errResult, nil
	}

	label, err := client.GetLabel(ctx, id)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get label: %w", err)), nil
	}

	return tools.JSONResult(label), nil
}

func handleDiscogsCollection(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	username := tools.GetStringParam(req, "username")
	folderID := tools.GetIntParam(req, "folder_id", 0)
	page := tools.GetIntParam(req, "page", 1)

	items, pagination, err := client.GetCollection(ctx, username, folderID, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get collection: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"items":      items,
		"count":      len(items),
		"pagination": pagination,
	}), nil
}

func handleDiscogsCollectionFolders(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	username := tools.GetStringParam(req, "username")

	folders, err := client.GetCollectionFolders(ctx, username)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get folders: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"folders": folders,
		"count":   len(folders),
	}), nil
}

func handleDiscogsWantlist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	username := tools.GetStringParam(req, "username")
	page := tools.GetIntParam(req, "page", 1)

	items, pagination, err := client.GetWantlist(ctx, username, page)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get wantlist: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"items":      items,
		"count":      len(items),
		"pagination": pagination,
	}), nil
}

func handleDiscogsMarketplace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Discogs client: %w", err)), nil
	}

	releaseID, errResult := tools.RequireIntParam(req, "release_id")
	if errResult != nil {
		return errResult, nil
	}

	currency := tools.OptionalStringParam(req, "currency", "USD")

	listings, err := client.SearchMarketplace(ctx, releaseID, currency)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to search marketplace: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"release_id": releaseID,
		"currency":   currency,
		"listings":   listings,
		"count":      len(listings),
	}), nil
}
