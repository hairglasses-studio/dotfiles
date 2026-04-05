// Package resolume_plugins provides Resolume plugin management MCP tools.
package resolume_plugins

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Resolume plugin management
type Module struct{}

func (m *Module) Name() string {
	return "resolume_plugins"
}

func (m *Module) Description() string {
	return "Resolume plugin management: scan, sync, upload, and download FFGL/ISF plugins via S3"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// Discovery Tools
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_list",
				mcp.WithDescription("List all installed Resolume plugins (JuiceBar + Extra Effects + built-in)."),
				mcp.WithString("type", mcp.Description("Filter by type: ffgl_effect, ffgl_source, isf_shader, juicebar")),
				mcp.WithBoolean("include_builtin", mcp.Description("Include built-in plugins (default: false)")),
			),
			Handler:             handleList,
			Category:            "resolume_plugins",
			Subcategory:         "discovery",
			Tags:                []string{"resolume", "plugins", "list", "ffgl", "isf", "juicebar"},
			UseCases:            []string{"Browse installed plugins", "Inventory local plugins"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "resolume",
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_scan",
				mcp.WithDescription("Deep scan all Resolume plugin directories with metadata extraction."),
			),
			Handler:             handleScan,
			Category:            "resolume_plugins",
			Subcategory:         "discovery",
			Tags:                []string{"resolume", "plugins", "scan", "metadata"},
			UseCases:            []string{"Full plugin audit", "Get detailed plugin info"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "resolume",
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_s3_list",
				mcp.WithDescription("List plugins available in the S3 bucket."),
				mcp.WithString("prefix", mcp.Description("Filter by S3 prefix (e.g., 'ffgl/', 'isf/', 'juicebar-backup/')")),
			),
			Handler:             handleS3List,
			Category:            "resolume_plugins",
			Subcategory:         "discovery",
			Tags:                []string{"resolume", "plugins", "s3", "cloud", "list"},
			UseCases:            []string{"Browse cloud plugins", "See available downloads"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "resolume",
		},

		// Sync Tools
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_upload",
				mcp.WithDescription("Upload a plugin to S3."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Local path to the plugin file")),
				mcp.WithString("version", mcp.Description("Version string (e.g., '1.0.0')")),
				mcp.WithString("tags", mcp.Description("Comma-separated tags")),
				mcp.WithBoolean("is_juicebar", mcp.Description("Mark as JuiceBar backup")),
			),
			Handler:             handleUpload,
			Category:            "resolume_plugins",
			Subcategory:         "sync",
			Tags:                []string{"resolume", "plugins", "upload", "s3"},
			UseCases:            []string{"Backup plugin to cloud", "Share plugin"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "resolume",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_download",
				mcp.WithDescription("Download a plugin from S3 and install to Extra Effects."),
				mcp.WithString("key", mcp.Required(), mcp.Description("S3 key of the plugin (from s3_list)")),
			),
			Handler:             handleDownload,
			Category:            "resolume_plugins",
			Subcategory:         "sync",
			Tags:                []string{"resolume", "plugins", "download", "s3", "install"},
			UseCases:            []string{"Download cloud plugin", "Restore from backup"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "resolume",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_sync",
				mcp.WithDescription("Compare local and S3 plugins to identify sync opportunities."),
			),
			Handler:             handleSync,
			Category:            "resolume_plugins",
			Subcategory:         "sync",
			Tags:                []string{"resolume", "plugins", "sync", "compare"},
			UseCases:            []string{"Check sync status", "Find missing plugins"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "resolume",
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_backup",
				mcp.WithDescription("Backup all local plugins to S3."),
				mcp.WithBoolean("include_juicebar", mcp.Description("Include JuiceBar plugins (default: true)")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview what would be uploaded")),
			),
			Handler:             handleBackup,
			Category:            "resolume_plugins",
			Subcategory:         "sync",
			Tags:                []string{"resolume", "plugins", "backup", "s3", "all"},
			UseCases:            []string{"Full plugin backup", "Disaster recovery prep"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "resolume",
			IsWrite:             true,
		},

		// Installation Tools
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_install",
				mcp.WithDescription("Install a plugin file to Resolume Extra Effects folder."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to plugin file (.bundle/.dll/.isf)")),
			),
			Handler:             handleInstall,
			Category:            "resolume_plugins",
			Subcategory:         "installation",
			Tags:                []string{"resolume", "plugins", "install", "local"},
			UseCases:            []string{"Install new plugin", "Add FFGL effect"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "resolume",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_uninstall",
				mcp.WithDescription("Remove a plugin from Resolume Extra Effects folder."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name (partial match supported)")),
			),
			Handler:             handleUninstall,
			Category:            "resolume_plugins",
			Subcategory:         "installation",
			Tags:                []string{"resolume", "plugins", "uninstall", "remove"},
			UseCases:            []string{"Remove plugin", "Clean up effects"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "resolume",
			IsWrite:             true,
		},

		// Management Tools
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_health",
				mcp.WithDescription("Check Resolume plugin system health and get recommendations."),
			),
			Handler:             handleHealth,
			Category:            "resolume_plugins",
			Subcategory:         "health",
			Tags:                []string{"resolume", "plugins", "health", "status"},
			UseCases:            []string{"Diagnose plugin issues", "System check"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "resolume",
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_paths",
				mcp.WithDescription("Show all Resolume plugin directory paths."),
			),
			Handler:             handlePaths,
			Category:            "resolume_plugins",
			Subcategory:         "config",
			Tags:                []string{"resolume", "plugins", "paths", "directories"},
			UseCases:            []string{"Find plugin folders", "Check paths"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "resolume",
		},

		// Search Tools
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_search",
				mcp.WithDescription("Search for plugins in S3 by name, type, or tags."),
				mcp.WithString("query", mcp.Description("Search query (matches plugin name)")),
				mcp.WithString("type", mcp.Description("Filter by type: ffgl_effect, ffgl_source, isf_shader")),
				mcp.WithString("tags", mcp.Description("Comma-separated tags to filter by")),
			),
			Handler:             handleSearch,
			Category:            "resolume_plugins",
			Subcategory:         "discovery",
			Tags:                []string{"resolume", "plugins", "search", "s3", "filter"},
			UseCases:            []string{"Find specific plugin", "Filter by category"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "resolume",
		},

		// Batch Operations
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_batch_upload",
				mcp.WithDescription("Upload multiple plugins to S3 at once."),
				mcp.WithString("paths", mcp.Required(), mcp.Description("Comma-separated local paths to plugin files")),
				mcp.WithBoolean("is_juicebar", mcp.Description("Mark as JuiceBar backups")),
			),
			Handler:             handleBatchUpload,
			Category:            "resolume_plugins",
			Subcategory:         "sync",
			Tags:                []string{"resolume", "plugins", "upload", "batch", "s3"},
			UseCases:            []string{"Upload multiple plugins", "Bulk backup"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "resolume",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_batch_download",
				mcp.WithDescription("Download multiple plugins from S3 at once."),
				mcp.WithString("keys", mcp.Required(), mcp.Description("Comma-separated S3 keys to download")),
			),
			Handler:             handleBatchDownload,
			Category:            "resolume_plugins",
			Subcategory:         "sync",
			Tags:                []string{"resolume", "plugins", "download", "batch", "s3"},
			UseCases:            []string{"Download multiple plugins", "Bulk restore"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "resolume",
			IsWrite:             true,
		},

		// Google Drive Import
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_import_gdrive",
				mcp.WithDescription("Import a plugin from Google Drive and optionally upload to S3."),
				mcp.WithString("file_id", mcp.Required(), mcp.Description("Google Drive file ID")),
				mcp.WithBoolean("upload_to_s3", mcp.Description("Also upload to S3 after installing (default: true)")),
			),
			Handler:             handleImportGDrive,
			Category:            "resolume_plugins",
			Subcategory:         "import",
			Tags:                []string{"resolume", "plugins", "gdrive", "import"},
			UseCases:            []string{"Import from shared Drive", "Download plugin from link"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "resolume",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_import_gdrive_folder",
				mcp.WithDescription("Import all plugins from a Google Drive folder."),
				mcp.WithString("folder_id", mcp.Required(), mcp.Description("Google Drive folder ID")),
				mcp.WithBoolean("upload_to_s3", mcp.Description("Also upload to S3 after installing (default: true)")),
			),
			Handler:             handleImportGDriveFolder,
			Category:            "resolume_plugins",
			Subcategory:         "import",
			Tags:                []string{"resolume", "plugins", "gdrive", "import", "batch"},
			UseCases:            []string{"Import folder of plugins", "Bulk import from Drive"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "resolume",
			IsWrite:             true,
		},

		// ISF Shader Browser
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_isf_browse",
				mcp.WithDescription("Browse ISF shader collections from the community."),
				mcp.WithString("category", mcp.Description("Filter by category: effects, generators, glitch, patterns")),
				mcp.WithNumber("limit", mcp.Description("Maximum results to return")),
			),
			Handler:             handleISFBrowse,
			Category:            "resolume_plugins",
			Subcategory:         "isf",
			Tags:                []string{"resolume", "plugins", "isf", "shaders", "browse"},
			UseCases:            []string{"Discover ISF shaders", "Browse shader collections"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "resolume",
		},

		// Resolume Integration
		{
			Tool: mcp.NewTool("aftrs_resolume_plugins_resolume_status",
				mcp.WithDescription("Check if Resolume is running and get connection info."),
			),
			Handler:             handleResolumeStatus,
			Category:            "resolume_plugins",
			Subcategory:         "integration",
			Tags:                []string{"resolume", "plugins", "status", "connection"},
			UseCases:            []string{"Check Resolume status", "Verify connection"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "resolume",
		},
	}
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// getClient creates a new Resolume plugins client
func getClient() (*clients.ResolumePluginsClient, error) {
	return clients.NewResolumePluginsClient()
}

// Handlers

func handleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	plugins, err := client.ScanLocalPlugins(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	typeFilter := tools.GetStringParam(req, "type")
	includeBuiltin := tools.GetBoolParam(req, "include_builtin", false)

	var filtered []clients.ResolumePluginInfo
	for _, p := range plugins {
		// Filter by type
		if typeFilter != "" && string(p.Type) != typeFilter {
			continue
		}
		// Filter built-in
		if !includeBuiltin && (strings.Contains(p.LocalPath, "/Applications/") ||
			strings.Contains(p.LocalPath, "Program Files")) {
			continue
		}
		filtered = append(filtered, p)
	}

	if len(filtered) == 0 {
		return tools.TextResult("No plugins found matching criteria."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Resolume Plugins (%d)\n\n", len(filtered)))

	// Group by source
	bySource := make(map[clients.ResolumePluginSource][]clients.ResolumePluginInfo)
	for _, p := range filtered {
		bySource[p.Source] = append(bySource[p.Source], p)
	}

	for source, plugins := range bySource {
		sb.WriteString(fmt.Sprintf("## %s (%d)\n\n", formatSource(source), len(plugins)))
		sb.WriteString("| Name | Type | Size | Modified |\n")
		sb.WriteString("|------|------|------|----------|\n")

		for _, p := range plugins {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				p.Name,
				p.Type,
				formatSize(p.Size),
				p.ModifiedAt.Format("2006-01-02"),
			))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleScan(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	plugins, err := client.ScanLocalPlugins(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Plugin Scan Results\n\n"))
	sb.WriteString(fmt.Sprintf("**Total plugins found:** %d\n\n", len(plugins)))

	// Statistics
	stats := make(map[clients.ResolumePluginType]int)
	var totalSize int64
	juicebarCount := 0

	for _, p := range plugins {
		stats[p.Type]++
		totalSize += p.Size
		if p.IsJuiceBar {
			juicebarCount++
		}
	}

	sb.WriteString("## Statistics\n\n")
	sb.WriteString("| Type | Count |\n")
	sb.WriteString("|------|-------|\n")
	for pType, count := range stats {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", pType, count))
	}
	sb.WriteString(fmt.Sprintf("| **Total** | **%d** |\n", len(plugins)))
	sb.WriteString(fmt.Sprintf("\n**Total size:** %s\n", formatSize(totalSize)))
	sb.WriteString(fmt.Sprintf("**JuiceBar plugins:** %d\n\n", juicebarCount))

	// Detailed list
	sb.WriteString("## Plugin Details\n\n")
	for i, p := range plugins {
		if i >= 50 {
			sb.WriteString(fmt.Sprintf("\n... and %d more plugins\n", len(plugins)-50))
			break
		}
		sb.WriteString(fmt.Sprintf("### %s\n", p.Name))
		sb.WriteString(fmt.Sprintf("- **File:** `%s`\n", p.Filename))
		sb.WriteString(fmt.Sprintf("- **Type:** %s\n", p.Type))
		sb.WriteString(fmt.Sprintf("- **Source:** %s\n", p.Source))
		sb.WriteString(fmt.Sprintf("- **Size:** %s\n", formatSize(p.Size)))
		sb.WriteString(fmt.Sprintf("- **Path:** `%s`\n", p.LocalPath))
		if len(p.SHA256) >= 16 {
			sb.WriteString(fmt.Sprintf("- **SHA256:** `%s`\n", p.SHA256[:16]+"..."))
		} else if p.SHA256 != "" {
			sb.WriteString(fmt.Sprintf("- **SHA256:** `%s`\n", p.SHA256))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleS3List(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	prefix := tools.GetStringParam(req, "prefix")
	plugins, err := client.ListS3Plugins(ctx, prefix)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(plugins) == 0 {
		return tools.TextResult("No plugins found in S3 bucket."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# S3 Plugin Catalog (%d)\n\n", len(plugins)))
	sb.WriteString("| Name | Type | Size | S3 Key |\n")
	sb.WriteString("|------|------|------|--------|\n")

	for _, p := range plugins {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | `%s` |\n",
			p.Name,
			p.Type,
			formatSize(p.Size),
			p.S3Key,
		))
	}

	return tools.TextResult(sb.String()), nil
}

func handleUpload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	// Parse tags
	var tags []string
	if tagsStr := tools.GetStringParam(req, "tags"); tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	opts := clients.UploadOptions{
		Version:    tools.GetStringParam(req, "version"),
		Tags:       tags,
		IsJuiceBar: tools.GetBoolParam(req, "is_juicebar", false),
		Source:     clients.ResolumePluginSourceLocal,
	}

	result, err := client.UploadPlugin(ctx, path, opts)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Plugin Uploaded\n\n")
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", result.Name))
	sb.WriteString(fmt.Sprintf("**S3 Key:** `%s`\n", result.S3Key))
	sb.WriteString(fmt.Sprintf("**Size:** %s\n", formatSize(result.Size)))
	if len(result.SHA256) >= 32 {
		sb.WriteString(fmt.Sprintf("**SHA256:** `%s`\n", result.SHA256[:32]+"..."))
	} else if result.SHA256 != "" {
		sb.WriteString(fmt.Sprintf("**SHA256:** `%s`\n", result.SHA256))
	}

	return tools.TextResult(sb.String()), nil
}

func handleDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	key, errResult := tools.RequireStringParam(req, "key")
	if errResult != nil {
		return errResult, nil
	}

	destPath, err := client.DownloadPlugin(ctx, key)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("**Downloaded and installed:** `%s`\n\nRestart Resolume to load the new plugin.", destPath)), nil
}

func handleSync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.SyncPlugins(ctx, "compare")
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Plugin Sync Status\n\n")

	if len(result.ToUpload) > 0 {
		sb.WriteString(fmt.Sprintf("## Local Only (%d) → Upload to S3\n\n", len(result.ToUpload)))
		for _, p := range result.ToUpload {
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", p.Name, p.Type))
		}
		sb.WriteString("\n")
	}

	if len(result.ToDownload) > 0 {
		sb.WriteString(fmt.Sprintf("## S3 Only (%d) → Download to Local\n\n", len(result.ToDownload)))
		for _, p := range result.ToDownload {
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", p.Name, p.Type))
		}
		sb.WriteString("\n")
	}

	if len(result.LocalNewer) > 0 {
		sb.WriteString(fmt.Sprintf("## Local Newer (%d) → Re-upload\n\n", len(result.LocalNewer)))
		for _, p := range result.LocalNewer {
			sb.WriteString(fmt.Sprintf("- %s\n", p.Name))
		}
		sb.WriteString("\n")
	}

	if len(result.S3Newer) > 0 {
		sb.WriteString(fmt.Sprintf("## S3 Newer (%d) → Re-download\n\n", len(result.S3Newer)))
		for _, p := range result.S3Newer {
			sb.WriteString(fmt.Sprintf("- %s\n", p.Name))
		}
		sb.WriteString("\n")
	}

	if len(result.ToUpload) == 0 && len(result.ToDownload) == 0 &&
		len(result.LocalNewer) == 0 && len(result.S3Newer) == 0 {
		sb.WriteString("**All plugins are in sync!**\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleBackup(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	includeJuiceBar := tools.GetBoolParam(req, "include_juicebar", true)
	dryRun := tools.GetBoolParam(req, "dry_run", false)

	if dryRun {
		// Just show what would be uploaded
		plugins, err := client.ScanLocalPlugins(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		var toUpload []clients.ResolumePluginInfo
		for _, p := range plugins {
			if p.IsJuiceBar && !includeJuiceBar {
				continue
			}
			if strings.Contains(p.LocalPath, "/Applications/") ||
				strings.Contains(p.LocalPath, "Program Files") {
				continue
			}
			toUpload = append(toUpload, p)
		}

		var sb strings.Builder
		sb.WriteString("# Backup Preview (Dry Run)\n\n")
		sb.WriteString(fmt.Sprintf("**Would upload:** %d plugins\n\n", len(toUpload)))

		for _, p := range toUpload {
			sb.WriteString(fmt.Sprintf("- %s (%s, %s)\n", p.Name, p.Type, formatSize(p.Size)))
		}

		return tools.TextResult(sb.String()), nil
	}

	result, err := client.BackupAllPlugins(ctx, includeJuiceBar)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Backup Complete\n\n")
	sb.WriteString(fmt.Sprintf("**Uploaded:** %d plugins\n\n", len(result.Uploaded)))

	if len(result.Uploaded) > 0 {
		sb.WriteString("## Successful\n\n")
		for _, p := range result.Uploaded {
			sb.WriteString(fmt.Sprintf("- %s → `%s`\n", p.Name, p.S3Key))
		}
		sb.WriteString("\n")
	}

	if len(result.Errors) > 0 {
		sb.WriteString("## Errors\n\n")
		for _, e := range result.Errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e.Error()))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleInstall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	destPath, err := client.InstallPlugin(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("**Installed:** `%s`\n\nRestart Resolume to load the new plugin.", destPath)), nil
}

func handleUninstall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.UninstallPlugin(ctx, name); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("**Uninstalled:** %s\n\nRestart Resolume to apply changes.", name)), nil
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

	var sb strings.Builder
	sb.WriteString("# Resolume Plugin Health\n\n")

	statusEmoji := "✅"
	if health.Status == "degraded" {
		statusEmoji = "⚠️"
	} else if health.Status == "critical" {
		statusEmoji = "❌"
	}

	sb.WriteString(fmt.Sprintf("**Health Score:** %d/100 %s\n", health.Score, statusEmoji))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", health.Status))

	sb.WriteString("## Metrics\n\n")
	sb.WriteString(fmt.Sprintf("- **Local plugins:** %d\n", health.LocalCount))
	sb.WriteString(fmt.Sprintf("- **S3 plugins:** %d\n", health.S3Count))

	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
	}

	if len(health.Recommendations) > 0 {
		sb.WriteString("\n## Recommendations\n\n")
		for _, rec := range health.Recommendations {
			sb.WriteString(fmt.Sprintf("- %s\n", rec))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handlePaths(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	paths := client.GetPluginPaths()

	var sb strings.Builder
	sb.WriteString("# Resolume Plugin Paths\n\n")

	sb.WriteString("| Location | Path | Exists |\n")
	sb.WriteString("|----------|------|--------|\n")

	for name, path := range paths {
		exists := "❌"
		if _, err := os.Stat(path); err == nil {
			exists = "✅"
		}
		sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", name, path, exists))
	}

	return tools.TextResult(sb.String()), nil
}

// Helper functions

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatSource(source clients.ResolumePluginSource) string {
	switch source {
	case clients.ResolumePluginSourceJuiceBar:
		return "JuiceBar"
	case clients.ResolumePluginSourceSpackOMat:
		return "Spack-O-Mat"
	case clients.ResolumePluginSourceISF:
		return "ISF Shaders"
	case clients.ResolumePluginSourceS3:
		return "S3 Cloud"
	default:
		return "Local/Custom"
	}
}

// New handler functions

func handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	query := tools.GetStringParam(req, "query")
	pluginType := tools.GetStringParam(req, "type")

	var tags []string
	if tagsStr := tools.GetStringParam(req, "tags"); tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			tags = append(tags, strings.TrimSpace(t))
		}
	}

	plugins, err := client.SearchS3Plugins(ctx, query, pluginType, tags)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(plugins) == 0 {
		return tools.TextResult("No plugins found matching search criteria."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Search Results (%d)\n\n", len(plugins)))
	sb.WriteString("| Name | Type | Size | S3 Key |\n")
	sb.WriteString("|------|------|------|--------|\n")

	for _, p := range plugins {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | `%s` |\n",
			p.Name, p.Type, formatSize(p.Size), p.S3Key))
	}

	return tools.TextResult(sb.String()), nil
}

func handleBatchUpload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	pathsStr, errResult := tools.RequireStringParam(req, "paths")
	if errResult != nil {
		return errResult, nil
	}

	var paths []string
	for _, p := range strings.Split(pathsStr, ",") {
		paths = append(paths, strings.TrimSpace(p))
	}

	opts := clients.UploadOptions{
		IsJuiceBar: tools.GetBoolParam(req, "is_juicebar", false),
		Source:     clients.ResolumePluginSourceLocal,
	}

	uploaded, errors := client.BatchUpload(ctx, paths, opts)

	var sb strings.Builder
	sb.WriteString("# Batch Upload Results\n\n")
	sb.WriteString(fmt.Sprintf("**Uploaded:** %d / %d\n\n", len(uploaded), len(paths)))

	if len(uploaded) > 0 {
		sb.WriteString("## Successful\n\n")
		for _, p := range uploaded {
			sb.WriteString(fmt.Sprintf("- %s → `%s`\n", p.Name, p.S3Key))
		}
		sb.WriteString("\n")
	}

	if len(errors) > 0 {
		sb.WriteString("## Errors\n\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e.Error()))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleBatchDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	keysStr, errResult := tools.RequireStringParam(req, "keys")
	if errResult != nil {
		return errResult, nil
	}

	var keys []string
	for _, k := range strings.Split(keysStr, ",") {
		keys = append(keys, strings.TrimSpace(k))
	}

	downloaded, errors := client.BatchDownload(ctx, keys)

	var sb strings.Builder
	sb.WriteString("# Batch Download Results\n\n")
	sb.WriteString(fmt.Sprintf("**Downloaded:** %d / %d\n\n", len(downloaded), len(keys)))

	if len(downloaded) > 0 {
		sb.WriteString("## Successful\n\n")
		for _, path := range downloaded {
			sb.WriteString(fmt.Sprintf("- `%s`\n", path))
		}
		sb.WriteString("\n")
	}

	if len(errors) > 0 {
		sb.WriteString("## Errors\n\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e.Error()))
		}
	}

	sb.WriteString("\n**Restart Resolume to load the new plugins.**")

	return tools.TextResult(sb.String()), nil
}

func handleImportGDrive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	fileID, errResult := tools.RequireStringParam(req, "file_id")
	if errResult != nil {
		return errResult, nil
	}

	uploadToS3 := tools.GetBoolParam(req, "upload_to_s3", true)

	// Get Google Drive client
	gdriveClient, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	plugin, err := client.ImportFromGDrive(ctx, gdriveClient, fileID, uploadToS3)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Plugin Imported from Google Drive\n\n")
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", plugin.Name))
	sb.WriteString(fmt.Sprintf("**Installed to:** `%s`\n", plugin.LocalPath))

	if plugin.S3Key != "" {
		sb.WriteString(fmt.Sprintf("**Uploaded to S3:** `%s`\n", plugin.S3Key))
	}

	sb.WriteString("\n**Restart Resolume to load the new plugin.**")

	return tools.TextResult(sb.String()), nil
}

func handleImportGDriveFolder(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	folderID, errResult := tools.RequireStringParam(req, "folder_id")
	if errResult != nil {
		return errResult, nil
	}

	uploadToS3 := tools.GetBoolParam(req, "upload_to_s3", true)

	// Get Google Drive client
	gdriveClient, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	imported, errors := client.ImportFolderFromGDrive(ctx, gdriveClient, folderID, uploadToS3)

	var sb strings.Builder
	sb.WriteString("# Google Drive Folder Import Results\n\n")
	sb.WriteString(fmt.Sprintf("**Imported:** %d plugins\n\n", len(imported)))

	if len(imported) > 0 {
		sb.WriteString("## Successful\n\n")
		for _, p := range imported {
			if p.S3Key != "" {
				sb.WriteString(fmt.Sprintf("- %s → S3: `%s`\n", p.Name, p.S3Key))
			} else {
				sb.WriteString(fmt.Sprintf("- %s → `%s`\n", p.Name, p.LocalPath))
			}
		}
		sb.WriteString("\n")
	}

	if len(errors) > 0 {
		sb.WriteString("## Errors\n\n")
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e.Error()))
		}
	}

	sb.WriteString("\n**Restart Resolume to load the new plugins.**")

	return tools.TextResult(sb.String()), nil
}

func handleISFBrowse(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	category := tools.GetStringParam(req, "category")
	limit := tools.GetIntParam(req, "limit", 0)

	shaders, err := client.BrowseISFShaders(ctx, category, limit)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(shaders) == 0 {
		return tools.TextResult("No ISF shader collections found."), nil
	}

	var sb strings.Builder
	sb.WriteString("# ISF Shader Collections\n\n")
	sb.WriteString("These are curated ISF shader collections from the community.\n\n")

	for _, shader := range shaders {
		sb.WriteString(fmt.Sprintf("## %s\n", shader.Title))
		sb.WriteString(fmt.Sprintf("**Author:** %s\n", shader.Author))
		sb.WriteString(fmt.Sprintf("**Description:** %s\n", shader.Description))
		sb.WriteString(fmt.Sprintf("**Categories:** %s\n", strings.Join(shader.Categories, ", ")))
		sb.WriteString(fmt.Sprintf("**Download:** %s\n\n", shader.DownloadURL))
	}

	sb.WriteString("---\n")
	sb.WriteString("*Download the ZIP files and extract ISF shaders to your Resolume Extra Effects folder.*\n")

	return tools.TextResult(sb.String()), nil
}

func handleResolumeStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	info, err := client.GetResolumeInfo(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Resolume Status\n\n")

	running, ok := info["running"].(bool)
	if ok && running {
		sb.WriteString("**Status:** Connected\n")
		if version, ok := info["version"].(string); ok && version != "" {
			sb.WriteString(fmt.Sprintf("**Version:** %s\n", version))
		}
	} else {
		sb.WriteString("**Status:** Not Running\n")
	}

	sb.WriteString(fmt.Sprintf("**Platform:** %s\n\n", info["platform"]))

	// Show plugin paths
	if paths, ok := info["paths"].(map[string]string); ok {
		sb.WriteString("## Plugin Paths\n\n")
		sb.WriteString("| Location | Path | Exists |\n")
		sb.WriteString("|----------|------|--------|\n")

		for name, path := range paths {
			exists := "No"
			if _, err := os.Stat(path); err == nil {
				exists = "Yes"
			}
			sb.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", name, path, exists))
		}
	}

	return tools.TextResult(sb.String()), nil
}
