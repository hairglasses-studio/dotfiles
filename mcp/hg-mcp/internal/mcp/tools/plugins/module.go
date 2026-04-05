// Package plugins provides GitHub-based plugin registry tools for hg-mcp.
package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for plugin registry
type Module struct{}

func (m *Module) Name() string {
	return "plugins"
}

func (m *Module) Description() string {
	return "GitHub-based plugin registry for TouchDesigner and Resolume plugins"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_plugin_list",
				mcp.WithDescription("List all available plugins from the registry."),
				mcp.WithString("filter", mcp.Description("Filter by name, type, or tag")),
				mcp.WithString("type", mcp.Description("Filter by plugin type: touchdesigner_tox, touchdesigner_component, resolume_effect, resolume_source")),
			),
			Handler:             handleList,
			Category:            "plugins",
			Subcategory:         "discovery",
			Tags:                []string{"plugins", "list", "registry", "touchdesigner", "resolume"},
			UseCases:            []string{"Browse available plugins", "Find TD or Resolume plugins"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "plugins",
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_search",
				mcp.WithDescription("Search for plugins by keyword."),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query (matches name, description, tags)")),
			),
			Handler:             handleSearch,
			Category:            "plugins",
			Subcategory:         "discovery",
			Tags:                []string{"plugins", "search", "find"},
			UseCases:            []string{"Find plugins by keyword", "Search for specific functionality"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "plugins",
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_info",
				mcp.WithDescription("Get detailed information about a plugin."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name")),
			),
			Handler:             handleInfo,
			Category:            "plugins",
			Subcategory:         "discovery",
			Tags:                []string{"plugins", "info", "details"},
			UseCases:            []string{"View plugin details", "Check plugin compatibility"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "plugins",
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_versions",
				mcp.WithDescription("List available versions for a plugin."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name")),
			),
			Handler:             handleVersions,
			Category:            "plugins",
			Subcategory:         "discovery",
			Tags:                []string{"plugins", "versions", "releases"},
			UseCases:            []string{"Check available versions", "View release history"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "plugins",
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_install",
				mcp.WithDescription("Install a plugin from the registry."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name to install")),
				mcp.WithString("version", mcp.Description("Version to install (default: latest)")),
			),
			Handler:             handleInstall,
			Category:            "plugins",
			Subcategory:         "installation",
			Tags:                []string{"plugins", "install", "download"},
			UseCases:            []string{"Install a plugin", "Add new functionality"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "plugins",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_uninstall",
				mcp.WithDescription("Uninstall an installed plugin."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name to uninstall")),
			),
			Handler:             handleUninstall,
			Category:            "plugins",
			Subcategory:         "installation",
			Tags:                []string{"plugins", "uninstall", "remove"},
			UseCases:            []string{"Remove a plugin", "Clean up unused plugins"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "plugins",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_update",
				mcp.WithDescription("Update a plugin to the latest version."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name to update")),
			),
			Handler:             handleUpdate,
			Category:            "plugins",
			Subcategory:         "installation",
			Tags:                []string{"plugins", "update", "upgrade"},
			UseCases:            []string{"Update plugin to latest", "Get new features"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "plugins",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_installed",
				mcp.WithDescription("List all installed plugins."),
			),
			Handler:             handleInstalled,
			Category:            "plugins",
			Subcategory:         "status",
			Tags:                []string{"plugins", "installed", "local"},
			UseCases:            []string{"View installed plugins", "Check what's installed"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "plugins",
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_check_updates",
				mcp.WithDescription("Check for available plugin updates."),
			),
			Handler:             handleCheckUpdates,
			Category:            "plugins",
			Subcategory:         "status",
			Tags:                []string{"plugins", "updates", "check"},
			UseCases:            []string{"Check for updates", "See available upgrades"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "plugins",
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_sources",
				mcp.WithDescription("List configured plugin sources."),
			),
			Handler:             handleSources,
			Category:            "plugins",
			Subcategory:         "config",
			Tags:                []string{"plugins", "sources", "config"},
			UseCases:            []string{"View plugin sources", "Check configured repositories"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "plugins",
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_refresh",
				mcp.WithDescription("Refresh plugin cache from sources."),
			),
			Handler:             handleRefresh,
			Category:            "plugins",
			Subcategory:         "maintenance",
			Tags:                []string{"plugins", "refresh", "cache"},
			UseCases:            []string{"Refresh plugin list", "Update cache"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "plugins",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_plugin_health",
				mcp.WithDescription("Check plugin system health."),
			),
			Handler:             handleHealth,
			Category:            "plugins",
			Subcategory:         "health",
			Tags:                []string{"plugins", "health", "status"},
			UseCases:            []string{"Check plugin system health", "Diagnose issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "plugins",
		},
	}
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Helper function
var getClient = tools.LazyClient(clients.NewPluginClient)

// Handlers

func handleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	filter := tools.GetStringParam(req, "filter")
	pluginType := tools.GetStringParam(req, "type")

	if pluginType != "" && filter == "" {
		filter = pluginType
	}

	plugins, err := client.ListPlugins(ctx, filter)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(plugins) == 0 {
		return tools.TextResult("No plugins found"), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d plugins:\n\n", len(plugins)))

	for _, p := range plugins {
		status := ""
		if p.Installed {
			status = " [installed]"
		}
		if p.UpdateAvail {
			status = " [update available]"
		}
		sb.WriteString(fmt.Sprintf("- %s (%s)%s\n", p.Name, p.Type, status))
		sb.WriteString(fmt.Sprintf("  %s\n", p.Description))
		if len(p.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("  Tags: %s\n", strings.Join(p.Tags, ", ")))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	plugins, err := client.SearchPlugins(ctx, query)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(plugins) == 0 {
		return tools.TextResult(fmt.Sprintf("No plugins found matching '%s'", query)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d plugins matching '%s':\n\n", len(plugins), query))

	for _, p := range plugins {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", p.Name, p.Type))
		sb.WriteString(fmt.Sprintf("  %s\n\n", p.Description))
	}

	return tools.TextResult(sb.String()), nil
}

func handleInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	plugin, err := client.GetPlugin(ctx, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(plugin), nil
}

func handleVersions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	versions, err := client.GetPluginVersions(ctx, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(versions) == 0 {
		return tools.TextResult(fmt.Sprintf("No releases found for %s", name)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Versions for %s:\n\n", name))

	for _, v := range versions {
		prerelease := ""
		if v.Prerelease {
			prerelease = " (pre-release)"
		}
		sb.WriteString(fmt.Sprintf("- %s%s\n", v.Version, prerelease))
		sb.WriteString(fmt.Sprintf("  Released: %s\n", v.ReleaseDate.Format("2006-01-02")))
		if len(v.Assets) > 0 {
			sb.WriteString(fmt.Sprintf("  Assets: %s\n", strings.Join(v.Assets, ", ")))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleInstall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	version := tools.OptionalStringParam(req, "version", "latest")

	result, err := client.InstallPlugin(ctx, name, version)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if result.Success {
		return tools.TextResult(fmt.Sprintf("Successfully installed %s v%s to %s",
			result.Plugin, result.Version, result.InstallPath)), nil
	}

	return tools.ErrorResult(fmt.Errorf("installation failed: %s", result.Error)), nil
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

	return tools.TextResult(fmt.Sprintf("Successfully uninstalled %s", name)), nil
}

func handleUpdate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	result, err := client.UpdatePlugin(ctx, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if result.Success {
		return tools.TextResult(fmt.Sprintf("Successfully updated %s to v%s",
			result.Plugin, result.Version)), nil
	}

	return tools.ErrorResult(fmt.Errorf("update failed: %s", result.Error)), nil
}

func handleInstalled(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	installed, err := client.ListInstalled(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(installed) == 0 {
		return tools.TextResult("No plugins installed"), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Installed plugins (%d):\n\n", len(installed)))

	for _, p := range installed {
		sb.WriteString(fmt.Sprintf("- %s v%s (%s)\n", p.Name, p.Version, p.Type))
		sb.WriteString(fmt.Sprintf("  Path: %s\n", p.InstallPath))
		sb.WriteString(fmt.Sprintf("  Installed: %s\n\n", p.InstalledAt.Format("2006-01-02 15:04")))
	}

	return tools.TextResult(sb.String()), nil
}

func handleCheckUpdates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	updates, err := client.CheckUpdates(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(updates) == 0 {
		return tools.TextResult("All plugins are up to date"), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d plugin(s) have updates available:\n\n", len(updates)))

	for _, p := range updates {
		sb.WriteString(fmt.Sprintf("- %s: %s -> latest\n", p.Name, p.Version))
	}

	return tools.TextResult(sb.String()), nil
}

func handleSources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	sources, err := client.GetSources(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(sources) == 0 {
		return tools.TextResult("No sources configured"), nil
	}

	var sb strings.Builder
	sb.WriteString("Plugin sources:\n\n")

	for _, s := range sources {
		sb.WriteString(fmt.Sprintf("- %s\n", s))
	}

	return tools.TextResult(sb.String()), nil
}

func handleRefresh(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Reload manifest
	manifest, err := client.LoadManifest(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Refreshed plugin cache. %d plugins available.", len(manifest.Plugins))), nil
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
