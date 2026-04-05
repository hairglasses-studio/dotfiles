---
description: Capture a point-in-time workspace snapshot for recovery checkpointing.
user_invocable: true
---

Execute in parallel:
1. `mcp__dotfiles__claude_workspace_snapshot` — Capture full workspace state
2. `mcp__dotfiles__claude_repo_status` — Git status for current repo

Display: snapshot ID, timestamp, repos captured, uncommitted changes, unpushed commits.
Use this before risky operations as a recovery checkpoint.
