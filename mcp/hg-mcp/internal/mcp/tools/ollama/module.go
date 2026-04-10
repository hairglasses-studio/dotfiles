// Package ollama provides MCP tools for local LLM inference using Ollama.
package ollama

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Ollama tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "ollama"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Local LLM inference using Ollama for text generation and chat"
}

// Tools returns the Ollama tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ollama_status",
				mcp.WithDescription("Get Ollama service status including version and loaded models"),
			),
			Handler:             handleOllamaStatus,
			Category:            "ollama",
			Subcategory:         "status",
			Tags:                []string{"ollama", "llm", "status", "ai"},
			UseCases:            []string{"Check service status", "View loaded models", "Verify connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ollama",
		},
		{
			Tool: mcp.NewTool("aftrs_ollama_models",
				mcp.WithDescription("List available Ollama models installed locally"),
			),
			Handler:             handleOllamaModels,
			Category:            "ollama",
			Subcategory:         "models",
			Tags:                []string{"ollama", "models", "list", "ai"},
			UseCases:            []string{"List installed models", "Check model sizes", "View available options"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ollama",
		},
		{
			Tool: mcp.NewTool("aftrs_ollama_generate",
				mcp.WithDescription("Generate text completion using a local LLM model"),
				mcp.WithString("prompt", mcp.Required(), mcp.Description("The prompt to generate completion for")),
				mcp.WithString("model", mcp.Description("Model to use (default: qwen3:8b)")),
				mcp.WithString("system", mcp.Description("Optional system prompt to set context")),
				mcp.WithNumber("temperature", mcp.Description("Sampling temperature (0.0-2.0, default: 0.8)")),
				mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens to generate")),
			),
			Handler:             handleOllamaGenerate,
			Category:            "ollama",
			Subcategory:         "generation",
			Tags:                []string{"ollama", "generate", "completion", "ai"},
			UseCases:            []string{"Generate text", "Complete prompts", "Creative writing"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ollama",
		},
		{
			Tool: mcp.NewTool("aftrs_ollama_chat",
				mcp.WithDescription("Chat with a local LLM model using conversation history"),
				mcp.WithString("message", mcp.Required(), mcp.Description("The message to send")),
				mcp.WithString("model", mcp.Description("Model to use (default: qwen3:8b)")),
				mcp.WithString("system", mcp.Description("Optional system prompt to set assistant behavior")),
				mcp.WithArray("history", mcp.Description("Previous conversation messages")),
			),
			Handler:             handleOllamaChat,
			Category:            "ollama",
			Subcategory:         "chat",
			Tags:                []string{"ollama", "chat", "conversation", "ai"},
			UseCases:            []string{"Chat with AI", "Multi-turn conversation", "Interactive Q&A"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ollama",
		},
		{
			Tool: mcp.NewTool("aftrs_ollama_health",
				mcp.WithDescription("Check Ollama service health and get troubleshooting recommendations"),
			),
			Handler:             handleOllamaHealth,
			Category:            "ollama",
			Subcategory:         "status",
			Tags:                []string{"ollama", "health", "diagnostics", "troubleshooting"},
			UseCases:            []string{"Diagnose issues", "Check service health", "Get troubleshooting tips"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ollama",
		},
	}
}

var getOllamaClient = tools.LazyClient(clients.NewOllamaClient)

func handleOllamaStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getOllamaClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ollama client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleOllamaModels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getOllamaClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ollama client: %w", err)), nil
	}

	models, err := client.ListModels(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list models: %w", err)), nil
	}

	result := map[string]interface{}{
		"models": models,
		"count":  len(models),
	}
	return tools.JSONResult(result), nil
}

func handleOllamaGenerate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getOllamaClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ollama client: %w", err)), nil
	}

	prompt, errResult := tools.RequireStringParam(req, "prompt")
	if errResult != nil {
		return errResult, nil
	}

	model := tools.OptionalStringParam(req, "model", clients.DefaultOllamaChatModel())

	system := tools.GetStringParam(req, "system")
	temperature := tools.GetFloatParam(req, "temperature", 0.8)
	maxTokens := tools.GetIntParam(req, "max_tokens", 0)

	generateReq := &clients.GenerateRequest{
		Model:  model,
		Prompt: prompt,
		System: system,
		Options: map[string]interface{}{
			"temperature": temperature,
		},
	}

	if maxTokens > 0 {
		generateReq.Options["num_predict"] = maxTokens
	}

	response, err := client.Generate(ctx, generateReq)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to generate: %w", err)), nil
	}

	result := map[string]interface{}{
		"model":    response.Model,
		"response": response.Response,
		"done":     response.Done,
		"stats": map[string]interface{}{
			"total_duration_ms": response.TotalDuration / 1000000,
			"prompt_eval_count": response.PromptEvalCount,
			"eval_count":        response.EvalCount,
		},
	}
	return tools.JSONResult(result), nil
}

func handleOllamaChat(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getOllamaClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ollama client: %w", err)), nil
	}

	message, errResult := tools.RequireStringParam(req, "message")
	if errResult != nil {
		return errResult, nil
	}

	model := tools.OptionalStringParam(req, "model", clients.DefaultOllamaChatModel())

	system := tools.GetStringParam(req, "system")

	// Build messages
	var messages []clients.ChatMessage

	// Add system message if provided
	if system != "" {
		messages = append(messages, clients.ChatMessage{
			Role:    "system",
			Content: system,
		})
	}

	// Parse history if provided
	args, ok := req.Params.Arguments.(map[string]interface{})
	if ok {
		if historyRaw, exists := args["history"]; exists {
			if historyArray, ok := historyRaw.([]interface{}); ok {
				for _, h := range historyArray {
					if hMap, ok := h.(map[string]interface{}); ok {
						role, _ := hMap["role"].(string)
						content, _ := hMap["content"].(string)
						if role != "" && content != "" {
							messages = append(messages, clients.ChatMessage{
								Role:    role,
								Content: content,
							})
						}
					}
				}
			}
		}
	}

	// Add current message
	messages = append(messages, clients.ChatMessage{
		Role:    "user",
		Content: message,
	})

	chatReq := &clients.ChatRequest{
		Model:    model,
		Messages: messages,
	}

	response, err := client.Chat(ctx, chatReq)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to chat: %w", err)), nil
	}

	result := map[string]interface{}{
		"model":   response.Model,
		"message": response.Message,
		"done":    response.Done,
		"stats": map[string]interface{}{
			"total_duration_ms": response.TotalDuration / 1000000,
			"prompt_eval_count": response.PromptEvalCount,
			"eval_count":        response.EvalCount,
		},
	}
	return tools.JSONResult(result), nil
}

func handleOllamaHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getOllamaClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ollama client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
