---
paths:
  - "mcp/mapitall/**"
---

mapitall conventions:
- Keep entrypoints in `cmd/` and application logic in `internal/`; avoid leaking internal packages across sibling modules.
- Preserve a clear boundary between search/catalog behavior and write or synchronization actions.
- Keep request/response payloads bounded and favor follow-up detail calls over oversized single-tool output.
