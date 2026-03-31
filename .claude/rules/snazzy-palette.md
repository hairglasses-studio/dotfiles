---
paths:
  - "**/*.css"
  - "**/*.scss"
  - "**/*.conf"
  - "**/*.toml"
  - "eww/**"
  - "mako/**"
  - "wofi/**"
  - "wlogout/**"
  - "waybar/**"
---

# Snazzy Palette Enforcement

Only use colors from the Snazzy-on-Black palette:

| Name | Hex | RGBA |
|------|-----|------|
| Cyan | #57c7ff | rgba(87, 199, 255, ...) |
| Magenta | #ff6ac1 | rgba(255, 106, 193, ...) |
| Green | #5af78e | rgba(90, 247, 142, ...) |
| Yellow | #f3f99d | rgba(243, 249, 157, ...) |
| Red | #ff5c57 | rgba(255, 92, 87, ...) |
| Gray | #686868 | rgba(104, 104, 104, ...) |
| Light cyan | #9aedfe | (ANSI bright cyan) |
| Light fg | #eff0eb | (ANSI bright white) |
| Foreground | #f1f1f0 | |
| Background | #000000 | |
| Dark bg | #1a1a1a / #1a1b26 | |

Do not introduce colors outside this palette unless the user explicitly requests it.
