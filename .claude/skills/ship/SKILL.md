---
name: ship
description: Gate-check, commit, push, and create a PR in one command. Blocks if tests fail.
allowed-tools: mcp__dotfiles__ops_ship, mcp__dotfiles__ops_pre_push, mcp__dotfiles__ops_ci_status, mcp__dotfiles__ops_changed_files
---

Ship the current changes: run pre-push gate, commit, push, and create a PR. `$ARGUMENTS` is the commit message (required). Append `--draft` for a draft PR, `--dry-run` to preview without executing.

## Workflow

1. **Parse arguments**:
   - Commit message: everything before `--draft` or `--dry-run` flags
   - `--draft`: create as draft PR
   - `--dry-run`: preview only (DEFAULT if no flags specified — you must explicitly say "not dry run" or "execute" to ship for real)

2. **Preview changes** — run `ops_changed_files()` and display what will be committed

3. **Run the ship** — call `ops_ship(message="<commit msg>", draft=<bool>, dry_run=<bool>)`

4. **Display results** based on outcome:

### If `shipped`:
```
## Shipped!

Commit: {sha} — {message}
PR: {url} (#{number})
Branch: {head} -> {base}
Files: {files_staged count} | +{insertions} -{deletions}

Run /ci-check to monitor GitHub Actions.
```

### If `blocked`:
```
## Blocked at: {blocked_at}

Pre-push gate failed at step: {failed_step}

{Show errors/failures table from pre_push output}

Fix the issues and run /iterate, then try /ship again.
```

### If `dry_run`:
```
## Ship Preview (dry run)

Would commit: {message}
Would create PR: {title}
Files to commit: {files_staged}
Changes: +{insertions} -{deletions}

Pre-push gate: {overall}
{Show step results table}

Run `/ship <message>` with `--execute` to ship for real.
```

5. **After shipping** — if not dry-run, automatically run `ops_ci_status(wait=false)` to show initial CI state
