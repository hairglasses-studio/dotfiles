---
description: "Sync CI workflow files across all repos from canonical sources. $ARGUMENTS: (empty)=dry-run audit, 'execute'=deploy changes"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)** or **"audit"**: Call `mcp__dotfiles__dotfiles_workflow_sync` with `dry_run=true` — show which repos have stale workflows
- **"execute"**: Call `mcp__dotfiles__dotfiles_workflow_sync` with `execute=true` — deploy canonical workflows to all repos

Display: repo | workflow file | status (current/stale/missing) | action taken
