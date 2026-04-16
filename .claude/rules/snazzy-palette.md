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

# Hairglasses Neon Palette

Only use colors from the unified "Hairglasses Neon" palette (source: `theme/palette.env`):

| Name | Hex | Token | ANSI slot |
|------|-----|-------|-----------|
| Primary (cyan) | #29f0ff | `THEME_PRIMARY` | color6 |
| Secondary (magenta) | #ff47d1 | `THEME_SECONDARY` | color5 |
| Tertiary (green) | #3dffb5 | `THEME_TERTIARY` | color2 |
| Warning (yellow) | #ffe45e | `THEME_WARNING` | color3 |
| Danger (red) | #ff5c8a | `THEME_DANGER` | color1 |
| Blue | #4aa8ff | `THEME_BLUE` | color4 |
| Muted | #66708f | `THEME_MUTED` | color8 |
| Foreground | #f7fbff | `THEME_FG` | color7 |
| Background | #05070d | `THEME_BG` | color0 |
| Surface | #0f1219 | `THEME_SURFACE` | |
| Surface alt | #161c2b | `THEME_SURFACE_ALT` | |
| Panel | #0d1018 | `THEME_PANEL` | |
| Border | #2a3246 | `THEME_BORDER` | |
| Border strong | #46506d | `THEME_BORDER_STRONG` | |

Do not introduce colors outside this palette unless the user explicitly requests it.
