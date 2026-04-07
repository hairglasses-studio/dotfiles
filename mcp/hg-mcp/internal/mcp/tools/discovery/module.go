// Package discovery provides tool discovery and catalog features for hg-mcp.
package discovery

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Cached tool data for performance
var (
	cachedToolList    []tools.ToolDefinition
	cachedToolMap     map[string]tools.ToolDefinition
	cachedSortedTools []tools.ToolDefinition
	cacheOnce         sync.Once
)

// initCache initializes the cached tool data
func initCache() {
	cacheOnce.Do(func() {
		registry := tools.GetRegistry()
		cachedToolList = registry.GetAllToolDefinitions()

		// Build lookup map
		cachedToolMap = make(map[string]tools.ToolDefinition, len(cachedToolList))
		for _, td := range cachedToolList {
			cachedToolMap[td.Tool.Name] = td
		}

		// Build sorted list (excluding discovery tools)
		cachedSortedTools = make([]tools.ToolDefinition, 0, len(cachedToolList))
		for _, td := range cachedToolList {
			if td.Category != "discovery" {
				cachedSortedTools = append(cachedSortedTools, td)
			}
		}
		sort.Slice(cachedSortedTools, func(i, j int) bool {
			if cachedSortedTools[i].Category != cachedSortedTools[j].Category {
				return cachedSortedTools[i].Category < cachedSortedTools[j].Category
			}
			return cachedSortedTools[i].Tool.Name < cachedSortedTools[j].Tool.Name
		})
	})
}

func getCachedTools() []tools.ToolDefinition {
	initCache()
	return cachedSortedTools
}

func getCachedToolMap() map[string]tools.ToolDefinition {
	initCache()
	return cachedToolMap
}

// Module implements the ToolModule interface for tool discovery
type Module struct{}

func (m *Module) Name() string {
	return "discovery"
}

func (m *Module) Description() string {
	return "Tool discovery and catalog features"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// Progressive discovery tools (cobb pattern)
		{
			Tool: mcp.NewTool("aftrs_tool_discover",
				mcp.WithDescription("Browse tools with configurable detail level for token efficiency. Use 'names' for ~500 tokens, 'signatures' for ~800, 'descriptions' for ~2000, or 'full' for complete schemas."),
				mcp.WithString("detail_level",
					mcp.Description("Detail level: 'names' (tool names only), 'signatures' (name + required params), 'descriptions' (name + description), 'full' (complete schemas). Default: descriptions"),
				),
				mcp.WithString("category",
					mcp.Description("Filter by category (e.g., 'discord', 'ableton', 'resolume')"),
				),
				mcp.WithString("runtime_group",
					mcp.Description("Filter by runtime group: dj_music, vj_video, lighting, audio_production, show_control, infrastructure, messaging, inventory, streaming, platform"),
				),
				mcp.WithString("search",
					mcp.Description("Filter tools by name or description keyword"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Maximum tools to return (default: 50, max: 200)"),
				),
				mcp.WithNumber("offset",
					mcp.Description("Offset for pagination (default: 0)"),
				),
			),
			Handler:             handleToolDiscover,
			Category:            "discovery",
			Subcategory:         "progressive",
			Tags:                []string{"discovery", "browse", "progressive", "token-efficient"},
			UseCases:            []string{"Browse tools efficiently", "Progressive detail loading", "Token-optimized discovery"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
			OutputSchema: tools.ObjectOutputSchema(map[string]interface{}{
				"detail_level": map[string]interface{}{
					"type":        "string",
					"description": "Detail level used (names, signatures, descriptions, full)",
				},
				"showing": map[string]interface{}{
					"type":        "string",
					"description": "Range of results shown (e.g., 1-50 of 880)",
				},
				"total": map[string]interface{}{
					"type":        "integer",
					"description": "Total number of matching tools",
				},
				"tools": map[string]interface{}{
					"type":        "array",
					"description": "List of discovered tools",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name":          map[string]interface{}{"type": "string", "description": "Tool name"},
							"description":   map[string]interface{}{"type": "string", "description": "Tool description (when detail_level >= descriptions)"},
							"category":      map[string]interface{}{"type": "string", "description": "Tool category (when detail_level = full)"},
							"runtime_group": map[string]interface{}{"type": "string", "description": "Runtime group (when detail_level = full)"},
							"required":      map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Required parameters (when detail_level >= signatures)"},
						},
					},
				},
			}, []string{"total", "tools"}),
		},
		{
			Tool: mcp.NewTool("aftrs_tool_schema",
				mcp.WithDescription("Get complete schema for specific tool(s) on-demand. Use after aftrs_tool_discover to load only needed schemas."),
				mcp.WithString("tool_names",
					mcp.Required(),
					mcp.Description("Comma-separated tool names (e.g., 'aftrs_gmail_send,aftrs_discord_post')"),
				),
				mcp.WithBoolean("required_only",
					mcp.Description("Only show required parameters (saves 20-30% tokens). Default: false"),
				),
			),
			Handler:             handleToolSchema,
			Category:            "discovery",
			Subcategory:         "progressive",
			Tags:                []string{"discovery", "schema", "on-demand"},
			UseCases:            []string{"Load full schema on-demand", "Get parameter details"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
		},
		{
			Tool: mcp.NewTool("aftrs_tool_search",
				mcp.WithDescription("Search for tools by keyword. Returns matching tools with descriptions."),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("Search query (e.g., 'discord', 'status', 'send message')"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Maximum results to return (default: 10, max: 50)"),
				),
			),
			Handler:             handleToolSearch,
			Category:            "discovery",
			Subcategory:         "search",
			Tags:                []string{"discovery", "search", "help"},
			UseCases:            []string{"Find tools by keyword", "Discover available functionality"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
		},
		{
			Tool: mcp.NewTool("aftrs_tool_catalog",
				mcp.WithDescription("Browse the complete tool catalog organized by category."),
				mcp.WithString("category",
					mcp.Description("Filter by category (e.g., 'discord', 'touchdesigner')"),
				),
				mcp.WithString("format",
					mcp.Description("Output format: 'compact' (names only) or 'full' (with descriptions). Default: compact"),
				),
			),
			Handler:             handleToolCatalog,
			Category:            "discovery",
			Subcategory:         "browse",
			Tags:                []string{"discovery", "catalog", "browse"},
			UseCases:            []string{"Browse available tools", "Explore categories"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
		},
		{
			Tool: mcp.NewTool("aftrs_tool_stats",
				mcp.WithDescription("Get statistics about the tool registry: total tools, categories, and modules."),
			),
			Handler:             handleToolStats,
			Category:            "discovery",
			Subcategory:         "stats",
			Tags:                []string{"discovery", "stats"},
			UseCases:            []string{"Understand tool distribution", "Get registry overview"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
		},
		// Phase 29A: Enhanced Tool Discovery (6 new tools)
		{
			Tool: mcp.NewTool("aftrs_tools_related",
				mcp.WithDescription("Find tools related to a given tool by shared category and tags."),
				mcp.WithString("tool_name",
					mcp.Required(),
					mcp.Description("Name of the tool to find relations for"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Maximum related tools to return (default: 10)"),
				),
			),
			Handler:             handleToolsRelated,
			Category:            "discovery",
			Subcategory:         "navigation",
			Tags:                []string{"discovery", "related", "navigation"},
			UseCases:            []string{"Find similar tools", "Discover related functionality"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
		},
		{
			Tool: mcp.NewTool("aftrs_tools_workflow",
				mcp.WithDescription("Get recommended tool sequences for common goals and workflows."),
				mcp.WithString("goal",
					mcp.Description("Goal or workflow name (e.g., 'bpm_sync', 'troubleshoot_audio'). Leave empty to list all."),
				),
			),
			Handler:             handleToolsWorkflow,
			Category:            "discovery",
			Subcategory:         "navigation",
			Tags:                []string{"discovery", "workflow", "sequence"},
			UseCases:            []string{"Get recommended tool order", "Learn common patterns"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
		},
		{
			Tool: mcp.NewTool("aftrs_tools_by_system",
				mcp.WithDescription("List all tools for a specific system (resolume, ableton, obs, etc.)."),
				mcp.WithString("system",
					mcp.Description("System name (e.g., 'resolume', 'ableton', 'obs'). Leave empty to list systems."),
				),
			),
			Handler:             handleToolsBySystem,
			Category:            "discovery",
			Subcategory:         "navigation",
			Tags:                []string{"discovery", "system", "filter"},
			UseCases:            []string{"Find all tools for a system", "Explore system capabilities"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
		},
		{
			Tool: mcp.NewTool("aftrs_tools_recent",
				mcp.WithDescription("List recently used tools in this session."),
				mcp.WithNumber("limit",
					mcp.Description("Maximum tools to return (default: 10)"),
				),
			),
			Handler:             handleToolsRecent,
			Category:            "discovery",
			Subcategory:         "session",
			Tags:                []string{"discovery", "recent", "history"},
			UseCases:            []string{"See recent tool usage", "Quick access to used tools"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
		},
		{
			Tool: mcp.NewTool("aftrs_tools_favorites",
				mcp.WithDescription("Manage favorite tools for quick access."),
				mcp.WithString("action",
					mcp.Description("Action: 'list' (default), 'add', or 'remove'"),
				),
				mcp.WithString("tool_name",
					mcp.Description("Tool name (required for add/remove)"),
				),
			),
			Handler:             handleToolsFavorites,
			Category:            "discovery",
			Subcategory:         "session",
			Tags:                []string{"discovery", "favorites", "quick-access"},
			UseCases:            []string{"Mark frequently used tools", "Quick access to favorites"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
		},
		{
			Tool: mcp.NewTool("aftrs_tools_alias",
				mcp.WithDescription("Create short aliases for frequently used tools."),
				mcp.WithString("action",
					mcp.Description("Action: 'list' (default), 'set', 'remove', or 'resolve'"),
				),
				mcp.WithString("alias",
					mcp.Description("Short alias name (required for set/remove/resolve)"),
				),
				mcp.WithString("tool_name",
					mcp.Description("Full tool name (required for 'set' action)"),
				),
			),
			Handler:             handleToolsAlias,
			Category:            "discovery",
			Subcategory:         "session",
			Tags:                []string{"discovery", "alias", "shortcuts"},
			UseCases:            []string{"Create tool shortcuts", "Simplify tool names"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "discovery",
		},
	}
}

// handleToolDiscover handles progressive tool discovery with configurable detail levels
func handleToolDiscover(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	detailLevel := tools.OptionalStringParam(req, "detail_level", "descriptions")
	category := tools.GetStringParam(req, "category")
	runtimeGroup := tools.GetStringParam(req, "runtime_group")
	search := tools.GetStringParam(req, "search")
	limit := tools.GetIntParam(req, "limit", 50)
	offset := tools.GetIntParam(req, "offset", 0)

	if limit > 200 {
		limit = 200
	}
	if limit < 1 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Get filtered tools
	allTools := getCachedTools()
	var filtered []tools.ToolDefinition

	searchLower := strings.ToLower(search)

	for _, td := range allTools {
		// Category filter
		if category != "" && !strings.EqualFold(td.Category, category) {
			continue
		}
		// Runtime group filter
		if runtimeGroup != "" && !strings.EqualFold(td.RuntimeGroup, runtimeGroup) {
			continue
		}
		// Search filter
		if search != "" {
			nameLower := strings.ToLower(td.Tool.Name)
			descLower := strings.ToLower(td.Tool.Description)
			if !strings.Contains(nameLower, searchLower) && !strings.Contains(descLower, searchLower) {
				// Check tags
				found := false
				for _, tag := range td.Tags {
					if strings.Contains(strings.ToLower(tag), searchLower) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
		}
		filtered = append(filtered, td)
	}

	total := len(filtered)

	// Apply pagination
	if offset >= total {
		filtered = nil
	} else {
		end := offset + limit
		if end > total {
			end = total
		}
		filtered = filtered[offset:end]
	}

	var sb strings.Builder
	sb.WriteString("# Tool Discovery\n\n")
	sb.WriteString(fmt.Sprintf("**Detail Level:** %s | **Showing:** %d-%d of %d", detailLevel, offset+1, offset+len(filtered), total))
	if category != "" {
		sb.WriteString(fmt.Sprintf(" | **Category:** %s", category))
	}
	if runtimeGroup != "" {
		sb.WriteString(fmt.Sprintf(" | **RuntimeGroup:** %s", runtimeGroup))
	}
	if search != "" {
		sb.WriteString(fmt.Sprintf(" | **Search:** %s", search))
	}
	sb.WriteString("\n\n")

	if len(filtered) == 0 {
		sb.WriteString("No tools found matching criteria.\n")
		return tools.TextResult(sb.String()), nil
	}

	switch detailLevel {
	case "names":
		// ~500 tokens - names only
		sb.WriteString("**Tools:**\n")
		for _, td := range filtered {
			sb.WriteString(fmt.Sprintf("- %s\n", td.Tool.Name))
		}

	case "signatures":
		// ~800 tokens - name + required params
		for _, td := range filtered {
			sb.WriteString(fmt.Sprintf("**%s**", td.Tool.Name))
			// Extract required params from schema
			if td.Tool.InputSchema.Properties != nil {
				var required []string
				for _, reqName := range td.Tool.InputSchema.Required {
					required = append(required, reqName)
				}
				if len(required) > 0 {
					sb.WriteString(fmt.Sprintf("(%s)", strings.Join(required, ", ")))
				}
			}
			sb.WriteString("\n")
		}

	case "descriptions":
		// ~2000 tokens - name + description (default)
		for _, td := range filtered {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", td.Tool.Name, td.Tool.Description))
		}

	case "full":
		// Full schemas
		for _, td := range filtered {
			sb.WriteString(fmt.Sprintf("## %s\n", td.Tool.Name))
			sb.WriteString(fmt.Sprintf("**Category:** %s", td.Category))
			if td.Subcategory != "" {
				sb.WriteString(fmt.Sprintf(" / %s", td.Subcategory))
			}
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("**Description:** %s\n", td.Tool.Description))
			if len(td.Tags) > 0 {
				sb.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(td.Tags, ", ")))
			}
			sb.WriteString(fmt.Sprintf("**Complexity:** %s | **Write:** %v\n", td.Complexity, td.IsWrite))
			if td.Deprecated {
				sb.WriteString(fmt.Sprintf("**⚠️ DEPRECATED** → Use `%s` instead\n", td.Successor))
			}
			// Parameters
			if td.Tool.InputSchema.Properties != nil {
				sb.WriteString("\n**Parameters:**\n")
				for name, prop := range td.Tool.InputSchema.Properties {
					isRequired := false
					for _, reqName := range td.Tool.InputSchema.Required {
						if reqName == name {
							isRequired = true
							break
						}
					}
					reqStr := ""
					if isRequired {
						reqStr = " *(required)*"
					}
					// Extract description from property
					if propMap, ok := prop.(map[string]interface{}); ok {
						desc := ""
						if d, ok := propMap["description"].(string); ok {
							desc = d
						}
						sb.WriteString(fmt.Sprintf("- `%s`%s: %s\n", name, reqStr, desc))
					} else {
						sb.WriteString(fmt.Sprintf("- `%s`%s\n", name, reqStr))
					}
				}
			}
			sb.WriteString("\n")
		}

	default:
		return tools.ErrorResult(fmt.Errorf("invalid detail_level '%s'. Use: names, signatures, descriptions, full", detailLevel)), nil
	}

	// Pagination hint
	if offset+len(filtered) < total {
		sb.WriteString(fmt.Sprintf("\n---\n*Use `offset=%d` to see more*\n", offset+limit))
	}
	sb.WriteString("\n*Use `aftrs_tool_schema(tool_names=\"...\")` for full schemas on-demand*\n")

	return tools.TextResult(sb.String()), nil
}

// handleToolSchema returns complete schemas for specific tools on-demand
func handleToolSchema(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolNames, errResult := tools.RequireStringParam(req, "tool_names")
	if errResult != nil {
		return errResult, nil
	}
	requiredOnly := tools.GetBoolParam(req, "required_only", false)

	names := strings.Split(toolNames, ",")
	for i := range names {
		names[i] = strings.TrimSpace(names[i])
	}

	toolMap := getCachedToolMap()

	var sb strings.Builder
	sb.WriteString("# Tool Schemas\n\n")

	found := 0
	notFound := []string{}

	for _, name := range names {
		td, ok := toolMap[name]
		if !ok {
			notFound = append(notFound, name)
			continue
		}
		found++

		sb.WriteString(fmt.Sprintf("## %s\n\n", td.Tool.Name))
		sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", td.Tool.Description))
		sb.WriteString(fmt.Sprintf("**Category:** %s", td.Category))
		if td.Subcategory != "" {
			sb.WriteString(fmt.Sprintf(" / %s", td.Subcategory))
		}
		sb.WriteString("\n")
		if len(td.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(td.Tags, ", ")))
		}
		sb.WriteString(fmt.Sprintf("**Complexity:** %s | **Write:** %v\n", td.Complexity, td.IsWrite))
		if td.Deprecated {
			sb.WriteString(fmt.Sprintf("**⚠️ DEPRECATED** → Use `%s` instead\n", td.Successor))
		}
		if len(td.UseCases) > 0 {
			sb.WriteString(fmt.Sprintf("**Use Cases:** %s\n", strings.Join(td.UseCases, "; ")))
		}

		// Parameters
		if td.Tool.InputSchema.Properties != nil {
			sb.WriteString("\n### Parameters\n\n")
			requiredSet := make(map[string]bool)
			for _, reqName := range td.Tool.InputSchema.Required {
				requiredSet[reqName] = true
			}

			// Sort parameter names for consistent output
			paramNames := make([]string, 0, len(td.Tool.InputSchema.Properties))
			for name := range td.Tool.InputSchema.Properties {
				paramNames = append(paramNames, name)
			}
			sort.Strings(paramNames)

			for _, name := range paramNames {
				prop := td.Tool.InputSchema.Properties[name]
				isRequired := requiredSet[name]

				// Skip non-required params if requiredOnly
				if requiredOnly && !isRequired {
					continue
				}

				reqStr := ""
				if isRequired {
					reqStr = " **required**"
				}

				// Extract type and description from property
				propType := "any"
				desc := ""
				if propMap, ok := prop.(map[string]interface{}); ok {
					if t, ok := propMap["type"].(string); ok {
						propType = t
					}
					if d, ok := propMap["description"].(string); ok {
						desc = d
					}
					if enum, ok := propMap["enum"].([]interface{}); ok {
						enumStrs := make([]string, len(enum))
						for i, e := range enum {
							enumStrs[i] = fmt.Sprintf("%v", e)
						}
						desc += fmt.Sprintf(" (options: %s)", strings.Join(enumStrs, ", "))
					}
				}

				sb.WriteString(fmt.Sprintf("- `%s` (%s)%s: %s\n", name, propType, reqStr, desc))
			}
		} else {
			sb.WriteString("\n*No parameters*\n")
		}
		sb.WriteString("\n---\n\n")
	}

	if len(notFound) > 0 {
		sb.WriteString(fmt.Sprintf("**Not found:** %s\n", strings.Join(notFound, ", ")))
	}

	sb.WriteString(fmt.Sprintf("\n*Loaded %d/%d tool schemas*\n", found, len(names)))

	return tools.TextResult(sb.String()), nil
}

// handleToolSearch handles the aftrs_tool_search tool
func handleToolSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 10)
	if limit > 50 {
		limit = 50
	}
	if limit < 1 {
		limit = 10
	}

	registry := tools.GetRegistry()
	results := registry.SearchTools(query)

	if len(results) > limit {
		results = results[:limit]
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Tool Search: \"%s\"\n\n", query))
	sb.WriteString(fmt.Sprintf("Found %d matching tools:\n\n", len(results)))

	for _, r := range results {
		sb.WriteString(fmt.Sprintf("## %s\n", r.Tool.Tool.Name))
		sb.WriteString(fmt.Sprintf("**Category:** %s", r.Tool.Category))
		if r.Tool.Subcategory != "" {
			sb.WriteString(fmt.Sprintf(" / %s", r.Tool.Subcategory))
		}
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("**Description:** %s\n", r.Tool.Tool.Description))
		if len(r.Tool.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(r.Tool.Tags, ", ")))
		}
		sb.WriteString("\n")
	}

	if len(results) == 0 {
		sb.WriteString("No tools found matching your query.\n\n")
		sb.WriteString("Try:\n")
		sb.WriteString("- Different keywords\n")
		sb.WriteString("- `aftrs_tool_catalog` to browse all tools\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleToolCatalog handles the aftrs_tool_catalog tool
func handleToolCatalog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category := tools.GetStringParam(req, "category")
	format := tools.OptionalStringParam(req, "format", "compact")

	registry := tools.GetRegistry()
	catalog := registry.GetToolCatalog()

	var sb strings.Builder
	sb.WriteString("# Aftrs MCP Tool Catalog\n\n")

	stats := registry.GetToolStats()
	sb.WriteString(fmt.Sprintf("**Total:** %d tools across %d modules\n\n", stats.TotalTools, stats.ModuleCount))

	// Get sorted categories
	categories := make([]string, 0, len(catalog))
	for cat := range catalog {
		if category == "" || strings.EqualFold(cat, category) {
			categories = append(categories, cat)
		}
	}
	sort.Strings(categories)

	if len(categories) == 0 && category != "" {
		sb.WriteString(fmt.Sprintf("No category found matching '%s'\n\n", category))
		sb.WriteString("Available categories:\n")
		for cat := range catalog {
			sb.WriteString(fmt.Sprintf("- %s\n", cat))
		}
		return tools.TextResult(sb.String()), nil
	}

	for _, cat := range categories {
		subcats := catalog[cat]
		sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(cat)))

		// Get sorted subcategories
		subcatNames := make([]string, 0, len(subcats))
		for subcat := range subcats {
			subcatNames = append(subcatNames, subcat)
		}
		sort.Strings(subcatNames)

		for _, subcat := range subcatNames {
			toolList := subcats[subcat]
			if subcat != "general" {
				sb.WriteString(fmt.Sprintf("### %s\n", subcat))
			}

			for _, td := range toolList {
				if format == "full" {
					sb.WriteString(fmt.Sprintf("- **%s**: %s\n", td.Tool.Name, td.Tool.Description))
				} else {
					sb.WriteString(fmt.Sprintf("- %s\n", td.Tool.Name))
				}
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("---\n")
	sb.WriteString("Use `aftrs_tool_search(query=\"keyword\")` to search for specific tools.\n")

	return tools.TextResult(sb.String()), nil
}

// handleToolStats handles the aftrs_tool_stats tool
func handleToolStats(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	registry := tools.GetRegistry()
	stats := registry.GetToolStats()

	var sb strings.Builder
	sb.WriteString("# Tool Registry Statistics\n\n")

	sb.WriteString(fmt.Sprintf("**Total Tools:** %d\n", stats.TotalTools))
	sb.WriteString(fmt.Sprintf("**Total Modules:** %d\n\n", stats.ModuleCount))

	sb.WriteString("## Tools by Category\n\n")
	categories := make([]string, 0, len(stats.ByCategory))
	for cat := range stats.ByCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	for _, cat := range categories {
		count := stats.ByCategory[cat]
		sb.WriteString(fmt.Sprintf("- **%s:** %d tools\n", cat, count))
	}

	sb.WriteString("\n## Tools by Complexity\n\n")
	sb.WriteString(fmt.Sprintf("- **Simple:** %d\n", stats.ByComplexity["simple"]))
	sb.WriteString(fmt.Sprintf("- **Moderate:** %d\n", stats.ByComplexity["moderate"]))
	sb.WriteString(fmt.Sprintf("- **Complex:** %d\n", stats.ByComplexity["complex"]))

	sb.WriteString("\n## Read/Write Distribution\n\n")
	sb.WriteString(fmt.Sprintf("- **Read-only:** %d tools\n", stats.ReadOnlyCount))
	sb.WriteString(fmt.Sprintf("- **Write:** %d tools\n", stats.WriteToolsCount))

	sb.WriteString("\n---\n")
	sb.WriteString("Use `aftrs_tool_catalog` to browse all tools.\n")

	return tools.TextResult(sb.String()), nil
}

// handleToolsRelated finds tools related to a given tool
func handleToolsRelated(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolName, errResult := tools.RequireStringParam(req, "tool_name")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Get the source tool
	registry := tools.GetRegistry()
	sourceTool, ok := registry.GetTool(toolName)
	if !ok {
		return tools.ErrorResult(fmt.Errorf("tool '%s' not found", toolName)), nil
	}

	// Build tool info map for relation finding
	allTools := make(map[string]clients.ToolInfo)
	for _, td := range registry.GetAllToolDefinitions() {
		allTools[td.Tool.Name] = clients.ToolInfo{
			Category:    td.Category,
			Description: td.Tool.Description,
			Tags:        td.Tags,
		}
	}

	// Find related tools
	discoveryClient := clients.GetDiscoveryClient()
	relations := discoveryClient.FindRelatedTools(toolName, sourceTool.Category, sourceTool.Tags, allTools, limit)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Tools Related to %s\n\n", toolName))
	sb.WriteString(fmt.Sprintf("**Category:** %s\n", sourceTool.Category))
	if len(sourceTool.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("**Tags:** %s\n\n", strings.Join(sourceTool.Tags, ", ")))
	}

	if len(relations) == 0 {
		sb.WriteString("No related tools found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found %d related tools:\n\n", len(relations)))
	for _, rel := range relations {
		sb.WriteString(fmt.Sprintf("## %s\n", rel.ToolName))
		sb.WriteString(fmt.Sprintf("**Category:** %s | **Relevance:** %.0f%%\n", rel.Category, rel.Relevance*100))
		if len(rel.SharedTags) > 0 {
			sb.WriteString(fmt.Sprintf("**Shared Tags:** %s\n", strings.Join(rel.SharedTags, ", ")))
		}
		sb.WriteString(fmt.Sprintf("%s\n\n", rel.Description))
	}

	return tools.TextResult(sb.String()), nil
}

// handleToolsWorkflow shows recommended tool sequences for goals
func handleToolsWorkflow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	goal := tools.GetStringParam(req, "goal")

	discoveryClient := clients.GetDiscoveryClient()

	var sb strings.Builder
	sb.WriteString("# Tool Workflows\n\n")

	if goal == "" {
		// List all workflows
		workflows := discoveryClient.GetWorkflows()
		sb.WriteString("Available workflow templates:\n\n")
		for _, name := range workflows {
			toolSeq, _ := discoveryClient.GetWorkflow(name)
			sb.WriteString(fmt.Sprintf("## %s\n", name))
			sb.WriteString(fmt.Sprintf("**Tools:** %d steps\n", len(toolSeq)))
			for i, t := range toolSeq {
				sb.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, t))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("---\n")
		sb.WriteString("Use `aftrs_tools_workflow(goal=\"workflow_name\")` for details.\n")
	} else {
		// Search for matching workflows
		matches := discoveryClient.GetWorkflowsForGoal(goal)
		if len(matches) == 0 {
			// Try exact match
			toolSeq, ok := discoveryClient.GetWorkflow(goal)
			if ok {
				matches = append(matches, clients.WorkflowTemplate{
					Name:  goal,
					Tools: toolSeq,
				})
			}
		}

		if len(matches) == 0 {
			sb.WriteString(fmt.Sprintf("No workflows found matching '%s'\n\n", goal))
			sb.WriteString("Available workflows:\n")
			for _, name := range discoveryClient.GetWorkflows() {
				sb.WriteString(fmt.Sprintf("- %s\n", name))
			}
		} else {
			sb.WriteString(fmt.Sprintf("Workflows matching '%s':\n\n", goal))
			for _, wf := range matches {
				sb.WriteString(fmt.Sprintf("## %s\n", wf.Name))
				if wf.Description != "" {
					sb.WriteString(fmt.Sprintf("%s\n\n", wf.Description))
				}
				sb.WriteString("**Recommended sequence:**\n")
				for i, t := range wf.Tools {
					sb.WriteString(fmt.Sprintf("%d. `%s`\n", i+1, t))
				}
				if len(wf.Tags) > 0 {
					sb.WriteString(fmt.Sprintf("\n**Tags:** %s\n", strings.Join(wf.Tags, ", ")))
				}
				sb.WriteString("\n")
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleToolsBySystem lists tools for a specific system
func handleToolsBySystem(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	system := tools.GetStringParam(req, "system")

	discoveryClient := clients.GetDiscoveryClient()

	var sb strings.Builder
	sb.WriteString("# Tools by System\n\n")

	if system == "" {
		// List all systems
		systems := discoveryClient.GetSystems()
		sb.WriteString("Available systems:\n\n")
		sb.WriteString("| System | Tools |\n")
		sb.WriteString("|--------|-------|\n")
		for _, sys := range systems {
			toolList := discoveryClient.GetToolsForSystem(sys)
			sb.WriteString(fmt.Sprintf("| **%s** | %d |\n", sys, len(toolList)))
		}
		sb.WriteString("\n---\n")
		sb.WriteString("Use `aftrs_tools_by_system(system=\"name\")` to see tools for a system.\n")
	} else {
		toolList := discoveryClient.GetToolsForSystem(system)
		if toolList == nil {
			sb.WriteString(fmt.Sprintf("System '%s' not found.\n\n", system))
			sb.WriteString("Available systems:\n")
			for _, sys := range discoveryClient.GetSystems() {
				sb.WriteString(fmt.Sprintf("- %s\n", sys))
			}
		} else {
			sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(system)))
			sb.WriteString(fmt.Sprintf("**%d tools available:**\n\n", len(toolList)))

			registry := tools.GetRegistry()
			for _, toolName := range toolList {
				if td, ok := registry.GetTool(toolName); ok {
					sb.WriteString(fmt.Sprintf("- **%s**: %s\n", toolName, td.Tool.Description))
				} else {
					sb.WriteString(fmt.Sprintf("- %s\n", toolName))
				}
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleToolsRecent shows recently used tools
func handleToolsRecent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(req, "limit", 10)
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	discoveryClient := clients.GetDiscoveryClient()
	recent := discoveryClient.GetRecentTools(limit)

	var sb strings.Builder
	sb.WriteString("# Recently Used Tools\n\n")

	if len(recent) == 0 {
		sb.WriteString("No tools have been used yet in this session.\n\n")
		sb.WriteString("Tool usage is tracked automatically. Use any tool and it will appear here.\n")
	} else {
		sb.WriteString(fmt.Sprintf("Last %d tools used:\n\n", len(recent)))
		sb.WriteString("| Tool | Category | Time |\n")
		sb.WriteString("|------|----------|------|\n")
		for _, entry := range recent {
			ago := formatTimeAgo(entry.Timestamp)
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", entry.ToolName, entry.Category, ago))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleToolsFavorites manages favorite tools
func handleToolsFavorites(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.OptionalStringParam(req, "action", "list")
	toolName := tools.GetStringParam(req, "tool_name")

	discoveryClient := clients.GetDiscoveryClient()
	registry := tools.GetRegistry()

	var sb strings.Builder
	sb.WriteString("# Favorite Tools\n\n")

	switch action {
	case "list":
		favorites := discoveryClient.GetFavorites()
		if len(favorites) == 0 {
			sb.WriteString("No favorite tools set.\n\n")
			sb.WriteString("Add favorites with `aftrs_tools_favorites(action=\"add\", tool_name=\"tool_name\")`\n")
		} else {
			sb.WriteString(fmt.Sprintf("**%d favorite tools:**\n\n", len(favorites)))
			for _, name := range favorites {
				if td, ok := registry.GetTool(name); ok {
					sb.WriteString(fmt.Sprintf("- **%s**: %s\n", name, td.Tool.Description))
				} else {
					sb.WriteString(fmt.Sprintf("- %s (not found)\n", name))
				}
			}
		}

	case "add":
		if toolName == "" {
			return tools.ErrorResult(fmt.Errorf("tool_name is required for 'add' action")), nil
		}
		if _, ok := registry.GetTool(toolName); !ok {
			return tools.ErrorResult(fmt.Errorf("tool '%s' not found", toolName)), nil
		}
		discoveryClient.AddFavorite(toolName)
		sb.WriteString(fmt.Sprintf("✓ Added **%s** to favorites.\n", toolName))

	case "remove":
		if toolName == "" {
			return tools.ErrorResult(fmt.Errorf("tool_name is required for 'remove' action")), nil
		}
		discoveryClient.RemoveFavorite(toolName)
		sb.WriteString(fmt.Sprintf("✓ Removed **%s** from favorites.\n", toolName))

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action '%s'. Use 'list', 'add', or 'remove'", action)), nil
	}

	return tools.TextResult(sb.String()), nil
}

// handleToolsAlias manages tool aliases
func handleToolsAlias(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.OptionalStringParam(req, "action", "list")
	alias := tools.GetStringParam(req, "alias")
	toolName := tools.GetStringParam(req, "tool_name")

	discoveryClient := clients.GetDiscoveryClient()
	registry := tools.GetRegistry()

	var sb strings.Builder
	sb.WriteString("# Tool Aliases\n\n")

	switch action {
	case "list":
		aliases := discoveryClient.GetAliases()
		if len(aliases) == 0 {
			sb.WriteString("No aliases defined.\n\n")
			sb.WriteString("Create aliases with `aftrs_tools_alias(action=\"set\", alias=\"short\", tool_name=\"full_name\")`\n")
		} else {
			sb.WriteString("| Alias | Tool |\n")
			sb.WriteString("|-------|------|\n")
			// Sort aliases for consistent output
			aliasList := make([]string, 0, len(aliases))
			for a := range aliases {
				aliasList = append(aliasList, a)
			}
			sort.Strings(aliasList)
			for _, a := range aliasList {
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` |\n", a, aliases[a]))
			}
		}

	case "set":
		if alias == "" {
			return tools.ErrorResult(fmt.Errorf("alias is required for 'set' action")), nil
		}
		if toolName == "" {
			return tools.ErrorResult(fmt.Errorf("tool_name is required for 'set' action")), nil
		}
		if _, ok := registry.GetTool(toolName); !ok {
			return tools.ErrorResult(fmt.Errorf("tool '%s' not found", toolName)), nil
		}
		discoveryClient.SetAlias(alias, toolName)
		sb.WriteString(fmt.Sprintf("✓ Alias **%s** → `%s`\n", alias, toolName))

	case "remove":
		if alias == "" {
			return tools.ErrorResult(fmt.Errorf("alias is required for 'remove' action")), nil
		}
		discoveryClient.RemoveAlias(alias)
		sb.WriteString(fmt.Sprintf("✓ Removed alias **%s**\n", alias))

	case "resolve":
		if alias == "" {
			return tools.ErrorResult(fmt.Errorf("alias is required for 'resolve' action")), nil
		}
		if resolved, ok := discoveryClient.ResolveAlias(alias); ok {
			sb.WriteString(fmt.Sprintf("**%s** → `%s`\n", alias, resolved))
		} else {
			sb.WriteString(fmt.Sprintf("Alias '%s' not found.\n", alias))
		}

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action '%s'. Use 'list', 'set', 'remove', or 'resolve'", action)), nil
	}

	return tools.TextResult(sb.String()), nil
}

// formatTimeAgo formats a timestamp as "X ago" string
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", mins)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
