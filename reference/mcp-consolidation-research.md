# MCP Server Consolidation Research

Conducted 2026-04-02. Research into consolidating 5 MCP servers into 1.

## Key Finding: Token Savings Are Zero

Claude Code uses **Tool Search (deferred loading) by default**. Only tool names (~8.7K tokens
for 50+ tools) are loaded into context. Full schemas are fetched on-demand via ToolSearch.

Whether 86 tools are in 1 server or 5, the token cost is identical.

Source: https://www.atcyrus.com/stories/mcp-tool-search-claude-code-context-pollution-guide

## Tool Search Mechanism

1. Session start: only tool *names* loaded (minimal tokens)
2. Claude uses `ToolSearch` tool to find relevant tools by keyword/regex
3. Matching tool schemas returned inline (3-5 at a time)
4. Only discovered tools get full definitions in context
5. Prompt cache preserved — prefix stays unchanged

### Control

```bash
ENABLE_TOOL_SEARCH=true        # All tools deferred (default)
ENABLE_TOOL_SEARCH=auto        # Threshold: upfront if <10% context
ENABLE_TOOL_SEARCH=auto:5      # Threshold at 5%
ENABLE_TOOL_SEARCH=false       # All tools loaded upfront
```

## Real Benefits of Consolidation

1. Fewer processes (5 → 1): less memory, simpler lifecycle
2. Single binary: one `go build`, one deploy
3. Shared state: compositor detection, dotfiles paths, atomic writes
4. Simpler .mcp.json: one entry vs five bash wrappers
5. Eliminates Node.js dependency (sway-mcp)
6. Unified mcpkit error handling everywhere

## MCP Tool Best Practices

### Naming
```
Pattern: <domain>_<resource>_<action>
Examples: hypr_list_windows, shader_set, bt_connect
```

### Descriptions
- Lead with capability (first sentence answers "what does this do?")
- Include search keywords
- Max 2KB per tool (Claude Code truncates beyond that)

### Error Handling
```go
// Structured error codes from mcpkit
handler.ErrInvalidParam  // "INVALID_PARAM"
handler.ErrNotFound      // "NOT_FOUND"
handler.ErrAPIError      // "API_ERROR"
```

## Architecture: mcpkit Multi-Module Pattern

```go
// Registry natively supports multiple modules
reg := registry.NewToolRegistry()
reg.RegisterModule(&DotfilesModule{})    // 26 tools
reg.RegisterModule(&HyprlandModule{})    // 12 tools
reg.RegisterModule(&InputModule{})       // 26 tools
reg.RegisterModule(&ShaderModule{})      // 10 tools
reg.RegisterModule(&SwayModule{})        // 12 tools
// Total: 86 tools, 1 server, 1 process
```

## Sources
- https://www.atcyrus.com/stories/mcp-tool-search-claude-code-context-pollution-guide
- https://unified.to/blog/scaling_mcp_tools_with_anthropic_defer_loading
- https://claudefa.st/blog/tools/mcp-extensions/mcp-tool-search
- https://code.claude.com/docs/en/mcp
- https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-search-tool
- https://modelcontextprotocol.io/docs/learn/architecture
