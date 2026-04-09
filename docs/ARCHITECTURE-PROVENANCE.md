# Architecture & Provenance

This note is the fast orientation map for the three surfaces that matter most in `dotfiles`: installer/bootstrap flow, workstation runtime config, and the bundled `mcp/` subtree.

## Repo Split

### 1. Installer and deployment surface

These files decide how the repo is installed, linked, and verified on a machine.

- `install.sh`: idempotent entrypoint for user-scoped symlinks, package bootstrap, and platform-aware setup.
- `dotfiles.toml`: feature flags and profile switches that shape what gets enabled.
- `Brewfile`, `Pacfile`, `metapac/`, `paru/`, `topgrade/`: package and update surfaces.
- `etc/`: repo-tracked system config intended to deploy into `/etc`.
- `systemd/`: repo-managed user services linked into `~/.config/systemd/user/`.
- `scripts/etc-deploy.sh`, `scripts/greetd-deploy.sh`, `scripts/logiops-deploy.sh`, `scripts/refind-deploy.sh`, `scripts/plymouth-deploy.sh`: privileged or machine-scoped deployment helpers.

### 2. Workstation runtime config

These directories are the day-to-day desktop and shell runtime state managed by the installer.

- Desktop control plane: `hyprland/`, `hyprshell/`, `hypr-dock/`, `hyprdynamicmonitors/`, `hyprland-autoname-workspaces/`, `eww/`, `swaync/`, `wofi/`, `wlogout/`, `greetd/`.
- Terminal and shell stack: `kitty/`, `ghostty/`, `tmux/`, `zsh/`, `starship/`, `nvim/`.
- TUI theming and utilities: `bat/`, `btop/`, `cava/`, `glow/`, `k9s/`, `lazygit/`, `yazi/`, `fastfetch/`, `gh/`.
- Device or hardware config: `keyboard/`, `makima/`, `logiops/`, `solaar/`, `environment.d/`, `udev/`.
- Visual and boot stack: `wallpaper-shaders/`, `refind/`, `plymouth/`.

### 3. Automation and MCP surface

These paths are the reusable operator and tooling layer that other repos consume or mirror.

- `scripts/` and `scripts/lib/`: canonical shell automation for repo setup, parity sync, workflow sync, docs generation, health checks, and workstation utilities.
- `.agents/skills/`: canonical workflow skill surface for this repo.
- `.codex/agents/`: Codex delegation-role surface.
- `mcp/`: bundled Go and JS MCP modules managed as a workspace, with this repo acting as an integration surface rather than only a rice repo.

## Provenance Rules

- `AGENTS.md` is the canonical instruction file. `CLAUDE.md`, `GEMINI.md`, and `.github/copilot-instructions.md` are compatibility mirrors.
- `.agents/skills/` is the source of truth for reusable workflow skills. `.claude/skills/` is generated compatibility output and should be refreshed by `scripts/hg-skill-surface-sync.sh`.
- `scripts/` and `scripts/lib/` should stay ahead of one-off shell fragments. If a flow is repeated, the reusable script is the canonical source.
- `systemd/` is user-scoped unless a deployment helper explicitly installs into `/etc` or another machine-global location.
- `mcp/` contains bundled modules that may also exist as standalone publish mirrors. Treat mirror parity as an explicit maintenance task, not an assumption.

## Ownership Heuristics

Use these heuristics when deciding where a change belongs.

- If it changes bootstrap, linking, or deployment semantics, it belongs in `install.sh`, `dotfiles.toml`, `etc/`, `systemd/`, or a deploy helper.
- If it changes how the desktop or shell behaves after install, it belongs in the runtime config directories.
- If it affects repeated operator flows, repo sync, provider parity, or MCP packaging, it belongs in `scripts/`, `.agents/skills/`, or `mcp/`.

## Verification Ladder

Start narrow and escalate only when the change crosses surfaces.

1. `bash -n scripts/*.sh scripts/lib/*.sh`
2. `bash ./scripts/hg-agent-docs.sh --source auto .`
3. `bash ./scripts/hg-workflow-sync.sh --dry-run`
4. `bash ./scripts/hg-codex-audit.sh --write-workspace-cache --write-wiki-docs --write-json`
5. Targeted BATS or CI smoke coverage under `tests/` and `.github/workflows/`

## Search Hints

If you are searching the repo for a change target, start with these anchors.

- Installer/bootstrap: `install.sh`, `dotfiles.toml`, `print_link_specs`, `create_symlinks`
- Workstation runtime: `hyprland/`, `eww/`, `kitty/`, `ghostty/`, `systemd/`
- Operator tooling: `scripts/hg*`, `scripts/lib/hg-*`, `.agents/skills/`
- MCP packaging and parity: `mcp/`, `docs/MCP-MIRROR-PARITY.md`, `scripts/hg-mcp-mirror-parity.sh`, `.mcp.json`, `scripts/hg-codex-mcp-sync.sh`, `scripts/hg-codex-audit.sh`

## Why this note exists

The repo is no longer just a personal dotfile dump. It is an installer, a workstation control plane, and an MCP integration surface. This note exists so future automation and cross-repo research can locate the right ownership boundary without reconstructing it from raw code every time.
