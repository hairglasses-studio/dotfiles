# Dotfiles

This repo uses [AGENTS.md](AGENTS.md) as the canonical instruction file.

Cross-platform development environment (macOS + Manjaro Linux) managed with symlinks. Config files live here and are symlinked into `~/.config/` and `~/Library/Application Support/` by `install.sh`.

## Architecture

### Shader Pipeline
Source GLSL shaders live in `ghostty/shaders/` (legacy path, kept as the canonical source directory). Shaders are transpiled to CRTty (GLSL 330 core) and Hypr-DarkWindow formats for kitty. No `#include` support — each `.glsl` must be self-contained.

- **`ghostty/shaders/`** — 138 GLSL shaders (source directory)
- **`ghostty/shaders/shaders.toml`** — Central manifest (single source of truth for shader metadata)
- **`ghostty/shaders/lib/`** — Shared GLSL libraries (inlined by preprocessor)
- **`ghostty/shaders/bin/`** — Management scripts:
  - `shader-meta.sh` — Query/validate the manifest
  - `shader-build.sh` — Preprocessor: inlines `// #include "lib/X.glsl"` with `BEGIN/END` markers (idempotent)
  - `shader-test.sh` — Compilation testing via glslangValidator
  - `shader-cycle.sh` — Curated shader rotation
  - `shader-random.sh` — Random shader selection
  - `shader-benchmark.sh` — Performance profiling
  - `shader-playlist.sh` — Fisher-Yates shuffled playlist engine
  - `shader-auto-rotate.sh` — Timed rotation

### Config Symlinks
```
~/.config/kitty      -> dotfiles/kitty
~/.config/aerospace  -> dotfiles/aerospace
```

### Visual Stack (layered, bottom to top)
1. **kitty** — terminal with CRTty/Hypr-DarkWindow shaders (GPU-rendered)

### Package Management
- **macOS:** Homebrew via `Brewfile`
- **Linux:** metapac (declarative, paru backend) with 12 groups in `metapac/groups/`. Run `metapac sync` to install all packages, or edit group TOMLs to add/remove. Falls back to `Pacfile` if metapac is not installed.

### Window Management
- **macOS:** AeroSpace tiling + SketchyBar + JankyBorders
- **Linux:** Hyprland + eww bar + swaync notifications + wofi launcher + wlogout

### Shared Libraries
- **`scripts/lib/compositor.sh`** — Compositor detection & IPC abstraction. Functions: `compositor_type`, `compositor_msg`, `compositor_query`, `compositor_output`, `compositor_subscribe`, `compositor_reload`, `compositor_workspace`. Detects Hyprland/AeroSpace via env vars.
- **`scripts/lib/config.sh`** — Atomic config operations. Functions: `config_atomic_write`, `config_sed_replace`, `config_backup`, `config_reload_service`. All scripts that modify configs should source this.

### Claude Code Integration
- **PostToolUse hook** — Auto-reloads Hyprland/swaync/eww when Claude writes config files
- **MCP servers** — `dotfiles-mcp` (82 tools, consolidated from 4 servers), `systemd-mcp`, `tmux-mcp`, `process-mcp` — Go binaries in sibling repos, registered in `.mcp.json`.

### Wallpaper Shaders
Live animated wallpapers via Shadertoy-compatible GLSL rendered by `shaderbg`:
- **`wallpaper-shaders/`** — 5 procgen GLSL fragments (cyber-rain, neon-grid, plasma-flow, fractal-pulse, particle-aurora)
- **`scripts/shader-wallpaper.sh`** — Rotation engine: `shader-wallpaper.sh [next|random|set <path>|stop|list|static]`
- Keybinds: `$mod+Shift+W` (next), `$mod+Shift+Ctrl+W` (random), `$mod+Shift+Alt+W` (static fallback)

### Boot Stack
- **rEFInd** — UEFI bootloader with Matrix cyberpunk theme (config tracked in `refind/`, deployed to `/boot/efi/EFI/refind/` via copy)
- **Plymouth** — Animated boot splash (proxzima cyberpunk theme) between rEFInd and login
- **Kernel params:** `quiet splash loglevel=3 nvidia_drm.modeset=1 nvidia.NVreg_PreserveVideoMemoryAllocations=1`

### MCP Server (unified dotfiles-mcp)
- **`dotfiles-mcp`** — 82 tools across 8 modules: dotfiles config management (30), Hyprland desktop control (12), input devices/BT/controllers/MIDI (26), shader pipeline (14). Single Go binary, stdio transport.
- **`systemd-mcp`** — systemd unit management
- **`tmux-mcp`** — tmux session management
- **`process-mcp`** — process management

### Claude Code Skills & Agents
- **Skills:** `/rice-check` (validate rice), `/screenshot-review` (visual analysis), `/shader-browse` (shader explorer)
- **Agents:** `rice-developer` (autonomous cyberpunk rice iteration, opus), `config-validator` (fast syntax check, haiku)
- **Rules:** `hyprland.md` (0.54 block windowrule syntax), `shaders.md` (self-contained GLSL), `snazzy-palette.md` (color enforcement)

## Key Patterns

### kitty Shader Rotation
Shader rotation is managed by `kitty-shader-playlist.sh`. kitty reloads config on `SIGUSR1` or via `kitty @ set-colors` remote control.

### Hyprland Plugin Reset
If plugins stop working (dispatchers return "Invalid dispatcher" in log), do a clean cycle:
```bash
# Unload ALL plugins
for so in /var/cache/hyprpm/hg/*/*.so; do hyprctl plugin unload "$so"; done
sleep 2
# Reload from clean state
hyprpm reload -n
```
Do NOT run partial `hyprpm enable/disable` during a session — it corrupts the dispatcher registry. The `exec-once = hyprpm reload -n` in hyprland.conf handles boot-time loading.

### AeroSpace Float Rules
No "ignore app" mode exists. Use `[[on-window-detected]]` with `layout floating`. Apps with no bundle ID (like glslViewer) must be matched by `if.app-name-regex-substring`.

## Shared Libraries

All standalone scripts should `set -euo pipefail` and source the appropriate libraries.

| Library | Functions | Purpose |
|---------|-----------|---------|
| `scripts/lib/hg-core.sh` | `hg_info`, `hg_ok`, `hg_error`, `hg_die`, `hg_require`, Snazzy colors | CLI framework for hg-* scripts |
| `scripts/lib/compositor.sh` | `compositor_type`, `compositor_msg`, `compositor_query`, `compositor_reload`, `compositor_subscribe` | Cross-compositor IPC (Hyprland/AeroSpace) |
| `scripts/lib/config.sh` | `config_atomic_write`, `config_sed_replace`, `config_backup`, `config_reload_service`, `config_reload_parallel` | Atomic config writes with `mktemp + mv`, parallel service reloads |
| `scripts/lib/notify.sh` | `hg_notify_low`, `hg_notify_normal`, `hg_notify_critical` | Desktop notifications via notify-send |
| `scripts/lib/kitty-config.sh` | `kitty_get_shader_path`, `kitty_get_shader_name`, `kitty_get_shader_animation` | Shared kitty config queries (eliminates inline grep/sed) |
| `scripts/lib/agent-post-tool-audit.sh` | PostToolUse hook | Reloads services on config edits, validates hyprland/eww/systemd errors, checks metapac coverage, enforces Snazzy palette |
| `scripts/lib/agent-pre-tool-validate.sh` | PreToolUse hook | Validates .yuck paren balance and .scss syntax before writes |
| `scripts/lib/claude-post-tool-reload.sh` | Compatibility shim | Delegates to `agent-post-tool-reload.sh` for legacy Claude hook configs |
| `scripts/lib/claude-pre-tool-validate.sh` | Compatibility shim | Delegates to `agent-pre-tool-validate.sh` for legacy Claude hook configs |
| `scripts/lib/prompt-capture.sh` | UserPromptSubmit hook | Captures multi-line prompts to docs/prompts/ with TOML frontmatter + SQLite indexing |

## Scripts Reference

### Desktop/WM
| Script | Description |
|--------|-------------|
| `hypr-keybinds.sh` | Generate keybind reference from live Hyprland config |
| `hypr-boot-log.sh` | Capture Hyprland boot errors for post-boot review |
| `hypr-bt-boot.sh` | Boot-time bluetooth device connection |
| `dropdown-terminal.sh` | Yakuake-style toggle for ralphglasses + claude code |
| `app-switcher.sh` | Wofi-based window switcher |

### Eww Widgets
| Script | Description |
|--------|-------------|
| `eww-workspaces.sh` | Workspace listener for eww bar |
| `eww-activewindow.sh` | Active window title listener |
| `eww-mode.sh` | Compositor submap/mode listener |
| `eww-volume.sh` | Event-driven volume daemon |
| `eww-cava.sh` | Audio visualization streamer |
| `eww-calendar.sh` | Calendar grid JSON generator |
| `eww-calendar-sync.sh` | Google Calendar sync |
| `eww-events.sh` | Upcoming events for sidebar |
| `eww-weather.sh` | Weather data via wttr.in |
| `eww-updates.sh` | System update checker |
| `eww-theme-gen.sh` | Color overrides from wallpaper via matugen |

### Fleet/Repo Management
| Script | Description |
|--------|-------------|
| `hg-pipeline.sh` | Build+test pipeline (Go/Node/Python) |
| `hg-health.sh` | Org-wide repo health dashboard |
| `hg-fleet-health.sh` | Fleet status (CI, commits, tests) |
| `hg-go-sync.sh` | Sync Go version across repos |
| `hg-dep-audit.sh` | Dependency version skew audit |
| `hg-new-repo.sh` | Scaffold new repo with standard files |
| `hg-workflow-sync.sh` | Sync CI workflows across repos |
| `hg-onboard-repo.sh` | Onboard repo with standard config |
| `hg-agent-docs.sh` | Generate compatibility docs from canonical AGENTS.md or legacy CLAUDE.md |
| `hg-codex-audit.sh` | Repo inventory and parity audit |

### System/Boot
| Script | Description |
|--------|-------------|
| `cyberboot.sh` | Cyberpunk terminal boot sequence (sourced) |
| `refind-deploy.sh` | Full rEFInd deployment |
| `refind-kernel-sync.sh` | Validate loader paths after kernel changes |
| `refind-boot-guard.sh` | Restore rEFInd as first boot entry |
| `plymouth-deploy.sh` | Install Cybernet Plymouth theme |
| `greetd-deploy.sh` | Deploy greetd config |
| `etc-deploy.sh` | Deploy tracked /etc/ configs |

### Utilities
| Script | Description |
|--------|-------------|
| `shader-wallpaper.sh` | Procgen shader wallpaper engine |
| `wallpaper-cycle.sh` | Animated wallpaper rotation via swww |
| `screenshot-crop.sh` | Crop-select screenshot to clipboard |
| `mx-battery.sh` | MX Master 4 battery for status bar |
| `mx-battery-notify.sh` | Low battery desktop notification |
| `agent-session-picker.sh` | Focus active agent session via wofi |
| `ccg.sh` | Global Claude Code session browser (FZF picker, preview, resume across all repos) |
| `vlm-analyze.sh` | Screenshot analysis via Claude vision |

## Aliases
```
shader-meta, shader-build, shader-test, shader-cycle, shader-bench
peek (peekaboo screen capture)
ccg (global Claude Code session browser — browse/resume sessions across all repos)
```

## Testing
```bash
shader-test                    # compile all 138 shaders via glslangValidator
shader-meta validate           # manifest <-> .glsl file consistency
shader-build --check           # preprocessor dry-run
```
