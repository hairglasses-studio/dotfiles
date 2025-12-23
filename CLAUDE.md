# Dotfiles - Claude Code Instructions

## Project Overview
Unified configuration repository: shell tools, ecosystem wiki, bootstrap scripts, and development environment configurations for the AFTRS ecosystem.

## Repository Structure

```
dotfiles/
├── aftrs_wiki/              # Ecosystem documentation
│   ├── infrastructure/      # Network, backups, OPNsense
│   ├── projects/            # 15+ project docs
│   ├── patterns/            # Design patterns, AI
│   └── roadmap/             # Features, planning
├── aftrs-shell/             # Shell tools
│   ├── scripts/             # Automation scripts
│   └── submodules/          # antidote, oh-my-zsh
├── aftrs-init/              # Bootstrap system
│   ├── config/              # Config seeds
│   └── bin/                 # Init binaries
├── aftrs-cline/             # Cline AI tools
│   ├── mcp-server-projects/ # MCP servers
│   └── awesome-clinerules/  # Cline rules
├── aftrs-obsidianmd/        # Obsidian plugins
│   ├── obsidian-api/        # API docs
│   └── obsidian-plugins/    # Plugin dev
└── ricectl                  # CLI utility
```

## Code Standards

### Shell Scripts
- Use bash for portability
- Include shebang: `#!/bin/bash`
- Use `set -euo pipefail` for strict mode
- Validate with shellcheck
- Keep scripts POSIX-compatible when possible

### Documentation
- Use Markdown for all docs
- Follow existing wiki structure
- Include cross-references
- Add diagrams where helpful

### Configuration Files
- Use YAML/TOML when possible
- Include comments explaining options
- Provide example configurations
- Document all environment variables

## Key Files

### Shell Configuration
- `aftrs-shell/scripts/install.sh` - Plugin installation
- `aftrs-shell/completions/` - Shell completions
- `aftrs-shell/.zshrc` - Zsh configuration

### Bootstrap
- `aftrs-init/bin/bootstrap.sh` - Main bootstrap script
- `aftrs-init/config/` - Configuration seeds

### Wiki
- `aftrs_wiki/projects/` - Project documentation
- `aftrs_wiki/patterns/` - Design patterns
- `aftrs_wiki/infrastructure/` - Infra docs

## Common Tasks

### Update shell plugins
```bash
cd aftrs-shell && ./scripts/update.sh
```

### Bootstrap new host
```bash
./aftrs-init/bin/bootstrap.sh
```

### Validate shell scripts
```bash
shellcheck aftrs-shell/scripts/*.sh
shellcheck aftrs-init/bin/*.sh
```

## Environment Variables

| Variable | Purpose |
|----------|---------|
| `AFTRS_SHELL_HOME` | Shell tools root |
| `AFTRS_LOCAL_ROOT` | Local config root |
| `AFTRS_COMPLETIONS_DIR` | Completions directory |

## Documented Projects

The wiki documents 15+ projects in the AFTRS ecosystem:
- **AgentCTL** - AI wiki system (Flask, PostgreSQL)
- **AFTRS CLI** - Network automation toolkit
- **TSCTL** - Tailscale network management
- **CR8 CLI** - DJ media processing
- **OAI CLI** - OpenAI content archiving
- **UNRAID Scripts** - Server automation
- **Console-Hax** - PS2 homebrew

## Important Notes

- Shell scripts must work on both Linux and macOS
- Test bootstrap scripts in clean environment
- Keep wiki synchronized with actual project state
- Use oh-my-posh for consistent terminal theming
- Maintain completion files for all CLI tools
