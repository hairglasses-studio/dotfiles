# hg-mcp — Claude Code Instructions — Gemini CLI Instructions



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


## Key Conventions

- Return `tools.ErrorResult(err), nil` for expected errors (bad input, not found)
- Return `nil, err` only for truly unexpected failures
- Use `tools.CodedErrorResult(code, msg)` when error codes matter
- **OTel middleware** (`mcpkit/observability`) — tracing spans + Prometheus metrics (invocations, duration, errors, active count)
- **Audit middleware** (`mcpkit/security`) — logs tool calls and completions
- **Rate limit middleware** (`mcpkit/resilience`) — per-service token bucket (keyed on `CircuitBreakerGroup`)
- **Circuit breaker middleware** (`mcpkit/resilience`) — per-service circuit breaker
- **Built-in** (mcpkit registry core) — 30s timeout, panic recovery, response truncation


## Shared Research Repository

Cross-project research lives at `~/hairglasses-studio/docs/` (git: hairglasses-studio/docs). When launching research agents, check existing docs first and write reusable research outputs back to the shared repo rather than local docs/.
