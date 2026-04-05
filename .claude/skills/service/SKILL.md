---
description: "Manage systemd services. $ARGUMENTS format: '<action> <unit>' where action is: status, restart, start, stop, logs, failed, timers, units"
user_invocable: true
---

Parse `$ARGUMENTS` to determine the action:

- **"status <unit>"**: Call `mcp__dotfiles__systemd_status` with `unit=<unit>`
- **"restart <unit>"**: Call `mcp__dotfiles__systemd_restart` with `unit=<unit>`, then verify with `mcp__dotfiles__systemd_status`. Report success/failure with last 10 log lines via `mcp__dotfiles__systemd_logs`.
- **"start <unit>"**: Call `mcp__dotfiles__systemd_start` with `unit=<unit>`
- **"stop <unit>"**: Call `mcp__dotfiles__systemd_stop` with `unit=<unit>`
- **"logs <unit>"**: Call `mcp__dotfiles__systemd_logs` with `unit=<unit>` and `lines=50`
- **"failed"** (no unit): Call `mcp__dotfiles__systemd_failed` to show all failed units
- **"timers"** (no unit): Call `mcp__dotfiles__systemd_list_timers` to show active timers
- **"units"** (no unit): Call `mcp__dotfiles__systemd_list_units` to show all loaded units
- **(empty)**: Call `mcp__dotfiles__systemd_failed` to show failed units as a quick health check

Default scope is `user`. All operations use `--user` flag.
