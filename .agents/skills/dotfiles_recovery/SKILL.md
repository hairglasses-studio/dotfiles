---
name: dotfiles_recovery
description: Claude-session recovery, forensic analysis, and handoff workflow for the hairglasses dotfiles repo.
allowed-tools:
  - Bash
  - Read
  - Write
  - Grep
  - mcp__dotfiles__claude_crash_detect
  - mcp__dotfiles__claude_recovery_report
  - mcp__dotfiles__claude_repo_diff
  - mcp__dotfiles__claude_session_search
  - mcp__dotfiles__claude_session_detail
  - mcp__dotfiles__claude_session_replay
  - mcp__dotfiles__claude_repo_status
---

# Dotfiles Recovery

Use this skill for recovering interrupted Claude/Codex sessions, reconstructing prior context, and producing durable handoff artifacts.

## Default loop

1. Identify dead or incomplete sessions before touching any repo state.
2. Reconstruct the relevant conversation and current git state before proposing next steps.
3. Keep recovery output organized as session context, repo status, salvageable work, and recommended action.
4. Write handoff or autopsy output when the recovery needs to survive another interruption.

Read `references/workflows.md` for the compressed workflow catalog that replaces the previous recovery-focused Claude skill set.
