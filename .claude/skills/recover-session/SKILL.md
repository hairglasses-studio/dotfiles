---
description: "Resume a specific Claude Code session with full context restoration. $ARGUMENTS: session ID or search term"
user_invocable: true
---

Parse `$ARGUMENTS` as session ID or search term:
1. If looks like a session ID: Call `mcp__dotfiles__claude_session_detail` with the ID
2. If looks like a search term: Call `mcp__dotfiles__claude_session_search` to find matching sessions

Then for the target session:
3. `mcp__dotfiles__claude_session_replay` — Reconstruct conversation thread
4. `mcp__dotfiles__claude_repo_status` — Show current git state of the session's repo

Display: session context, last 5 interactions, git status, and the resume command to copy.
