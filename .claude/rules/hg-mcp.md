---
paths:
  - "mcp/hg-mcp/**"
---

hg-mcp conventions:
- Add tool modules under `internal/mcp/tools/<module>/` and keep module registration patterns aligned with the existing registry facade.
- Keep CLI and transport entrypoints in `cmd/` thin; domain logic belongs under `internal/`.
- Prefer stdio and streamable transports over deprecated SSE behavior when touching runtime defaults or docs.
- Treat embedded web assets and observability code as separate concerns from tool-module changes.
