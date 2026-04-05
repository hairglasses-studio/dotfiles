---
name: review
description: Review a PR's changes, check CI status, and provide structured feedback with approve/request-changes decision
allowed-tools: Bash, Read, Grep, Glob, mcp__dotfiles__ops_changed_files, mcp__dotfiles__ops_ci_status, mcp__dotfiles__ops_analyze_failures, mcp__dotfiles__ops_build
---

Review a pull request. `$ARGUMENTS` should be a PR number (e.g., `42`) or URL. If empty, review the current branch's changes against main.

## Workflow

1. **Get PR details** — run `gh pr view $ARGUMENTS --json title,body,headRefName,baseRefName,additions,deletions,changedFiles,reviewDecision,commits,url` via Bash
2. **Check CI** — run `ops_ci_status(pr=$ARGUMENTS)` for check status
3. **Get changed files** — run `ops_changed_files()` to understand the diff
4. **Read key files** — for each changed file, read the relevant sections to understand the changes
5. **Build check** — run `ops_build()` to verify compilation

6. **Report** as structured markdown:

```
## PR Review: #{number} — {title}

**Branch:** {head} → {base} | **CI:** {pass/fail/pending}
**Changes:** +{additions} -{deletions} across {changedFiles} files

### Changed Files
| File | +/- | Assessment |
|------|-----|------------|
| {path} | +{ins} -{del} | {brief assessment} |

### Issues Found
| Severity | File | Line | Issue |
|----------|------|------|-------|
| {warn/error} | {file} | {line} | {description} |

### Summary
{1-3 sentence overall assessment}

### Decision: {APPROVE / REQUEST_CHANGES / COMMENT}
{reasoning for decision}
```

7. If `$ARGUMENTS` includes `--approve` and no issues found, run `gh pr review $PR --approve` via Bash
