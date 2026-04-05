---
description: Run a quick system health check. Execute these in parallel:
user_invocable: true
---

1. `mcp__dotfiles__system_temps` — CPU/GPU/NVMe temperatures and fan RPMs
2. `mcp__dotfiles__system_gpu` — NVIDIA GPU utilization, memory, power draw
3. `mcp__dotfiles__system_memory` — RAM and swap usage
4. `mcp__dotfiles__system_disk` — Per-mount disk usage
5. `mcp__dotfiles__system_uptime` — Uptime, load averages, last boot
6. `mcp__dotfiles__system_updates` — Pending pacman/AUR updates
7. `mcp__dotfiles__network_status` — Network connectivity and active connections
8. `mcp__dotfiles__audio_status` — Audio sink/source, volume, mute state

After gathering all results, present a dashboard:
- Flag any CPU temp > 85C, GPU temp > 90C, memory > 85%, disk > 90% as WARN
- Flag any CPU temp > 95C, GPU temp > 100C, memory > 95%, disk > 98% as CRIT
- Show overall status: OK / WARN / CRIT
- List warnings and recommended actions
