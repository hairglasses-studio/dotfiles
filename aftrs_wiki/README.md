# AFTRS Wiki - Comprehensive Development Documentation

## Overview
This wiki serves as a comprehensive knowledge base for all development projects, CLI tools, infrastructure, and systems developed by the AFTRS team. It provides persistent context for AI agents and human developers working across the entire ecosystem.

## 🔬 Detailed Project Breakdown (2025-09-14)

### AgentCTL (AI Wiki System & Network Management)
- Git-based, AI-friendly documentation system with persistent memory for all projects and decisions
- Flask web app, PostgreSQL DB, Nginx reverse proxy, Supervisor, Docker containerization
- Multi-model AI support: OpenAI, Anthropic, Ollama (auto-selection)
- Integrated network management (aftrs_cli): asset-driven inventory, real-time monitoring, diagnostics, web dashboard
- Features: semantic search, auto-cross-referencing, smart commit messages, template-driven docs, backup/recovery, health monitoring, dynamic test generation, automated documentation
- Unified CLI (`ai-wiki`), automated hooks, progress tracking, cross-referencing
- Troubleshooting: DB health, web app logs, AI model connectivity, asset loader, dashboard status
- [Full details](./projects/agentctl.md)

### AFTRS CLI (Asset-Driven Network Management)
- Centralized network automation/monitoring toolkit for aftrs.void infrastructure
- YAML asset inventory, dynamic test generation, real-time monitoring, web dashboard, scheduled automation
- Tailscale/OPNSense/UNRAID integration, advanced diagnostics, performance optimization, asset import/export
- Bulk GitHub repo management (ghorg + gum): interactive/headless workflows, persistent sync to ~/Docs
- Features: asset management, network testing, diagnostics, dashboard, Tailscale management, remote site management, UNRAID integration, documentation generation, automation pipeline
- Model management: agentic coding models, flexible version/context size
- Troubleshooting: diag, optimize, connectivity tests, logs
- [Full details](./projects/aftrs_cli.md)

### TSCTL (Tailscale Network Management)
- CLI tool for Tailscale network config, status, tag management, health diagnostics
- YAML config, direct Tailscale API integration, automated hooks, shell completions
- Features: interactive setup, tag verification, health diagnostics, bootstrap/tag automation
- Daily ops: status, tag check, health monitor; troubleshooting: doctor, setup, unhook
- Security: private repo, tag-based access, audit trail
- Integration: AgentCTL, AFTRS CLI, UNRAID Scripts, OPNSense
- [Full details](./projects/tsctl.md)

### CR8 CLI (Media Processing & Content Management)
- Modular toolkit for DJ crates, playlist archives, metadata sync, YouTube/SoundCloud conversion
- Bash library system, PostgreSQL DB, Flask web portal, Docker containerization, authentication system
- Features: multi-platform sync, premium auth, crate management, DJ software integration, cloud backup, analytics
- Core commands: media download, DB management, crate export/import, Rekordbox integration, vault/auth, sync/monitoring, system management
- Web portal: playlist/track views, crate upload, sync/vault actions, diff tool, animated charts
- Troubleshooting: vault, DB, download, web portal, debug mode
- Integration: AgentCTL, UNRAID Scripts, OAI CLI, AFTRS CLI
- [Full details](./projects/cr8_cli.md)

### OAI CLI (OpenAI Content Archiving)
- Archiving/backup for OpenAI videos, images, chat conversations; web dashboard, DB tracking, automated workflows
- Multi-content archiving, PostgreSQL DB, Flask dashboard, Docker, browser automation (Chrome/Chromium)
- Features: image/video/chat archiving, search/discovery, analytics, security/logging
- Core commands: image/video/chat archiving, DB backup/restore, search, system health, monitoring, cleanup
- Web dashboard: real-time monitoring, search, content browsing, stats
- Troubleshooting: DB, Chrome, download, dashboard, debug mode
- Integration: AgentCTL, CR8 CLI, UNRAID Scripts, AFTRS CLI
- [Full details](./projects/oai_cli.md)

### UNRAID Scripts (Server Automation & Maintenance)
- Modular automation/maintenance for UNRAID server: system maintenance, backup, media processing, health monitoring
- Bash/Python scripts, cron jobs, Docker, FFmpeg integration
- Features: automated maintenance, backup management, media processing, Sora integration, health monitoring, CLI tools
- Core commands: maintenance, health check, backup, media processing, Sora backup, CLI tools
- Advanced workflows: daily/weekly/monthly ops, batch media, deduplication, organization, backup
- Troubleshooting: backup, media, Sora, performance, debug mode
- Integration: AgentCTL, CR8 CLI, OAI CLI, AFTRS CLI
- [Full details](./projects/unraid_scripts.md)

### Console-Hax PS2 Interoperability
- Unified workflow for building/running PS2 homebrew across ~/Docs/console-hax
- Shared scripts: docker_build.sh, run_ps2link.sh, fetch_irx.sh, common.sh
- Makefile targets: docker-build, run-ps2link; IRX module management
- Notes: reproducible builds, environment variables, network run requirements
- [Full details](./projects/console-hax-ps2-interoperability.md)

---
All project documentation, architecture, workflows, and integration points are now fully detailed and cross-referenced. This ensures future-proofing and prevents code loss during consolidation, migration, or archival.

## 📚 Org Journals & Inventories

- Console-Hax Inventory (TSV): ./projects/console-hax-inventory.tsv
- Console-Hax Consolidation Notes: ./projects/console-hax-consolidation.md
- Multi-Repo Workflow Assessment: ./infrastructure/multi-repo-workflows.md
- AFTRS-AI Inventory (TSV): ./projects/aftrs-ai-inventory.tsv
- AFTRS-AI Org Breakdown: ./projects/aftrs-ai.md
- Hairglasses Inventory (TSV): ./projects/hairglasses-inventory.tsv
- Hairglasses Org Breakdown: ./projects/hairglasses.md
- Secretstudios Inventory (TSV): ./projects/secretstudios-inventory.tsv
- Secretstudios Org Breakdown: ./projects/secretstudios.md
- Org-Admin Inventory (TSV): ./projects/org-admin-inventory.tsv
- Org-Admin Org Breakdown: ./projects/org-admin.md
## 🏗️ Architecture Overview

### Core Systems
- **AgentCTL**: AI-powered wiki system with Docker deployment
- **AFTRS CLI**: Network management, automation tools, and bulk GitHub repo management (ghorg + gum)
- **TSCTL**: Tailscale network management and automation
- **CR8 CLI**: Media processing and content management
- **OAI CLI**: OpenAI API integration and content scraping
- **UNRAID Scripts**: Server automation and maintenance

#### Bulk GitHub Repo Management (ghorg + gum)
- Bulk clone, sync, check status, autopush, and reset all repos from your GitHub orgs/users
- All repos are always cloned and synced to `~/Docs` (your home Docs directory). The script will create and populate this directory if it doesn't exist or is empty.
- Interactive (gum) or headless (YAML config) workflows
- All install scripts ensure `ghorg` and `gum` are present for automation
- See [AFTRS CLI](./projects/aftrs_cli.md) for full details

### Infrastructure
- **UNRAID Server**: High-performance storage and compute
- **Ollama Model Hub**: Local registry bootstrapped with six open LLM agents (`mistral-agent`, `phi3-agent`, `llama3-agent`, `deepseek-agent`, `wizard-agent`, `qwen-agent`). Use `open-webui/setup_ollama_models.sh` to refresh/repair.
- **OPNSense Router**: Network routing and firewall
- **Tailscale Network**: Secure VPN and remote access
- **Comcast Modem**: Internet connectivity

## 📚 Documentation Structure

### Projects Index
- [AgentCTL](./projects/agentctl.md) - AI Wiki System
- [AFTRS CLI](./projects/aftrs_cli.md) - Network Management
- [TSCTL](./projects/tsctl.md) - Tailscale Management
- [CR8 CLI](./projects/cr8_cli.md) - Media Processing
- [OAI CLI](./projects/oai_cli.md) - OpenAI Integration
- [UNRAID Scripts](./projects/unraid_scripts.md) - Server Automation

### Infrastructure
- [Network Architecture](./infrastructure/network.md) - Physical and logical network design
- [UNRAID Server](./infrastructure/unraid.md) - Server configuration and management
- [Tailscale Configuration](./infrastructure/tailscale.md) - VPN setup and management
- [OPNSense Router](./infrastructure/opnsense.md) - Router configuration

### Development Patterns
- [CLI Design Patterns](./patterns/cli_design.md) - Common CLI tool patterns
- [Docker Deployment](./patterns/docker_deployment.md) - Container deployment strategies
- [Git Workflow](./patterns/git_workflow.md) - Version control practices
- [AI Integration](./patterns/ai_integration.md) - AI model integration patterns

### Future Roadmap
- [Centralization Plans](./roadmap/centralization.md) - UNRAID server consolidation
- [Feature Backlog](./roadmap/features.md) - Planned features and improvements
- [Known Issues](./roadmap/issues.md) - Current problems and workarounds

## 🚀 Quick Start

### For AI Agents
1. Review the [Project Index](./projects/) to understand the codebase
2. Check [Development Patterns](./patterns/) for design norms
3. Consult [Infrastructure](./infrastructure/) for system context
4. Reference [Future Roadmap](./roadmap/) for planned work

### For Human Developers
1. Start with [Network Architecture](./infrastructure/network.md)
2. Review [CLI Design Patterns](./patterns/cli_design.md)
3. Check [Known Issues](./roadmap/issues.md) for current problems
4. Explore [Future Roadmap](./roadmap/features.md) for next steps

## 🔧 Development Environment

### Prerequisites
- Linux/WSL2 environment
- Docker and Docker Compose
- Python 3.9+
- Git
- Tailscale client
- Access to UNRAID server

### Common Commands
```bash
# Start AgentCTL wiki
cd agentctl && ./deploy.sh

# Bulk clone/sync all GitHub org repos (interactive, always in ~/Docs)
cd aftrs_cli && ./aftrs.sh repos clone

# Bulk clone/sync all GitHub org repos (headless, always in ~/Docs)
cd aftrs_cli && ./aftrs.sh repos clone --headless

# Check repo sync status (in ~/Docs)
cd aftrs_cli && ./aftrs.sh repos status

# Access Tailscale network
tsctl status

# Backup UNRAID config
cd unraid_config_backup && ./backup_unraid_enhanced.sh

# Process media with CR8
cd cr8_cli && ./cr8 process
```

## 📊 System Status

### Active Projects
- ✅ AgentCTL - Production ready with Docker
- ✅ TSCTL - Tailscale management operational
- ✅ CR8 CLI - Media processing functional
- ✅ OAI CLI - OpenAI integration working
- 🔄 AFTRS CLI - Network management in development
- 🔄 UNRAID Scripts - Automation in progress

### Infrastructure Health
- ✅ UNRAID Server - Operational with NVMe storage
- ✅ OPNSense Router - Configured and running
- ✅ Tailscale Network - Secure VPN active
- ✅ Comcast Modem - Internet connectivity stable

## 🤖 AI Agent Context

This wiki is designed to provide comprehensive context for AI agents working on the AFTRS ecosystem. Key design principles:

1. **Persistent Memory**: All decisions, patterns, and context preserved
2. **Structured Documentation**: Consistent templates and metadata
3. **Cross-References**: Links between related systems and decisions
4. **Version Control**: Git-based tracking of all changes
5. **AI-Friendly**: Markdown with YAML frontmatter for parsing

## 📝 Contributing

### For AI Agents
- Update documentation when making changes
- Follow established patterns and templates
- Cross-reference related systems and decisions
- Maintain git history with meaningful commits

### For Human Developers
- Review and approve AI-generated changes
- Add context and explanations where needed
- Update roadmap and known issues
- Maintain system architecture documentation

---

*Last Updated: 2025-01-14*
*Maintained by: AFTRS Development Team*

## aftrs-shell Org Consolidation Journal (2025-09-14)

- All forked public repos in `aftrs-shell` have been archived using GitHub CLI to prevent confusion with internal codebase and documentation.
- Unique internal repos identified: `dotfiles`, `cbonsai`, `org-admin` (excluding `.github` and `aftrs-shell` meta-repo).
- GitHub CLI does not support direct org-to-org repo transfer; migration to `aftrs-void` must be performed via GitHub web UI or API.
- All findings, actions, and recommendations are documented here for future reference and audit.
- Proceeding with org-wide consolidation and documentation for remaining orgs as planned.
## 📊 Repository Statistics (Updated: 2025-09-22)

**Total Repositories:** 30

### By Category
- **Infrastructure:** 16 repositories
- **Miscellaneous:** 7 repositories
- **Projects:** 7 repositories

### By Language
- **Bash:** 18 repositories
- **Python:** 18 repositories
- **Yaml:** 16 repositories
- **Javascript:** 4 repositories

## Phase 2 Consolidations

*Additional consolidations completed on 2025-09-23*

### Infrastructure Backups
- [Configuration Backups](infrastructure/backups/README.md) - Consolidated config backup repositories

### Utilities
- [Small Utilities](projects/utilities/README.md) - Consolidated single-purpose utility repositories

## Consolidated Project Plans

*The following projects were consolidated from separate repositories on 2025-09-23*

### Project Plans
- [aftrs-terraform](projects/aftrs-terraform.md)
- [secret_cli](projects/secret_cli.md)
- [wizardstower-unraid](projects/wizardstower-unraid.md)
- [tailscale-acl](projects/tailscale-acl.md)
- [bench_cli](projects/bench_cli.md)
