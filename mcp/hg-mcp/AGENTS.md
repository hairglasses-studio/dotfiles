# hg-mcp — Claude Code Instructions — Agent Instructions



## Build & Test

```bash
# Full build (web UI + Go binary)
make build

# Just Go binary (web UI must already exist at internal/web/dist/)
make build-web  # if needed
go build -o hg-mcp-bin ./cmd/hg-mcp

# Run
./hg-mcp-bin                              # stdio mode (for Claude Code)
MCP_MODE=sse PORT=8080 ./hg-mcp-bin       # SSE mode (deprecated April 2026)
MCP_MODE=streamable PORT=8080 ./hg-mcp-bin # Streamable HTTP (MCP 2025-03-26 spec)
MCP_MODE=web PORT=8080 ./hg-mcp-bin       # Web UI mode

# Tests
go test ./...           # all tests
go test -short ./...    # skip slow/integration tests
make lint               # go vet ./...
```

## Architecture

```
hg-mcp/
├── cmd/
│   ├── hg-mcp/main.go          # Main MCP server entry point
│   ├── aftrs/main.go            # CLI entry point
│   ├── beatport-sync/main.go    # Standalone sync tool
│   └── secrets/main.go          # Secrets management
├── internal/
│   ├── mcp/
│   │   ├── tools/               # Tool modules (one package per domain)
│   │   │   ├── registry.go      # ToolRegistry facade wrapping mcpkit + hg-mcp features
│   │   │   ├── helpers.go       # Shim: delegates to mcpkit/handler
│   │   │   ├── clientutil.go    # Shim: delegates to mcpkit/client
│   │   │   ├── compat.go       # Type aliases from mcpkit/registry
│   │   │   ├── inventory/       # Example: inventory module (47 tools)
│   │   │   ├── discord/         # Discord bot tools
│   │   │   ├── nanoleaf/        # Nanoleaf light panels (8 tools)
│   │   │   ├── hue/             # Philips Hue bridge (10 tools)
│   │   │   ├── resolume/        # VJ software tools
│   │   │   └── ...              # ~120 module packages
│   │   └── tasks/               # Background task management
│   ├── clients/                 # Service clients (lazy-initialized)
│   ├── bot/                     # Discord bot
│   ├── bridge/                  # Cross-system bridges
│   ├── chains/                  # Tool chaining/workflows
│   ├── config/                  # Config audit
│   ├── sync/                    # Sync pipelines (Rekordbox, SoundCloud, etc.)
│   └── web/                     # Web UI (embedded React app)
├── pkg/
│   ├── secrets/                 # Secret providers (AWS, 1Password)
│   └── security/                # RBAC, audit logging
├── web/                         # React + Vite source
├── observability/               # Prometheus + Grafana + OTel stack
├── docs/                        # Architecture and operational docs
├── Makefile                     # Build automation
└── .github/workflows/ci.yml    # CI pipeline
```

## Key Conventions

### Adding a New Tool Module

1. Create `internal/mcp/tools/<name>/module.go`
2. Implement the `Module` interface:

```go
package mymodule

import (
    "context"
    "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
    "github.com/mark3labs/mcp-go/mcp"
)

type Module struct{}

func init() {
    tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string        { return "mymodule" }
func (m *Module) Description() string { return "My module description" }

func (m *Module) Tools() []tools.ToolDefinition {
    return []tools.ToolDefinition{
        {
            Tool: mcp.Tool{
                Name:        "aftrs_mymodule_action",
                Description: "Does something useful",
                InputSchema: mcp.ToolInputSchema{
                    Type: "object",
                    Properties: map[string]interface{}{
                        "param": map[string]interface{}{
                            "type":        "string",
                            "description": "A required parameter",
                        },
                    },
                    Required: []string{"param"},
                },
            },
            Handler:  handleAction,
            Category: "mymodule",
        },
    }
}

func handleAction(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    param := tools.GetStringParam(req, "param")
    if param == "" {
        return tools.ErrorResult(fmt.Errorf("param is required")), nil
    }
    // ... business logic via client ...
    return tools.JSONResult(map[string]interface{}{"result": "ok"}), nil
}
```

3. Import the package in `cmd/hg-mcp/main.go`:
```go
_ "github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/mymodule"
```

### Handler Helpers (`internal/mcp/tools/helpers.go`)

```go
tools.ErrorResult(err)                         // Error response
tools.JSONResult(data)                         // JSON response
tools.TextResult(text)                         // Plain text response
tools.GetStringParam(req, "name")              // String param (empty if missing)
tools.GetIntParam(req, "count", 10)            // Int param with default
tools.GetFloatParam(req, "price", 0)           // Float param with default
tools.GetBoolParam(req, "confirm", false)      // Bool param with default
tools.GetStringArrayParam(req, "tags")         // String array param
```

### Error Handling

- Return `tools.ErrorResult(err), nil` for expected errors (bad input, not found)
- Return `nil, err` only for truly unexpected failures
- Use `tools.CodedErrorResult(code, msg)` when error codes matter

### Testing Pattern

```go
func TestMyHandler(t *testing.T) {
    // Override client with in-memory mock
    clients.TestOverrideInventoryClient(mockClient)
    defer clients.TestOverrideInventoryClient(nil)

    req := mcp.CallToolRequest{...}
    result, err := handleMyAction(context.Background(), req)
    // assertions...
}
```

### Observability

All tool handlers are wrapped by mcpkit's middleware chain (configured in `registry.go` `ConfigureMiddleware()`):
- **OTel middleware** (`mcpkit/observability`) — tracing spans + Prometheus metrics (invocations, duration, errors, active count)
- **Audit middleware** (`mcpkit/security`) — logs tool calls and completions
- **Rate limit middleware** (`mcpkit/resilience`) — per-service token bucket (keyed on `CircuitBreakerGroup`)
- **Circuit breaker middleware** (`mcpkit/resilience`) — per-service circuit breaker
- **Built-in** (mcpkit registry core) — 30s timeout, panic recovery, response truncation

Observability is initialized in `cmd/hg-mcp/main.go` via `mcpkit/observability.Init()` and passed to the registry via `tools.SetObservabilityProvider()` before `ConfigureMiddleware()`.

Module-specific metrics go in a `metrics.go` file (see `inventory/metrics.go` for example).

### Runtime Groups

Tools are auto-assigned to one of 10 runtime groups based on their category. The `RuntimeGroup` field on `ToolDefinition` is populated automatically by `RegisterModule()` via the `categoryToRuntimeGroup` map — no need to set it in module code.

Groups: `dj_music`, `vj_video`, `lighting`, `audio_production`, `show_control`, `infrastructure`, `messaging`, `inventory`, `streaming`, `platform`

Use `ListToolsByRuntimeGroup(group)` or `GetRuntimeGroupStats()` for programmatic access.


## Shared Research Repository

Cross-project research lives at `~/hairglasses-studio/docs/` (git: hairglasses-studio/docs). When launching research agents, check existing docs first and write reusable research outputs back to the shared repo rather than local docs/.
