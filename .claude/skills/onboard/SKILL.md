---
description: "Scaffold new repos and onboard existing ones with standard files. $ARGUMENTS: 'new <name>'=create new repo, 'onboard <path>'=add standard files, 'score <path>'=check OSS readiness"
user_invocable: true
allowed-tools: mcp__dotfiles__dotfiles_create_repo, mcp__dotfiles__dotfiles_onboard_repo, mcp__dotfiles__dotfiles_oss_score
---

Parse `$ARGUMENTS`:

- **"new <name>"**: Call `mcp__dotfiles__dotfiles_create_repo` with `name=<name>` — scaffold a new repo with standard files (.editorconfig, CI, LICENSE, README, CLAUDE.md, .gitignore)
- **"onboard <path>"**: Call `mcp__dotfiles__dotfiles_onboard_repo` with `repo=<path>` — add missing standard files to an existing repo
- **"score <path>"**: Call `mcp__dotfiles__dotfiles_oss_score` with `repo=<path>` — score open-source readiness 0-100 across 8 categories
- **(empty)**: Show usage:
  ```
  Usage: /onboard <command>

  Commands:
    new <name>      Create a new repo with standard scaffolding
    onboard <path>  Add standard files to an existing repo
    score <path>    Score open-source readiness (0-100)
  ```

For `new` and `onboard`, display the files created/added. For `score`, display the per-category breakdown with pass/fail and action items.
