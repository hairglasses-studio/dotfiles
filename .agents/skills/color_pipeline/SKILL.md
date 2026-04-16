---
name: color_pipeline
description: Render Hairglasses Neon palette across all 12 consumers via envsubst template fan-out. Use when the palette changes, when wiring a new consumer, or when diagnosing palette drift.
allowed-tools:
  - Bash
  - Read
  - Write
  - Edit
  - Grep
  - Glob
  - mcp__dotfiles__color_pipeline_apply
  - mcp__dotfiles__dotfiles_rice_check
  - mcp__dotfiles__dotfiles_reload_service
  - mcp__dotfiles__dotfiles_cascade_reload
  - mcp__dotfiles__hypr_reload_config
  - mcp__dotfiles__hypr_screenshot
  - mcp__dotfiles__hypr_screenshot_window
  - mcp__dotfiles__screen_screenshot
  - mcp__dotfiles__kitty_load_config
  - mcp__dotfiles__wallpaper_set
  - mcp__dotfiles__wallpaper_list
  - mcp__dotfiles__dotfiles_desktop_status
---

# color_pipeline — palette render + fan-out

A single palette (`theme/palette.env`) flows through a set of envsubst templates
(`matugen/templates/*`) to every themed consumer on the stack. The pipeline is
event-driven (keybind, chezmoi hook, or manual invocation); each target has a
post-hook so reloads are narrow.

This is novel: no public Claude Code / agent skill exists for this workflow.
See `.claude/rules/snazzy-palette.md` for the palette spec and
`matugen/README.md` for template registry.

## The render pipeline

```
theme/palette.env
  │
  ▼
scripts/palette-propagate.sh [--wallpaper] [--no-reload] [--dry-run]
  │
  ├─► envsubst(templates/gtk-colors.css)         ──► ironbar/swaync/wofi/wlogout/hyprshell CSS
  ├─► envsubst(templates/kitty-colors.conf)      ──► ~/.config/kitty/cyberpunk-neon.conf  + pkill -SIGUSR1 kitty
  ├─► envsubst(templates/hyprland-colors.conf)   ──► ~/.config/hypr/colors.conf
  ├─► envsubst(templates/hyprlock-colors.conf)   ──► ~/.config/hypr/hyprlock-colors.conf
  ├─► envsubst(templates/btop-theme.theme)       ──► ~/.config/btop/themes/hairglasses-neon.theme
  ├─► envsubst(templates/yazi-theme.toml)        ──► ~/.config/yazi/theme.toml
  ├─► envsubst(templates/zsh-fzf-colors.sh)      ──► ~/.config/fzf/hg-colors.sh
  └─► envsubst(templates/cava-colors.conf)       ──► ~/.config/cava/config  + pkill -USR2 cava
```

## When to run

- After editing `theme/palette.env` (palette swap)
- After editing `matugen/templates/*` (template tweak)
- When a new consumer is added (see "Adding a consumer" below)
- When diagnosing "why did my palette change not appear in X"
- As part of `rice-reload.sh` (already integrated)

**Event-driven auto-fire**: `dotfiles-rice-watch.service` runs a persistent
`inotifywait` on `theme/palette.env` + `matugen/templates/` and fires
`palette-propagate.sh` automatically on close_write/move/create. Start with
`systemctl --user start dotfiles-rice-watch.service`.

## The workflow

### 1. Edit the palette

```bash
# Source of truth
cd ~/hairglasses-studio/dotfiles
$EDITOR theme/palette.env
```

Palette tokens are documented in `.claude/rules/snazzy-palette.md`. Only change
the HEX values inside the existing tokens; don't add new tokens without also
adding them to the templates.

### 2. Run the propagation

```bash
# Dry-run first — shows which targets would be written
scripts/palette-propagate.sh --dry-run

# Live render + fire reload hooks
scripts/palette-propagate.sh

# Wallpaper-derived accents (opt-in — overrides primary/secondary/tertiary
# with MD3 HCT-extracted colors from current swww wallpaper)
scripts/palette-propagate.sh --wallpaper

# Just render, don't reload (useful for batch testing)
scripts/palette-propagate.sh --no-reload
```

Or via the MCP tool:

```
mcp__dotfiles__color_pipeline_apply                        # fixed palette
mcp__dotfiles__color_pipeline_apply use_wallpaper=true     # matugen-derived
mcp__dotfiles__color_pipeline_apply dry_run=true           # preview only
```

### 3. Verify

```bash
# Status snapshot of all visual services
mcp__dotfiles__dotfiles_rice_check

# Screenshot target apps to eyeball the change
mcp__dotfiles__hypr_screenshot_window class=kitty
mcp__dotfiles__hypr_screenshot_window class=ironbar     # actually layer-shell; use screen_screenshot
mcp__dotfiles__screen_screenshot
```

### 4. Cascade reload if something drifted

```bash
mcp__dotfiles__dotfiles_cascade_reload   # runs rice-reload.sh under the hood
```

## Adding a consumer

1. Create `matugen/templates/<app>-colors.<ext>` with `${THEME_*}` variables
2. Register target path + post-hook in `scripts/palette-propagate.sh` (add a
   `_handle` call in the consumer block)
3. If the consumer has a config file that sources the rendered output (Hyprland
   `source = ...` pattern), add the include line to the app's main config
4. Run `scripts/palette-propagate.sh --dry-run` to confirm the target appears
5. Run live and diff against prior config to verify no unintended drift
6. Update `matugen/README.md` registry table

## Wallpaper-derived mode (optional)

When the user wants accents to adapt to their wallpaper:

```bash
swww img /path/to/new-wallpaper.jpg --transition-type fade
scripts/palette-propagate.sh --wallpaper
```

matugen extracts MD3 HCT colors; only primary/secondary/tertiary are overridden.
Neutrals (bg, fg, surface, panel, border) stay locked to palette.env so the
cyberpunk-dark aesthetic holds.

State check for wallpaper-derived mode:

```bash
ls -la ~/.local/state/swww/current     # current wallpaper path
cat ~/.cache/matugen/colors.json | jq  # last extracted palette (if wallpaper mode was used)
```

## Common failure modes

| Symptom | Cause | Fix |
|---|---|---|
| "Missing template" warning | New template added to consumer list but file missing | Create template in `matugen/templates/` |
| App doesn't pick up new colors | Post-hook didn't fire / reload signal wrong | Check the `_handle` call's post-hook; test manually |
| ironbar CSS unchanged visually | systemd restart failed | `systemctl --user status ironbar` |
| kitty colors stuck | kitty cached — SIGUSR1 not enough | `$mod+CTRL+G` or full kitty restart |
| Wallpaper-mode accents unchanged | matugen or jq missing; swww state missing | Install matugen+jq; ensure swww is running |

## Related

- `.claude/rules/snazzy-palette.md` — palette spec
- `.claude/rules/nvidia-wayland.md` — NVIDIA/Wayland tuning (adjacent concern)
- `matugen/README.md` — template registry
- `scripts/rice-reload.sh` — full desktop reload (calls palette-propagate)
- `scripts/theme-sync.sh` — palette-propagate + gsettings/Qt/Plasma runtime state
- `rice_iteration` skill — screenshot → edit → reload feedback loop (composes with this one)
