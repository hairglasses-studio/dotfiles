---
description: "Manage eww widgets and desktop services. $ARGUMENTS: (empty)=status, 'restart'=restart daemon, 'get <var>'=query variable, 'reload'=cascade reload all themed services"
user_invocable: true
allowed-tools: mcp__dotfiles__dotfiles_eww_status, mcp__dotfiles__dotfiles_eww_restart, mcp__dotfiles__dotfiles_eww_get, mcp__dotfiles__dotfiles_cascade_reload
---

Parse `$ARGUMENTS`:

- **(empty)** or **"status"**: Call `mcp__dotfiles__dotfiles_eww_status` — show daemon health, active windows, key variables
- **"restart"**: Call `mcp__dotfiles__dotfiles_eww_restart` — kill and restart eww daemon with both bars
- **"get <var>"**: Call `mcp__dotfiles__dotfiles_eww_get` with `variable=<var>` — query current value of an eww variable (e.g., `volume`, `workspace`, `battery`)
- **"reload"**: Call `mcp__dotfiles__dotfiles_cascade_reload` — ordered multi-service reload with health verification (eww, mako/swaync, waybar, hyprland)

Present status as a dashboard showing daemon state, active windows, and any errors.
