// Package tools provides the core tool registry and interfaces for the hg-mcp server.
package tools

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hairglasses-studio/mcpkit/registry"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hairglasses-studio/hg-mcp/internal/observability"
	musicsync "github.com/hairglasses-studio/hg-mcp/internal/sync"
	"github.com/hairglasses-studio/hg-mcp/pkg/ratelimit"
	"github.com/hairglasses-studio/hg-mcp/pkg/security"
)

// ToolHandlerFunc is the function signature for tool handlers.
type ToolHandlerFunc = registry.ToolHandlerFunc

// ToolComplexity indicates the complexity level of a tool.
type ToolComplexity = registry.ToolComplexity

const (
	ComplexitySimple   ToolComplexity = "simple"   // Quick lookups, status checks
	ComplexityModerate ToolComplexity = "moderate" // Multi-step operations
	ComplexityComplex  ToolComplexity = "complex"  // Deep investigation, analysis
)

// ToolDefinition represents a complete tool with metadata.
// Aliased from mcpkit/registry so types are interchangeable.
type ToolDefinition = registry.ToolDefinition

// RuntimeGroup constants for high-level functional grouping.
const (
	RuntimeGroupDJMusic         = "dj_music"
	RuntimeGroupVJVideo         = "vj_video"
	RuntimeGroupLighting        = "lighting"
	RuntimeGroupAudioProduction = "audio_production"
	RuntimeGroupShowControl     = "show_control"
	RuntimeGroupInfrastructure  = "infrastructure"
	RuntimeGroupMessaging       = "messaging"
	RuntimeGroupInventory       = "inventory"
	RuntimeGroupStreaming       = "streaming"
	RuntimeGroupPlatform        = "platform"
)

// categoryToRuntimeGroup maps per-module categories to high-level runtime groups.
var categoryToRuntimeGroup = map[string]string{
	// DJ / Music
	"rekordbox": RuntimeGroupDJMusic, "serato": RuntimeGroupDJMusic, "traktor": RuntimeGroupDJMusic,
	"soundcloud": RuntimeGroupDJMusic, "spotify": RuntimeGroupDJMusic, "beatport": RuntimeGroupDJMusic,
	"samples": RuntimeGroupDJMusic, "stems": RuntimeGroupDJMusic, "fingerprint": RuntimeGroupDJMusic,
	"cr8": RuntimeGroupDJMusic, "setlist": RuntimeGroupDJMusic, "prolink": RuntimeGroupDJMusic,
	"mixcloud": RuntimeGroupDJMusic, "bandcamp": RuntimeGroupDJMusic, "boomkat": RuntimeGroupDJMusic,
	"juno": RuntimeGroupDJMusic, "traxsource": RuntimeGroupDJMusic, "tidal": RuntimeGroupDJMusic,
	"discogs": RuntimeGroupDJMusic, "ytmusic": RuntimeGroupDJMusic, "music_discovery": RuntimeGroupDJMusic,
	"sync": RuntimeGroupDJMusic,

	// VJ / Video
	"resolume": RuntimeGroupVJVideo, "resolume_plugins": RuntimeGroupVJVideo,
	"touchdesigner": RuntimeGroupVJVideo, "ndicv": RuntimeGroupVJVideo,
	"video": RuntimeGroupVJVideo, "videoai": RuntimeGroupVJVideo, "videorouting": RuntimeGroupVJVideo,
	"vj_clips": RuntimeGroupVJVideo, "ffgl": RuntimeGroupVJVideo, "vimix": RuntimeGroupVJVideo,
	"mapmap": RuntimeGroupVJVideo, "gpushare": RuntimeGroupVJVideo, "avsync": RuntimeGroupVJVideo,
	"ptz": RuntimeGroupVJVideo, "ptztrack": RuntimeGroupVJVideo, "retrogaming": RuntimeGroupVJVideo,

	// Lighting
	"lighting": RuntimeGroupLighting, "wled": RuntimeGroupLighting, "ledfx": RuntimeGroupLighting,
	"sacn": RuntimeGroupLighting, "ola": RuntimeGroupLighting, "qlcplus": RuntimeGroupLighting,
	"grandma3": RuntimeGroupLighting, "xlights": RuntimeGroupLighting, "opc": RuntimeGroupLighting,
	"companion": RuntimeGroupLighting, "linuxshowplayer": RuntimeGroupLighting,
	"nanoleaf": RuntimeGroupLighting, "hue": RuntimeGroupLighting,

	// Audio Production
	"ableton": RuntimeGroupAudioProduction, "supercollider": RuntimeGroupAudioProduction,
	"ardour": RuntimeGroupAudioProduction, "puredata": RuntimeGroupAudioProduction,
	"maxforlive": RuntimeGroupAudioProduction, "midi": RuntimeGroupAudioProduction,
	"dante": RuntimeGroupAudioProduction, "whisper": RuntimeGroupAudioProduction,

	// Show Control
	"showcontrol": RuntimeGroupShowControl, "showkontrol": RuntimeGroupShowControl,
	"chains": RuntimeGroupShowControl, "snapshots": RuntimeGroupShowControl,
	"timecodesync": RuntimeGroupShowControl, "triggersync": RuntimeGroupShowControl,
	"bpmsync": RuntimeGroupShowControl, "paramsync": RuntimeGroupShowControl,
	"workflows": RuntimeGroupShowControl, "workflow_automation": RuntimeGroupShowControl,
	"consolidated": RuntimeGroupShowControl, "studio": RuntimeGroupShowControl,
	"atem": RuntimeGroupShowControl, "chataigne": RuntimeGroupShowControl,
	"ossia": RuntimeGroupShowControl, "streamdeck": RuntimeGroupShowControl,

	// Infrastructure
	"unraid": RuntimeGroupInfrastructure, "opnsense": RuntimeGroupInfrastructure,
	"tailscale": RuntimeGroupInfrastructure, "system": RuntimeGroupInfrastructure,
	"hwmonitor": RuntimeGroupInfrastructure, "backup": RuntimeGroupInfrastructure,
	"usb": RuntimeGroupInfrastructure, "homeassistant": RuntimeGroupInfrastructure,
	"mqtt": RuntimeGroupInfrastructure, "rclone": RuntimeGroupInfrastructure,

	// Messaging
	"discord": RuntimeGroupMessaging, "discord_admin": RuntimeGroupMessaging,
	"slack": RuntimeGroupMessaging, "telegram": RuntimeGroupMessaging,
	"gmail": RuntimeGroupMessaging, "notion": RuntimeGroupMessaging,
	"pages": RuntimeGroupMessaging, "calendar": RuntimeGroupMessaging,
	"gtasks": RuntimeGroupMessaging,

	// Inventory
	"inventory": RuntimeGroupInventory,

	// Streaming
	"streaming": RuntimeGroupStreaming, "twitch": RuntimeGroupStreaming,
	"youtube_live": RuntimeGroupStreaming,

	// Platform
	"discovery": RuntimeGroupPlatform, "router": RuntimeGroupPlatform,
	"dashboard": RuntimeGroupPlatform, "security": RuntimeGroupPlatform,
	"analytics": RuntimeGroupPlatform, "graph": RuntimeGroupPlatform,
	"vault": RuntimeGroupPlatform, "memory": RuntimeGroupPlatform,
	"learning": RuntimeGroupPlatform, "healing": RuntimeGroupPlatform,
	"gateway": RuntimeGroupPlatform, "federation": RuntimeGroupPlatform,
	"swarm": RuntimeGroupPlatform, "tasks": RuntimeGroupPlatform,
	"archive": RuntimeGroupPlatform, "data_migration": RuntimeGroupPlatform,
	"gdrive": RuntimeGroupPlatform, "plugins": RuntimeGroupPlatform,
}

// ToolModule is the interface that tool modules implement
type ToolModule interface {
	// Name returns the module name (e.g., "discord", "touchdesigner")
	Name() string

	// Description returns a brief description of the module
	Description() string

	// Tools returns all tool definitions in this module
	Tools() []ToolDefinition
}

// ToolRegistry manages tool registration and lookup
type ToolRegistry struct {
	mu       sync.RWMutex
	modules  map[string]ToolModule
	tools    map[string]ToolDefinition
	deferred map[string]ToolDefinition // tools not yet registered with MCP server
	server   *server.MCPServer         // set by RegisterWithServer for deferred loading
	profile  string                    // active profile (set before RegisterWithServer)
}

// Global registry instance
var (
	globalRegistry     *ToolRegistry
	globalRegistryOnce sync.Once
)

// GetRegistry returns the global tool registry
func GetRegistry() *ToolRegistry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewToolRegistry()
	})
	return globalRegistry
}

// NewToolRegistry creates a new tool registry with pre-allocated maps
// to avoid rehashing during startup when ~120 modules register ~880 tools.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		modules:  make(map[string]ToolModule, 128),
		tools:    make(map[string]ToolDefinition, 1024),
		deferred: make(map[string]ToolDefinition, 512),
	}
}

// RegisterModule registers a tool module
func (r *ToolRegistry) RegisterModule(module ToolModule) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.modules[module.Name()] = module

	// Register all tools from the module, auto-assigning RuntimeGroup and IsWrite if not set
	for _, tool := range module.Tools() {
		if tool.RuntimeGroup == "" {
			if group, ok := categoryToRuntimeGroup[tool.Category]; ok {
				tool.RuntimeGroup = group
			}
		}
		if !tool.IsWrite {
			tool.IsWrite = registry.InferIsWrite(tool.Tool.Name)
		}
		r.tools[tool.Tool.Name] = tool
	}
}

// GetTool returns a tool definition by name
func (r *ToolRegistry) GetTool(name string) (ToolDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

// GetModule returns a module by name
func (r *ToolRegistry) GetModule(name string) (ToolModule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	module, ok := r.modules[name]
	return module, ok
}

// ListModules returns all registered module names
func (r *ToolRegistry) ListModules() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.modules))
	for name := range r.modules {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListTools returns all registered tool names
func (r *ToolRegistry) ListTools() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListToolsByCategory returns tools filtered by category
func (r *ToolRegistry) ListToolsByCategory(category string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name, tool := range r.tools {
		if tool.Category == category {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// ListToolsByRuntimeGroup returns tools filtered by runtime group
func (r *ToolRegistry) ListToolsByRuntimeGroup(group string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name, tool := range r.tools {
		if tool.RuntimeGroup == group {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// GetRuntimeGroupStats returns tool counts per runtime group
func (r *ToolRegistry) GetRuntimeGroupStats() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]int)
	for _, tool := range r.tools {
		group := tool.RuntimeGroup
		if group == "" {
			group = "unassigned"
		}
		stats[group]++
	}
	return stats
}

// GetAllToolDefinitions returns all registered tool definitions
func (r *ToolRegistry) GetAllToolDefinitions() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allTools := make([]ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		allTools = append(allTools, tool)
	}
	return allTools
}

// ToolCount returns the number of registered tools
func (r *ToolRegistry) ToolCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// ModuleCount returns the number of registered modules
func (r *ToolRegistry) ModuleCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.modules)
}

// ToolSearchResult represents a tool match with relevance score
type ToolSearchResult struct {
	Tool      ToolDefinition
	Score     int    // Higher is better match
	MatchType string // "name", "tag", "category", "description"
}

// SearchTools searches for tools matching a query string
func (r *ToolRegistry) SearchTools(query string) []ToolSearchResult {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query = strings.ToLower(query)
	queryWords := strings.Fields(query)
	var results []ToolSearchResult

	for _, tool := range r.tools {
		score := 0
		matchType := ""

		// Check tool name (highest priority)
		toolNameLower := strings.ToLower(tool.Tool.Name)
		if strings.Contains(toolNameLower, query) {
			score += 100
			matchType = "name"
		}
		for _, word := range queryWords {
			if strings.Contains(toolNameLower, word) {
				score += 20
			}
		}

		// Check tags
		for _, tag := range tool.Tags {
			tagLower := strings.ToLower(tag)
			if tagLower == query {
				score += 80
				if matchType == "" {
					matchType = "tag"
				}
			} else if strings.Contains(tagLower, query) {
				score += 40
				if matchType == "" {
					matchType = "tag"
				}
			}
		}

		// Check category
		if strings.Contains(strings.ToLower(tool.Category), query) {
			score += 50
			if matchType == "" {
				matchType = "category"
			}
		}

		// Check runtime group
		if tool.RuntimeGroup != "" && strings.Contains(strings.ToLower(tool.RuntimeGroup), query) {
			score += 60
			if matchType == "" {
				matchType = "runtime_group"
			}
		}

		// Check description
		descLower := strings.ToLower(tool.Tool.Description)
		for _, word := range queryWords {
			if len(word) > 2 && strings.Contains(descLower, word) {
				score += 10
				if matchType == "" {
					matchType = "description"
				}
			}
		}

		if score > 0 {
			results = append(results, ToolSearchResult{
				Tool:      tool,
				Score:     score,
				MatchType: matchType,
			})
		}
	}

	// Sort by score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

// ToolStats holds statistics about registered tools
type ToolStats struct {
	TotalTools      int            `json:"total_tools"`
	ModuleCount     int            `json:"module_count"`
	ByCategory      map[string]int `json:"by_category"`
	ByComplexity    map[string]int `json:"by_complexity"`
	ByRuntimeGroup  map[string]int `json:"by_runtime_group"`
	WriteToolsCount int            `json:"write_tools_count"`
	ReadOnlyCount   int            `json:"read_only_count"`
	DeprecatedCount int            `json:"deprecated_count"`
}

// GetToolStats returns statistics about the registered tools
func (r *ToolRegistry) GetToolStats() ToolStats {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := ToolStats{
		TotalTools:     len(r.tools),
		ModuleCount:    len(r.modules),
		ByCategory:     make(map[string]int),
		ByComplexity:   make(map[string]int),
		ByRuntimeGroup: make(map[string]int),
	}

	for _, tool := range r.tools {
		stats.ByCategory[tool.Category]++
		stats.ByComplexity[string(tool.Complexity)]++
		group := tool.RuntimeGroup
		if group == "" {
			group = "unassigned"
		}
		stats.ByRuntimeGroup[group]++
		if tool.IsWrite {
			stats.WriteToolsCount++
		} else {
			stats.ReadOnlyCount++
		}
		if tool.Deprecated {
			stats.DeprecatedCount++
		}
	}

	return stats
}

// GetToolCatalog returns a structured catalog of all tools organized by category
func (r *ToolRegistry) GetToolCatalog() map[string]map[string][]ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	catalog := make(map[string]map[string][]ToolDefinition)
	for _, tool := range r.tools {
		if catalog[tool.Category] == nil {
			catalog[tool.Category] = make(map[string][]ToolDefinition)
		}
		subcategory := tool.Subcategory
		if subcategory == "" {
			subcategory = "general"
		}
		catalog[tool.Category][subcategory] = append(catalog[tool.Category][subcategory], tool)
	}
	return catalog
}

// GetUnconfiguredCategories returns categories that have registered tools but
// whose backing service lacks configuration (env vars). Returns map[category]toolCount.
func (r *ToolRegistry) GetUnconfiguredCategories() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Count tools per category
	catCounts := make(map[string]int)
	for _, tool := range r.tools {
		catCounts[tool.Category]++
	}

	// Check which categories have unconfigured circuit breaker groups.
	// A category is "unconfigured" if its tools have a CircuitBreakerGroup
	// but the service is not reachable (heuristic: env var not set).
	// For now, return empty — the config audit covers this via AuditConfig.
	result := make(map[string]int)
	configuredGroups := make(map[string]bool)
	for _, tool := range r.tools {
		if tool.CircuitBreakerGroup != "" {
			configuredGroups[tool.CircuitBreakerGroup] = true
		}
	}
	// Categories without any circuit breaker group are considered unconfigured
	for cat, count := range catCounts {
		hasGroup := false
		for _, tool := range r.tools {
			if tool.Category == cat && tool.CircuitBreakerGroup != "" {
				hasGroup = true
				break
			}
		}
		if !hasGroup {
			result[cat] = count
		}
	}
	return result
}

// DefaultToolTimeout is the maximum time a tool handler can run
const DefaultToolTimeout = 30 * time.Second

// SetProfile sets the active profile before RegisterWithServer is called.
// If not called, defaults to the HG_MCP_PROFILE environment variable.
func (r *ToolRegistry) SetProfile(profile string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.profile = profile
}

// GetProfile returns the active profile name.
func (r *ToolRegistry) GetProfile() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.profile == "" {
		return hgToolProfile()
	}
	return r.profile
}

// RegisterWithServer registers eager tools with an MCP server and stores
// deferred tools for on-demand loading via LoadDomain.
func (r *ToolRegistry) RegisterWithServer(s *server.MCPServer) {
	start := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	r.server = s
	profile := r.profile
	if profile == "" {
		profile = hgToolProfile()
	}

	var eagerCount, deferredCount int
	for _, tool := range r.tools {
		if shouldDeferTool(profile, tool) {
			r.deferred[tool.Tool.Name] = tool
			deferredCount++
		} else {
			toolWithAnnotations := registry.ApplyToolMetadata(tool, "aftrs_", false)
			s.AddTool(toolWithAnnotations.Tool, server.ToolHandlerFunc(r.wrapHandler(tool.Tool.Name, tool.Handler)))
			eagerCount++
		}
	}

	slog.Info("registered tools with MCP server",
		"profile", profile,
		"eager", eagerCount,
		"deferred", deferredCount,
		"total", len(r.tools),
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

// DeferredToolCount returns the number of tools waiting for on-demand loading.
func (r *ToolRegistry) DeferredToolCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.deferred)
}

// DeferredGroups returns runtime groups that have deferred tools, with tool counts.
func (r *ToolRegistry) DeferredGroupCounts() map[string]int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	counts := make(map[string]int)
	for _, td := range r.deferred {
		group := td.RuntimeGroup
		if group == "" {
			group = "unassigned"
		}
		counts[group]++
	}
	return counts
}

// LoadDomain registers all deferred tools for a given runtime group with the
// MCP server. Returns the number of tools loaded and any error.
// Safe to call multiple times — already-loaded domains are a no-op.
func (r *ToolRegistry) LoadDomain(group string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.server == nil {
		return 0, fmt.Errorf("MCP server not set; call RegisterWithServer first")
	}

	var loaded int
	for name, td := range r.deferred {
		rg := td.RuntimeGroup
		if rg == "" {
			rg = "unassigned"
		}
		if rg == group {
			toolWithAnnotations := registry.ApplyToolMetadata(td, "aftrs_", false)
			r.server.AddTool(toolWithAnnotations.Tool, server.ToolHandlerFunc(r.wrapHandler(td.Tool.Name, td.Handler)))
			delete(r.deferred, name)
			loaded++
		}
	}

	if loaded > 0 {
		slog.Info("loaded deferred domain",
			"domain", group,
			"tools_loaded", loaded,
			"remaining_deferred", len(r.deferred),
		)
	}

	return loaded, nil
}

// LoadAllDeferred registers all remaining deferred tools with the MCP server.
// Returns the total number of tools loaded.
func (r *ToolRegistry) LoadAllDeferred() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.server == nil {
		return 0
	}

	total := len(r.deferred)
	for name, td := range r.deferred {
		toolWithAnnotations := registry.ApplyToolMetadata(td, "aftrs_", false)
		r.server.AddTool(toolWithAnnotations.Tool, server.ToolHandlerFunc(r.wrapHandler(td.Tool.Name, td.Handler)))
		delete(r.deferred, name)
	}

	if total > 0 {
		slog.Info("loaded all deferred tools", "tools_loaded", total)
	}

	return total
}

// applyMCPAnnotations applies MCP 2025 annotations based on tool metadata
func applyMCPAnnotations(td ToolDefinition) ToolDefinition {
	// Generate human-readable title from tool name
	td.Tool.Annotations.Title = toolNameToTitle(td.Tool.Name)

	// Set hints based on IsWrite flag
	readOnly := !td.IsWrite
	destructive := td.IsWrite
	idempotent := !td.IsWrite
	openWorld := true // Most tools interact with external systems

	td.Tool.Annotations.ReadOnlyHint = &readOnly
	td.Tool.Annotations.DestructiveHint = &destructive
	td.Tool.Annotations.IdempotentHint = &idempotent
	td.Tool.Annotations.OpenWorldHint = &openWorld

	return td
}

// toolNameToTitle converts a tool name like "aftrs_gmail_send" to "Gmail Send"
func toolNameToTitle(name string) string {
	// Remove prefix
	name = strings.TrimPrefix(name, "aftrs_")

	// Replace underscores with spaces
	name = strings.ReplaceAll(name, "_", " ")

	// Title case each word
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

// wrapHandler wraps a tool handler with panic recovery, timeout, and observability.
// IMPORTANT: Caller must hold r.mu (read or write lock).
func (r *ToolRegistry) wrapHandler(toolName string, handler ToolHandlerFunc) ToolHandlerFunc {
	// Get the tool definition to access category — use direct map access
	// since the caller already holds the lock.
	tool := r.tools[toolName]
	category := tool.Category
	if category == "" {
		category = "unknown"
	}

	// Use per-tool timeout if set, otherwise default
	timeout := tool.Timeout
	if timeout == 0 {
		timeout = DefaultToolTimeout
	}

	return func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
		// Enforce timeout (per-tool or default)
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}

		// Start tracing span
		ctx, span := observability.StartSpan(ctx, toolName)
		if span != nil {
			defer span.End()
		}

		// Record start of execution
		observability.StartToolExecution(ctx, toolName, category)
		defer observability.EndToolExecution(ctx, toolName, category)

		start := time.Now()

		// Audit logging — extract params for structured logging
		var auditParams map[string]any
		if request.Params.Arguments != nil {
			if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
				auditParams = args
			}
		}
		security.LogToolInvocation(ctx, "", toolName, auditParams)

		// Panic recovery
		defer func() {
			duration := time.Since(start)
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				err = fmt.Errorf("panic in %s: %v\n%s", toolName, r, stack)
				result = &mcp.CallToolResult{
					Content: []mcp.Content{
						mcp.TextContent{
							Type: "text",
							Text: fmt.Sprintf("Internal error in %s: recovered from panic", toolName),
						},
					},
					IsError: true,
				}
				// Record panic as error
				observability.RecordToolInvocation(ctx, toolName, category, duration, err)
			}
		}()

		// Execute handler (optionally wrapped by rate limiter + circuit breaker)
		if tool.CircuitBreakerGroup != "" {
			// Rate limit before circuit breaker
			if waitErr := ratelimit.Get(tool.CircuitBreakerGroup).Wait(ctx); waitErr != nil {
				return ErrorResult(fmt.Errorf("rate limited: %w", waitErr)), nil
			}

			cb := musicsync.GlobalCircuitBreakers.Get(tool.CircuitBreakerGroup)
			cbErr := cb.Execute(ctx, func(cbCtx context.Context) error {
				result, err = handler(cbCtx, request)
				if err != nil {
					return err
				}
				if result != nil && result.IsError {
					return errors.New("tool returned error result")
				}
				return nil
			})
			// If circuit is open, return a user-friendly error
			if cbErr != nil && errors.Is(cbErr, musicsync.ErrCircuitOpen) {
				result = &mcp.CallToolResult{
					Content: []mcp.Content{
						mcp.TextContent{
							Type: "text",
							Text: fmt.Sprintf("[CIRCUIT_OPEN] %s service is temporarily unavailable (circuit breaker open)", tool.CircuitBreakerGroup),
						},
					},
					IsError: true,
				}
				err = nil
			}
		} else {
			result, err = handler(ctx, request)
		}

		// Record metrics
		duration := time.Since(start)
		observability.RecordToolInvocation(ctx, toolName, category, duration, err)

		// Log execution with error code extraction
		if err != nil {
			slog.Error("tool failed", "tool", toolName, "duration", duration, "error", err)
		} else if result != nil && result.IsError {
			// Extract error code from coded errors for structured logging
			for _, content := range result.Content {
				if tc, ok := content.(mcp.TextContent); ok && len(tc.Text) > 1 && tc.Text[0] == '[' {
					if idx := strings.Index(tc.Text, "]"); idx > 0 {
						code := tc.Text[1:idx]
						slog.Warn("tool error", "tool", toolName, "error_code", code, "duration", duration)
					}
				}
			}
		}

		// Audit log completion
		security.LogToolCompletion(ctx, "", toolName, duration, err)

		return result, err
	}
}
