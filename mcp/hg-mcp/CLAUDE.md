# hg-mcp — Claude Code Instructions

## Project Overview

Go MCP (Model Context Protocol) server with 1,190+ tools across 119 modules, organized into 10 runtime groups. Aggregates infrastructure, creative, DJ/music, lighting, video, and inventory management tools into a single MCP endpoint.

## Project Structure

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

## Build & Run

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

## mcpkit Dependency

hg-mcp depends on `github.com/hairglasses-studio/mcpkit` (local replace in `go.mod`). mcpkit provides the production-grade MCP infrastructure: registry, handler helpers, client pools, resilience (circuit breaker, rate limiter, cache), security (RBAC, audit), observability (OTel + Prometheus), and sanitize.

**Architecture:** `internal/mcp/tools/` is a thin shim layer that delegates to mcpkit. All 119 tool modules import from `internal/mcp/tools/` (not mcpkit directly), so mcpkit upgrades require zero module changes.

**Key shim files:**
- `helpers.go` — delegates `TextResult`, `ErrorResult`, `JSONResult`, param getters to `mcpkit/handler`
- `clientutil.go` — delegates `LazyClient[T]` to `mcpkit/client`
- `compat.go` — re-exports MCP SDK types (`Tool`, `CallToolRequest`, etc.) from `mcpkit/registry`
- `registry.go` — wraps `mcpkit/registry.ToolRegistry`, adds hg-mcp-specific `SearchTools`, `LogConfigStatus`, runtime group mapping, rate limit defaults

**Keeping mcpkit up to date:**
- Periodically check `../mcpkit` for updates: `cd ../mcpkit && git log --oneline -10`
- After pulling mcpkit changes: `go mod tidy && go build ./... && go test -short ./...`
- If mcpkit adds new middleware or features, wire them in `registry.go`'s `ConfigureMiddleware()`
- If mcpkit changes exported APIs, only the shim files need updating (not the 119 modules)

**What lives where:**
| Concern | Location | Notes |
|---------|----------|-------|
| Core registry, middleware chain | mcpkit/registry | Shared infrastructure |
| Handler helpers, param extraction | mcpkit/handler | Shared infrastructure |
| HTTP client pool (Fast/Standard/Slow) | mcpkit/client | Used via `httpclient` alias in `internal/clients/` |
| Circuit breaker, rate limiter, cache | mcpkit/resilience | Cache used in `internal/clients/` |
| OTel + Prometheus init | mcpkit/observability | Initialized in `cmd/hg-mcp/main.go` |
| Audit logging middleware | mcpkit/security | Wired in `registry.go` `ConfigureMiddleware()` |
| Input validators | mcpkit/sanitize | `SortField` inlined in `internal/clients/system.go` |
| RBAC users/globals, audit tools | pkg/security (hg-mcp) | hg-mcp-specific user config + admin tools |
| Sync pipeline circuit breakers | internal/sync (hg-mcp) | Separate from registry middleware |
| Runtime group mapping | internal/mcp/tools/registry.go | hg-mcp-specific category→group map |
| Rate limit per-service defaults | internal/mcp/tools/registry.go | hg-mcp-specific service configs |

## Code Standards

- **Go 1.25+** required
- Use `go vet` and `go fmt` before committing
- Modules auto-register via `init()` — no central tool list to maintain
- Clients are lazy-initialized on first use (see `LazyClient[T]` in `clientutil.go`)

## Key Patterns

### Adding a New Tool Module

1. Create `internal/mcp/tools/<name>/module.go`
2. Implement the `Module` interface:

```go
package mymodule

import (
    "context"
    "github.com/aftrs-studio/hg-mcp/internal/mcp/tools"
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
_ "github.com/aftrs-studio/hg-mcp/internal/mcp/tools/mymodule"
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

## Important Directories

| Path | Purpose |
|------|---------|
| `internal/mcp/tools/` | All tool modules — one package per domain |
| `internal/clients/` | Service clients (HTTP, OSC, WebSocket, etc.) |
| `pkg/security/` | RBAC roles and audit logging |
| `observability/` | Docker Compose stack for Prometheus/Grafana/OTel |
| `web/` | React + Vite frontend source |

## Environment

- Copy `.env.example` to `.env` and fill in credentials
- Key vars: `MCP_MODE`, `PORT`, `INVENTORY_SPREADSHEET_ID`, `DISCORD_BOT_TOKEN`, `NANOLEAF_HOST`, `HUE_BRIDGE_IP`
- See `.env.example` for the full list (~100 vars across all integrations)

## Known DRY Opportunities

These are real duplication patterns found across the codebase. They are documented here for future cleanup sessions:

1. **368 manual `json.MarshalIndent` calls** — Many modules do `json.MarshalIndent(data, "", "  ")` + `TextResult(string(bytes))` instead of using `tools.JSONResult(data)` which does exactly this. Mechanical fix across ~42 files.

2. **~486 repeated parameter validation blocks** — Pattern: `param := tools.GetStringParam(req, "x"); if param == "" { return tools.CodedErrorResult(...) }`. Could be a `RequireStringParam(req, "name") (string, *CallToolResult)` helper.

3. **208 scattered `os.Getenv` calls** in client constructors — Could be a `config.GetRequired("VAR")` helper.

4. **Identical test boilerplate** across ~91 modules — `TestModuleInfo` function is copy-pasted. Could be a shared test helper.

## References

- **MCP Protocol**: [spec](https://spec.modelcontextprotocol.io/), [Go SDK](https://github.com/mark3labs/mcp-go)
- **Go patterns**: [Effective Go](https://go.dev/doc/effective_go), [Go Blog](https://go.dev/blog/)
- **Observability**: [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/), [Prometheus Go client](https://github.com/prometheus/client_golang)
- **Resilience**: [Circuit breaker pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker), [Rate limiting](https://pkg.go.dev/golang.org/x/time/rate)
