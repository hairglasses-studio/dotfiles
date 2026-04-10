// Package ollama provides MCP tools for local LLM inference using Ollama.
package ollama

import (
	"context"
	"encoding/json"
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
	return "Local LLM inference using Ollama for text generation, chat, and native model capabilities"
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
			Tool: mcp.NewTool("aftrs_ollama_loaded",
				mcp.WithDescription("List currently loaded Ollama models from /api/ps including residency metadata"),
			),
			Handler:             handleOllamaLoaded,
			Category:            "ollama",
			Subcategory:         "status",
			Tags:                []string{"ollama", "models", "loaded", "residency"},
			UseCases:            []string{"Inspect loaded models", "Check residency metadata", "Verify keep-alive behavior"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ollama",
		},
		{
			Tool: mcp.NewTool("aftrs_ollama_show",
				mcp.WithDescription("Show detailed metadata for a specific Ollama model using /api/show"),
				mcp.WithString("model", mcp.Required(), mcp.Description("Model name or alias to inspect")),
			),
			Handler:             handleOllamaShow,
			Category:            "ollama",
			Subcategory:         "models",
			Tags:                []string{"ollama", "models", "metadata", "inspect"},
			UseCases:            []string{"Inspect model metadata", "Check capabilities", "Verify model details"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ollama",
		},
		{
			Tool: mcp.NewTool("aftrs_ollama_generate",
				mcp.WithDescription("Generate text completion using a local LLM model"),
				mcp.WithString("prompt", mcp.Required(), mcp.Description("The prompt to generate completion for")),
				mcp.WithString("model", mcp.Description("Model to use (default: code-primary)")),
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
			Tool: mcp.NewTool("aftrs_ollama_structured",
				mcp.WithDescription("Generate schema-constrained structured output using Ollama's native format support"),
				mcp.WithString("prompt", mcp.Required(), mcp.Description("Prompt to answer with structured JSON")),
				mcp.WithObject("schema", mcp.Required(), mcp.Description("JSON schema object to enforce via Ollama format")),
				mcp.WithString("model", mcp.Description("Model to use (default: code-primary)")),
				mcp.WithString("system", mcp.Description("Optional system prompt to set context")),
				mcp.WithNumber("temperature", mcp.Description("Sampling temperature (0.0-2.0, default: 0.0)")),
				mcp.WithNumber("max_tokens", mcp.Description("Maximum tokens to generate")),
			),
			Handler:             handleOllamaStructured,
			Category:            "ollama",
			Subcategory:         "generation",
			Tags:                []string{"ollama", "structured", "json", "schema"},
			UseCases:            []string{"Generate JSON output", "Enforce schemas", "Build deterministic local parsers"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "ollama",
		},
		{
			Tool: mcp.NewTool("aftrs_ollama_chat",
				mcp.WithDescription("Chat with a local LLM model using conversation history"),
				mcp.WithString("message", mcp.Required(), mcp.Description("The message to send")),
				mcp.WithString("model", mcp.Description("Model to use (default: code-primary)")),
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
			Tool: mcp.NewTool("aftrs_ollama_tool_chat",
				mcp.WithDescription("Run a native Ollama chat request with explicit tool definitions and inspect returned tool calls"),
				mcp.WithString("message", mcp.Required(), mcp.Description("The message to send")),
				mcp.WithString("model", mcp.Description("Model to use (default: code-primary)")),
				mcp.WithString("system", mcp.Description("Optional system prompt to set assistant behavior")),
				mcp.WithArray("tools", mcp.Required(), mcp.Description("Tool definitions to expose to the model"), func(schema map[string]any) {
					schema["items"] = map[string]any{"type": "object"}
				}),
				mcp.WithArray("history", mcp.Description("Previous conversation messages"), func(schema map[string]any) {
					schema["items"] = map[string]any{"type": "object"}
				}),
			),
			Handler:             handleOllamaToolChat,
			Category:            "ollama",
			Subcategory:         "chat",
			Tags:                []string{"ollama", "chat", "tools", "function-calling"},
			UseCases:            []string{"Inspect tool calls", "Prototype local tool use", "Validate native tool schemas"},
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
		{
			Tool: mcp.NewTool("aftrs_ollama_readiness",
				mcp.WithDescription("Return the live Ollama readiness report including required models, managed code-* alias state, and pull guidance"),
				mcp.WithBoolean("require_heavy", mcp.Description("Also require the heavy local coding lane (default false)")),
			),
			Handler:             handleOllamaReadiness,
			Category:            "ollama",
			Subcategory:         "status",
			Tags:                []string{"ollama", "readiness", "models", "aliases", "diagnostics"},
			UseCases:            []string{"Check local model readiness", "Verify managed aliases", "Get pull guidance before coding sessions"},
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

func handleOllamaLoaded(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getOllamaClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ollama client: %w", err)), nil
	}

	models, err := client.ListRunningModels(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list running models: %w", err)), nil
	}

	result := map[string]interface{}{
		"models": models,
		"count":  len(models),
	}
	return tools.JSONResult(result), nil
}

func handleOllamaShow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getOllamaClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ollama client: %w", err)), nil
	}

	model, errResult := tools.RequireStringParam(req, "model")
	if errResult != nil {
		return errResult, nil
	}

	info, err := client.GetModelInfo(ctx, model)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to inspect model: %w", err)), nil
	}

	return tools.JSONResult(info), nil
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

func handleOllamaStructured(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getOllamaClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ollama client: %w", err)), nil
	}

	prompt, errResult := tools.RequireStringParam(req, "prompt")
	if errResult != nil {
		return errResult, nil
	}

	args, _ := req.Params.Arguments.(map[string]interface{})
	schema, ok := args["schema"].(map[string]interface{})
	if !ok || len(schema) == 0 {
		return tools.ErrorResult(fmt.Errorf("schema must be a non-empty JSON object")), nil
	}

	model := tools.OptionalStringParam(req, "model", clients.DefaultOllamaCodeModel())
	system := tools.GetStringParam(req, "system")
	temperature := tools.GetFloatParam(req, "temperature", 0.0)
	maxTokens := tools.GetIntParam(req, "max_tokens", 0)

	generateReq := &clients.GenerateRequest{
		Model:  model,
		Prompt: prompt,
		System: system,
		Format: schema,
		Options: map[string]interface{}{
			"temperature": temperature,
		},
	}
	if maxTokens > 0 {
		generateReq.Options["num_predict"] = maxTokens
	}

	response, err := client.Generate(ctx, generateReq)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to generate structured output: %w", err)), nil
	}

	result := map[string]interface{}{
		"model":               response.Model,
		"response":            response.Response,
		"parsed_json":         parseJSON(response.Response),
		"done":                response.Done,
		"prompt_eval_count":   response.PromptEvalCount,
		"completion_tokens":   response.EvalCount,
		"total_duration_msec": response.TotalDuration / 1000000,
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

	messages := buildChatMessages(req, system, message)

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

func handleOllamaToolChat(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getOllamaClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ollama client: %w", err)), nil
	}

	message, errResult := tools.RequireStringParam(req, "message")
	if errResult != nil {
		return errResult, nil
	}

	model := tools.OptionalStringParam(req, "model", clients.DefaultOllamaCodeModel())
	system := tools.GetStringParam(req, "system")
	args, _ := req.Params.Arguments.(map[string]interface{})
	toolsArray := parseTools(args["tools"])
	if len(toolsArray) == 0 {
		return tools.ErrorResult(fmt.Errorf("tools must contain at least one valid function definition")), nil
	}

	chatReq := &clients.ChatRequest{
		Model:    model,
		Messages: buildChatMessages(req, system, message),
		Tools:    toolsArray,
	}

	response, err := client.Chat(ctx, chatReq)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to run tool chat: %w", err)), nil
	}

	result := map[string]interface{}{
		"model":      response.Model,
		"message":    response.Message,
		"tool_calls": response.Message.ToolCalls,
		"done":       response.Done,
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

func handleOllamaReadiness(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getOllamaClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Ollama client: %w", err)), nil
	}

	requireHeavy := tools.GetBoolParam(req, "require_heavy", false)
	readiness, err := client.GetReadiness(ctx, requireHeavy)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get readiness: %w", err)), nil
	}

	return tools.JSONResult(readiness), nil
}

func buildChatMessages(req mcp.CallToolRequest, system, message string) []clients.ChatMessage {
	var messages []clients.ChatMessage
	if system != "" {
		messages = append(messages, clients.ChatMessage{
			Role:    "system",
			Content: system,
		})
	}

	if args, ok := req.Params.Arguments.(map[string]interface{}); ok {
		if historyRaw, exists := args["history"]; exists {
			if historyArray, ok := historyRaw.([]interface{}); ok {
				for _, h := range historyArray {
					hMap, ok := h.(map[string]interface{})
					if !ok {
						continue
					}
					role, _ := hMap["role"].(string)
					content, _ := hMap["content"].(string)
					toolName, _ := hMap["tool_name"].(string)
					if role == "" || content == "" {
						continue
					}
					messages = append(messages, clients.ChatMessage{
						Role:      role,
						Content:   content,
						ToolName:  toolName,
						ToolCalls: parseTools(hMap["tool_calls"]),
					})
				}
			}
		}
	}

	messages = append(messages, clients.ChatMessage{
		Role:    "user",
		Content: message,
	})
	return messages
}

func parseTools(raw interface{}) []clients.OllamaTool {
	items, ok := raw.([]interface{})
	if !ok {
		return nil
	}

	toolsOut := make([]clients.OllamaTool, 0, len(items))
	for _, item := range items {
		toolMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		functionMap, ok := toolMap["function"].(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := functionMap["name"].(string)
		if name == "" {
			continue
		}
		description, _ := functionMap["description"].(string)
		parameters, _ := functionMap["parameters"].(map[string]interface{})
		arguments, _ := functionMap["arguments"].(map[string]interface{})

		toolType, _ := toolMap["type"].(string)
		if toolType == "" {
			toolType = "function"
		}

		tool := clients.OllamaTool{
			Type: toolType,
			Function: clients.OllamaToolFunction{
				Name:        name,
				Description: description,
				Parameters:  parameters,
				Arguments:   arguments,
			},
		}
		if indexValue, ok := toolMap["index"].(float64); ok {
			tool.Function.Index = int(indexValue)
		} else if indexValue, ok := functionMap["index"].(float64); ok {
			tool.Function.Index = int(indexValue)
		}
		toolsOut = append(toolsOut, tool)
	}
	return toolsOut
}

func parseJSON(raw string) interface{} {
	var parsed interface{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil
	}
	return parsed
}
