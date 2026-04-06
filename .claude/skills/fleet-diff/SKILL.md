---
name: fleet-diff
description: Show what changed across all repos since a date — commit counts, churn, types, authors. The "what happened while I was away" dashboard.
allowed-tools: mcp__dotfiles__ops_fleet_diff
---

Show fleet-wide changes since a date or relative time. `$ARGUMENTS` is the time range (e.g., `3d`, `1w`, `2024-04-01`). Defaults to `7d` if empty.

## Workflow

1. Call `ops_fleet_diff(since="$ARGUMENTS")`
2. Display results as a dashboard:

```
## Fleet Activity: since {since}

**{active_repos} active repos** | {total_commits} commits | +{total_insertions} -{total_deletions}

### Most Active
{most_active as numbered list}

### Per-Repo Breakdown

| Repo | Commits | +/- | Types | Authors |
|------|---------|-----|-------|---------|
| {repo} | {commits} | +{ins} -{del} | feat:{n} fix:{n} chore:{n} | {authors} |

```

If no repos have activity, display:
```
## Fleet Activity: since {since}

No changes detected across the fleet. Everything is quiet.
```
