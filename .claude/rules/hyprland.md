---
paths:
  - "hypr/**"
  - "hyprland/**"
---

Hyprland config conventions:
- Always use US QWERTY layout: `kb_layout = us`
- Use `monitor=desc:` syntax for monitor identification (not serial numbers)
- 3-monitor layout (verified 2026-04-23 via `hyprctl monitors all`), all scale=1:
  - HDMI-A-1: XEC ES-G32C1Q, 2560x1440@60Hz, portrait transform=1, pos 4596,0 (left, logical 1440x2560)
  - DP-3: Samsung LC49G95T, 5120x1440@239.76Hz native (DSC), landscape transform=0, pos 6036,560 (center, logical 5120x1440)
  - DP-2: XEC ES-G32C1Q, 2560x1440@180Hz, portrait transform=3, pos 11156,0 (right, logical 1440x2560)
  - Logical x layout: HDMI 4596..6036, DP-3 6036..11156, DP-2 11156..12596 — edges meet, no overlap.
  - At scale=1 the rotated portraits (1440x2560) are taller than Samsung (1440), so Samsung anchors y=560 to vertically center inside the portraits' 2560-tall canvases.
  - Samsung's native 5120x1440@240Hz DSC mode is the persisted state. If DSC drops out (Samsung stuck at 0x0 or only 2560x1440 in availableModes), see nvidia-wayland.md for the hardware power-cycle recovery.
- After editing: `hyprctl reload` fires automatically via PostToolUse hook
