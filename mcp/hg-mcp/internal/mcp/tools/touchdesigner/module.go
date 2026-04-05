// Package touchdesigner provides TouchDesigner control tools for hg-mcp.
package touchdesigner

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewTouchDesignerClient)

// Module implements the ToolModule interface for TouchDesigner integration
type Module struct{}

func (m *Module) Name() string {
	return "touchdesigner"
}

func (m *Module) Description() string {
	return "TouchDesigner project control and monitoring"
}

func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_td_status",
				mcp.WithDescription("Get TouchDesigner project status including FPS, cook time, and error counts."),
			),
			Handler:     handleStatus,
			Category:    "touchdesigner",
			Subcategory: "status",
			Tags:        []string{"touchdesigner", "status", "fps", "performance"},
			UseCases:    []string{"Check TD performance", "Monitor project health"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_operators",
				mcp.WithDescription("List operators in a TouchDesigner network."),
				mcp.WithString("path",
					mcp.Description("Network path to list (default: /project1)"),
				),
				mcp.WithString("type",
					mcp.Description("Filter by operator type (e.g., 'TOP', 'CHOP', 'DAT')"),
				),
			),
			Handler:     handleOperators,
			Category:    "touchdesigner",
			Subcategory: "operators",
			Tags:        []string{"touchdesigner", "operators", "network"},
			UseCases:    []string{"Browse operator network", "Find specific operators"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_parameters",
				mcp.WithDescription("Get or set parameters on a TouchDesigner operator."),
				mcp.WithString("operator",
					mcp.Required(),
					mcp.Description("Operator path (e.g., /project1/geo1)"),
				),
				mcp.WithString("param",
					mcp.Description("Parameter name to get/set (omit to list all)"),
				),
				mcp.WithString("value",
					mcp.Description("Value to set (omit to just read)"),
				),
			),
			Handler:     handleParameters,
			Category:    "touchdesigner",
			Subcategory: "parameters",
			Tags:        []string{"touchdesigner", "parameters", "control"},
			UseCases:    []string{"Read operator parameters", "Modify operator settings"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_td_execute",
				mcp.WithDescription("Execute Python code in TouchDesigner's context."),
				mcp.WithString("script",
					mcp.Required(),
					mcp.Description("Python script to execute"),
				),
			),
			Handler:     handleExecute,
			Category:    "touchdesigner",
			Subcategory: "scripting",
			Tags:        []string{"touchdesigner", "python", "scripting", "execute"},
			UseCases:    []string{"Run custom Python scripts", "Automate TD operations"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_td_network_health",
				mcp.WithDescription("Get health score and analysis for a TouchDesigner operator network."),
				mcp.WithString("path",
					mcp.Description("Network path to analyze (default: /project1)"),
				),
			),
			Handler:     handleNetworkHealth,
			Category:    "touchdesigner",
			Subcategory: "health",
			Tags:        []string{"touchdesigner", "health", "performance", "analysis"},
			UseCases:    []string{"Analyze network health", "Find performance issues"},
			Complexity:  tools.ComplexityModerate,
		},
		// v0.7 Advanced Tools
		{
			Tool: mcp.NewTool("aftrs_td_textures",
				mcp.WithDescription("List texture TOPs with resolution and memory usage."),
				mcp.WithString("path",
					mcp.Description("Network path to scan (default: /project1)"),
				),
			),
			Handler:     handleTextures,
			Category:    "touchdesigner",
			Subcategory: "textures",
			Tags:        []string{"touchdesigner", "textures", "top", "memory"},
			UseCases:    []string{"Find large textures", "Audit texture memory usage"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_performance",
				mcp.WithDescription("Get performance profiling data including cook times and frame timing."),
			),
			Handler:     handlePerformance,
			Category:    "touchdesigner",
			Subcategory: "performance",
			Tags:        []string{"touchdesigner", "performance", "profiling", "cook"},
			UseCases:    []string{"Profile performance", "Find bottlenecks"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_gpu_memory",
				mcp.WithDescription("Get GPU memory usage breakdown."),
			),
			Handler:     handleGPUMemory,
			Category:    "touchdesigner",
			Subcategory: "gpu",
			Tags:        []string{"touchdesigner", "gpu", "memory", "vram"},
			UseCases:    []string{"Monitor GPU memory", "Prevent VRAM exhaustion"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_errors",
				mcp.WithDescription("Get current errors and warnings from TouchDesigner."),
			),
			Handler:     handleErrors,
			Category:    "touchdesigner",
			Subcategory: "errors",
			Tags:        []string{"touchdesigner", "errors", "warnings", "debug"},
			UseCases:    []string{"View error log", "Debug issues"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_backup",
				mcp.WithDescription("Create a backup of the current TouchDesigner project."),
				mcp.WithString("destination",
					mcp.Description("Backup file path (optional, auto-generates timestamp)"),
				),
			),
			Handler:     handleBackup,
			Category:    "touchdesigner",
			Subcategory: "project",
			Tags:        []string{"touchdesigner", "backup", "save", "project"},
			UseCases:    []string{"Create project backup", "Save before changes"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_td_preset_recall",
				mcp.WithDescription("Recall a parameter preset to an operator."),
				mcp.WithString("preset",
					mcp.Required(),
					mcp.Description("Preset file path or name"),
				),
				mcp.WithString("operator",
					mcp.Description("Target operator path (optional)"),
				),
			),
			Handler:     handlePresetRecall,
			Category:    "touchdesigner",
			Subcategory: "presets",
			Tags:        []string{"touchdesigner", "preset", "recall", "parameters"},
			UseCases:    []string{"Load saved presets", "Recall show settings"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_td_cue_trigger",
				mcp.WithDescription("Trigger a cue or jump to position on a timeline."),
				mcp.WithString("timeline",
					mcp.Required(),
					mcp.Description("Timeline operator path"),
				),
				mcp.WithString("cue",
					mcp.Description("Cue name or frame number"),
				),
			),
			Handler:     handleCueTrigger,
			Category:    "touchdesigner",
			Subcategory: "timeline",
			Tags:        []string{"touchdesigner", "cue", "timeline", "trigger"},
			UseCases:    []string{"Trigger show cues", "Jump to timeline position"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_td_timelines",
				mcp.WithDescription("List timeline CHOPs with playback status."),
				mcp.WithString("path",
					mcp.Description("Network path to scan (default: /project1)"),
				),
			),
			Handler:     handleTimelines,
			Category:    "touchdesigner",
			Subcategory: "timeline",
			Tags:        []string{"touchdesigner", "timeline", "chop", "playback"},
			UseCases:    []string{"View timeline status", "Check playback state"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_components",
				mcp.WithDescription("List container COMPs in a network."),
				mcp.WithString("path",
					mcp.Description("Network path to scan (default: /project1)"),
				),
			),
			Handler:     handleComponents,
			Category:    "touchdesigner",
			Subcategory: "components",
			Tags:        []string{"touchdesigner", "components", "comp", "containers"},
			UseCases:    []string{"Browse components", "Navigate project structure"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_project_info",
				mcp.WithDescription("Get detailed information about the current TouchDesigner project."),
			),
			Handler:     handleProjectInfo,
			Category:    "touchdesigner",
			Subcategory: "project",
			Tags:        []string{"touchdesigner", "project", "info", "statistics"},
			UseCases:    []string{"View project details", "Get operator counts"},
			Complexity:  tools.ComplexitySimple,
		},
		// Additional tools for expanded coverage
		{
			Tool: mcp.NewTool("aftrs_td_chops",
				mcp.WithDescription("List CHOP operators in a network with channel counts."),
				mcp.WithString("path", mcp.Description("Network path to scan (default: /project1)")),
			),
			Handler:     handleCHOPs,
			Category:    "touchdesigner",
			Subcategory: "chops",
			Tags:        []string{"touchdesigner", "chop", "channels", "audio"},
			UseCases:    []string{"List CHOPs", "Find audio/control operators"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_chop_channels",
				mcp.WithDescription("Get channel values from a specific CHOP operator."),
				mcp.WithString("chop", mcp.Required(), mcp.Description("CHOP operator path")),
			),
			Handler:     handleCHOPChannels,
			Category:    "touchdesigner",
			Subcategory: "chops",
			Tags:        []string{"touchdesigner", "chop", "channels", "values"},
			UseCases:    []string{"Read channel values", "Monitor CHOP output"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_dats",
				mcp.WithDescription("List DAT operators in a network."),
				mcp.WithString("path", mcp.Description("Network path to scan (default: /project1)")),
			),
			Handler:     handleDATs,
			Category:    "touchdesigner",
			Subcategory: "dats",
			Tags:        []string{"touchdesigner", "dat", "table", "text"},
			UseCases:    []string{"List DATs", "Find data operators"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_dat_content",
				mcp.WithDescription("Get or set the content of a DAT operator."),
				mcp.WithString("dat", mcp.Required(), mcp.Description("DAT operator path")),
				mcp.WithString("content", mcp.Description("Content to set (omit to read)")),
			),
			Handler:     handleDATContent,
			Category:    "touchdesigner",
			Subcategory: "dats",
			Tags:        []string{"touchdesigner", "dat", "content", "text"},
			UseCases:    []string{"Read DAT content", "Update DAT data"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_td_render_settings",
				mcp.WithDescription("Get or set render output settings (resolution, format, etc)."),
				mcp.WithNumber("width", mcp.Description("Output width")),
				mcp.WithNumber("height", mcp.Description("Output height")),
				mcp.WithNumber("fps", mcp.Description("Output FPS")),
			),
			Handler:     handleRenderSettings,
			Category:    "touchdesigner",
			Subcategory: "render",
			Tags:        []string{"touchdesigner", "render", "output", "resolution"},
			UseCases:    []string{"Check render settings", "Change output resolution"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_td_custom_pars",
				mcp.WithDescription("List custom parameters on an operator."),
				mcp.WithString("operator", mcp.Required(), mcp.Description("Operator path")),
			),
			Handler:     handleCustomPars,
			Category:    "touchdesigner",
			Subcategory: "parameters",
			Tags:        []string{"touchdesigner", "custom", "parameters", "interface"},
			UseCases:    []string{"View custom parameters", "Check interface settings"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_td_tox_export",
				mcp.WithDescription("Export an operator/component as a TOX file."),
				mcp.WithString("operator", mcp.Required(), mcp.Description("Operator path to export")),
				mcp.WithString("destination", mcp.Description("Destination file path")),
			),
			Handler:     handleToxExport,
			Category:    "touchdesigner",
			Subcategory: "tox",
			Tags:        []string{"touchdesigner", "tox", "export", "component"},
			UseCases:    []string{"Export component", "Share TOX file"},
			Complexity:  tools.ComplexityModerate,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_td_variables",
				mcp.WithDescription("Get or set project variables."),
				mcp.WithString("name", mcp.Description("Variable name (omit to list all)")),
				mcp.WithString("value", mcp.Description("Value to set (omit to read)")),
			),
			Handler:     handleVariables,
			Category:    "touchdesigner",
			Subcategory: "project",
			Tags:        []string{"touchdesigner", "variables", "project", "global"},
			UseCases:    []string{"View project variables", "Set global values"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_td_pulse",
				mcp.WithDescription("Pulse a parameter on an operator."),
				mcp.WithString("operator", mcp.Required(), mcp.Description("Operator path")),
				mcp.WithString("param", mcp.Required(), mcp.Description("Parameter name to pulse")),
			),
			Handler:     handlePulse,
			Category:    "touchdesigner",
			Subcategory: "parameters",
			Tags:        []string{"touchdesigner", "pulse", "trigger", "parameter"},
			UseCases:    []string{"Trigger pulse parameters", "Reset operators"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
		{
			Tool: mcp.NewTool("aftrs_td_reset",
				mcp.WithDescription("Reset an operator to its default state."),
				mcp.WithString("operator", mcp.Required(), mcp.Description("Operator path to reset")),
			),
			Handler:     handleReset,
			Category:    "touchdesigner",
			Subcategory: "operators",
			Tags:        []string{"touchdesigner", "reset", "defaults", "operator"},
			UseCases:    []string{"Reset operator", "Restore defaults"},
			Complexity:  tools.ComplexitySimple,
			IsWrite:     true,
		},
	}

	// Apply circuit breaker to all tools — network-dependent (WebSocket)
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "touchdesigner"
	}

	return allTools
}

// getClient creates a new TouchDesigner client

// handleStatus handles the aftrs_td_status tool
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# TouchDesigner Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** ❌ Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**Target:** %s:%s\n\n", client.Host(), client.Port()))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("TouchDesigner needs a WebServer DAT configured to enable API access.\n\n")
		sb.WriteString("**Quick Setup:**\n")
		sb.WriteString("1. Create a WebServer DAT in your project\n")
		sb.WriteString("2. Set the port (default: 9980)\n")
		sb.WriteString("3. Enable the server\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export TD_HOST=localhost\n")
		sb.WriteString("export TD_PORT=9980\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** ✅ Connected\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s\n", status.ProjectName))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n\n", status.Version))

	sb.WriteString("## Performance\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Value |\n"))
	sb.WriteString(fmt.Sprintf("|--------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| FPS | %.1f |\n", status.FPS))
	sb.WriteString(fmt.Sprintf("| Real-time FPS | %.1f |\n", status.RealTimeFPS))
	sb.WriteString(fmt.Sprintf("| Cook Time | %.2f ms |\n", status.CookTime))
	sb.WriteString(fmt.Sprintf("| GPU Memory | %s |\n", status.GPUMemory))
	sb.WriteString(fmt.Sprintf("| CPU Usage | %.1f%% |\n", status.CPUUsage))

	sb.WriteString("\n## Health\n\n")
	if status.ErrorCount > 0 {
		sb.WriteString(fmt.Sprintf("⚠️ **Errors:** %d\n", status.ErrorCount))
	} else {
		sb.WriteString("✅ **Errors:** 0\n")
	}
	if status.WarningCount > 0 {
		sb.WriteString(fmt.Sprintf("⚠️ **Warnings:** %d\n", status.WarningCount))
	} else {
		sb.WriteString("✅ **Warnings:** 0\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleOperators handles the aftrs_td_operators tool
func handleOperators(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := tools.GetStringParam(req, "path")
	typeFilter := tools.GetStringParam(req, "type")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	operators, err := client.GetOperators(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Apply type filter if specified
	if typeFilter != "" {
		typeFilter = strings.ToUpper(typeFilter)
		filtered := []clients.TDOperator{}
		for _, op := range operators {
			if strings.ToUpper(op.Family) == typeFilter || strings.HasSuffix(strings.ToUpper(op.Type), typeFilter) {
				filtered = append(filtered, op)
			}
		}
		operators = filtered
	}

	var sb strings.Builder
	if path == "" {
		path = "/project1"
	}
	sb.WriteString(fmt.Sprintf("# Operators in %s\n\n", path))

	if len(operators) == 0 {
		sb.WriteString("No operators found")
		if typeFilter != "" {
			sb.WriteString(fmt.Sprintf(" matching type '%s'", typeFilter))
		}
		sb.WriteString(".\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found %d operators:\n\n", len(operators)))
	sb.WriteString("| Name | Type | Family | Status |\n")
	sb.WriteString("|------|------|--------|--------|\n")

	for _, op := range operators {
		status := "✅"
		if op.HasErrors {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", op.Name, op.Type, op.Family, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleParameters handles the aftrs_td_parameters tool
func handleParameters(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	operatorPath, errResult := tools.RequireStringParam(req, "operator")
	if errResult != nil {
		return errResult, nil
	}

	paramName := tools.GetStringParam(req, "param")
	value := tools.GetStringParam(req, "value")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If value is provided, set the parameter
	if value != "" && paramName != "" {
		err := client.SetParameter(ctx, operatorPath, paramName, value)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Set %s.%s = %s", operatorPath, paramName, value)), nil
	}

	// Otherwise, get parameters
	params, err := client.GetParameters(ctx, operatorPath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Parameters: %s\n\n", operatorPath))

	if paramName != "" {
		// Get specific parameter
		if val, ok := params[paramName]; ok {
			sb.WriteString(fmt.Sprintf("**%s:** %v\n", paramName, val))
		} else {
			sb.WriteString(fmt.Sprintf("Parameter '%s' not found.\n", paramName))
		}
	} else {
		// List all parameters
		if len(params) == 0 {
			sb.WriteString("No parameters found.\n")
		} else {
			sb.WriteString("| Parameter | Value |\n")
			sb.WriteString("|-----------|-------|\n")
			for name, val := range params {
				sb.WriteString(fmt.Sprintf("| %s | %v |\n", name, val))
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleExecute handles the aftrs_td_execute tool
func handleExecute(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	script, errResult := tools.RequireStringParam(req, "script")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.ExecutePython(ctx, script)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Python Execution Result\n\n")

	if result.Success {
		sb.WriteString("**Status:** ✅ Success\n\n")
		if result.Output != "" {
			sb.WriteString("**Output:**\n```\n")
			sb.WriteString(result.Output)
			sb.WriteString("\n```\n")
		}
	} else {
		sb.WriteString("**Status:** ❌ Failed\n\n")
		if result.Error != "" {
			sb.WriteString("**Error:**\n```\n")
			sb.WriteString(result.Error)
			sb.WriteString("\n```\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleNetworkHealth handles the aftrs_td_network_health tool
func handleNetworkHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := tools.GetStringParam(req, "path")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	health, err := client.GetNetworkHealth(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if path == "" {
		path = "/project1"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Network Health: %s\n\n", path))

	// Status emoji
	statusEmoji := "✅"
	if health.Status == "degraded" {
		statusEmoji = "⚠️"
	} else if health.Status == "critical" {
		statusEmoji = "❌"
	}

	sb.WriteString(fmt.Sprintf("**Health Score:** %d/100 %s\n", health.Score, statusEmoji))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", health.Status))

	sb.WriteString("## Metrics\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Value |\n"))
	sb.WriteString(fmt.Sprintf("|--------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| Total Operators | %d |\n", health.TotalOperators))
	sb.WriteString(fmt.Sprintf("| Errors | %d |\n", health.ErrorCount))
	sb.WriteString(fmt.Sprintf("| Warnings | %d |\n", health.WarningCount))
	sb.WriteString(fmt.Sprintf("| Slow Operators | %d |\n", len(health.SlowOperators)))

	if len(health.SlowOperators) > 0 {
		sb.WriteString("\n## Slow Operators (>5ms cook time)\n\n")
		sb.WriteString("| Operator | Cook Time |\n")
		sb.WriteString("|----------|-----------|\n")
		for _, op := range health.SlowOperators {
			sb.WriteString(fmt.Sprintf("| %s | %.2f ms |\n", op.Path, op.CookTime))
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

// handleTextures handles the aftrs_td_textures tool
func handleTextures(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := tools.GetStringParam(req, "path")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	textures, err := client.GetTextures(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if path == "" {
		path = "/project1"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Textures in %s\n\n", path))

	if len(textures) == 0 {
		sb.WriteString("No texture TOPs found.\n\n")
		sb.WriteString("*Note: Requires TouchDesigner API connection to list textures.*\n")
		return tools.TextResult(sb.String()), nil
	}

	var totalMemory float64
	for _, tex := range textures {
		totalMemory += tex.MemoryMB
	}

	sb.WriteString(fmt.Sprintf("Found **%d** textures (%.1f MB total):\n\n", len(textures), totalMemory))
	sb.WriteString("| Name | Resolution | Format | Memory |\n")
	sb.WriteString("|------|------------|--------|--------|\n")

	for _, tex := range textures {
		sb.WriteString(fmt.Sprintf("| %s | %dx%d | %s | %.1f MB |\n", tex.Name, tex.Width, tex.Height, tex.Format, tex.MemoryMB))
	}

	return tools.TextResult(sb.String()), nil
}

// handlePerformance handles the aftrs_td_performance tool
func handlePerformance(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	perf, err := client.GetPerformance(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# TouchDesigner Performance\n\n")

	sb.WriteString("## Frame Timing\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| FPS | %.1f |\n", perf.FPS))
	sb.WriteString(fmt.Sprintf("| Frame Time | %.2f ms |\n", perf.FrameTime))
	sb.WriteString(fmt.Sprintf("| Cook Time | %.2f ms |\n", perf.CookTime))
	sb.WriteString(fmt.Sprintf("| GPU Time | %.2f ms |\n", perf.GPUTime))
	sb.WriteString(fmt.Sprintf("| CPU Usage | %.1f%% |\n", perf.CPUUsage))

	if len(perf.TopCookTimes) > 0 {
		sb.WriteString("\n## Top Cook Times\n\n")
		sb.WriteString("| Operator | Cook Time |\n")
		sb.WriteString("|----------|-----------|\n")
		for _, op := range perf.TopCookTimes {
			sb.WriteString(fmt.Sprintf("| %s | %.2f ms |\n", op.Path, op.CookTime))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleGPUMemory handles the aftrs_td_gpu_memory tool
func handleGPUMemory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	mem, err := client.GetGPUMemory(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# GPU Memory Usage\n\n")

	// Calculate usage percentage
	usagePercent := (mem.UsedMB / mem.TotalMB) * 100

	// Status emoji
	statusEmoji := "✅"
	if usagePercent > 90 {
		statusEmoji = "❌"
	} else if usagePercent > 75 {
		statusEmoji = "⚠️"
	}

	sb.WriteString(fmt.Sprintf("**Status:** %s %.0f%% used\n\n", statusEmoji, usagePercent))

	sb.WriteString("## Memory Breakdown\n\n")
	sb.WriteString("| Category | Usage |\n")
	sb.WriteString("|----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total | %.0f MB |\n", mem.TotalMB))
	sb.WriteString(fmt.Sprintf("| Used | %.0f MB |\n", mem.UsedMB))
	sb.WriteString(fmt.Sprintf("| Free | %.0f MB |\n", mem.FreeMB))
	sb.WriteString(fmt.Sprintf("| Textures | %.0f MB |\n", mem.TextureMB))
	sb.WriteString(fmt.Sprintf("| Buffers | %.0f MB |\n", mem.BufferMB))
	sb.WriteString(fmt.Sprintf("| GPU Utilization | %.0f%% |\n", mem.Utilization))

	return tools.TextResult(sb.String()), nil
}

// handleErrors handles the aftrs_td_errors tool
func handleErrors(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	errors, err := client.GetErrors(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# TouchDesigner Errors\n\n")

	if len(errors) == 0 {
		sb.WriteString("✅ No errors or warnings.\n")
		return tools.TextResult(sb.String()), nil
	}

	// Count by severity
	errorCount := 0
	warningCount := 0
	for _, e := range errors {
		if e.Severity == "error" {
			errorCount++
		} else {
			warningCount++
		}
	}

	sb.WriteString(fmt.Sprintf("Found **%d errors** and **%d warnings**:\n\n", errorCount, warningCount))
	sb.WriteString("| Severity | Operator | Message |\n")
	sb.WriteString("|----------|----------|---------|\n")

	for _, e := range errors {
		icon := "⚠️"
		if e.Severity == "error" {
			icon = "❌"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", icon, e.Operator, e.Message))
	}

	return tools.TextResult(sb.String()), nil
}

// handleBackup handles the aftrs_td_backup tool
func handleBackup(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	destination := tools.GetStringParam(req, "destination")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	path, err := client.BackupProject(ctx, destination)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Project backed up to: %s", path)), nil
}

// handlePresetRecall handles the aftrs_td_preset_recall tool
func handlePresetRecall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	preset, errResult := tools.RequireStringParam(req, "preset")
	if errResult != nil {
		return errResult, nil
	}
	operator := tools.GetStringParam(req, "operator")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.RecallPreset(ctx, preset, operator)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	msg := fmt.Sprintf("✅ Recalled preset: %s", preset)
	if operator != "" {
		msg += fmt.Sprintf(" to %s", operator)
	}
	return tools.TextResult(msg), nil
}

// handleCueTrigger handles the aftrs_td_cue_trigger tool
func handleCueTrigger(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	timeline, errResult := tools.RequireStringParam(req, "timeline")
	if errResult != nil {
		return errResult, nil
	}
	cue := tools.GetStringParam(req, "cue")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.TriggerCue(ctx, timeline, cue)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	msg := fmt.Sprintf("✅ Triggered timeline: %s", timeline)
	if cue != "" {
		msg += fmt.Sprintf(" (cue: %s)", cue)
	}
	return tools.TextResult(msg), nil
}

// handleTimelines handles the aftrs_td_timelines tool
func handleTimelines(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := tools.GetStringParam(req, "path")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	timelines, err := client.GetTimelines(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if path == "" {
		path = "/project1"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Timelines in %s\n\n", path))

	if len(timelines) == 0 {
		sb.WriteString("No timeline CHOPs found.\n\n")
		sb.WriteString("*Note: Requires TouchDesigner API connection to list timelines.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** timelines:\n\n", len(timelines)))
	sb.WriteString("| Name | Length | Position | Status | Loop |\n")
	sb.WriteString("|------|--------|----------|--------|------|\n")

	for _, tl := range timelines {
		status := "⏸️ Paused"
		if tl.Playing {
			status = "▶️ Playing"
		}
		loop := "No"
		if tl.Loop {
			loop = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %.0f | %.0f | %s | %s |\n", tl.Name, tl.Length, tl.Position, status, loop))
	}

	return tools.TextResult(sb.String()), nil
}

// handleComponents handles the aftrs_td_components tool
func handleComponents(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := tools.GetStringParam(req, "path")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	components, err := client.GetComponents(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if path == "" {
		path = "/project1"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Components in %s\n\n", path))

	if len(components) == 0 {
		sb.WriteString("No container COMPs found.\n\n")
		sb.WriteString("*Note: Requires TouchDesigner API connection to list components.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** components:\n\n", len(components)))
	sb.WriteString("| Name | Type | Children | Status |\n")
	sb.WriteString("|------|------|----------|--------|\n")

	for _, comp := range components {
		status := "✅"
		if comp.HasErrors {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n", comp.Name, comp.Type, comp.Children, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleProjectInfo handles the aftrs_td_project_info tool
func handleProjectInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	info, err := client.GetProjectInfo(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# TouchDesigner Project Info\n\n")

	sb.WriteString("## General\n\n")
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", info.Name))
	sb.WriteString(fmt.Sprintf("**Path:** %s\n", info.Path))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n", info.Version))
	sb.WriteString(fmt.Sprintf("**Build:** %s\n", info.BuildVersion))

	if !info.SaveTime.IsZero() {
		sb.WriteString(fmt.Sprintf("**Last Saved:** %s\n", info.SaveTime.Format("2006-01-02 15:04:05")))
	}

	sb.WriteString("\n## Operator Counts\n\n")
	sb.WriteString("| Family | Count |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total | %d |\n", info.OperatorCount))
	sb.WriteString(fmt.Sprintf("| COMP | %d |\n", info.CompCount))
	sb.WriteString(fmt.Sprintf("| TOP | %d |\n", info.TOPCount))
	sb.WriteString(fmt.Sprintf("| CHOP | %d |\n", info.CHOPCount))
	sb.WriteString(fmt.Sprintf("| DAT | %d |\n", info.DATCount))
	sb.WriteString(fmt.Sprintf("| SOP | %d |\n", info.SOPCount))
	sb.WriteString(fmt.Sprintf("| MAT | %d |\n", info.MATCount))

	return tools.TextResult(sb.String()), nil
}

// handleCHOPs handles the aftrs_td_chops tool
func handleCHOPs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := tools.GetStringParam(req, "path")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	chops, err := client.GetCHOPs(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if path == "" {
		path = "/project1"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# CHOPs in %s\n\n", path))

	if len(chops) == 0 {
		sb.WriteString("No CHOP operators found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** CHOPs:\n\n", len(chops)))
	sb.WriteString("| Name | Type | Channels | Length | Sample Rate |\n")
	sb.WriteString("|------|------|----------|--------|-------------|\n")

	for _, chop := range chops {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %.0f Hz |\n", chop.Name, chop.Type, chop.NumChans, chop.Length, chop.SampleRate))
	}

	return tools.TextResult(sb.String()), nil
}

// handleCHOPChannels handles the aftrs_td_chop_channels tool
func handleCHOPChannels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	chopPath, errResult := tools.RequireStringParam(req, "chop")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	chop, err := client.GetCHOPChannels(ctx, chopPath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Channels: %s\n\n", chopPath))

	if len(chop.Channels) == 0 {
		sb.WriteString("No channels found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**Type:** %s | **Channels:** %d | **Length:** %d\n\n", chop.Type, chop.NumChans, chop.Length))
	sb.WriteString("| Channel | Value | Min | Max |\n")
	sb.WriteString("|---------|-------|-----|-----|\n")

	for _, ch := range chop.Channels {
		sb.WriteString(fmt.Sprintf("| %s | %.4f | %.4f | %.4f |\n", ch.Name, ch.Value, ch.Min, ch.Max))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDATs handles the aftrs_td_dats tool
func handleDATs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := tools.GetStringParam(req, "path")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	dats, err := client.GetDATs(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if path == "" {
		path = "/project1"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# DATs in %s\n\n", path))

	if len(dats) == 0 {
		sb.WriteString("No DAT operators found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** DATs:\n\n", len(dats)))
	sb.WriteString("| Name | Type | Rows | Cols |\n")
	sb.WriteString("|------|------|------|------|\n")

	for _, dat := range dats {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d |\n", dat.Name, dat.Type, dat.NumRows, dat.NumCols))
	}

	return tools.TextResult(sb.String()), nil
}

// handleDATContent handles the aftrs_td_dat_content tool
func handleDATContent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	datPath, errResult := tools.RequireStringParam(req, "dat")
	if errResult != nil {
		return errResult, nil
	}
	content := tools.GetStringParam(req, "content")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If content provided, set it
	if content != "" {
		err := client.SetDATContent(ctx, datPath, content)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("✅ Set content on %s", datPath)), nil
	}

	// Otherwise read content
	dat, err := client.GetDATContent(ctx, datPath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# DAT Content: %s\n\n", datPath))
	sb.WriteString(fmt.Sprintf("**Type:** %s | **Rows:** %d | **Cols:** %d\n\n", dat.Type, dat.NumRows, dat.NumCols))
	sb.WriteString("```\n")
	sb.WriteString(dat.Content)
	sb.WriteString("\n```\n")

	return tools.TextResult(sb.String()), nil
}

// handleRenderSettings handles the aftrs_td_render_settings tool
func handleRenderSettings(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	width := tools.GetIntParam(req, "width", 0)
	height := tools.GetIntParam(req, "height", 0)
	fps := tools.GetIntParam(req, "fps", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If any settings provided, update them
	if width > 0 || height > 0 || fps > 0 {
		settings, _ := client.GetRenderSettings(ctx)
		if width > 0 {
			settings.Width = width
		}
		if height > 0 {
			settings.Height = height
		}
		if fps > 0 {
			settings.FPS = float64(fps)
		}
		err := client.SetRenderSettings(ctx, settings)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("✅ Updated render settings: %dx%d @ %.0f FPS", settings.Width, settings.Height, settings.FPS)), nil
	}

	// Otherwise read settings
	settings, err := client.GetRenderSettings(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Render Settings\n\n")
	sb.WriteString("| Setting | Value |\n")
	sb.WriteString("|---------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Resolution | %dx%d |\n", settings.Width, settings.Height))
	sb.WriteString(fmt.Sprintf("| FPS | %.0f |\n", settings.FPS))
	sb.WriteString(fmt.Sprintf("| Pixel Format | %s |\n", settings.PixelFormat))
	sb.WriteString(fmt.Sprintf("| Antialiasing | %dx |\n", settings.Antialiasing))

	return tools.TextResult(sb.String()), nil
}

// handleCustomPars handles the aftrs_td_custom_pars tool
func handleCustomPars(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	operatorPath, errResult := tools.RequireStringParam(req, "operator")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	pars, err := client.GetCustomPars(ctx, operatorPath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Custom Parameters: %s\n\n", operatorPath))

	if len(pars) == 0 {
		sb.WriteString("No custom parameters found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** custom parameters:\n\n", len(pars)))
	sb.WriteString("| Name | Label | Type | Value | Default |\n")
	sb.WriteString("|------|-------|------|-------|--------|\n")

	for _, par := range pars {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %v | %v |\n", par.Name, par.Label, par.Type, par.Value, par.Default))
	}

	return tools.TextResult(sb.String()), nil
}

// handleToxExport handles the aftrs_td_tox_export tool
func handleToxExport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	operatorPath, errResult := tools.RequireStringParam(req, "operator")
	if errResult != nil {
		return errResult, nil
	}
	destination := tools.GetStringParam(req, "destination")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	tox, err := client.ExportTox(ctx, operatorPath, destination)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Exported %s to: %s", operatorPath, tox.Path)), nil
}

// handleVariables handles the aftrs_td_variables tool
func handleVariables(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := tools.GetStringParam(req, "name")
	value := tools.GetStringParam(req, "value")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If name and value provided, set variable
	if name != "" && value != "" {
		err := client.SetVariable(ctx, name, value)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("✅ Set variable %s = %s", name, value)), nil
	}

	// Otherwise list variables
	vars, err := client.GetVariables(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Project Variables\n\n")

	if len(vars) == 0 {
		sb.WriteString("No project variables found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** variables:\n\n", len(vars)))
	sb.WriteString("| Name | Value | Type |\n")
	sb.WriteString("|------|-------|------|\n")

	for _, v := range vars {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", v.Name, v.Value, v.Type))
	}

	return tools.TextResult(sb.String()), nil
}

// handlePulse handles the aftrs_td_pulse tool
func handlePulse(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	operatorPath, errResult := tools.RequireStringParam(req, "operator")
	if errResult != nil {
		return errResult, nil
	}
	paramName, errResult := tools.RequireStringParam(req, "param")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.PulseParameter(ctx, operatorPath, paramName)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Pulsed %s.%s", operatorPath, paramName)), nil
}

// handleReset handles the aftrs_td_reset tool
func handleReset(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	operatorPath, errResult := tools.RequireStringParam(req, "operator")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.ResetOperator(ctx, operatorPath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Reset %s to defaults", operatorPath)), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
