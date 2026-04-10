package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/ollama"
	"github.com/mark3labs/mcp-go/mcp"
)

type evalRequest struct {
	ToolName string                 `json:"tool_name"`
	Args     map[string]interface{} `json:"args"`
}

type evalResponse struct {
	ToolName string      `json:"tool_name"`
	Result   interface{} `json:"result"`
}

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		die("read stdin: %v", err)
	}

	var req evalRequest
	if err := json.Unmarshal(input, &req); err != nil {
		die("decode request: %v", err)
	}

	toolName := strings.TrimSpace(req.ToolName)
	if toolName == "" {
		die("tool_name is required")
	}

	module := &ollama.Module{}
	var handler func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error)
	for _, tool := range module.Tools() {
		if tool.Tool.Name == toolName {
			handler = tool.Handler
			break
		}
	}
	if handler == nil {
		die("unsupported tool %q", toolName)
	}

	timeout := 90 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	mcpReq := mcp.CallToolRequest{}
	mcpReq.Params.Name = toolName
	mcpReq.Params.Arguments = req.Args

	result, err := handler(ctx, mcpReq)
	if err != nil {
		die("handler error: %v", err)
	}

	if result.IsError {
		die("tool returned error: %v", result.Content)
	}

	var parsedResult interface{}
	if len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			_ = json.Unmarshal([]byte(textContent.Text), &parsedResult)
			if parsedResult == nil {
				parsedResult = textContent.Text
			}
		} else {
			parsedResult = result.Content
		}
	}

	resp := evalResponse{
		ToolName: toolName,
		Result:   parsedResult,
	}
	if err := json.NewEncoder(os.Stdout).Encode(resp); err != nil {
		die("encode response: %v", err)
	}
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}