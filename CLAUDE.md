# Dotfiles

macOS development environment managed with symlinks. Config files live here and are symlinked into `~/.config/` and `~/Library/Application Support/` by `install.sh`.

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
- **AeroSpace** — i3-style tiling (`aerospace/aerospace.toml`)
- **SketchyBar** — status bar (40px top gap reserved)
- **JankyBorders** — window borders

## Key Patterns

### Ghostty Config Updates
Ghostty auto-reloads via FSEvents. Scripts use atomic writes (`mktemp + mv`) to avoid partial reads:
```bash
tmp="$(mktemp "${CONFIG}.XXXXXX")"
sed -e "s|^custom-shader = .*|custom-shader = $path|" "$CONFIG" > "$tmp"
mv -f "$tmp" "$CONFIG"
```

### Tattoy Config Toggling
Tattoy watches its config with 100ms debounce. Section-scoped sed to avoid hitting `[text_contrast].enabled`:
```bash
sed -e "/^\[shader\]/,/^\[/ s|^enabled = .*|enabled = true|"
```

### AeroSpace Float Rules
No "ignore app" mode exists. Use `[[on-window-detected]]` with `layout floating`. Apps with no bundle ID (like glslViewer) must be matched by `if.app-name-regex-substring`.

### Shared Libraries
- **`scripts/lib/compositor.sh`** — Compositor detection & IPC abstraction. Detects AeroSpace (macOS) via `uname -s`.
- **`scripts/lib/config.sh`** — Atomic config operations (`config_atomic_write`, `config_sed_replace`, `config_backup`).

### Keyboard Firmware
- **`keyboard/`** — Drop CTRL v1/v2 + Keychron V1 Ultra VIA keymaps
- PrtSc/ScrLk/Pause trio remapped to F13/Enter/F14 for window focus control

### MCP Servers
- **`dotfiles-mcp`** — 4 tools: list configs, validate TOML/JSON, reload service, check symlinks
- **`shader-mcp`** — 5 tools: list/set/random/test/get_state for 132 shaders

### Claude Code Skills & Agents
- **Skills:** `/shader-browse` (shader explorer)
- **Agents:** `config-validator` (fast syntax check, haiku)
- **Rules:** `snazzy-palette.md` (color enforcement), `shaders.md` (self-contained GLSL)

## Aliases
```
shader-meta, shader-build, shader-test, shader-cycle, shader-bench
peek (peekaboo screen capture)
crt-on, crt-off, crt-toggle (RetroVisor)
viz (cava audio visualizer)
hack, dashboard, scan, deploy, briefing (cyberpunk functions)
cc, ccr, ccc (Claude Code)
```

## Testing
```bash
shader-test                    # compile all 132 shaders via glslangValidator
shader-meta validate           # manifest <-> .glsl file consistency
shader-build --check           # preprocessor dry-run
```
