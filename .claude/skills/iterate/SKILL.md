---
name: iterate
description: Run the SDLC build/test/analyze loop and report what to fix next. The core feedback tool for autonomous development.
allowed-tools: mcp__dotfiles__ops_iterate, mcp__dotfiles__ops_session_status, mcp__dotfiles__ops_session_create, mcp__dotfiles__ops_changed_files
---

Run a build+test+analyze iteration on the current project. `$ARGUMENTS` is an optional session ID to continue a previous session. If empty, auto-creates a new session.

## Workflow

1. **Check changed files** first — run `ops_changed_files()` to understand what was modified
2. **Run the iteration** — call `ops_iterate(session_id="$ARGUMENTS")` (or no session_id to auto-create)
3. **Display results** based on the status:

### If `all_pass`:
```
## Iteration {N}: ALL PASS

Build: {duration}ms | Tests: {passed} passed, 0 failed
Session: {session_id} | Changed: {files}

Ready to ship! Run /ship to commit, push, and create a PR.
```

### If `build_fail`:
```
## Iteration {N}: BUILD FAILED

{error_count} compile error(s):

| File | Line | Error |
|------|------|-------|
| {file} | {line} | {message} |

### Fix Order
{suggested_fix_order as numbered list}

Fix these errors and run /iterate again.
```

### If `test_fail`:
```
## Iteration {N}: TESTS FAILED

Build: OK | Tests: {passed} passed, {failed} failed

| Package | Test | Category |
|---------|------|----------|
| {package} | {test} | {category} |

### Analysis
{analysis.summary}

### Next Actions
{next_actions as numbered list}

Fix the failing tests and run /iterate again.
```

4. **Show convergence** if 3+ iterations exist:
   - Error trend: `{error_trend}` (e.g., `[5, 3, 1]`)
   - Converging: yes/no
   - Total time: `{total_time_ms}ms` across all iterations
