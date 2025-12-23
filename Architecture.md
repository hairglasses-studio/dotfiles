# Dotfiles Architecture

## Overview

Unified configuration and ecosystem documentation repository consolidating shell tools, wiki system, bootstrap scripts, and development environment configurations.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Dotfiles System                            │
├─────────────────────────────────────────────────────────────────┤
│  ┌───────────────────────────────────────────────────────────┐  │
│  │                    Wiki Layer                             │  │
│  │  ┌──────────────────────────────────────────────────────┐ │  │
│  │  │                 aftrs_wiki/                          │ │  │
│  │  │  ├── infrastructure/  (OPNsense, backups, network)   │ │  │
│  │  │  ├── projects/        (15+ project docs)             │ │  │
│  │  │  ├── patterns/        (Design patterns, AI)          │ │  │
│  │  │  ├── roadmap/         (Features, issues)             │ │  │
│  │  │  └── miscellaneous/   (General reference)            │ │  │
│  │  └──────────────────────────────────────────────────────┘ │  │
│  └───────────────────────────────────────────────────────────┘  │
│                          │                                      │
│  ┌───────────────────────▼───────────────────────────────────┐  │
│  │                   Shell Tools Layer                       │  │
│  │  ┌──────────────────────────────────────────────────────┐ │  │
│  │  │                  aftrs-shell/                        │ │  │
│  │  │  ├── scripts/       (Automation, updates)            │ │  │
│  │  │  ├── submodules/    (antidote, oh-my-zsh, use-omz)   │ │  │
│  │  │  └── completions/   (zsh, bash, fish)                │ │  │
│  │  └──────────────────────────────────────────────────────┘ │  │
│  └───────────────────────────────────────────────────────────┘  │
│                          │                                      │
│  ┌───────────────────────▼───────────────────────────────────┐  │
│  │                   Bootstrap Layer                         │  │
│  │  ┌──────────────────────────────────────────────────────┐ │  │
│  │  │                   aftrs-init/                        │ │  │
│  │  │  ├── config/        (Config seeds)                   │ │  │
│  │  │  └── bin/           (Initialization binaries)        │ │  │
│  │  └──────────────────────────────────────────────────────┘ │  │
│  └───────────────────────────────────────────────────────────┘  │
│                          │                                      │
│  ┌───────────────────────▼───────────────────────────────────┐  │
│  │                 Development Tools Layer                   │  │
│  │  ┌─────────────────────────┐ ┌────────────────────────┐   │  │
│  │  │       aftrs-cline/      │ │   aftrs-obsidianmd/    │   │  │
│  │  │  ├── mcp-server-projects│ │  ├── obsidian-api     │   │  │
│  │  │  └── awesome-clinerules │ │  ├── obsidian-plugins │   │  │
│  │  │                         │ │  └── obsidian-importer│   │  │
│  │  └─────────────────────────┘ └────────────────────────┘   │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### AFTRS Wiki (`aftrs_wiki/`)
Comprehensive ecosystem documentation:

| Directory | Content |
|-----------|---------|
| `infrastructure/` | OPNsense, backups, networking patterns |
| `projects/` | 15+ detailed project docs (AgentCTL, CR8 CLI, TSCTL, etc.) |
| `patterns/` | Design patterns, AI integration, CLI patterns |
| `roadmap/` | Features, issues, planning |
| `miscellaneous/` | General reference |

### AFTRS Shell (`aftrs-shell/`)
Shell plugin meta-repository:

- **scripts/** - Automation and update scripts
- **submodules/** - antidote, oh-my-zsh, use-omz
- **completions/** - Aggregated completions (zsh, bash, fish)
- oh-my-posh powerline theme integration

**Environment Variables**:
- `AFTRS_SHELL_HOME` - Shell tools root
- `AFTRS_LOCAL_ROOT` - Local config root
- `AFTRS_COMPLETIONS_DIR` - Completions directory

### AFTRS Init (`aftrs-init/`)
Bootstrap system for new hosts:

- Unattended host configuration for aftrs.void network
- Config seeds in `aftrs_init/config`
- Automatic unpacking to `~/.config/...`
- Part of host provisioning pipeline

### AFTRS Cline (`aftrs-cline/`)
AI development tools:

- Cline IDE configuration and rules
- 5 MCP server projects
- awesome-clinerules collection
- Multi-assistant support

### AFTRS ObsidianMD (`aftrs-obsidianmd/`)
Obsidian plugin ecosystem:

- Official API docs
- Plugin development resources
- Data migration tools (importer)
- 8+ Obsidian utilities

## Consolidated Repositories

This monorepo consolidates 6+ previously separate repositories:

1. `aftrs_wiki` - Ecosystem documentation
2. `aftrs-shell` - Shell tools meta-repo
3. `aftrs-init` - Bootstrap system
4. `aftrs-cline` - Cline AI tools
5. `aftrs-obsidianmd` - Obsidian plugins
6. `.github` - Organization config

## Technology Stack

| Layer | Technology |
|-------|------------|
| Shell | zsh, bash, fish |
| Theme | oh-my-posh, Powerlevel10k |
| Plugin Manager | antidote |
| Docs | Markdown, Obsidian |
| AI Tools | Cline, MCP servers |
| Bootstrap | Bash scripts |

## Integration Points

### Documented Projects
- AgentCTL - AI wiki system
- AFTRS CLI - Network automation
- TSCTL - Tailscale management
- CR8 CLI - DJ media processing
- OAI CLI - OpenAI archiving
- UNRAID Scripts - Server automation
- Console-Hax - PS2 homebrew

### Organization Consolidation
Migrated from 6+ orgs to 3 focused orgs:
- aftrs-ai → aftrs-studio
- aftrs-shell → aftrs-studio
- secretstudios → aftrs-studio
- hairglasses → aftrs-studio
- console-hax → aftrs-studio
- org-admin → aftrs-studio
