// Package main is the entry point for the aftrs CLI.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"

	// Import modules to trigger registration via init()
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ableton"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/analytics"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/atem"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/backup"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/beatport"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/bpmsync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/calendar"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/chains"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/consolidated"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/cr8"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/dante"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/dashboard"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/discord"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/discord_admin"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/discovery"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/federation"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ffgl"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/fingerprint"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/gateway"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/gdrive"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/gmail"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/grandma3"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/graph"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/gtasks"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/healing"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/homeassistant"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/hwmonitor"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/learning"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ledfx"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/lighting"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/maxforlive"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/memory"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/midi"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/mqtt"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ndicv"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/notion"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/obs"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/paramsync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/plugins"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ptz"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ptztrack"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/rclone"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/rekordbox"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/resolume"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/resolume_plugins"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/retrogaming"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/router"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/security"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/serato"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/setlist"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/showcontrol"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/showkontrol"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/snapshots"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/spotify"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/stems"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/streamdeck"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/streaming"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/studio"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/swarm"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/sync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/tailscale"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/tasks"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/timecodesync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/touchdesigner"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/traktor"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/triggersync"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/twitch"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/unraid"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/usb"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/vault"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/video"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/videoai"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/videorouting"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/vj_clips"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/whisper"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/wled"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/workflows"
	_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/youtube_live"
)

var rootCmd = &cobra.Command{
	Use:   "aftrs",
	Short: "Aftrs Studio Operations CLI",
	Long:  `A CLI for managing Aftrs AudioVisual Studio operations.`,
}

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "Tool management commands",
}

var toolListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available tools",
	Run: func(cmd *cobra.Command, args []string) {
		registry := tools.GetRegistry()
		stats := registry.GetToolStats()

		fmt.Printf("Aftrs MCP Tools: %d tools across %d modules\n\n", stats.TotalTools, stats.ModuleCount)

		for cat, count := range stats.ByCategory {
			fmt.Printf("  %s: %d tools\n", cat, count)
		}
	},
}

var toolSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for tools by keyword",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		registry := tools.GetRegistry()
		results := registry.SearchTools(args[0])

		fmt.Printf("Found %d tools matching '%s':\n\n", len(results), args[0])

		for _, r := range results {
			fmt.Printf("  %s (%s)\n", r.Tool.Tool.Name, r.Tool.Category)
			fmt.Printf("    %s\n\n", r.Tool.Tool.Description)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aftrs version 0.1.0")
	},
}

var toolDocsCmd = &cobra.Command{
	Use:   "docs [output-dir]",
	Short: "Generate markdown documentation for all tools",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputDir := "docs/tools"
		if len(args) > 0 {
			outputDir = args[0]
		}
		if err := generateDocs(outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating docs: %v\n", err)
			os.Exit(1)
		}
	},
}

func generateDocs(outputDir string) error {
	registry := tools.GetRegistry()
	stats := registry.GetToolStats()

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	// Group tools by module
	moduleTools := make(map[string][]tools.ToolDefinition)
	for _, tool := range registry.GetAllToolDefinitions() {
		moduleTools[tool.Category] = append(moduleTools[tool.Category], tool)
	}

	// Sort modules
	modules := make([]string, 0, len(moduleTools))
	for m := range moduleTools {
		modules = append(modules, m)
	}
	sort.Strings(modules)

	// Generate index page
	var indexContent strings.Builder
	indexContent.WriteString("# AFTRS MCP Tools Reference\n\n")
	indexContent.WriteString(fmt.Sprintf("> Auto-generated on %s\n\n", time.Now().Format("2006-01-02")))
	indexContent.WriteString(fmt.Sprintf("**%d tools** across **%d modules**\n\n", stats.TotalTools, stats.ModuleCount))
	indexContent.WriteString("## Modules\n\n")
	indexContent.WriteString("| Module | Tools | Description |\n")
	indexContent.WriteString("|--------|-------|-------------|\n")

	for _, moduleName := range modules {
		toolList := moduleTools[moduleName]
		sort.Slice(toolList, func(i, j int) bool {
			return toolList[i].Tool.Name < toolList[j].Tool.Name
		})

		// Get module description
		mod, _ := registry.GetModule(moduleName)
		desc := ""
		if mod != nil {
			desc = mod.Description()
		}

		indexContent.WriteString(fmt.Sprintf("| [%s](./%s.md) | %d | %s |\n",
			moduleName, moduleName, len(toolList), desc))

		// Generate module page
		if err := generateModulePage(outputDir, moduleName, toolList, mod); err != nil {
			return err
		}
	}

	// Write index
	indexPath := filepath.Join(outputDir, "README.md")
	if err := os.WriteFile(indexPath, []byte(indexContent.String()), 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	fmt.Printf("Generated documentation for %d tools across %d modules\n", stats.TotalTools, len(modules))
	fmt.Printf("Output: %s\n", outputDir)
	return nil
}

func generateModulePage(outputDir, moduleName string, toolList []tools.ToolDefinition, mod tools.ToolModule) error {
	var content strings.Builder

	desc := ""
	if mod != nil {
		desc = mod.Description()
	}

	content.WriteString(fmt.Sprintf("# %s\n\n", moduleName))
	if desc != "" {
		content.WriteString(fmt.Sprintf("> %s\n\n", desc))
	}
	content.WriteString(fmt.Sprintf("**%d tools**\n\n", len(toolList)))

	// Table of contents
	content.WriteString("## Tools\n\n")
	for _, tool := range toolList {
		anchor := strings.ReplaceAll(tool.Tool.Name, "_", "-")
		content.WriteString(fmt.Sprintf("- [`%s`](#%s)\n", tool.Tool.Name, anchor))
	}
	content.WriteString("\n---\n\n")

	// Tool details
	for _, tool := range toolList {
		content.WriteString(fmt.Sprintf("## %s\n\n", tool.Tool.Name))
		content.WriteString(fmt.Sprintf("%s\n\n", tool.Tool.Description))

		// Metadata
		if tool.Complexity != "" {
			content.WriteString(fmt.Sprintf("**Complexity:** %s\n\n", tool.Complexity))
		}
		if len(tool.Tags) > 0 {
			content.WriteString(fmt.Sprintf("**Tags:** `%s`\n\n", strings.Join(tool.Tags, "`, `")))
		}
		if len(tool.UseCases) > 0 {
			content.WriteString("**Use Cases:**\n")
			for _, uc := range tool.UseCases {
				content.WriteString(fmt.Sprintf("- %s\n", uc))
			}
			content.WriteString("\n")
		}

		// Parameters
		if tool.Tool.InputSchema.Properties != nil && len(tool.Tool.InputSchema.Properties) > 0 {
			content.WriteString("### Parameters\n\n")
			content.WriteString("| Name | Type | Required | Description |\n")
			content.WriteString("|------|------|----------|-------------|\n")

			// Get required params
			required := make(map[string]bool)
			for _, r := range tool.Tool.InputSchema.Required {
				required[r] = true
			}

			// Sort params
			params := make([]string, 0, len(tool.Tool.InputSchema.Properties))
			for p := range tool.Tool.InputSchema.Properties {
				params = append(params, p)
			}
			sort.Strings(params)

			for _, paramName := range params {
				prop := tool.Tool.InputSchema.Properties[paramName]
				propMap, ok := prop.(map[string]interface{})
				if !ok {
					continue
				}

				paramType := "any"
				if t, ok := propMap["type"].(string); ok {
					paramType = t
				}

				paramDesc := ""
				if d, ok := propMap["description"].(string); ok {
					paramDesc = d
				}

				reqStr := ""
				if required[paramName] {
					reqStr = "Yes"
				}

				content.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n",
					paramName, paramType, reqStr, paramDesc))
			}
			content.WriteString("\n")
		}

		// Example (JSON schema)
		if tool.Tool.InputSchema.Properties != nil && len(tool.Tool.InputSchema.Properties) > 0 {
			example := make(map[string]interface{})
			for p, prop := range tool.Tool.InputSchema.Properties {
				propMap, ok := prop.(map[string]interface{})
				if !ok {
					continue
				}
				if t, ok := propMap["type"].(string); ok {
					switch t {
					case "string":
						example[p] = "example"
					case "integer", "number":
						example[p] = 0
					case "boolean":
						example[p] = false
					case "array":
						example[p] = []interface{}{}
					case "object":
						example[p] = map[string]interface{}{}
					}
				}
			}
			if len(example) > 0 {
				content.WriteString("### Example\n\n```json\n")
				jsonBytes, _ := json.MarshalIndent(example, "", "  ")
				content.WriteString(string(jsonBytes))
				content.WriteString("\n```\n\n")
			}
		}

		content.WriteString("---\n\n")
	}

	// Write file
	filePath := filepath.Join(outputDir, moduleName+".md")
	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}

	return nil
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show server status, tool counts, and runtime group breakdown",
	Run: func(cmd *cobra.Command, args []string) {
		registry := tools.GetRegistry()
		stats := registry.GetToolStats()
		groupStats := registry.GetRuntimeGroupStats()

		fmt.Printf("hg-mcp Status\n")
		fmt.Printf("═════════════\n\n")
		fmt.Printf("  Tools:    %d\n", stats.TotalTools)
		fmt.Printf("  Modules:  %d\n", stats.ModuleCount)
		fmt.Printf("  Write:    %d\n", stats.WriteToolsCount)
		fmt.Printf("  ReadOnly: %d\n", stats.ReadOnlyCount)
		fmt.Printf("\n")

		fmt.Printf("Runtime Groups\n")
		fmt.Printf("──────────────\n")

		// Sort groups for consistent output
		groups := make([]string, 0, len(groupStats))
		for g := range groupStats {
			groups = append(groups, g)
		}
		sort.Strings(groups)
		for _, g := range groups {
			fmt.Printf("  %-20s %d tools\n", g, groupStats[g])
		}
	},
}

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "List all tools grouped by category (shortcut for 'tool list')",
	Run: func(cmd *cobra.Command, args []string) {
		registry := tools.GetRegistry()
		stats := registry.GetToolStats()

		fmt.Printf("hg-mcp: %d tools across %d modules\n\n", stats.TotalTools, stats.ModuleCount)

		// Sort categories
		cats := make([]string, 0, len(stats.ByCategory))
		for cat := range stats.ByCategory {
			cats = append(cats, cat)
		}
		sort.Strings(cats)

		for _, cat := range cats {
			fmt.Printf("  %-25s %3d tools\n", cat, stats.ByCategory[cat])
		}
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show which services are configured via environment variables",
	Run: func(cmd *cobra.Command, args []string) {
		report := config.AuditConfig()

		fmt.Printf("Service Configuration\n")
		fmt.Printf("═════════════════════\n\n")

		if len(report.Configured) > 0 {
			fmt.Printf("Configured (%d):\n", len(report.Configured))
			sort.Strings(report.Configured)
			for _, s := range report.Configured {
				fmt.Printf("  ✓ %s\n", s)
			}
		}

		if len(report.Missing) > 0 {
			fmt.Printf("\nNot configured (%d):\n", len(report.Missing))
			sort.Strings(report.Missing)
			for _, s := range report.Missing {
				fmt.Printf("  · %s\n", s)
			}
		}

		fmt.Printf("\nTotal: %d/%d services configured\n",
			len(report.Configured), len(report.Configured)+len(report.Missing))
	},
}

func init() {
	toolCmd.AddCommand(toolListCmd)
	toolCmd.AddCommand(toolSearchCmd)
	toolCmd.AddCommand(toolDocsCmd)
	rootCmd.AddCommand(toolCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(toolsCmd)
	rootCmd.AddCommand(configCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
