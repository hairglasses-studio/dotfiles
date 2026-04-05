---
name: fix-ci
description: Analyze CI failures, diagnose root causes, and optionally auto-fix and push
allowed-tools: Bash, Read, Grep, Glob, mcp__dotfiles__ops_ci_status, mcp__dotfiles__ops_analyze_failures, mcp__dotfiles__ops_iterate, mcp__dotfiles__ops_build, mcp__dotfiles__ops_test_smart, mcp__dotfiles__ops_commit
---

Analyze and fix CI failures for the current branch. `$ARGUMENTS` can be:
- Empty: check current branch CI and analyze failures
- A PR number: check that PR's CI
- `--auto-fix`: attempt automatic fixes and push

## Workflow

1. **Poll CI** — run `ops_ci_status(wait=true)` to wait for checks to complete (up to 5 min)

2. **If all pass:**
```
## CI Status: ALL PASS

| Check | Status | Duration |
|-------|--------|----------|
{checks table}

No action needed.
```

3. **If failures exist:**
   - Fetch failed check logs: `gh run view {run_id} --log-failed` via Bash
   - Parse the log output for error patterns
   - Run `ops_analyze_failures(build_errors=..., test_failures=...)` on extracted errors
   - Display:

```
## CI Failures: {failed_count} check(s) failed

### {check_name}
**Category:** {compile_error/test_assertion/timeout/etc.}
**Root cause:** {analysis}
**Files affected:** {list}

### Suggested Fixes
1. {fix suggestion with file and line}
2. {fix suggestion}
```

4. **If `--auto-fix` mode:**
   - Apply suggested fixes using Read/Edit tools
   - Run `ops_iterate()` to verify fixes locally
   - If passing, run `ops_commit(message="fix: resolve CI failures", execute=true)` and push
   - Re-check CI: `ops_ci_status(wait=true)`
   - Report outcome
