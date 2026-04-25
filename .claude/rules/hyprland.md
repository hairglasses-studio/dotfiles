---
paths:
  - "hypr/**"
  - "hyprland/**"
---

Hyprland config conventions:
- Always use US QWERTY layout: `kb_layout = us`
- Use `monitor=desc:` syntax for monitor identification (not serial numbers)
- 3-monitor layout (verified 2026-04-23 via `hyprctl monitors all`):
  - HDMI-A-1: XEC ES-G32C1Q, 2560x1440@60Hz, portrait transform=1, scale=2, pos 5316,80 (left)
  - DP-3: Samsung LC49G95T, 5120x1440@239.76Hz native (DSC), landscape transform=0, scale=1, pos 6036,0 (center)
  - DP-2: XEC ES-G32C1Q, 2560x1440@180Hz, portrait transform=3, scale=2, pos 11156,80 (right)
  - Logical x layout: HDMI 5316..6036, DP-3 6036..11156, DP-2 11156..11876 — edges meet, no overlap.
  - Samsung is the tallest panel (1440 vs portraits' 1280), so it anchors y=0 and the flanking portraits sit at y=80 to vertically center inside it.
  - Samsung's native 5120x1440@240Hz DSC mode is the persisted state. If DSC drops out (Samsung stuck at 0x0 or only 2560x1440 in availableModes), see nvidia-wayland.md for the hardware power-cycle recovery.
- After editing: `hyprctl reload` fires automatically via PostToolUse hook
