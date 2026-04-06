---
name: auto-heal
description: Self-healing iteration loop — detect failures, auto-fix mechanical issues, verify. Closes the loop for missing deps, unused vars, missing imports.
allowed-tools: mcp__dotfiles__ops_iterate, mcp__dotfiles__ops_auto_fix, mcp__dotfiles__ops_changed_files, mcp__dotfiles__ops_session_status
---

Run a self-healing development loop: iterate -> detect failures -> auto-fix mechanical issues -> re-iterate to verify. `$ARGUMENTS` is an optional session ID.

## Workflow

1. **Check changed files** — run `ops_changed_files()` to understand scope
2. **Run first iteration** — call `ops_iterate(session_id="$ARGUMENTS")`
3. **If failures detected**, check if auto-fixable:
   - Call `ops_auto_fix(issues=<iterate.analysis.issues>, execute=false)` to preview patches
   - If `patch_count > 0`, show the preview:

```
## Auto-Fix Preview

{patch_count} fixable issue(s) found:

| File | Action | Description |
|------|--------|-------------|
| {file} | {action} | {before} -> {after} |

{remaining_issues count} issue(s) require manual fix.
```

4. **Apply fixes** — call `ops_auto_fix(issues=<same>, execute=true)`
5. **Verify** — call `ops_iterate()` again to confirm fixes worked

### If all pass after auto-fix:
```
## Auto-Heal: SUCCESS

Round 1: {error_count} errors detected
Auto-fix: {applied_count} patches applied ({action types})
Round 2: ALL PASS

Ready to ship! Run /ship to commit and create a PR.
```

### If issues remain:
```
## Auto-Heal: PARTIAL

Round 1: {error_count} errors
Auto-fix: {applied_count}/{patch_count} patches applied
Round 2: {remaining} errors remain

### Remaining Issues (manual fix needed)
| File | Line | Category | Message |
|------|------|----------|---------|
| {file} | {line} | {category} | {message} |

Fix these manually and run /iterate again.
```

### If all_pass on first iteration:
```
## Auto-Heal: Already green!

Build: OK | Tests: {passed} passed, 0 failed
No fixes needed. Run /ship to commit.
```
