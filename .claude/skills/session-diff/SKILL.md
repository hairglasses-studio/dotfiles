---
description: "Compare progress between two Claude Code sessions. $ARGUMENTS: <session_id_1> <session_id_2>"
user_invocable: true
---

Parse `$ARGUMENTS` to extract two session IDs (space-separated).

Call `mcp__dotfiles__claude_session_compare` with both session IDs.

Display: side-by-side comparison of commits, files changed, tools used, errors encountered, and overall progress delta.
