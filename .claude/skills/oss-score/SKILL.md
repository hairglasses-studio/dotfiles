---
name: oss-score
description: Score a repo's open-source readiness against GitHub best practices (0-100)
allowed-tools: Bash, Read, Grep, Glob, mcp__dotfiles__dotfiles_oss_score, mcp__dotfiles__dotfiles_oss_check
---

Score a repository's open-source readiness. $ARGUMENTS should be the repo path (e.g., `~/hairglasses-studio/mcpkit`). If no path given, use the current working directory.

1. Run `dotfiles_oss_score(repo_path="<resolved absolute path>")` to get the full scoring report
2. Display results as a markdown table:

```
## OSS Readiness: <repo-name> — <score>/100 (Grade: <grade>)

| Category | Score | Checks |
|----------|-------|--------|
| Community Files | X/20 | README (pass/fail), LICENSE, ... |
| README Quality | X/15 | badges, install, usage, ... |
| ... | | |

### Top Action Items
1. [+N pts] suggestion...
2. [+N pts] suggestion...
```

3. For any category scoring below 50%, run `dotfiles_oss_check(repo_path="...", category="<name>")` for detailed per-check breakdown and specific suggestions
4. End with a summary: total score, grade, and the single most impactful action item to improve the score
