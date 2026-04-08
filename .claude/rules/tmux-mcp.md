---
paths:
  - "mcp/tmux-mcp/**"
---

tmux-mcp conventions:
- Handle "no tmux server" cases gracefully; list-style operations should degrade cleanly.
- Preserve `session:window.pane` target parsing semantics and detached session creation defaults.
- Keep tmux command composition readable and explicit so pane/session side effects stay auditable.
