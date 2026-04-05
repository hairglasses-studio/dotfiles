---
description: Verify that all recently recovered sessions completed successfully.
user_invocable: true
---

Execute in sequence:
1. `mcp__dotfiles__claude_session_scan` — Find all sessions
2. For each session with recent recovery events, call `mcp__dotfiles__claude_session_health`
3. `mcp__dotfiles__claude_recovery_history` — Check recovery audit trail

Display: session ID | recovered? | current health | repo status | verification result (pass/fail)
Flag any sessions that were recovered but are now failing again.
