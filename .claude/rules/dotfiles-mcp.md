---
paths:
  - "mcp/dotfiles-mcp/**"
---

dotfiles-mcp conventions:
- Keep the server as a validated dotfiles-management surface; writes should preserve config safety and explicit reload behavior.
- Use `DOTFILES_DIR` as the source of truth for repo paths rather than hardcoding alternate workstation paths.
- Keep batch or write operations explicit about dry-run versus live execution.
- Prefer compact tool-of-tools helpers only when they materially reduce repetitive multi-step workstation flows.
