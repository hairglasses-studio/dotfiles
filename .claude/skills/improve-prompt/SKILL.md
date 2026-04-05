---
name: improve-prompt
description: Analyze, score, and improve a prompt using the 13-stage enhancement pipeline. Shows quality report and enhanced version.
allowed-tools: mcp__dotfiles__prompt_capture mcp__dotfiles__prompt_score mcp__dotfiles__prompt_improve mcp__dotfiles__prompt_search mcp__dotfiles__prompt_get
---

Analyze and improve a prompt. `$ARGUMENTS` is the prompt text to improve, or `last` to improve the most recently captured prompt.

## Workflow

1. **Score** the prompt using `prompt_score`:
   - If `$ARGUMENTS` is `last`, use `prompt_search` with `limit: 1` to find the most recent, then `prompt_score` with its hash
   - Otherwise, score the raw text: `prompt_score(prompt="$ARGUMENTS")`

2. **Display** the quality report as a markdown table:

```
## Prompt Quality: {score}/100 (Grade: {grade})

| Dimension | Score | Grade | Top Suggestion |
|-----------|-------|-------|----------------|
| Clarity | X | A | ... |
| Specificity | X | B | ... |
| ... | | | |

Lint issues: {lint_count}
```

3. **Improve** if score < 75:
   - Run `prompt_improve(prompt="$ARGUMENTS")` (or with hash if from registry)
   - Show the stages that ran and improvements made
   - Display the enhanced prompt in a code block

4. **Capture** both versions to the registry:
   - `prompt_capture(prompt="<original>", repo="<current repo>")` if not already stored
   - The improved version is auto-captured by `prompt_improve`

5. **Summary**: Show score delta: **Original: {score}/100 -> Enhanced: {new_score}/100** (+{delta} points)
