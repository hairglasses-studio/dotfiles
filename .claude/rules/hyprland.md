---
paths:
  - "hypr/**"
  - "hyprland/**"
---

Hyprland config conventions:
- Always use US QWERTY layout: `kb_layout = us`
- Use `monitor=desc:` syntax for monitor identification (not serial numbers)
- Dual monitor: DP-3 (Samsung LC49G95T, 5120x1440@240Hz ultrawide, left) + DP-2 (XEC ES-G32C1Q, 2560x1440@180Hz portrait, right)
- After editing: `hyprctl reload` fires automatically via PostToolUse hook
