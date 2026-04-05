---
description: "Manage the prompt registry. $ARGUMENTS: (empty)=stats, 'search <query>'=find prompts, 'score <hash>'=quality score, 'improve <hash>'=enhance, 'export'=export all"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)**: Call `mcp__dotfiles__prompt_stats` — aggregate stats by repo, status, task type, grade
- **"search <query>"**: Call `mcp__dotfiles__prompt_search` with `query=<query>` — full-text search
- **"score <hash>"**: Call `mcp__dotfiles__prompt_score` with `hash=<hash>` — 10-dimension quality report
- **"improve <hash>"**: Call `mcp__dotfiles__prompt_improve` with `hash=<hash>` — 13-stage enhancement
- **"get <hash>"**: Call `mcp__dotfiles__prompt_get` with `hash=<hash>` — retrieve full prompt
- **"tag <hash> <tags>"**: Call `mcp__dotfiles__prompt_tag` with tags to add
- **"export"**: Call `mcp__dotfiles__prompt_export` — export all prompts as markdown
