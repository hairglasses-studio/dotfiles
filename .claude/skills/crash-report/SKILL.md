---
description: Generate a comprehensive crash report for debugging and post-mortem analysis.
user_invocable: true
---

Execute in parallel:
1. `mcp__dotfiles__claude_crash_detect` — Find all dead sessions
2. `mcp__dotfiles__claude_recovery_history` — Recovery audit trail

For each crashed session:
3. `mcp__dotfiles__claude_session_logs` with `lines=50` — Recent logs
4. `mcp__dotfiles__claude_repo_diff` — Uncommitted/unpushed work at risk

Generate report: crash count, affected repos, data at risk, error patterns, recommended recovery order.
