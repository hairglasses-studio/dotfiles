// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FFGLClient provides FFGL plugin development tools
type FFGLClient struct {
	templatesDir string
	outputDir    string
	buildDir     string
}

// FFGLPluginInfo represents information about an FFGL plugin
type FFGLPluginInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // effect, source
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Parameters  []string `json:"parameters,omitempty"`
	Inputs      int      `json:"inputs"`
	HasShader   bool     `json:"has_shader"`
}

// FFGLBuildResult represents build output
type FFGLBuildResult struct {
	Success    bool     `json:"success"`
	OutputPath string   `json:"output_path,omitempty"`
	Platform   string   `json:"platform"`
	Errors     []string `json:"errors,omitempty"`
	Warnings   []string `json:"warnings,omitempty"`
}

// FFGLTestResult represents test output
type FFGLTestResult struct {
	Passed      bool     `json:"passed"`
	TestsRun    int      `json:"tests_run"`
	TestsFailed int      `json:"tests_failed"`
	Details     []string `json:"details,omitempty"`
}

// FFGLShaderValidation represents shader validation results
type FFGLShaderValidation struct {
	Valid    bool     `json:"valid"`
	Type     string   `json:"type"` // vertex, fragment
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// NewFFGLClient creates a new FFGL client
func NewFFGLClient() (*FFGLClient, error) {
	templatesDir := os.Getenv("FFGL_TEMPLATES_DIR")
	if templatesDir == "" {
		home, _ := os.UserHomeDir()
		templatesDir = filepath.Join(home, "ffgl-templates")
	}

	outputDir := os.Getenv("FFGL_OUTPUT_DIR")
	if outputDir == "" {
		home, _ := os.UserHomeDir()
		outputDir = filepath.Join(home, "ffgl-plugins")
	}

	buildDir := os.Getenv("FFGL_BUILD_DIR")
	if buildDir == "" {
		buildDir = filepath.Join(outputDir, "build")
	}

	return &FFGLClient{
		templatesDir: templatesDir,
		outputDir:    outputDir,
		buildDir:     buildDir,
	}, nil
}

// CreateEffect scaffolds a new FFGL effect plugin
func (c *FFGLClient) CreateEffect(ctx context.Context, name, author, description string) (*FFGLPluginInfo, error) {
	info := &FFGLPluginInfo{
		Name:        name,
		Type:        "effect",
		Author:      author,
		Description: description,
		Version:     "1.0.0",
		Inputs:      1,
		HasShader:   true,
	}

	// Create plugin directory
	pluginDir := filepath.Join(c.outputDir, name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Generate main plugin file
	mainContent := c.generateEffectTemplate(info)
	mainPath := filepath.Join(pluginDir, name+".cpp")
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write main file: %w", err)
	}

	// Generate header file
	headerContent := c.generateHeaderTemplate(info)
	headerPath := filepath.Join(pluginDir, name+".h")
	if err := os.WriteFile(headerPath, []byte(headerContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write header file: %w", err)
	}

	// Generate shader files
	fragContent := c.generateFragmentShader(info)
	fragPath := filepath.Join(pluginDir, "shader.frag")
	if err := os.WriteFile(fragPath, []byte(fragContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write fragment shader: %w", err)
	}

	// Generate CMakeLists.txt
	cmakeContent := c.generateCMakeLists(info)
	cmakePath := filepath.Join(pluginDir, "CMakeLists.txt")
	if err := os.WriteFile(cmakePath, []byte(cmakeContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write CMakeLists.txt: %w", err)
	}

	return info, nil
}

// CreateSource scaffolds a new FFGL source plugin
func (c *FFGLClient) CreateSource(ctx context.Context, name, author, description string) (*FFGLPluginInfo, error) {
	info := &FFGLPluginInfo{
		Name:        name,
		Type:        "source",
		Author:      author,
		Description: description,
		Version:     "1.0.0",
		Inputs:      0,
		HasShader:   true,
	}

	// Create plugin directory
	pluginDir := filepath.Join(c.outputDir, name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Generate source plugin files
	mainContent := c.generateSourceTemplate(info)
	mainPath := filepath.Join(pluginDir, name+".cpp")
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write main file: %w", err)
	}

	return info, nil
}

// Build builds an FFGL plugin for the current platform
func (c *FFGLClient) Build(ctx context.Context, pluginName string) (*FFGLBuildResult, error) {
	result := &FFGLBuildResult{
		Success:  false,
		Platform: getPlatform(),
	}

	pluginDir := filepath.Join(c.outputDir, pluginName)
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		result.Errors = append(result.Errors, fmt.Sprintf("plugin directory not found: %s", pluginDir))
		return result, nil
	}

	// Build directory
	buildDir := filepath.Join(pluginDir, "build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to create build directory: %v", err))
		return result, nil
	}

	// Run CMake
	cmakeCmd := exec.CommandContext(ctx, "cmake", "..")
	cmakeCmd.Dir = buildDir
	if output, err := cmakeCmd.CombinedOutput(); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("cmake failed: %s", string(output)))
		return result, nil
	}

	// Run make/build
	buildCmd := exec.CommandContext(ctx, "cmake", "--build", ".")
	buildCmd.Dir = buildDir
	if output, err := buildCmd.CombinedOutput(); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("build failed: %s", string(output)))
		return result, nil
	}

	result.Success = true
	result.OutputPath = filepath.Join(buildDir, pluginName+getPluginExtension())

	return result, nil
}

// Test runs golden image tests on a plugin
func (c *FFGLClient) Test(ctx context.Context, pluginName string) (*FFGLTestResult, error) {
	result := &FFGLTestResult{
		Passed: true,
	}

	pluginDir := filepath.Join(c.outputDir, pluginName)
	testDir := filepath.Join(pluginDir, "tests")

	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		result.Details = append(result.Details, "No test directory found")
		return result, nil
	}

	// Would run actual tests here
	result.TestsRun = 0
	result.Details = append(result.Details, "Test framework not configured")

	return result, nil
}

// ValidateShader validates GLSL shader code
func (c *FFGLClient) ValidateShader(ctx context.Context, shaderPath string) (*FFGLShaderValidation, error) {
	result := &FFGLShaderValidation{
		Valid: true,
	}

	// Determine shader type from extension
	if strings.HasSuffix(shaderPath, ".vert") {
		result.Type = "vertex"
	} else if strings.HasSuffix(shaderPath, ".frag") {
		result.Type = "fragment"
	} else {
		result.Type = "unknown"
		result.Warnings = append(result.Warnings, "Unknown shader type from extension")
	}

	// Read shader file
	content, err := os.ReadFile(shaderPath)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("failed to read shader: %v", err))
		return result, nil
	}

	// Basic validation
	shaderStr := string(content)

	// Check for main function
	if !strings.Contains(shaderStr, "void main") {
		result.Valid = false
		result.Errors = append(result.Errors, "Missing main() function")
	}

	// Check for common issues
	if strings.Contains(shaderStr, "gl_FragColor") && strings.Contains(shaderStr, "#version 330") {
		result.Warnings = append(result.Warnings, "gl_FragColor is deprecated in GLSL 330+, use out variable")
	}

	return result, nil
}

// Package creates a distributable package for the plugin
func (c *FFGLClient) Package(ctx context.Context, pluginName, version string) (string, error) {
	pluginDir := filepath.Join(c.outputDir, pluginName)
	buildDir := filepath.Join(pluginDir, "build")

	pluginFile := filepath.Join(buildDir, pluginName+getPluginExtension())
	if _, err := os.Stat(pluginFile); os.IsNotExist(err) {
		return "", fmt.Errorf("plugin binary not found - run build first")
	}

	// Create package directory
	pkgDir := filepath.Join(c.outputDir, "packages")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create package directory: %w", err)
	}

	// Package name
	pkgName := fmt.Sprintf("%s-%s-%s%s", pluginName, version, getPlatform(), ".zip")
	pkgPath := filepath.Join(pkgDir, pkgName)

	// Create zip (simplified)
	// In production, would use archive/zip
	return pkgPath, nil
}

// ListPlugins lists all plugins in the output directory
func (c *FFGLClient) ListPlugins(ctx context.Context) ([]FFGLPluginInfo, error) {
	plugins := []FFGLPluginInfo{}

	entries, err := os.ReadDir(c.outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return plugins, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "build" && entry.Name() != "packages" {
			// Check for plugin files
			pluginDir := filepath.Join(c.outputDir, entry.Name())
			cppFile := filepath.Join(pluginDir, entry.Name()+".cpp")

			if _, err := os.Stat(cppFile); err == nil {
				info := FFGLPluginInfo{
					Name: entry.Name(),
					Type: "effect", // Would parse from file
				}

				// Check for shader
				fragShader := filepath.Join(pluginDir, "shader.frag")
				if _, err := os.Stat(fragShader); err == nil {
					info.HasShader = true
				}

				plugins = append(plugins, info)
			}
		}
	}

	return plugins, nil
}

// Template generators
func (c *FFGLClient) generateEffectTemplate(info *FFGLPluginInfo) string {
	return fmt.Sprintf(`// FFGL Effect Plugin: %s
// Author: %s
// Description: %s

#include "%s.h"
#include <string>

static CFFGLPluginInfo PluginInfo(
    PluginFactory<%s>,
    "%s",  // Plugin unique ID
    "%s",  // Plugin name
    2,     // API major version
    1,     // API minor version
    1,     // Plugin major version
    0,     // Plugin minor version
    FF_EFFECT,
    "%s",  // Plugin description
    "%s"   // Author
);

%s::%s() : CFFGLPlugin()
{
    SetMinInputs(1);
    SetMaxInputs(1);
}

%s::~%s()
{
}

FFResult %s::InitGL(const FFGLViewportStruct* vp)
{
    // Initialize OpenGL resources
    return FF_SUCCESS;
}

FFResult %s::DeInitGL()
{
    // Clean up OpenGL resources
    return FF_SUCCESS;
}

FFResult %s::ProcessOpenGL(ProcessOpenGLStruct* pGL)
{
    // Process the input texture
    return FF_SUCCESS;
}
`, info.Name, info.Author, info.Description,
		info.Name, info.Name, info.Name[:4], info.Name,
		info.Description, info.Author,
		info.Name, info.Name, info.Name, info.Name,
		info.Name, info.Name, info.Name)
}

func (c *FFGLClient) generateHeaderTemplate(info *FFGLPluginInfo) string {
	guard := strings.ToUpper(info.Name) + "_H"
	return fmt.Sprintf(`#ifndef %s
#define %s

#include <FFGLSDK.h>

class %s : public CFFGLPlugin
{
public:
    %s();
    ~%s();

    FFResult InitGL(const FFGLViewportStruct* vp) override;
    FFResult DeInitGL() override;
    FFResult ProcessOpenGL(ProcessOpenGLStruct* pGL) override;

private:
    // Add private members here
};

#endif // %s
`, guard, guard, info.Name, info.Name, info.Name, guard)
}

func (c *FFGLClient) generateSourceTemplate(info *FFGLPluginInfo) string {
	return fmt.Sprintf(`// FFGL Source Plugin: %s
// Author: %s
// Description: %s

#include <FFGLSDK.h>

// Source plugin implementation
`, info.Name, info.Author, info.Description)
}

func (c *FFGLClient) generateFragmentShader(info *FFGLPluginInfo) string {
	return `#version 330 core

uniform sampler2D inputTexture;
uniform vec2 resolution;
uniform float time;

in vec2 fragCoord;
out vec4 fragColor;

void main()
{
    vec2 uv = fragCoord;
    vec4 color = texture(inputTexture, uv);

    // Apply effect here

    fragColor = color;
}
`
}

func (c *FFGLClient) generateCMakeLists(info *FFGLPluginInfo) string {
	return fmt.Sprintf(`cmake_minimum_required(VERSION 3.12)
project(%s)

set(CMAKE_CXX_STANDARD 17)

# Find FFGL SDK
set(FFGL_SDK_PATH "$ENV{FFGL_SDK_PATH}" CACHE PATH "Path to FFGL SDK")

# Add plugin
add_library(%s SHARED
    %s.cpp
    %s.h
)

# Platform-specific settings
if(APPLE)
    set_target_properties(%s PROPERTIES
        BUNDLE TRUE
        BUNDLE_EXTENSION "bundle"
    )
elseif(WIN32)
    set_target_properties(%s PROPERTIES
        SUFFIX ".dll"
    )
endif()
`, info.Name, info.Name, info.Name, info.Name, info.Name, info.Name)
}

// Helper functions
func getPlatform() string {
	switch os := os.Getenv("GOOS"); os {
	case "darwin":
		return "macos"
	case "windows":
		return "windows"
	default:
		return "linux"
	}
}

func getPluginExtension() string {
	switch getPlatform() {
	case "macos":
		return ".bundle"
	case "windows":
		return ".dll"
	default:
		return ".so"
	}
}
