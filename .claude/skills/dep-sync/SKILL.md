---
description: Audit and sync Go dependency versions across all repos. Shows version skew and optionally runs mcpkit version sync.
user_invocable: true
---

Execute in sequence:
1. `mcp__dotfiles__dotfiles_dep_audit` — Show Go dependency version skew across all repos
2. If skew detected, ask if user wants to sync mcpkit:
   - `mcp__dotfiles__dotfiles_mcpkit_version_sync` with `execute=false` (dry-run first)
   - Show which repos would be updated
   - If user confirms, run with `execute=true`

Display: dependency name | oldest version | newest version | repos affected
