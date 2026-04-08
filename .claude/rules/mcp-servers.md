---
paths:
  - "mcp/**"
---

All MCP servers use mcpkit framework:
- ToolModule interface: `Name()`, `Description()`, `Tools() []ToolDefinition`
- Registration: `init()` → `tools.GetRegistry().RegisterModule(&Module{})`
- Handler contract: always return `(*mcp.CallToolResult, nil)` — use `handler.CodedErrorResult()` for errors
- Integration tests: `mcptest.NewServer()` from mcpkit
- Each server is a separate Go module under `mcp/` with its own `go.mod`
- Workspace managed via `mcp/go.work`
