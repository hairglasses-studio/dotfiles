---
description: Display a comprehensive recovery status dashboard across all sessions and repos.
user_invocable: true
---

Execute in parallel:
1. `mcp__dotfiles__claude_session_scan` — Scan all sessions, check PID liveness
2. `mcp__dotfiles__claude_recovery_history` — Audit trail of crash/recovery events

For each session found, call `mcp__dotfiles__claude_session_health` to get health scores.

Display grid: session ID | repo | health score (0-100) | alive? | last activity | unpushed? | action needed
Color-code by health: green (>80), yellow (50-80), red (<50).
