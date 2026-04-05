---
description: Run a fleet health check across all repos in ~/hairglasses-studio.
user_invocable: true
---

Execute these in parallel:
1. `mcp__dotfiles__dotfiles_fleet_audit` — Per-repo language, Go version, CI status, test count, CLAUDE.md presence
2. `mcp__dotfiles__dotfiles_health_check` — Org-wide build status and pipeline.mk inclusion
3. `mcp__dotfiles__ops_session_list` — Active SDLC sessions across repos

Present results as a dashboard:
- Per-repo row: name | language | Go version | CI status | tests | last commit age | CLAUDE.md
- Highlight repos with failing CI, outdated Go version, or missing CLAUDE.md
- Show active SDLC sessions with iteration counts
- Calculate fleet convergence: % repos with all checks passing
