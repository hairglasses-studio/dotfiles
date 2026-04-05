---
description: "Sync all repos: pull, audit, clone missing. $ARGUMENTS can be: (empty)=full sync, 'pull'=pull only, 'audit'=audit only, 'clone'=clone missing"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)** or **"sync"**: Call `mcp__dotfiles__dotfiles_gh_full_sync` for complete pull+audit+clone
- **"pull"**: Call `mcp__dotfiles__dotfiles_gh_pull_all` to fetch/pull all repos
- **"audit"**: Call `mcp__dotfiles__dotfiles_gh_local_sync_audit` to check orphaned/missing/mismatched
- **"clone"**: Call `mcp__dotfiles__dotfiles_gh_bulk_clone` to clone all missing repos

Display per-repo status with dirty/detached warnings.
