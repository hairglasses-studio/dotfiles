---
paths:
  - "mcp/**"
---

Shared MCP workspace conventions:
- Prefer module-local `cmd/` entrypoints and `internal/` implementation packages over cross-module coupling.
- Keep modules compatible with the `mcp/go.work` workspace and preserve their local `go.mod` boundaries.
- Shared helper code belongs in `mcp/internal/` or mcpkit, not in ad hoc copies across modules.
- Use discovery-first tool surfaces for large modules and keep read/write actions clearly separated.
