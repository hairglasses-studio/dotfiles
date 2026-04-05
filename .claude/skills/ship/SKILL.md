---
name: ship
description: Gate-check, commit, push, and create a PR in one command. Blocks if tests fail.
allowed-tools: mcp__dotfiles__ops_ship, mcp__dotfiles__ops_pre_push, mcp__dotfiles__ops_ci_status, mcp__dotfiles__ops_changed_files
---

Ship the current changes: run pre-push gate, commit, push, and create a PR. `$ARGUMENTS` is the commit message (required). Append `--draft` for a draft PR, `--execute` to actually ship (dry-run by default).

## Workflow

1. **Parse arguments**:
   - Commit message: everything before `--draft` or `--execute` flags
   - `--draft`: create as draft PR
   - `--execute`: actually ship (DEFAULT is dry-run preview — you must pass `--execute` to ship for real)

2. **Preview changes** — run `ops_changed_files()` and display what will be committed

3. **Run the ship** — call `ops_ship(message="<commit msg>", draft=<bool>, execute=<bool>)`

4. **Display results** based on outcome:

### If `shipped`:
```
## Shipped!

Commit: {sha} — {message}
PR: {url} (#{number})
Branch: {head} -> {base}
Files: {files_staged count} | +{insertions} -{deletions}

Run /fix-ci to monitor GitHub Actions.
```

### If `blocked`:
```
## Blocked at: {blocked_at}

Pre-push gate failed at step: {failed_step}

{Show errors/failures table from pre_push output}

Fix the issues and run /iterate, then try /ship again.
```

### If dry-run (default):
```
## Ship Preview (dry run)

Would commit: {message}
Would create PR: {title}
Files to commit: {files_staged}
Changes: +{insertions} -{deletions}

Pre-push gate: {overall}
{Show step results table}

Run `/ship <message> --execute` to ship for real.
```

5. **After shipping** — if executed, automatically run `ops_ci_status(wait=false)` to show initial CI state
