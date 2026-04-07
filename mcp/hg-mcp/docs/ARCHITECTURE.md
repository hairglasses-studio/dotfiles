# hg-mcp Architecture Guide

This guide explains how the code is organized and how to add new features.

---

## How It Works (Simple Explanation)

**hg-mcp** is built like a toolbox:
- Each **module** is like a drawer in the toolbox (e.g., "discord", "touchdesigner")
- Each **tool** is something you can do (e.g., "send a message", "get status")
- The **registry** keeps track of all the tools
- The **MCP server** lets Claude use the tools

---

## Directory Structure

```
hg-mcp/
├── cmd/                            # Entry points (where the program starts)
│   ├── aftrs/main.go              # CLI tool - run from terminal
│   └── hg-mcp/main.go          # MCP server - connects to Claude
│
├── internal/                       # Internal code (the "engine")
│   ├── clients/                   # Code that talks to external services
│   │   ├── discord.go            # Discord bot connection
│   │   ├── touchdesigner.go      # TouchDesigner connection
│   │   └── ...
│   │
│   ├── commands/                  # CLI command definitions
│   │   └── root.go               # Main command setup
│   │
│   └── mcp/
│       └── tools/                 # All the MCP tools
│           ├── registry.go       # Keeps track of all tools
│           ├── discord/          # Discord module
│           │   └── module.go     # Discord tools defined here
│           ├── touchdesigner/    # TouchDesigner module
│           └── ...
│
├── docs/                          # Documentation
├── configs/                       # Configuration files
└── go.mod                         # Go module definition
```

---

## Key Concepts

### What is a Module?

A module is a group of related tools. For example, the `discord` module contains:
- `aftrs_discord_send` - Send a message
- `aftrs_discord_channels` - List channels
- `aftrs_discord_status` - Check if bot is connected

### What is the Registry?

The registry is like a phone book for tools. When Claude asks "what tools are available?", the registry provides the list.

### What is a Runtime Group?

A runtime group is a high-level functional category that spans multiple modules. For example, the `lighting` group includes WLED, grandMA3, Nanoleaf, Hue, LedFX, sACN, and QLC+ modules. There are 10 groups:

| Group | Description |
|-------|-------------|
| `dj_music` | DJ software, music platforms, samples |
| `vj_video` | VJ software, video processing, cameras |
| `lighting` | DMX, LED panels, smart lights |
| `audio_production` | DAWs, audio routing, MIDI |
| `show_control` | Cue management, automation, switching |
| `infrastructure` | Servers, networking, home automation |
| `messaging` | Discord, email, Notion, calendars |
| `inventory` | Hardware liquidation |
| `streaming` | Live streaming, Twitch, YouTube |
| `platform` | Tool discovery, dashboards, security |

Runtime groups are auto-assigned based on the module's `Category` field — you don't need to set them manually.

### What is a Client?

A client is code that talks to an external service. For example:
- `discord.go` - Talks to Discord's API
- `touchdesigner.go` - Talks to TouchDesigner

---

## mcpkit Integration

hg-mcp depends on [`mcpkit`](../mcpkit) (local replace in `go.mod`) for production-grade MCP infrastructure:

- **Registry & middleware chain** — tool registration, 30s timeout, panic recovery, response truncation
- **Handler helpers** — `TextResult`, `ErrorResult`, `JSONResult`, param extraction
- **Client pools** — HTTP client pools (Fast/Standard/Slow tiers)
- **Resilience** — circuit breaker, rate limiter, response cache
- **Observability** — OpenTelemetry tracing + Prometheus metrics
- **Security** — RBAC middleware, audit logging
- **Sanitize** — input validation (paths, usernames, device paths, etc.)

### Shim Layer

All 119 tool modules import from `internal/mcp/tools/` (not mcpkit directly). This shim layer re-exports mcpkit types and helpers so that mcpkit API changes only require updating 4 shim files:

| Shim File | Delegates To |
|-----------|-------------|
| `helpers.go` | `mcpkit/handler` — result builders, param getters |
| `clientutil.go` | `mcpkit/client` — `LazyClient[T]` |
| `compat.go` | `mcpkit/registry` — MCP SDK type aliases |
| `registry.go` | `mcpkit/registry` — wraps `ToolRegistry`, adds runtime groups |

mcpkit source is at `../mcpkit` relative to this repo.

---

## Adding a New Module (Step-by-Step)

Let's say you want to add a module for controlling OBS.

### Step 1: Create the Module Folder

```bash
# Create the folder
mkdir -p internal/mcp/tools/obs
```

### Step 2: Create the Module File

Create `internal/mcp/tools/obs/module.go`:

```go
// Package obs provides OBS control tools
package obs

import (
    "context"

    "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
    mcp "github.com/mark3labs/mcp-go"
)

// Module is the OBS tool module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
    return "obs"
}

// Description returns what this module does
func (m *Module) Description() string {
    return "OBS Studio control"
}

// Tools returns all tools in this module
func (m *Module) Tools() []tools.ToolDefinition {
    return []tools.ToolDefinition{
        {
            Tool: mcp.NewTool("aftrs_obs_status",
                mcp.WithDescription("Get OBS connection status"),
            ),
            Handler:     handleStatus,
            Category:    "obs",
            Subcategory: "status",
            Tags:        []string{"streaming", "status"},
            Complexity:  tools.Simple,
        },
    }
}

// handleStatus checks if OBS is connected
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Your code here to check OBS status
    return tools.TextResult("OBS is connected"), nil
}

// init registers this module when the program starts
func init() {
    tools.GetRegistry().RegisterModule(&Module{})
}
```

### Step 3: Import the Module

Add this import to `internal/mcp/tools/tools.go`:

```go
import (
    _ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/obs"
)
```

### Step 4: Build and Test

```bash
# Build the project
go build -o aftrs ./cmd/aftrs && go build -o hg-mcp ./cmd/hg-mcp

# Run to verify
./hg-mcp
```

---

## Adding a New Tool to an Existing Module

Let's add a new tool to the discord module.

### Step 1: Open the Module File

Edit `internal/mcp/tools/discord/module.go`

### Step 2: Add the Tool Definition

Add to the `Tools()` function:

```go
{
    Tool: mcp.NewTool("aftrs_discord_react",
        mcp.WithDescription("Add a reaction to a message"),
        mcp.WithString("message_id", mcp.Required(),
            mcp.Description("The message ID to react to")),
        mcp.WithString("emoji", mcp.Required(),
            mcp.Description("The emoji to add")),
    ),
    Handler:     handleReact,
    Category:    "discord",
    Subcategory: "messages",
    Tags:        []string{"reaction", "emoji"},
    Complexity:  tools.Simple,
},
```

### Step 3: Add the Handler Function

Add at the bottom of the file:

```go
// handleReact adds a reaction to a message
func handleReact(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Get the parameters
    messageID := req.Params.Arguments["message_id"].(string)
    emoji := req.Params.Arguments["emoji"].(string)

    // Your code here to add the reaction
    // ...

    return tools.TextResult("Reaction added!"), nil
}
```

### Step 4: Build and Test

```bash
go build -o aftrs ./cmd/aftrs && go build -o hg-mcp ./cmd/hg-mcp
```

---

## Creating a Client (Talking to External Services)

Clients are how we talk to external services like Discord or TouchDesigner.

### Example: Simple HTTP Client

Create `internal/clients/myservice.go`:

```go
package clients

import (
    "context"
    "net/http"
    "os"
)

// MyServiceClient talks to MyService
type MyServiceClient struct {
    host   string
    apiKey string
}

// NewMyServiceClient creates a new client
func NewMyServiceClient() (*MyServiceClient, error) {
    // Get settings from environment variables
    host := os.Getenv("MYSERVICE_HOST")
    if host == "" {
        host = "localhost:8080"  // Default value
    }

    apiKey := os.Getenv("MYSERVICE_API_KEY")

    return &MyServiceClient{
        host:   host,
        apiKey: apiKey,
    }, nil
}

// GetStatus checks if the service is running
func (c *MyServiceClient) GetStatus(ctx context.Context) (string, error) {
    resp, err := http.Get("http://" + c.host + "/status")
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode == 200 {
        return "connected", nil
    }
    return "disconnected", nil
}
```

### Using the Client in a Tool Handler

```go
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // Create the client
    client, err := clients.NewMyServiceClient()
    if err != nil {
        return tools.ErrorResult(err), nil
    }

    // Use the client
    status, err := client.GetStatus(ctx)
    if err != nil {
        return tools.ErrorResult(err), nil
    }

    return tools.TextResult("Status: " + status), nil
}
```

---

## Helper Functions

These helper functions make it easy to return results from tool handlers:

```go
// Return a success message
return tools.TextResult("Everything worked!"), nil

// Return an error
return tools.ErrorResult(err), nil

// Return JSON data
return tools.JSONResult(myData), nil
```

---

## Tool Naming Convention

All tools follow this pattern:

```
aftrs_<module>_<action>
```

Examples:
| Tool Name | Module | Action |
|-----------|--------|--------|
| `aftrs_discord_send` | discord | send |
| `aftrs_td_status` | td (TouchDesigner) | status |
| `aftrs_obs_scene_switch` | obs | scene_switch |

---

## Tool Complexity Levels

| Level | Description | Example |
|-------|-------------|---------|
| `Simple` | Just reads data, quick | Get status, list items |
| `Moderate` | Some processing needed | Send message, trigger action |
| `Complex` | Heavy processing, long running | Full health check, analysis |

---

## Common Mistakes and How to Fix Them

### "undefined: tools" error
You forgot to import the tools package:
```go
import "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
```

### "module not registered"
Make sure your `init()` function is at the bottom of your module file:
```go
func init() {
    tools.GetRegistry().RegisterModule(&Module{})
}
```

And import it in `tools.go`:
```go
import _ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/yourmodule"
```

### Tool not showing up
1. Did you build? `go build ./...`
2. Check for compile errors in your module file
3. Make sure the module is imported in `tools.go`

---

## Testing Your Changes

```bash
# Build everything
go build -o aftrs ./cmd/aftrs && go build -o hg-mcp ./cmd/hg-mcp

# Run tests
go test ./...

# Run the MCP server to verify
./hg-mcp
```

---

## Quick Reference

### Build Commands
```bash
go build -o aftrs ./cmd/aftrs && go build -o hg-mcp ./cmd/hg-mcp
```

### Run Commands
```bash
./hg-mcp                           # Stdio mode (Claude Code)
MCP_MODE=sse PORT=8080 ./hg-mcp    # SSE mode (remote)
```

### Format Code
```bash
go fmt ./...
```

### Run Tests
```bash
go test ./...
```
