// Package loader provides the hg_load_domain tool for on-demand domain loading.
// This module is always eager (platform group) and lets agents activate deferred
// tool domains at runtime without restarting the server.
package loader

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for deferred domain loading.
type Module struct{}

func (m *Module) Name() string        { return "loader" }
func (m *Module) Description() string { return "On-demand domain loader for deferred tool groups" }

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("hg_load_domain",
				mcp.WithDescription(
					"Load a deferred tool domain on demand. "+
						"In non-full profiles, most tool domains start unloaded to reduce context overhead. "+
						"Call this tool with a domain name (e.g. 'dj_music', 'vj_video', 'lighting') to activate "+
						"those tools in the current session. Use domain='all' to load everything. "+
						"Call with no domain to list available deferred domains and their tool counts.",
				),
				mcp.WithString("domain", mcp.Description(
					"Runtime group to load. Valid groups: dj_music, vj_video, lighting, "+
						"audio_production, show_control, infrastructure, messaging, inventory, "+
						"streaming, platform. Use 'all' to load every deferred domain. "+
						"Omit to list available deferred domains.",
				)),
			),
			Handler:     handleLoadDomain,
			Category:    "discovery",
			Subcategory: "loader",
			Tags:        []string{"loader", "domain", "deferred", "profile", "on-demand"},
			UseCases: []string{
				"Load DJ tools when starting a DJ session",
				"Activate lighting tools for show programming",
				"Load all tools for full access",
				"List which domains are available to load",
			},
			Complexity:   tools.ComplexitySimple,
			RuntimeGroup: tools.RuntimeGroupPlatform, // Always eager
		},
	}
}

func handleLoadDomain(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	domain := tools.GetStringParam(req, "domain")
	registry := tools.GetRegistry()

	// No domain specified: list available deferred domains.
	if domain == "" {
		return listDeferredDomains(registry), nil
	}

	// Load all deferred domains.
	if domain == "all" {
		loaded := registry.LoadAllDeferred()
		if loaded == 0 {
			return tools.TextResult("All domains are already loaded. No deferred tools remaining."), nil
		}
		return tools.TextResult(fmt.Sprintf(
			"Loaded all deferred domains: %d tools activated. All tools are now available.",
			loaded,
		)), nil
	}

	// Load a specific domain.
	loaded, err := registry.LoadDomain(domain)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrInvalidParam, err), nil
	}
	if loaded == 0 {
		// Check if this is a known group at all
		known := false
		for _, g := range tools.AllRuntimeGroups() {
			if g == domain {
				known = true
				break
			}
		}
		if !known {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf(
				"unknown domain %q; valid domains: %s",
				domain, strings.Join(tools.AllRuntimeGroups(), ", "),
			)), nil
		}
		return tools.TextResult(fmt.Sprintf(
			"Domain %q (%s) is already loaded or has no deferred tools.",
			domain, tools.RuntimeGroupLabel(domain),
		)), nil
	}

	remaining := registry.DeferredToolCount()
	return tools.TextResult(fmt.Sprintf(
		"Loaded domain %q (%s): %d tools activated. %d deferred tools remaining.",
		domain, tools.RuntimeGroupLabel(domain), loaded, remaining,
	)), nil
}

func listDeferredDomains(registry *tools.ToolRegistry) *mcp.CallToolResult {
	counts := registry.DeferredGroupCounts()
	total := registry.DeferredToolCount()
	profile := registry.GetProfile()

	if total == 0 {
		return tools.TextResult(fmt.Sprintf(
			"Profile: %s\nAll domains are loaded. No deferred tools remaining.\nTotal tools: %d",
			profile, registry.ToolCount(),
		))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Deferred Domains (profile: %s)\n\n", profile))
	sb.WriteString(fmt.Sprintf("**%d tools** deferred across **%d domains**\n\n", total, len(counts)))

	// Sort by group name
	groups := make([]string, 0, len(counts))
	for g := range counts {
		groups = append(groups, g)
	}
	sort.Strings(groups)

	sb.WriteString("| Domain | Label | Tools |\n")
	sb.WriteString("|--------|-------|-------|\n")
	for _, g := range groups {
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %d |\n", g, tools.RuntimeGroupLabel(g), counts[g]))
	}

	sb.WriteString("\nUse `hg_load_domain` with a domain name to activate, or `domain: \"all\"` for everything.")

	return tools.TextResult(sb.String())
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
