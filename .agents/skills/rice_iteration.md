---
name: rice_iteration
description: Visual feedback loop for desktop rice iteration — screenshot, analyze, change config, reload, verify. Uses MCP screenshot + OCR + reload tools.
triggers:
  - rice
  - rice iteration
  - visual change
  - desktop appearance
  - theme change
  - bar style
  - notification style
  - look and feel
  - visual feedback
---

# Rice Iteration Loop

A structured visual feedback loop for iterating on desktop appearance: screenshot the current state, analyze what needs changing, edit config, reload the affected service, and verify with another screenshot. No public Claude Code skill exists for this workflow.

## The Loop

```
1. CAPTURE  → screenshot current state
2. ANALYZE  → identify what to change (colors, spacing, fonts, opacity, etc.)
3. EDIT     → modify the config file
4. RELOAD   → restart only the affected service
5. VERIFY   → screenshot again and compare
6. REPEAT   → until satisfied
```

## MCP Tools by Phase

### 1. CAPTURE
- `hypr_screenshot` — full monitor or all monitors
- `hypr_screenshot_window` — single window by class (resized for LLM vision)
- `hypr_screenshot_region` — interactive region selection
- `hypr_screenshot_monitors` — separate per-monitor captures
- `screen_screenshot` — alternative full capture
- `screen_screenshot_annotated` — capture with UI element annotations

### 2. ANALYZE
- `screen_ocr` — extract text from a screenshot region (useful for checking font rendering, bar content)
- `dotfiles_rice_check` — compositor/shader/wallpaper/service status + palette compliance check
- `shader_status` — current shader, theme, visual label, playlist position
- `dotfiles_desktop_status` — full desktop control readiness

### 3. EDIT
Direct file editing with the Edit tool. Key config files by component:

| Component | Config path | Format |
|-----------|------------|--------|
| Hyprland | `hyprland/hyprland.conf` | Hyprland config |
| Ironbar | `ironbar/config.toml` + `ironbar/style.css` | TOML + CSS |
| swaync | `swaync/config.json` + `swaync/style.css` | JSON + CSS |
| kitty | `kitty/kitty.conf` | kitty config |
| wofi | `wofi/config` + `wofi/style.css` | INI + CSS |
| wlogout | `wlogout/layout` + `wlogout/style.css` | custom + CSS |
| starship | `starship/starship.toml` | TOML |
| btop | `btop/btop.conf` | custom |
| yazi | `yazi/theme.toml` | TOML |
| GTK | `gtk/settings.ini` | INI |
| Hyprshade | `hyprland/hyprshade.toml` | TOML |

### 4. RELOAD
Use the narrowest reload possible — cascade only when multiple services are affected:

| Scope | Tool | What it does |
|-------|------|------------|
| Single service | `dotfiles_reload_service` | Restart one service (ironbar, swaync, hyprland, tmux) |
| Hyprland config | `hypr_reload_config` | Hot-reload hyprland.conf only |
| kitty config | `kitty_load_config` | Hot-reload kitty.conf |
| Full stack | `dotfiles_cascade_reload` | Ordered multi-service reload with health verification |

### 5. VERIFY
Re-run CAPTURE tools and compare visually. Also use:
- `dotfiles_rice_check` — confirms palette compliance after changes
- `keybinds_list` — if keybinds were modified, verify they registered

## Palette Enforcement

All visual changes must use the Hairglasses Neon palette. Source of truth: `theme/palette.env`.

| Token | Hex | Usage |
|-------|-----|-------|
| `THEME_PRIMARY` | `#29f0ff` | Cyan — borders, highlights, active indicators |
| `THEME_SECONDARY` | `#ff47d1` | Magenta — accents, keybind display, urgent |
| `THEME_TERTIARY` | `#3dffb5` | Green — success, healthy, active |
| `THEME_WARNING` | `#ffe45e` | Yellow — warnings, pending |
| `THEME_DANGER` | `#ff5c8a` | Red — errors, critical |
| `THEME_BG` | `#05070d` | Near-black background |
| `THEME_SURFACE` | `#0f1219` | Panel/card background |
| `THEME_FG` | `#f7fbff` | Primary text |
| `THEME_MUTED` | `#66708f` | Disabled/secondary text |
| `THEME_BORDER` | `#2a3246` | Inactive borders |

Use `dotfiles://palette` MCP resource to get the full live palette as structured JSON.

## Component-Specific Tips

### Ironbar CSS
- Inspector: `GTK_DEBUG=interactive ironbar` (but requires restart)
- Classes follow `.workspaces`, `.clock`, `.tray`, `.custom` pattern
- Font sizes should stay proportional to Maple Mono NF CN 11

### Hyprland decoration
- `rounding`, `gaps_in`, `gaps_out`, `border_size` in `general {}` and `decoration {}`
- Active border uses gradient: `rgba(29f0ffee) rgba(ff47d1ee) 45deg`
- Shadow: `rgba(05070d88)` matches the BG color with alpha
- Test changes with `hyprctl keyword` before committing to config

### kitty shaders
- DarkWindow shaders in `kitty/shaders/darkwindow/` — 139 GLSL files
- Current shader visible via `shader_status`
- `shader_set` applies immediately without kitty restart
- Shaders are pre-validated: `shader_test` runs glslangValidator

### swaync notifications
- Style uses the same palette via CSS custom properties
- Test with `notify-send "Test" "Body text"` then screenshot
- DND toggle: `swaync-client -d`

## Anti-Patterns

- Do NOT `cascade_reload` for a single CSS change — use the targeted reload
- Do NOT guess at pixel values — screenshot first, measure in the image
- Do NOT introduce colors outside the palette without explicit user request
- Do NOT change font families — Maple Mono NF CN is standardized
- Do NOT edit generated files (`.claude/skills/` is generated from `.agents/skills/`)
