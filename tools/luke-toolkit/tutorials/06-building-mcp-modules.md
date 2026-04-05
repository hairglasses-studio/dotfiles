# Building MCP Modules

Create your own tools that Claude can use. This tutorial uses Go, the language aftrs-mcp is built with.

## Prerequisites

```bash
# Install Go
brew install go

# Verify installation
go version
```

---

## Project Structure

```
my-mcp-server/
├── go.mod
├── go.sum
├── main.go                    # Entry point
└── internal/
    └── mcp/
        └── tools/
            └── mymodule/
                └── module.go  # Your tools
```

---

## Step 1: Initialize Project

```bash
# Create directory
mkdir my-mcp-server
cd my-mcp-server

# Initialize Go module
go mod init my-mcp-server

# Get MCP library
go get github.com/mark3labs/mcp-go
```

---

## Step 2: Create Main Entry Point

Create `main.go`:

```go
package main

import (
    "log"
    "os"

    "github.com/mark3labs/mcp-go/server"
    _ "my-mcp-server/internal/mcp/tools/mymodule"
)

func main() {
    s := server.NewMCPServer(
        "My MCP Server",
        "1.0.0",
    )

    if err := server.ServeStdio(s); err != nil {
        log.Printf("Server error: %v", err)
        os.Exit(1)
    }
}
```

---

## Step 3: Create Your Module

Create `internal/mcp/tools/mymodule/module.go`:

```go
package mymodule

import (
    "context"
    "fmt"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

// Module definition
type Module struct{}

func (m *Module) Name() string {
    return "mymodule"
}

func (m *Module) Description() string {
    return "My custom tools"
}

// Define your tools
func (m *Module) Tools() []server.ServerTool {
    return []server.ServerTool{
        {
            Tool: mcp.Tool{
                Name:        "my_hello",
                Description: "Says hello to someone",
                InputSchema: mcp.ToolInputSchema{
                    Type: "object",
                    Properties: map[string]interface{}{
                        "name": map[string]interface{}{
                            "type":        "string",
                            "description": "Name to greet",
                        },
                    },
                    Required: []string{"name"},
                },
            },
            Handler: handleHello,
        },
    }
}

// Handler function
func handleHello(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Get parameters
    name, ok := request.Params.Arguments["name"].(string)
    if !ok {
        return nil, fmt.Errorf("name is required")
    }

    // Do something
    message := fmt.Sprintf("Hello, %s!", name)

    // Return result
    return mcp.NewToolResultText(message), nil
}

// Register module on import
func init() {
    // Your registration code here
}
```

---

## Step 4: Build and Test

```bash
# Build the server
go build -o my-mcp-server .

# Test it runs
./my-mcp-server
```

---

## Step 5: Add to Claude

Edit `~/.claude/settings.json`:

```json
{
  "mcpServers": {
    "my-server": {
      "command": "/path/to/my-mcp-server"
    }
  }
}
```

Restart Claude Code and test:

```
"Use the hello tool to greet Luke"
```

---

## Tool Patterns

### Simple Tool (No Parameters)

```go
{
    Tool: mcp.Tool{
        Name:        "my_status",
        Description: "Get current status",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{},
        },
    },
    Handler: func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        return mcp.NewToolResultText("All systems operational"), nil
    },
}
```

### Tool with Multiple Parameters

```go
{
    Tool: mcp.Tool{
        Name:        "my_calculate",
        Description: "Calculate two numbers",
        InputSchema: mcp.ToolInputSchema{
            Type: "object",
            Properties: map[string]interface{}{
                "a": map[string]interface{}{
                    "type":        "number",
                    "description": "First number",
                },
                "b": map[string]interface{}{
                    "type":        "number",
                    "description": "Second number",
                },
                "operation": map[string]interface{}{
                    "type":        "string",
                    "description": "add, subtract, multiply, divide",
                },
            },
            Required: []string{"a", "b", "operation"},
        },
    },
    Handler: handleCalculate,
}
```

### Tool with Optional Parameters

```go
Properties: map[string]interface{}{
    "required_param": map[string]interface{}{
        "type":        "string",
        "description": "This is required",
    },
    "optional_param": map[string]interface{}{
        "type":        "string",
        "description": "This is optional",
        "default":     "default_value",
    },
},
Required: []string{"required_param"},  // Only required ones
```

---

## Error Handling

```go
func handleMyTool(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Validate parameters
    param, ok := req.Params.Arguments["param"].(string)
    if !ok || param == "" {
        return nil, fmt.Errorf("param is required and must be a string")
    }

    // Do risky operation
    result, err := riskyOperation(param)
    if err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }

    return mcp.NewToolResultText(result), nil
}
```

---

## Calling External APIs

```go
import (
    "encoding/json"
    "net/http"
)

func handleAPICall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Make HTTP request
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        return nil, fmt.Errorf("API call failed: %w", err)
    }
    defer resp.Body.Close()

    // Parse response
    var data map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, fmt.Errorf("failed to parse response: %w", err)
    }

    // Return as JSON
    result, _ := json.MarshalIndent(data, "", "  ")
    return mcp.NewToolResultText(string(result)), nil
}
```

---

## Environment Variables

```go
import "os"

func handleWithConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    apiKey := os.Getenv("MY_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf("MY_API_KEY environment variable not set")
    }

    // Use apiKey...
}
```

Configure in settings:

```json
{
  "mcpServers": {
    "my-server": {
      "command": "/path/to/server",
      "env": {
        "MY_API_KEY": "secret-key-here"
      }
    }
  }
}
```

---

## Complete Example: File Counter

```go
package filetools

import (
    "context"
    "fmt"
    "os"
    "path/filepath"

    "github.com/mark3labs/mcp-go/mcp"
    "github.com/mark3labs/mcp-go/server"
)

type Module struct{}

func (m *Module) Name() string { return "filetools" }
func (m *Module) Description() string { return "File utility tools" }

func (m *Module) Tools() []server.ServerTool {
    return []server.ServerTool{
        {
            Tool: mcp.Tool{
                Name:        "count_files",
                Description: "Count files in a directory",
                InputSchema: mcp.ToolInputSchema{
                    Type: "object",
                    Properties: map[string]interface{}{
                        "path": map[string]interface{}{
                            "type":        "string",
                            "description": "Directory path to count files in",
                        },
                        "extension": map[string]interface{}{
                            "type":        "string",
                            "description": "Filter by extension (e.g., '.go')",
                        },
                    },
                    Required: []string{"path"},
                },
            },
            Handler: countFiles,
        },
    }
}

func countFiles(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    path, _ := req.Params.Arguments["path"].(string)
    ext, _ := req.Params.Arguments["extension"].(string)

    count := 0
    err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
        if err != nil {
            return nil  // Skip errors
        }
        if !info.IsDir() {
            if ext == "" || filepath.Ext(p) == ext {
                count++
            }
        }
        return nil
    })

    if err != nil {
        return nil, fmt.Errorf("failed to walk directory: %w", err)
    }

    return mcp.NewToolResultText(fmt.Sprintf("Found %d files", count)), nil
}
```

---

## Testing Tips

```bash
# Run with debug output
DEBUG=true ./my-mcp-server

# Test specific tool (create test JSON)
echo '{"method":"tools/call","params":{"name":"my_hello","arguments":{"name":"Luke"}}}' | ./my-mcp-server
```

---

## Next Steps

- [GitHub Workflows](07-github-workflows.md) - Automate testing
- [Discord Bots](08-discord-bots.md) - Build bots that use MCP
