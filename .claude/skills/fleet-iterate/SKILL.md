---
name: fleet-iterate
description: Run build+test across all repos in ~/hairglasses-studio and report a fleet health matrix
allowed-tools: mcp__dotfiles__ops_fleet_iterate
---

Run build and test across all project repos. `$ARGUMENTS` can be:
- Empty: test all Go/Node/Python repos (up to 20)
- `go`: only Go repos
- `node`: only Node.js repos
- `python`: only Python repos

## Workflow

1. Run `ops_fleet_iterate(language="$ARGUMENTS")` (or `all` if empty)

2. Display results as:

```
## Fleet Health: {passing}/{total} repos passing

| Repo | Language | Build | Tests | Errors | Status | Time |
|------|----------|-------|-------|--------|--------|------|
| {repo} | {lang} | {ok/fail} | {passed}/{failed} | {count} | {pass/fail} | {ms} |

### Summary
- Passing: {passing}
- Failing: {failing} 
- Skipped: {skipped}

### Failing Repos
{For each failing repo, show error details}
```

3. If all pass: "Fleet is healthy."
4. If any fail: suggest running `/iterate` on each failing repo individually.
