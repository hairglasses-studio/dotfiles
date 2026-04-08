---
paths:
  - "mcp/process-mcp/**"
---

process-mcp conventions:
- Keep process inspection read-oriented by default and make disruptive actions obviously separate from listing/status tools.
- Preserve safe handling for absent processes and transient process-state races.
- Keep shell/process execution paths explicit and auditable rather than hiding them behind broad helpers.
