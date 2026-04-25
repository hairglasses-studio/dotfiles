---
paths:
  - "hypr/**"
  - "hyprland/**"
---

Hyprland config conventions:
- Always use US QWERTY layout: `kb_layout = us`
- Use `monitor=desc:` syntax for monitor identification (not serial numbers)
- 3-monitor layout (verified 2026-04-23 via `hyprctl monitors all`):
  - HDMI-A-1: XEC ES-G32C1Q, 2560x1440@60Hz, portrait transform=1, scale=2 (left)
  - DP-3: Samsung LC49G95T, 2560x1440@60Hz fallback (DSC failing), landscape transform=0, scale=2 (center)
  - DP-2: XEC ES-G32C1Q, 2560x1440@180Hz, portrait transform=3, scale=2 (right)
  - Logical x layout: HDMI 5316..6036, DP-3 6036..7316, DP-2 7316..8036 — edges meet, no overlap.
  - Samsung's native 5120x1440@240Hz DSC mode is unavailable until DSC negotiation recovers — see nvidia-wayland.md.
- After editing: `hyprctl reload` fires automatically via PostToolUse hook
