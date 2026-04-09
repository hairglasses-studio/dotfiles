# Install and Operations Guide

This guide aligns the public-facing operational story with the scripts that actually exist in the repo today.

## Scope Split

### User-scoped bootstrap

`install.sh` is the main user-scoped bootstrap entrypoint.

What it does:

- bootstraps package tooling for the detected platform
- installs shell/editor helpers such as Oh My Zsh, TPM, and vim-plug
- creates the managed symlink surface into the user config directories
- supports `--check` for validation and `--print-link-specs` for machine-readable link inventory

What it does not do:

- deploy `/etc` configuration directly
- replace the standalone system deploy helpers
- manage hosted workflow sync

## Machine-global deployment helpers

Use the dedicated deploy scripts when the target is machine-global instead of user-scoped.

- `scripts/etc-deploy.sh`: tracked `/etc` config deployment
- `scripts/greetd-deploy.sh`: greetd deployment path
- `scripts/logiops-deploy.sh`: system-level Logitech config deployment
- `scripts/refind-deploy.sh`: rEFInd deployment
- `scripts/plymouth-deploy.sh`: Plymouth deployment

These are the right entrypoints when the change belongs under `/etc`, bootloader state, or another machine-global surface.

## Documentation parity helpers

Use the repo scripts instead of editing compatibility mirrors by hand.

- `scripts/hg-agent-docs.sh --source auto .`: regenerate `CLAUDE.md`, `GEMINI.md`, and `.github/copilot-instructions.md` from the canonical source file
- `scripts/hg-skill-surface-sync.sh`: refresh generated skill mirrors from `.agents/skills/`
- `scripts/hg-agent-home-sync.sh`: align `/home/hg` and `/root` provider homes, managed skill mirrors, and workspace-global overlays

The rule is simple:

- `AGENTS.md` is canonical when marked canonical.
- `CLAUDE.md`, `GEMINI.md`, and Copilot instructions are mirrors.
- `.agents/skills/` is canonical; `.claude/skills/` is generated compatibility output.

## Canonical Agent Launchers

Agent launch behavior is centralized in repo-managed launchers instead of shell-local one-offs.

- `scripts/hg-codex-launch.sh`
- `scripts/hg-claude-launch.sh`
- `scripts/hg-gemini-launch.sh`
- `scripts/hg-codex-worktree-prune.sh`

Operational contract:

- launchers re-exec through `sudo -H` when needed
- git repos launch into fresh managed worktrees under `/root/.codex/worktrees`
- dirty tracked and untracked non-ignored state is carried into the fresh worktree
- non-git directories run in place
- sessions already inside a managed worktree do not nest another worktree

Use these scripts from shell wrappers, tmux/bootstrap entrypoints, and automation instead of calling raw `codex`, `claude`, or `gemini` binaries directly.

## Workflow sync status

`scripts/hg-workflow-sync.sh` is intentionally retired as a hosted workflow mutator.

Current behavior:

- `--help` explains that hosted workflow sync is retired.
- `--dry-run` is informational only.
- the script does not create, update, commit, or push workflow files.

Use it as an informational surface, not as a repo-mutation command.

## Recommended Operator Checks

Use the smallest check that matches the surface you touched.

### Repo docs and parity

```bash
bash ./scripts/hg-agent-docs.sh --source auto .
bash ./scripts/hg-agent-home-sync.sh --check
```

### Workflow-sync messaging path

```bash
bash ./scripts/hg-workflow-sync.sh --dry-run
```

### Script syntax

```bash
bash -n scripts/*.sh scripts/lib/*.sh
```

### Installer link inventory

```bash
bash ./install.sh --print-link-specs
bash ./install.sh --check
```

### Canonical parity audit

```bash
bash ./scripts/hg-codex-audit.sh --write-workspace-cache --write-wiki-docs --write-json
```

## When to use `scripts/hg`

`scripts/hg` is the repo launcher for human and automation entrypoints that need the dotfiles module surface.

Use it when you want the repo-local command surface rather than a one-off direct script call.

Examples:

```bash
bash ./scripts/hg --help
bash ./scripts/hg doctor
bash ./scripts/hg config --help
```

## Why this guide exists

The repo now spans user-scoped install flow, machine-global deployment, provider-doc parity, and MCP packaging. This guide keeps the public operational story aligned with the actual script boundaries so future automation does not infer the wrong entrypoint from outdated assumptions.
