---
paths:
  - "**/*.css"
  - "**/*.scss"
  - "**/*.conf"
  - "**/*.toml"
  - "ironbar/**"
  - "wofi/**"
  - "wlogout/**"
  - "swaync/**"
  - "hyprshell/**"
---

# Hairglasses Palette — Token Rules

The palette is a **named playlist** — the active palette rotates via
`scripts/palette-playlist.sh`, but the **token names are invariant**. Write all
UI chrome against the tokens below, never against raw hex. The hex values in
this file are the `hairglasses-neon` defaults and shown only for reference;
every other palette in `theme/palettes/` fills the same tokens with different
hex values.

Source of truth: `theme/palette.env` (symlink → `theme/palettes/<active>.env`).

| Name | Token | ANSI slot | hairglasses-neon hex |
|------|-------|-----------|----------------------|
| Primary | `THEME_PRIMARY` | color6 | #29f0ff |
| Secondary | `THEME_SECONDARY` | color5 | #ff47d1 |
| Tertiary | `THEME_TERTIARY` | color2 | #3dffb5 |
| Warning | `THEME_WARNING` | color3 | #ffe45e |
| Danger | `THEME_DANGER` | color1 | #ff5c8a |
| Blue | `THEME_BLUE` | color4 | #4aa8ff |
| Muted | `THEME_MUTED` | color8 | #66708f |
| Foreground | `THEME_FG` | color7 | #f7fbff |
| Background | `THEME_BG` | color0 | #05070d |
| Surface | `THEME_SURFACE` | | #0f1219 |
| Surface alt | `THEME_SURFACE_ALT` | | #161c2b |
| Panel | `THEME_PANEL` | | #0d1018 |
| Panel strong | `THEME_PANEL_STRONG` | | #141a27 |
| Border | `THEME_BORDER` | | #2a3246 |
| Border strong | `THEME_BORDER_STRONG` | | #46506d |

## Rules

- Reference colors through `${THEME_*}` tokens in templates under
  `matugen/templates/`, not hardcoded hex in final configs.
- Do not introduce new tokens without also adding them to every palette in
  `theme/palettes/*.env` — missing entries will break `envsubst` rendering.
- Do not hardcode palette-specific hex (e.g. `#29f0ff`) outside
  `theme/palettes/<name>.env`. The rule holds for every palette, not just
  hairglasses-neon.
- Meaning stays consistent across palettes: `THEME_DANGER` is always the
  error/destructive accent; `THEME_PRIMARY` is always the headline accent.
  Palettes may shift hue but must preserve semantic role.

Available palettes: run `scripts/palette-playlist.sh list`.
