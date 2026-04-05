---
description: "Full desktop performance and service health check. Combines system monitoring, GPU stats, shader status, eww widgets, and failed systemd services."
user_invocable: true
---

Execute these in parallel:
1. `mcp__dotfiles__system_health_check` — Temps, GPU, memory, disk, uptime, updates with thresholds
2. `mcp__dotfiles__systemd_failed` — Any failed systemd user services
3. `mcp__dotfiles__ps_list` with `sort=cpu, limit=5` — Top 5 CPU hogs
4. `mcp__dotfiles__shader_status` — Current Ghostty shader and animation state
5. `mcp__dotfiles__dotfiles_eww_status` — Eww daemon health and widget state
6. `mcp__dotfiles__gpu_status` — Detailed NVIDIA GPU stats

Present as a desktop performance dashboard:
- System health: OK/WARN/CRIT with threshold violations
- GPU: temp, utilization, power draw, shader overhead
- Desktop services: eww, hyprland, hypridle, swaync status
- Process hogs: top 5 by CPU with PID and command
- Failed services: list with suggested fix commands
