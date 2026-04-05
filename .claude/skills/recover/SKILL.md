---
description: Run a full Claude Code crash recovery analysis across ~/hairglasses-studio.
user_invocable: true
---

Execute in sequence:
1. `mcp__dotfiles__claude_crash_detect` — Find dead sessions by PID liveness check
2. `mcp__dotfiles__claude_recovery_report` — Generate prioritized recovery queue
3. For each dead session with unpushed work, call `mcp__dotfiles__claude_repo_diff` to show uncommitted changes

Display: session ID | repo | status | unpushed commits | uncommitted files | recommended action
Sort by severity (sessions with uncommitted code first).
