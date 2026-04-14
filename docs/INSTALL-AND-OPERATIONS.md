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

Current `--check` contract:

- accepts the current repo checkout and the primary managed checkout when the check is run from another worktree of the same git repo
- distinguishes repo-owned symlinks such as `~/.config/ironbar` from writable config roots such as `~/.config/hyprshell`
- reports generated runtime state such as `~/.local/state/hypr/monitors.dynamic.conf` and `~/.local/state/kitty/sessions` separately from broken link targets

What it does not do:

- deploy `/etc` configuration directly
- replace the standalone system deploy helpers
- manage hosted workflow sync

## Machine-global deployment helpers

Use the dedicated deploy scripts when the target is machine-global instead of user-scoped.

- `scripts/etc-deploy.sh`: tracked `/etc` config deployment
- `scripts/greetd-deploy.sh`: greetd deployment path
- `scripts/juhradial-install.sh`: juhradial-mx install/build/deploy helper for the MX Master 4 stack
- `scripts/refind-deploy.sh`: rEFInd deployment
- `scripts/plymouth-deploy.sh`: Plymouth deployment

These are the right entrypoints when the change belongs under `/etc`, bootloader state, or another machine-global surface.

For the MX Master 4 stack specifically:

- `scripts/hg input status`: fast runtime status for juhradial, ydotool, transport, and config drift
- `scripts/hg input verify`: live runtime + patch-layer verification workflow
- `scripts/juhradial-wheel-apply.sh`: recovery-only wheel-state bridge; Solaar is intentionally limited to this repair path until juhradial restores the full HID++ wheel state itself

## Documentation parity helpers

Use the repo scripts instead of editing compatibility mirrors by hand.

- `scripts/hg-agent-docs.sh --source auto .`: regenerate `CLAUDE.md`, `GEMINI.md`, and `.github/copilot-instructions.md` from the canonical source file
- `codexkit skills sync|check|diff <repo_path>`: primary skill-surface engine for `.agents/skills/` to `.claude/skills/` and plugin mirrors
- `scripts/hg-skill-surface-sync.sh <repo_path>`: dotfiles convenience wrapper around the `codexkit skills` engine
- `scripts/hg-agent-home-sync.sh`: seed missing `/home/hg` and `/root` provider home docs, then align managed skill mirrors and workspace-global overlays

The rule is simple:

- `AGENTS.md` is canonical when marked canonical.
- `CLAUDE.md`, `GEMINI.md`, and Copilot instructions are mirrors.
- `.agents/skills/` is canonical; `.claude/skills/` is generated compatibility output.
- If a Claude command and Claude skill share a name, the full skill definition wins; global home sync does not overwrite canonical skill mirrors with the command shim.

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

## GitHub Stars Workflow

Use the repo-managed GitHub Stars entrypoints when the task is about starred repositories, GitHub star folders/lists, or installing the personal GitHub Stars MCP surface.

- `scripts/hg-github-stars.sh`: shell wrapper for list inspection, taxonomy audit/sync, cleanup candidates, bootstrap, and Codex MCP install
- `scripts/hg-github-official-mcp.sh`: wrapper for the official `github/github-mcp-server` image with token resolution that prefers `GITHUB_PAT` from `~/.env`

Typical operator checks:

```bash
bash ./scripts/hg-github-stars.sh summary --managed-prefix 'MCP / '
bash ./scripts/hg-github-stars.sh list-lists --include-items
bash ./scripts/hg-github-stars.sh audit-taxonomy --managed-prefix 'MCP / ' --bootstrap-defaults
```

Typical first-time setup:

```bash
bash ./scripts/hg-github-stars.sh bootstrap --install-codex-mcp --execute
```

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

### MX input stack

```bash
bash ./scripts/hg input status
bash ./scripts/hg input verify
```

## When to use `scripts/hg`

`scripts/hg` is the repo launcher for human and automation entrypoints that need the dotfiles module surface.

Use it when you want the repo-local command surface rather than a one-off direct script call.

Examples:

```bash
bash ./scripts/hg --help
bash ./scripts/hg doctor
bash ./scripts/hg config --help
bash ./scripts/hg desktop status
bash ./scripts/hg desktop preset save dual-dock
bash ./scripts/hg desktop layout save code-review --launch kitty='kitty --class kitty'
bash ./scripts/hg desktop project open ~/hairglasses-studio/dotfiles --monitor dual-dock --layout code-review --dry-run
```

## Why this guide exists

The repo now spans user-scoped install flow, machine-global deployment, provider-doc parity, and MCP packaging. This guide keeps the public operational story aligned with the actual script boundaries so future automation does not infer the wrong entrypoint from outdated assumptions.
