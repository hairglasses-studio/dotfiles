# Dotfiles

Cross-platform development environment (macOS + Manjaro Linux) managed with symlinks. Config files live here and are symlinked into `~/.config/` and `~/Library/Application Support/` by `install.sh`.

## Architecture

### Shader Pipeline
Ghostty transpiles GLSL -> SPIRV -> Metal at runtime. No `#include` support — each `.glsl` must be self-contained.

- **`ghostty/shaders/`** — 132+ GLSL shaders
- **`ghostty/shaders/shaders.toml`** — Central manifest (single source of truth for shader metadata)
- **`ghostty/shaders/lib/`** — Shared GLSL libraries (inlined by preprocessor)
- **`ghostty/shaders/bin/`** — Management scripts:
  - `shader-meta.sh` — Query/validate the manifest
  - `shader-build.sh` — Preprocessor: inlines `// #include "lib/X.glsl"` with `BEGIN/END` markers (idempotent)
  - `shader-test.sh` — Compilation testing via glslangValidator
  - `shader-cycle.sh` — Curated shader rotation + Tattoy mode switching
  - `shader-random.sh` — Random shader selection
  - `shader-benchmark.sh` — Performance profiling
  - `shader-playlist.sh` — Fisher-Yates shuffled playlist engine
  - `shader-auto-rotate.sh` — Timed rotation via launchd

### Config Symlinks
```
~/.config/ghostty    -> dotfiles/ghostty
~/.config/aerospace  -> dotfiles/aerospace
~/Library/Application Support/tattoy/tattoy.toml  (copied, not symlinked)
~/Library/Application Support/tattoy/shaders      -> dotfiles/ghostty/shaders
```

### Visual Stack (layered, bottom to top)
1. **Ghostty** — terminal with `custom-shader` (GPU-rendered at pixel resolution)
2. **Tattoy** — terminal compositor with shader + animated cursor layers (text-resolution via half-blocks)
3. **RetroVisor** — CRT overlay via ScreenCaptureKit + Metal (screen-level)

### Window Management
- **macOS:** AeroSpace tiling + SketchyBar + JankyBorders
- **Linux:** Hyprland (primary) or Sway (fallback) + eww bar + mako notifications + wofi launcher + wlogout

### Shared Libraries
- **`scripts/lib/compositor.sh`** — Compositor detection & IPC abstraction. Functions: `compositor_type`, `compositor_msg`, `compositor_query`, `compositor_output`, `compositor_subscribe`, `compositor_reload`, `compositor_workspace`. Detects Hyprland/Sway/AeroSpace via env vars.
- **`scripts/lib/config.sh`** — Atomic config operations. Functions: `config_atomic_write`, `config_sed_replace`, `config_backup`, `config_reload_service`. All scripts that modify configs should source this.

### Claude Code Integration
- **PostToolUse hook** — Auto-reloads Hyprland/mako/eww/waybar when Claude writes config files
- **MCP servers** — `sway-mcp` for Wayland desktop control (screenshot, input, windows). Future: `hyprland-mcp`, `dotfiles-mcp`, `shader-mcp` via mcpkit Go toolkit.

### Wallpaper Shaders
Live animated wallpapers via Shadertoy-compatible GLSL rendered by `shaderbg`:
- **`wallpaper-shaders/`** — 5 procgen GLSL fragments (cyber-rain, neon-grid, plasma-flow, fractal-pulse, particle-aurora)
- **`scripts/shader-wallpaper.sh`** — Rotation engine: `shader-wallpaper.sh [next|random|set <path>|stop|list|static]`
- Keybinds: `$mod+Shift+W` (next), `$mod+Shift+Ctrl+W` (random), `$mod+Shift+Alt+W` (static fallback)

### Boot Stack
- **rEFInd** — UEFI bootloader with Matrix cyberpunk theme (config tracked in `refind/`, deployed to `/boot/efi/EFI/refind/` via copy)
- **Plymouth** — Animated boot splash (proxzima cyberpunk theme) between rEFInd and login
- **Kernel params:** `quiet splash loglevel=3 nvidia_drm.modeset=1 nvidia.NVreg_PreserveVideoMemoryAllocations=1`

### MCP Servers (Go binaries, built separately)
- **`hyprland-mcp`** — 9 tools: screenshot, windows, workspaces, input (ydotool/wtype), config reload
- **`dotfiles-mcp`** — 4 tools: list configs, validate TOML/JSON, reload service, check symlinks
- **`shader-mcp`** — 5 tools: list/set/random/test/get_state for 132 shaders
- **`sway-mcp`** — Node.js: screenshot, windows, input, clipboard for Sway/Wayland

### Claude Code Skills & Agents
- **Skills:** `/rice-check` (validate rice), `/screenshot-review` (visual analysis), `/shader-browse` (shader explorer)
- **Agents:** `rice-developer` (autonomous cyberpunk rice iteration, opus), `config-validator` (fast syntax check, haiku)
- **Rules:** `hyprland.md` (0.54 block windowrule syntax), `shaders.md` (self-contained GLSL), `snazzy-palette.md` (color enforcement)

## Key Patterns

### Ghostty Config Updates
Ghostty auto-reloads via FSEvents. Scripts use atomic writes (`mktemp + mv`) to avoid partial reads:
```bash
tmp="$(mktemp "${CONFIG}.XXXXXX")"
sed -e "s|^custom-shader = .*|custom-shader = $path|" "$CONFIG" > "$tmp"
mv -f "$tmp" "$CONFIG"
```

### Tattoy Config Toggling
Tattoo watches its config with 100ms debounce. Section-scoped sed to avoid hitting `[text_contrast].enabled`:
```bash
sed -e "/^\[shader\]/,/^\[/ s|^enabled = .*|enabled = true|"
```

### AeroSpace Float Rules
No "ignore app" mode exists. Use `[[on-window-detected]]` with `layout floating`. Apps with no bundle ID (like glslViewer) must be matched by `if.app-name-regex-substring`.

## Aliases
```
shader-meta, shader-build, shader-test, shader-cycle, shader-bench
peek (peekaboo screen capture)
crt-on, crt-off, crt-toggle (RetroVisor)
```

## Testing
```bash
shader-test                    # compile all 132 shaders via glslangValidator
shader-meta validate           # manifest <-> .glsl file consistency
shader-build --check           # preprocessor dry-run
```
