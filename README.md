# dotfiles

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Shaders](https://img.shields.io/badge/GLSL_Shaders-139-purple)](kitty/shaders/)
[![MCP Tools](https://img.shields.io/badge/MCP_Tools-1,400+-blue)](mcp/)
[![WM](https://img.shields.io/badge/WM-Hyprland-cyan)](https://hyprland.org/)
[![CI](https://github.com/hairglasses-studio/dotfiles/actions/workflows/ci.yml/badge.svg)](https://github.com/hairglasses-studio/dotfiles/actions/workflows/ci.yml)
[![Go CI](https://github.com/hairglasses-studio/dotfiles/actions/workflows/ci-go-mcp.yml/badge.svg)](https://github.com/hairglasses-studio/dotfiles/actions/workflows/ci-go-mcp.yml)
[![Lint](https://github.com/hairglasses-studio/dotfiles/actions/workflows/ci-lint.yml/badge.svg)](https://github.com/hairglasses-studio/dotfiles/actions/workflows/ci-lint.yml)
[![Scorecard](https://api.securityscorecards.dev/projects/github.com/hairglasses-studio/dotfiles/badge)](https://securityscorecards.dev/viewer/?uri=github.com/hairglasses-studio/dotfiles)

Full-stack development environment for Manjaro Linux.

> **Engineering context:** The MCP integration (1,400+ tools across 7 Go servers) uses the same [mcpkit](https://github.com/hairglasses-studio/mcpkit) patterns as production server deployments — deferred tool loading, middleware chains, and composed operations.

Combines a wallpaper-aware `Hairglasses Neon` shell theme, Kitty-native visual rotation, declarative package management, and **1,400+ MCP tools** for desktop automation, fleet management, and AI agent infrastructure.

![Desktop — Hyprland + Quickshell menubar + tiled terminals (Hairglasses Neon)](.github/assets/desktop.png)

A QML ticker (`quickshell/modules/TickerBar.qml`) scrolls 39 live streams across the bottom of the primary monitor with a cyberpunk-effect stack — animated scanlines, ghost glitch, GPU-accelerated FBO render. Streams rotate through playlists in `ticker/content-playlists/*.txt`.

![Ticker demo — scrolling across 39 live streams (10s loop)](docs/assets/ticker-demo.gif)

### Technical Highlights

- **GPU Shaders**: 139 DarkWindow GLSL shaders paired with Kitty theme playlists for per-spawn visual rotation. The companion `kitty-playlist-validate` resolves every playlist entry against the bundled catalog with fuzzy-match suggestions so typos fail CI instead of silently skipping themes
- **Theme System**: Hairglasses Neon token pipeline for `quickshell`, `swaync`, `wofi`, and `wlogout`, with optional wallpaper-derived accent overlays via `theme-sync`. `palette-playlist list|next|random|set <name>` rotates the active palette across 9 curated envs (hairglasses-neon, amber, deep-purple, forest, ice, matrix, rose-pine, sunset, synthwave) — every palette fills the same 23 `THEME_*` tokens so templates render identically regardless of active palette
- **MCP Servers**: 2 Go modules under `mcp/` (dotfiles-mcp with ~430 tools + 25 resources, mapitall); desktop control, Bluetooth/MIDI, Kitty visual pipeline, GitHub org lifecycle, fleet auditing ([dotfiles-mcp](https://github.com/hairglasses-studio/dotfiles-mcp))
- **GitHub Stars Workflow**: taxonomy audit, GitHub list management, and Codex MCP install helpers via `scripts/hg-github-stars.sh`
- **Desktop Automation**: 19 Hyprland IPC tools, atomic config writes, compositor abstraction layer
- **Package Management**: Declarative metapac with 12 groups (paru backend)
- **Shell Framework**: Shared libraries for CLI utilities, notifications, config management, and Quickshell controls via `hg shell`; QML modules own the bar, dock, rotating ticker streams, modal menus, and notification daemon (the freedesktop.org Notifications interface — swaync stays installed as a rollback fallback only)
- **Terminal Launch Policy**: Hyprland's `$term` anchor points at `kitty-shell-launch` for plain shell windows, fresh instances, and no startup-session restore; `kitty-visual-launch` enforces unique top-level Kitty windows for raw launch surfaces; `kitty-dev-launch` remains the explicit tmux-backed dev-session entrypoint

The managed workstation alias `studio_desktop` now projects the desktop-focused `dotfiles-mcp` profile into Codex, Claude, and Gemini through the existing home-sync path.

Hyprland + Quickshell + swaync + kitty + Starship + Oh My Zsh + Neovim + tmux + btop + yazi + cava + lazygit.

## Install

```bash
git clone git@github.com:hairglasses-studio/dotfiles.git ~/hairglasses-studio/dotfiles
cd ~/hairglasses-studio/dotfiles
bash install.sh
```

The installer is idempotent — safe to run multiple times. Existing files are backed up to `~/.dotfiles-backup-*/`.

### What the installer does

1. Installs paru + metapac (declarative package management, 12 groups)
2. Installs Oh My Zsh + 5 community plugins
3. Bootstraps lazy.nvim for Neovim
4. Installs TPM (Tmux Plugin Manager)
5. Symlinks all 60+ configs to their expected locations
6. Links managed `~/.local/bin` wrappers for Kitty visuals, the shell-first `kitty-shell-launch`, the explicit tmux `kitty-dev-launch`, launcher fallback, app switcher, and the canonical Codex/Claude/Gemini launchers on Linux
7. Enables repo-managed systemd user services and packaged system services where applicable, while leaving the optional Kitty save-session timer disabled unless you explicitly opt in later
8. Syncs the shared shell theme into writable config targets and bootstraps Hyprland plugins via `hyprpm`
9. Builds bat theme cache

### Post-install

```bash
# Validate all symlinks
bash install.sh --check

# Install Neovim plugins (lazy.nvim auto-installs on first launch)
nvim --headless "+Lazy! sync" +qa

# Install tmux plugins — open tmux, press C-a then Shift-I
tmux new-session

# Sync the shared shell theme and Hyprland plugins
theme-sync
hyprpm-bootstrap

```

### Machine-specific config

- `hyprland/monitors.conf` — per-machine monitor layout and workspace orientation
- `hyprland/local.conf` — per-machine overrides (intentionally sparse)
- `git/gitconfig` — name, email
- `ssh/config` — 1Password SSH agent path

## What's Inside

| Config | Description |
|--------|-------------|
| `hyprland/` | Tiling WM — 113 keybinds, custom animations, plugin-based layout, wallpaper mode orchestration |
| `quickshell/` | QML shell — bar (every screen), bottom bar (DP-2), ticker, dock, modal menus, notification daemon, companion overlays (window label, fleet sparkline, lyrics banner) |
| `hyprdynamicmonitors/` | Dynamic monitor profiles that generate Hyprland includes into state storage |
| `hyprland-autoname-workspaces/` | Workspace naming and icon rules for cleaner shell surfaces |
| `swaync/` | Notification center + control surface themed from the shared token pipeline |
| `wofi/` | Fallback launcher/switcher/emoji styling when Quickshell menus are unavailable |
| `wlogout/` | Rollback power menu overlay aligned with the shell token system |
| `kitty/` | GPU terminal with DarkWindow shaders, theme playlists, shuffled visuals, and watcher-driven retheming |
| `zsh/` | Oh My Zsh, Starship prompt, 650+ aliases |
| `starship/` | Fill-based right alignment, git metrics, cloud context |
| `nvim/` | lazy.nvim, treesitter, LSP, telescope, Hairglasses Neon theme |
| `btop/` | System monitor with Hairglasses Neon theme |
| `yazi/` | Terminal file manager with Hairglasses Neon theme |
| `cava/` | Audio visualizer — 8-color Hairglasses Neon gradient |
| `k9s/` | Kubernetes TUI with Hairglasses Neon skin, 7 plugins, 18 aliases |
| `tmux/` | TPM, 7 plugins, vim-tmux-navigator, Hairglasses Neon status bar |
| `lazygit/` | Git TUI with Hairglasses Neon theme |
| `bat/` | Cat replacement with Hairglasses Neon syntax theme |
| `metapac/` | Declarative package management — 12 groups, paru backend |
| `topgrade/` | System update orchestration |
| `pypr/` | Hyprland scratchpads (terminal, volume, files) |
| `kanshi/` | Output profile management for Wayland |
| `glshell/` | GLSL overlay shader for terminal compositing |
| `wluma/` | Adaptive display brightness |
| `systemd/` | Repo-managed user services and timers |

### Directory layout

```
dotfiles/
├── hyprland/       → ~/.config/hypr (WM + pyprland + Hyprland companion config)
├── quickshell/     → ~/.config/quickshell (QML bar/ticker/dock/menus/notifications)
├── hyprdynamicmonitors/ → ~/.config/hyprdynamicmonitors (dynamic monitor profiles)
├── hyprland-autoname-workspaces/ → ~/.config/hyprland-autoname-workspaces
├── kitty/          → ~/.config/kitty (terminal + 139 shaders)
├── swaync/         → ~/.config/swaync (notifications fallback only — Quickshell owns the bus)
├── wofi/           → ~/.config/wofi (fallback launcher)
├── wlogout/        → ~/.config/wlogout (power menu)
├── nvim/           → ~/.config/nvim (editor)
├── bat/            → ~/.config/bat (cat replacement)
├── btop/           → ~/.config/btop (system monitor)
├── cava/           → ~/.config/cava (audio visualizer)
├── yazi/           → ~/.config/yazi (file manager)
├── k9s/            → ~/.config/k9s (kubernetes)
├── lazygit/        → ~/.config/lazygit (git TUI)
├── pypr/           → ~/.config/pypr (scratchpads)
├── kanshi/         → ~/.config/kanshi (output profile management)
├── glshell/        → GLSL overlay shader for terminal compositing
├── wluma/          → ~/.config/wluma (adaptive display brightness)
├── metapac/        → ~/.config/metapac (package groups)
├── topgrade/       → ~/.config/topgrade (system updates)
├── systemd/        → ~/.config/systemd/user/ (repo-managed user services and timers)
├── zsh/            → ~/.zshrc + ~/.zshenv
├── git/            → ~/.gitconfig + ~/.config/delta + ~/.config/git/ignore
├── tmux/           → ~/.tmux.conf
├── starship/       → ~/.config/starship.toml
├── scripts/        → 40+ utility scripts (selected launchers are linked into ~/.local/bin)
├── Pacfile         → fallback package list
├── install.sh      → symlink installer
└── Pacfile         → fallback package list for bootstrap installs
```

## MCP Servers

All MCP tools are consolidated under `mcp/` (7 Go modules via `go.work`; hg-mcp embeds an internal JS web UI but there are no standalone JS MCP servers). As of 2026-04-23, `dotfiles-mcp` alone exposes ~430 live tools + 25 resources + deferred tools; per-server totals vary and are authoritative via the runtime tool registry.

| Server | Tools | Description |
|--------|-------|-------------|
| `dotfiles-mcp` | ~430 | Desktop config management, Hyprland control, GitHub Stars taxonomy, Kitty visual pipeline, input devices, observability chassis |
| `hg-mcp` | 200+ | SDLC ops, fleet management, repo analysis, prompt pipeline |
| `systemd-mcp` | 10 | Systemd unit management |
| `tmux-mcp` | 11 | Tmux session management |
| `process-mcp` | 8 | Process debugging and port investigation |
| `mapitall` | 30+ | Controller/MIDI mapping engine and managed profile catalog |

All servers are built on [mcpkit](https://github.com/hairglasses-studio/mcpkit) and use stdio transport.
Mirrored MCP modules and the parity contract are tracked in [docs/MCP-MIRROR-PARITY.md](docs/MCP-MIRROR-PARITY.md).

### Install MCP Server Only

```bash
go install github.com/hairglasses-studio/dotfiles-mcp@latest
```

Add to `.mcp.json`:
```json
{
  "mcpServers": {
    "dotfiles": {
      "command": "dotfiles-mcp",
      "env": { "WAYLAND_DISPLAY": "wayland-1", "DOTFILES_DIR": "/path/to/dotfiles" }
    }
  }
}
```

### What Can You Do With This?

The MCP tools let AI agents control your Hyprland desktop directly:

- **"Set up my dev layout"** -- `hypr_switch_workspace` to WS 6, `kitty_launch` editor + terminal, `hypr_layout_save` for recall
- **"Cycle to a different shader"** -- `shader_random` picks a DarkWindow GLSL, `shader_status` confirms the visual
- **"Take a screenshot of this bug and file an issue"** -- `hypr_screenshot_window`, `screen_ocr` for the error text, `ops_pr_create`
- **"Try 8px gaps instead of 5"** -- `hypr_set_keyword` applies instantly, `hypr_screenshot` to compare, persist when happy
- **"Search my keybinds for screenshot"** -- `keybinds_search` returns all matching binds with modifiers and descriptions

### Skills

Claude Code skills that orchestrate the MCP tools into higher-level workflows:

| Skill | Description |
|-------|-------------|
| `keybinds` | Read, search, audit, and update Hyprland keybinds with ticker integration |
| `hypr_layouts` | Orchestrate multi-window layouts via live IPC dispatch and kitty spawning |
| `kitty_control` | Manage kitty sessions -- tabs, windows, themes, fonts, scrollback |
| `rice_iteration` | Screenshot-analyze-edit-reload-verify visual feedback loop |
| `screenshot_workflow` | Capture, OCR, annotate, and integrate screenshots into dev workflows |
| `hypr_config_tuning` | Live `hyprctl keyword` tuning before persisting to config |
| `github_stars_audit` | Audit GitHub stars for tools, cross-reference, produce implementation plans |
| `dotfiles_audit` | Comprehensive repo health and stale-reference audit |
| `dotfiles_ops` | Workstation operations, repo-tooling, onboarding, fleet maintenance |
| `dotfiles_ui` | Desktop UI, rice, shader, Ironbar, Hyprland, and screenshot workflows |
| `dotfiles_desktop_control` | Hyprland targeting, OCR-assisted inspection, input automation |
| `dotfiles_recovery` | Claude session recovery, forensic analysis, handoff artifacts |
| `dotfiles_git_hygiene` | Dry-run-first branch, worktree, and managed cleanup |
| `github_stars` | Organize starred repos into GitHub lists |

## Troubleshooting

**Shaders don't animate:** Check shader configuration in kitty config and verify DarkWindow shader pipeline.

**Symlinks point to wrong place:** Run `bash install.sh --check` to validate.

**tmux plugins not loading:** Inside tmux, press `C-a` then `shift-I` to trigger TPM install.

**New terminals still attach to tmux or reopen an old session:** Check that `hyprland/hyprland.conf` still points its `$term` anchor at `kitty-shell-launch`, `kitty/kitty.conf` still sets `startup_session none`, and any local save-session timer remains an explicit opt-in. Reserve `kitty-dev-launch` for the persistent tmux session on purpose.

**Hyprland plugins not working:** Re-run the repo-managed bootstrap:
```bash
hyprpm-bootstrap
```

## Font

[Maple Mono NF CN](https://github.com/subframe7536/maple-font) drives shell UI and window chrome. [Monaspace](https://monaspace.githubnext.com/) drives terminal/code surfaces. Install via `sudo pacman -S maplemono-nf-cn otf-monaspace otf-monaspace-nerd`.

## License

MIT
