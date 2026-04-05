// Package whisper provides MCP tools for audio transcription using Whisper.
package whisper

import (
	"context"
	"fmt"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Whisper tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "whisper"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Audio transcription and translation using OpenAI Whisper"
}

// Tools returns the Whisper tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_whisper_transcribe",
				mcp.WithDescription("Transcribe audio file to text using Whisper"),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the audio file to transcribe")),
				mcp.WithString("language", mcp.Description("Language code (e.g., 'en', 'es', 'fr') - auto-detected if not specified")),
			),
			Handler:             handleWhisperTranscribe,
			Category:            "whisper",
			Subcategory:         "transcription",
			Tags:                []string{"whisper", "transcribe", "speech-to-text", "audio"},
			UseCases:            []string{"Transcribe audio", "Convert speech to text", "Create captions"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "whisper",
			Timeout:             5 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_whisper_translate",
				mcp.WithDescription("Transcribe and translate audio to English"),
				mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the audio file to translate")),
			),
			Handler:             handleWhisperTranslate,
			Category:            "whisper",
			Subcategory:         "translation",
			Tags:                []string{"whisper", "translate", "speech-to-text", "audio"},
			UseCases:            []string{"Translate foreign audio", "Create English subtitles", "Multi-language support"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "whisper",
			Timeout:             5 * time.Minute,
		},
		{
			Tool: mcp.NewTool("aftrs_whisper_models",
				mcp.WithDescription("List available Whisper models with size and capability info"),
			),
			Handler:             handleWhisperModels,
			Category:            "whisper",
			Subcategory:         "configuration",
			Tags:                []string{"whisper", "models", "config"},
			UseCases:            []string{"View available models", "Compare model sizes", "Choose accuracy level"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "whisper",
		},
		{
			Tool: mcp.NewTool("aftrs_whisper_health",
				mcp.WithDescription("Check Whisper service health and get troubleshooting recommendations"),
			),
			Handler:             handleWhisperHealth,
			Category:            "whisper",
			Subcategory:         "status",
			Tags:                []string{"whisper", "health", "diagnostics", "troubleshooting"},
			UseCases:            []string{"Diagnose issues", "Check installation", "Verify API configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "whisper",
		},
	}
}

var getWhisperClient = tools.LazyClient(clients.NewWhisperClient)

func handleWhisperTranscribe(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getWhisperClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Whisper client: %w", err)), nil
	}

	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}

	language := tools.GetStringParam(req, "language")

	result, err := client.Transcribe(ctx, filePath, language)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to transcribe: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

func handleWhisperTranslate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getWhisperClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Whisper client: %w", err)), nil
	}

	filePath, errResult := tools.RequireStringParam(req, "file_path")
	if errResult != nil {
		return errResult, nil
	}

	result, err := client.Translate(ctx, filePath)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to translate: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

func handleWhisperModels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getWhisperClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Whisper client: %w", err)), nil
	}

	models, err := client.GetModels(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get models: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"models": models,
		"count":  len(models),
		"note":   "Use WHISPER_MODEL environment variable to set default model",
	}), nil
}

func handleWhisperHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getWhisperClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Whisper client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
