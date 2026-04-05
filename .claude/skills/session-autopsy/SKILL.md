---
description: "Perform forensic analysis on a dead Claude Code session. $ARGUMENTS: session ID"
user_invocable: true
---

Execute in sequence:
1. `mcp__dotfiles__claude_session_detail` with `session_id=$ARGUMENTS` — Deep inspection
2. `mcp__dotfiles__claude_session_logs` with `session_id=$ARGUMENTS, lines=100` — Last 100 log entries
3. `mcp__dotfiles__claude_session_health` with `session_id=$ARGUMENTS` — Health score across 5 dimensions
4. `mcp__dotfiles__claude_session_replay` with `session_id=$ARGUMENTS` — Conversation reconstruction

Analyze the logs for: error patterns, last successful tool call, crash point, memory/token usage.
Present: timeline of events, root cause hypothesis, and recovery recommendation.
