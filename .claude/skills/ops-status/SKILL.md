---
name: ops-status
description: Show SDLC session status, CI checks, and project health at a glance
allowed-tools: mcp__dotfiles__ops_session_status, mcp__dotfiles__ops_session_list, mcp__dotfiles__ops_ci_status, mcp__dotfiles__ops_build, mcp__dotfiles__ops_changed_files
---

Show the current SDLC status. `$ARGUMENTS` can be:
- Empty: show all active sessions + current branch CI status
- A session ID: show detailed status for that session
- `ci`: show only CI check status for the current branch

## Workflow

### Default (no arguments): Dashboard

Run these in parallel:
1. `ops_session_list()` — all active sessions
2. `ops_ci_status()` — CI for current branch
3. `ops_changed_files()` — uncommitted changes

Display as:

```
## SDLC Dashboard

### Active Sessions
| Session | Repo | Branch | Iterations | State | Error Trend |
|---------|------|--------|------------|-------|-------------|

### CI Status: {branch}
| Check | Status | Conclusion |
|-------|--------|------------|

### Working Tree
{total_changed} changed files (+{insertions} -{deletions})
Go packages affected: {go_packages}
```

### Session detail (session ID argument)

1. Run `ops_session_status(session_id="$ARGUMENTS")`
2. Display:

```
## Session {id} — {repo}

Branch: {branch} | Iterations: {count} | State: {current_state}
Total time: {total_time_ms}ms | Converging: {yes/no}

### Iteration History
| # | Status | Errors | Duration |
|---|--------|--------|----------|

### Error Trend
{error_trend as sparkline or list}
```

### CI only (`ci` argument)

1. Run `ops_ci_status(wait=false)`
2. Display check results as table with pass/fail/pending badges
