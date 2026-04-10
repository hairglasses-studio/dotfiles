---
name: github_stars
description: Organize GitHub starred repositories into GitHub lists using the dotfiles GitHub Stars MCP workflow and shell entrypoints.
allowed-tools:
  - Bash
  - Read
  - Write
  - Grep
  - mcp__dotfiles_workspace__dotfiles_tool_search
  - mcp__dotfiles_workspace__dotfiles_tool_schema
  - mcp__dotfiles_workspace__dotfiles_tool_catalog
  - mcp__dotfiles_workspace__dotfiles_server_health
  - mcp__dotfiles_workspace__dotfiles_gh_stars_list
  - mcp__dotfiles_workspace__dotfiles_gh_stars_summary
  - mcp__dotfiles_workspace__dotfiles_gh_star_lists_list
  - mcp__dotfiles_workspace__dotfiles_gh_star_lists_ensure
  - mcp__dotfiles_workspace__dotfiles_gh_star_lists_rename
  - mcp__dotfiles_workspace__dotfiles_gh_star_lists_delete
  - mcp__dotfiles_workspace__dotfiles_gh_stars_set
  - mcp__dotfiles_workspace__dotfiles_gh_star_membership_set
  - mcp__dotfiles_workspace__dotfiles_gh_stars_cleanup_candidates
  - mcp__dotfiles_workspace__dotfiles_gh_stars_taxonomy_suggest
  - mcp__dotfiles_workspace__dotfiles_gh_stars_taxonomy_audit
  - mcp__dotfiles_workspace__dotfiles_gh_stars_taxonomy_sync
  - mcp__dotfiles_workspace__dotfiles_gh_stars_audit_markdown
  - mcp__dotfiles_workspace__dotfiles_gh_stars_sync_markdown
  - mcp__dotfiles_workspace__dotfiles_gh_stars_bootstrap
  - mcp__dotfiles_workspace__dotfiles_gh_stars_install_codex_mcp
---

# GitHub Stars

Use this skill when the task is about starred GitHub repositories, GitHub star folders/lists, or installing the personal MCP surface for star organization.

## Default loop

1. Inspect current stars and GitHub lists first with `dotfiles_gh_stars_list` and `dotfiles_gh_star_lists_list`.
2. Use `dotfiles_gh_stars_summary` when you need a fast overview of list coverage, language clusters, or how much of the starred set is still unmanaged.
3. For new taxonomy ideas, use `dotfiles_gh_stars_taxonomy_suggest` before mutating anything.
4. Use `dotfiles_gh_stars_taxonomy_audit` before large syncs to find unmanaged repos, profile drift, and desired repos that are not yet starred.
5. Use `dotfiles_gh_stars_cleanup_candidates` to prune archived, forked, stale, or unlisted stars before reorganizing.
6. Prefer `dotfiles_gh_stars_taxonomy_sync` for repeatable repo-to-list reconciliation instead of one-off list edits.
7. Use `dotfiles_gh_stars_bootstrap` for the first-time MCP maturity setup (`MCP / Production`, `MCP / Experimental`, `MCP / Alpha`).
8. Use `dotfiles_gh_stars_audit_markdown` first when a markdown bundle or `github-reference-repos/index-files` directory should be checked against live GitHub list membership.
9. Use `dotfiles_gh_stars_sync_markdown` after the audit when that markdown corpus should become the source of truth for per-file star lists.
10. Use `dotfiles_gh_stars_install_codex_mcp` or `scripts/hg-github-stars.sh install-codex-mcp --execute` to install the global Codex MCP entries.

## Shell fallback

- `scripts/hg-github-stars.sh list-stars --include-lists`
- `scripts/hg-github-stars.sh summary --managed-prefix 'MCP / '`
- `scripts/hg-github-stars.sh list-lists --include-items`
- `scripts/hg-github-stars.sh rename-list --from 'Old Name' --to 'New Name' --execute`
- `scripts/hg-github-stars.sh suggest-taxonomy`
- `scripts/hg-github-stars.sh audit-taxonomy --managed-prefix 'MCP / ' --bootstrap-defaults`
- `scripts/hg-github-stars.sh cleanup-candidates --managed-prefix 'MCP / ' --require-unlisted`
- `scripts/hg-github-stars.sh audit-markdown --source-dir /home/hg/github-reference-repos/index-files`
- `scripts/hg-github-stars.sh set-membership --repos owner/name --lists list-a,list-b --operation replace --execute`
- `scripts/hg-github-stars.sh sync-markdown --source-dir /home/hg/github-reference-repos/index-files --execute`
- `scripts/hg-github-stars.sh bootstrap --install-codex-mcp --execute`

## Notes

- The workflow uses GitHub’s live GraphQL `UserList` surface, not browser automation.
- GitHub auth prefers `GITHUB_PAT` from `~/.env`, then falls back to existing shell vars and finally `gh auth token`.
- Mutating tools are dry-run by default. Set `execute=true` when you want the changes applied.
- `replace_managed` is the safest bulk mode when you only want to rewrite lists under a managed prefix like `MCP / `.
- The installed official wrapper passes through `GITHUB_TOOLSETS`, `GITHUB_TOOLS`, `GITHUB_DYNAMIC_TOOLSETS`, `GITHUB_READ_ONLY`, `GITHUB_INSIDERS`, `GITHUB_LOCKDOWN_MODE`, `GITHUB_MCP_SERVER_NAME`, and `GITHUB_MCP_SERVER_TITLE`.
