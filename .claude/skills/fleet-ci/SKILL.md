---
description: Check CI status across all repos and sync workflow files from canonical sources.
user_invocable: true
---

Execute in parallel:
1. `mcp__dotfiles__dotfiles_workflow_sync` with `dry_run=true` — Show which repos have stale CI workflows
2. `mcp__dotfiles__dotfiles_fleet_audit` — Get per-repo CI pass/fail status

Display: repo | CI status | workflow freshness | last run | action needed
Highlight repos with failing CI or outdated workflows.
