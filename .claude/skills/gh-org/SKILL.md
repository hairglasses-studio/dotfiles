---
description: "GitHub org lifecycle management. $ARGUMENTS: (empty)=audit local vs remote, 'list'=list org repos, 'sync'=full sync, 'clone'=clone missing, 'pull'=pull all, 'archive <name>'=archive repo, 'settings'=bulk apply settings, 'transfer'=transfer repos to org"
user_invocable: true
allowed-tools: mcp__dotfiles__dotfiles_gh_list_org_repos, mcp__dotfiles__dotfiles_gh_list_personal_repos, mcp__dotfiles__dotfiles_gh_local_sync_audit, mcp__dotfiles__dotfiles_gh_bulk_clone, mcp__dotfiles__dotfiles_gh_pull_all, mcp__dotfiles__dotfiles_gh_clean_stale, mcp__dotfiles__dotfiles_gh_full_sync, mcp__dotfiles__dotfiles_gh_bulk_archive, mcp__dotfiles__dotfiles_gh_bulk_settings, mcp__dotfiles__dotfiles_gh_transfer_repos, mcp__dotfiles__dotfiles_gh_recreate_forks, mcp__dotfiles__dotfiles_gh_onboard_repos
---

Parse `$ARGUMENTS`:

- **(empty)** or **"audit"**: Call `mcp__dotfiles__dotfiles_gh_local_sync_audit` — show orphaned, missing, and mismatched repos
- **"list"**: Call `mcp__dotfiles__dotfiles_gh_list_org_repos` — all org repos with local clone status. Supports filters: `list --missing`, `list --archived`, `list --lang go`
- **"personal"**: Call `mcp__dotfiles__dotfiles_gh_list_personal_repos` — personal repos with fork/visibility metadata
- **"sync"**: Call `mcp__dotfiles__dotfiles_gh_full_sync` — pull all + audit + clone missing in one shot
- **"clone"**: Call `mcp__dotfiles__dotfiles_gh_bulk_clone` — clone all missing org repos locally
- **"pull"**: Call `mcp__dotfiles__dotfiles_gh_pull_all` — fetch/pull all local repos (skips dirty/detached)
- **"clean"**: Call `mcp__dotfiles__dotfiles_gh_clean_stale` — remove orphaned local clones (checks uncommitted/unpushed first)
- **"archive <name>"**: Call `mcp__dotfiles__dotfiles_gh_bulk_archive` with `repos=[<name>]`
- **"settings"**: Call `mcp__dotfiles__dotfiles_gh_bulk_settings` — apply standard settings to all org repos (reports before/after)
- **"transfer"**: Call `mcp__dotfiles__dotfiles_gh_transfer_repos` — transfer non-fork personal repos to org
- **"onboard <url>"**: Call `mcp__dotfiles__dotfiles_gh_onboard_repos` with `repos=[<url>]` — fork, squash history, clone locally

Present results as tables. For audit: show separate sections for orphaned, missing, and mismatched.
