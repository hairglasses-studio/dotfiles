# hg-mcp — Claude Code Instructions

Go MCP server with 1,190+ tools across 119 modules, organized into 10 runtime groups. For detailed reference tables, code examples, and architecture docs, see [CLAUDE-DETAIL.md](./CLAUDE-DETAIL.md).

## Build & Run

```bash
make build                                    # Full build (web UI + Go binary)
go build -o hg-mcp-bin ./cmd/hg-mcp          # Just Go binary
./hg-mcp-bin                                  # stdio mode (for Claude Code)
MCP_MODE=streamable PORT=8080 ./hg-mcp-bin   # Streamable HTTP (MCP 2025-03-26 spec)
go test ./...                                 # all tests
go test -short ./...                          # skip slow/integration tests
make lint                                     # go vet ./...
```

## mcpkit Dependency

hg-mcp depends on `github.com/hairglasses-studio/mcpkit` (local replace in `go.mod`). `internal/mcp/tools/` is a thin shim layer — all 119 modules import from it, not mcpkit directly, so mcpkit upgrades require zero module changes.

After pulling mcpkit changes: `go mod tidy && go build ./... && go test -short ./...`

## Code Standards

- **Go 1.25+** required
- Use `go vet` and `go fmt` before committing
- Modules auto-register via `init()` — no central tool list to maintain
- Clients are lazy-initialized on first use (see `LazyClient[T]` in `clientutil.go`)

## Key Patterns

### Error Handling

- Return `tools.ErrorResult(err), nil` for expected errors (bad input, not found)
- Return `nil, err` only for truly unexpected failures
- Use `tools.CodedErrorResult(code, msg)` when error codes matter

### Handler Helpers

```go
tools.ErrorResult(err)                    // Error response
tools.JSONResult(data)                    // JSON response
tools.TextResult(text)                    // Plain text response
tools.GetStringParam(req, "name")         // String param (empty if missing)
tools.GetIntParam(req, "count", 10)       // Int param with default
tools.GetBoolParam(req, "confirm", false) // Bool param with default
```

### Runtime Groups

10 groups: `dj_music`, `vj_video`, `lighting`, `audio_production`, `show_control`, `infrastructure`, `messaging`, `inventory`, `streaming`, `platform`. Auto-assigned by category via `RegisterModule()`.

## Environment

- Copy `.env.example` to `.env` and fill in credentials
- Key vars: `MCP_MODE`, `PORT`, `INVENTORY_SPREADSHEET_ID`, `DISCORD_BOT_TOKEN`, `NANOLEAF_HOST`, `HUE_BRIDGE_IP`
