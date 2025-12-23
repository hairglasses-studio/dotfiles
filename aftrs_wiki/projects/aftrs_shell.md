# aftrs-shell Org Breakdown (2025-09-14)

## Overview
Meta-repository and automation for unified shell plugin/tool management, workspace generation, and prompt/completion setup. Focused on reproducibility, easy updates, and developer experience.

## Main Project: aftrs-shell Meta-Repo
- **Purpose:** Aggregate, install, and update all shell plugins/tools in the aftrs-shell org
- **Scripts:**
  - `install.sh`: Clone/update all org repos, run per-repo install, aggregate completions, install prompt theme
  - `aftrs_initial_unpack.sh`: Seed config templates to ~/.config
  - `generate_docs_workspace.sh`: Generate VS Code multi-root workspace for all repos in ~/Docs
  - `consolidate_plugins.sh`: (if present) for plugin aggregation
- **Submodules:**
  - `antidote/`, `oh-my-zsh/`, `use-omz/`: Upstream plugin sources
- **Config:**
  - `aftrs_init/config/`: Config templates for shell environment
  - `sync_fork_template.yml`: GitHub Actions workflow for fork sync
- **Completions & Prompt:**
  - Aggregated completions for zsh/bash/fish, powerline prompt via oh-my-posh
- **Workspace:**
  - Multi-root workspace generation for all git repos in ~/Docs
- **Documentation:**
  - `README.md`, `PLUGIN_MATRIX.md`, `PROJECT_NOTES.md`, `RICED_SHELL_SHOWCASE.md`

## Recommendations
- Use `install.sh` for all updates and initial setup
- Keep config templates backed up if customizing
- Use fork sync workflow for all forks to stay up to date with upstream
- Prefer script-based aggregation over submodules for flexibility
- Regenerate workspace after adding/removing repos

---
*Last updated: 2025-09-14*
*Prepared by: GitHub Copilot (automated agent)*
