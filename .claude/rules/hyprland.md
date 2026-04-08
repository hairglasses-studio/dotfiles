---
paths:
  - "hypr/**"
  - "hyprland/**"
---

Hyprland config conventions:
- Always use US QWERTY layout: `kb_layout = us`
- Use `monitor=desc:` syntax for monitor identification (not serial numbers)
- Dual monitor: DP-1 (2560x1440, left) + DP-2 (5120x1440 ultrawide, right)
- After editing: `hyprctl reload` fires automatically via PostToolUse hook
