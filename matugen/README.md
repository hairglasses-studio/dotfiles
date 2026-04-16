# matugen templates

Source-of-truth template set for the Hairglasses Neon palette pipeline.

## Two rendering modes

### Fixed mode (default)
`scripts/palette-propagate.sh` reads `theme/palette.env`, exports each
`THEME_*` variable, and runs `envsubst` across every `*.{conf,css,toml,theme,sh}`
file in `templates/`. Output goes to each app's real config path, then a
post-reload fires.

This is the mode that runs whenever `palette.env` changes, whenever
`rice-reload.sh` is invoked, and whenever `chezmoi run_onchange_` fires.
The "Hairglasses Neon" palette stays exact.

### Wallpaper mode (opt-in)
`scripts/palette-propagate.sh --wallpaper` runs `matugen image <current-wall>`
using `matugen/config.toml`. matugen's HCT algorithm extracts MD3 accent
colors from the image, overlays them on the locked neutral palette via
`custom_colors`, and renders a parallel `.matugen` template set into
`~/.cache/matugen/`. palette-propagate.sh then copies those cached files
over the fixed-mode output.

Templates for wallpaper mode live next to their fixed counterparts with a
`.matugen` suffix:

```
templates/kitty-colors.conf           # fixed-mode (envsubst)
templates/kitty-colors.conf.matugen   # wallpaper-mode (matugen syntax)
```

## Template registry

| Template | Output path | Consumer reload |
|---|---|---|
| `gtk-colors.css` | `~/.config/{ironbar,swaync,wofi,wlogout,hyprshell}/theme.generated.css` | `swaync-client -rs`, `systemctl --user restart ironbar` |
| `kitty-colors.conf` | `~/.config/kitty/cyberpunk-neon.conf` | `pkill -SIGUSR1 kitty` |
| `hyprland-colors.conf` | `~/.config/hypr/colors.conf` | `hyprctl reload` |
| `hyprlock-colors.conf` | `~/.config/hypr/hyprlock-colors.conf` | (inline, reloads on next lock) |
| `cava-colors.conf` | `~/.config/cava/config` (merged) | `pkill -USR2 cava` |
| `btop-theme.theme` | `~/.config/btop/themes/hairglasses-neon.theme` | (inline, reloads on btop restart) |
| `yazi-theme.toml` | `~/.config/yazi/theme.toml` | (inline, reloads on yazi restart) |
| `zsh-fzf-colors.sh` | `~/.config/fzf/hg-colors.sh` | (sourced in zshrc) |

## Variable surface

All templates reference env vars from `theme/palette.env`:

```
THEME_BG THEME_SURFACE THEME_SURFACE_ALT THEME_PANEL THEME_PANEL_STRONG
THEME_BORDER THEME_BORDER_STRONG THEME_FG THEME_MUTED
THEME_PRIMARY THEME_SECONDARY THEME_TERTIARY
THEME_WARNING THEME_DANGER THEME_BLUE
```

## Adding a new consumer

1. Create `templates/<app>-colors.<ext>` with `${THEME_*}` variables
2. Register target path + post-hook in `scripts/palette-propagate.sh`
3. Run `scripts/palette-propagate.sh` to verify output
4. If using wallpaper mode, also add a `.matugen` variant using matugen syntax
