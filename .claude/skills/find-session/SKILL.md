---
name: find-session
description: Find a past Claude Code session by topic/keyword and copy the resume command to clipboard
allowed-tools: mcp__dotfiles__claude_session_search, mcp__dotfiles__claude_session_detail, Bash
---

Find past Claude Code sessions by topic, keyword, or content across all repos.

1. Parse `$ARGUMENTS` as the search query. If empty, ask the user what to search for.

2. Call `claude_session_search(query="$ARGUMENTS")` to find matching sessions.

3. Display results as a markdown table:

```
| # | Status | Repo | Title | Age | Hits |
|---|--------|------|-------|-----|------|
```

4. Show the resume command for each result:
```
1. cd /path/to/repo && claude --resume <uuid>
```

5. Copy the top result's resume command to the clipboard:
   - macOS: `echo -n "<cmd>" | pbcopy`
   - Linux: `echo -n "<cmd>" | wl-copy`

6. If no results found, suggest broadening the search (shorter keywords, different terms).

7. If the user wants more detail on a specific result, call `claude_session_detail(session_id="<uuid>")` to show the full session info including recent prompts and tasks.
