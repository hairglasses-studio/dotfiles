// Package ffgl provides FFGL plugin development tools for hg-mcp.
package ffgl

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for FFGL development
type Module struct{}

func (m *Module) Name() string {
	return "ffgl"
}

func (m *Module) Description() string {
	return "FFGL plugin development scaffolding and build tools for Resolume"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ffgl_list",
				mcp.WithDescription("List FFGL plugins in the development directory."),
			),
			Handler:             handleList,
			Category:            "ffgl",
			Subcategory:         "plugins",
			Tags:                []string{"ffgl", "plugins", "list", "resolume"},
			UseCases:            []string{"View plugin projects", "Check development status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ffgl",
		},
		{
			Tool: mcp.NewTool("aftrs_ffgl_create_effect",
				mcp.WithDescription("Scaffold a new FFGL effect plugin with shader template."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name (no spaces)")),
				mcp.WithString("author", mcp.Description("Plugin author name")),
				mcp.WithString("description", mcp.Description("Plugin description")),
			),
			Handler:             handleCreateEffect,
			Category:            "ffgl",
			Subcategory:         "create",
			Tags:                []string{"ffgl", "effect", "create", "scaffold"},
			UseCases:            []string{"Create new effect plugin", "Start plugin development"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ffgl",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ffgl_create_source",
				mcp.WithDescription("Scaffold a new FFGL source plugin (video generator)."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name (no spaces)")),
				mcp.WithString("author", mcp.Description("Plugin author name")),
				mcp.WithString("description", mcp.Description("Plugin description")),
			),
			Handler:             handleCreateSource,
			Category:            "ffgl",
			Subcategory:         "create",
			Tags:                []string{"ffgl", "source", "create", "generator"},
			UseCases:            []string{"Create video source plugin", "Build generator"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ffgl",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ffgl_build",
				mcp.WithDescription("Build an FFGL plugin for the current platform."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name to build")),
			),
			Handler:             handleBuild,
			Category:            "ffgl",
			Subcategory:         "build",
			Tags:                []string{"ffgl", "build", "compile", "cmake"},
			UseCases:            []string{"Build plugin binary", "Compile for platform"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "ffgl",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_ffgl_test",
				mcp.WithDescription("Run golden image tests on an FFGL plugin."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name to test")),
			),
			Handler:             handleTest,
			Category:            "ffgl",
			Subcategory:         "test",
			Tags:                []string{"ffgl", "test", "golden", "validation"},
			UseCases:            []string{"Test plugin output", "Validate rendering"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ffgl",
		},
		{
			Tool: mcp.NewTool("aftrs_ffgl_validate_shader",
				mcp.WithDescription("Validate GLSL shader code for FFGL compatibility."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to shader file")),
			),
			Handler:             handleValidateShader,
			Category:            "ffgl",
			Subcategory:         "validate",
			Tags:                []string{"ffgl", "shader", "glsl", "validate"},
			UseCases:            []string{"Check shader syntax", "Validate GLSL code"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ffgl",
		},
		{
			Tool: mcp.NewTool("aftrs_ffgl_package",
				mcp.WithDescription("Package an FFGL plugin for distribution."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Plugin name to package")),
				mcp.WithString("version", mcp.Description("Version string (default: 1.0.0)")),
			),
			Handler:             handlePackage,
			Category:            "ffgl",
			Subcategory:         "package",
			Tags:                []string{"ffgl", "package", "distribute", "release"},
			UseCases:            []string{"Create distribution package", "Prepare for release"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ffgl",
			IsWrite:             true,
		},
	}
}

var getClient = tools.LazyClient(clients.NewFFGLClient)

// handleList handles the aftrs_ffgl_list tool
func handleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	plugins, err := client.ListPlugins(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# FFGL Plugin Projects\n\n")

	if len(plugins) == 0 {
		sb.WriteString("No plugin projects found.\n\n")
		sb.WriteString("Use `aftrs_ffgl_create_effect` or `aftrs_ffgl_create_source` to create a new plugin.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** plugins:\n\n", len(plugins)))
	sb.WriteString("| Name | Type | Shader | Version |\n")
	sb.WriteString("|------|------|--------|----------|\n")

	for _, p := range plugins {
		shader := "No"
		if p.HasShader {
			shader = "Yes"
		}
		version := p.Version
		if version == "" {
			version = "-"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			p.Name, p.Type, shader, version))
	}

	return tools.TextResult(sb.String()), nil
}

// handleCreateEffect handles the aftrs_ffgl_create_effect tool
func handleCreateEffect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	author := tools.OptionalStringParam(req, "author", "AFTRS Studio")

	description := tools.OptionalStringParam(req, "description", "FFGL Effect Plugin")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	info, err := client.CreateEffect(ctx, name, author, description)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# FFGL Effect Plugin Created\n\n")
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", info.Name))
	sb.WriteString(fmt.Sprintf("**Type:** %s\n", info.Type))
	sb.WriteString(fmt.Sprintf("**Author:** %s\n", info.Author))
	sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", info.Description))

	sb.WriteString("## Generated Files\n\n")
	sb.WriteString(fmt.Sprintf("- `%s.cpp` - Main plugin implementation\n", name))
	sb.WriteString(fmt.Sprintf("- `%s.h` - Plugin header\n", name))
	sb.WriteString("- `shader.frag` - Fragment shader template\n")
	sb.WriteString("- `CMakeLists.txt` - Build configuration\n\n")

	sb.WriteString("## Next Steps\n\n")
	sb.WriteString("1. Edit `shader.frag` to implement your effect\n")
	sb.WriteString("2. Add parameters in the header file\n")
	sb.WriteString("3. Build with `aftrs_ffgl_build`\n")

	return tools.TextResult(sb.String()), nil
}

// handleCreateSource handles the aftrs_ffgl_create_source tool
func handleCreateSource(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	author := tools.OptionalStringParam(req, "author", "AFTRS Studio")

	description := tools.OptionalStringParam(req, "description", "FFGL Source Plugin")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	info, err := client.CreateSource(ctx, name, author, description)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# FFGL Source Plugin Created\n\n")
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", info.Name))
	sb.WriteString(fmt.Sprintf("**Type:** %s (video generator)\n", info.Type))
	sb.WriteString(fmt.Sprintf("**Author:** %s\n\n", info.Author))

	sb.WriteString("## Next Steps\n\n")
	sb.WriteString("1. Implement video generation in the main file\n")
	sb.WriteString("2. Add parameters for control\n")
	sb.WriteString("3. Build with `aftrs_ffgl_build`\n")

	return tools.TextResult(sb.String()), nil
}

// handleBuild handles the aftrs_ffgl_build tool
func handleBuild(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Build(ctx, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Build: %s\n\n", name))

	if result.Success {
		sb.WriteString("**Status:** Success\n")
		sb.WriteString(fmt.Sprintf("**Platform:** %s\n", result.Platform))
		sb.WriteString(fmt.Sprintf("**Output:** `%s`\n", result.OutputPath))
	} else {
		sb.WriteString("**Status:** Failed\n\n")
		sb.WriteString("## Errors\n\n")
		for _, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("- %s\n", err))
		}
	}

	if len(result.Warnings) > 0 {
		sb.WriteString("\n## Warnings\n\n")
		for _, warn := range result.Warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", warn))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleTest handles the aftrs_ffgl_test tool
func handleTest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Test(ctx, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Test Results: %s\n\n", name))

	status := "Passed"
	if !result.Passed {
		status = "Failed"
	}
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", status))
	sb.WriteString(fmt.Sprintf("**Tests Run:** %d\n", result.TestsRun))
	sb.WriteString(fmt.Sprintf("**Tests Failed:** %d\n", result.TestsFailed))

	if len(result.Details) > 0 {
		sb.WriteString("\n## Details\n\n")
		for _, detail := range result.Details {
			sb.WriteString(fmt.Sprintf("- %s\n", detail))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleValidateShader handles the aftrs_ffgl_validate_shader tool
func handleValidateShader(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.ValidateShader(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Shader Validation\n\n")
	sb.WriteString(fmt.Sprintf("**File:** `%s`\n", path))
	sb.WriteString(fmt.Sprintf("**Type:** %s\n", result.Type))

	status := "Valid"
	if !result.Valid {
		status = "Invalid"
	}
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", status))

	if len(result.Errors) > 0 {
		sb.WriteString("## Errors\n\n")
		for _, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("- %s\n", err))
		}
	}

	if len(result.Warnings) > 0 {
		sb.WriteString("\n## Warnings\n\n")
		for _, warn := range result.Warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", warn))
		}
	}

	if result.Valid && len(result.Errors) == 0 && len(result.Warnings) == 0 {
		sb.WriteString("No issues found.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handlePackage handles the aftrs_ffgl_package tool
func handlePackage(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	version := tools.OptionalStringParam(req, "version", "1.0.0")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	pkgPath, err := client.Package(ctx, name, version)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Package Created\n\n")
	sb.WriteString(fmt.Sprintf("**Plugin:** %s\n", name))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n", version))
	sb.WriteString(fmt.Sprintf("**Package:** `%s`\n\n", pkgPath))

	sb.WriteString("## Installation\n\n")
	sb.WriteString("Copy the package contents to:\n")
	sb.WriteString("- **macOS:** `~/Documents/Resolume Arena/Extra Effects/`\n")
	sb.WriteString("- **Windows:** `C:\\Users\\<user>\\Documents\\Resolume Arena\\Extra Effects\\`\n")

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
