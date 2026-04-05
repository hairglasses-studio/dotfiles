---
description: "Process debugging and port investigation. $ARGUMENTS: (empty)=top processes, 'port <N>'=investigate port, 'tree <pid>'=process tree, 'kill <pid>'=send signal"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)**: Call `mcp__dotfiles__ps_list` with `sort=cpu, limit=10` — top 10 by CPU
- **"port <N>"**: Call `mcp__dotfiles__investigate_port` with `port=<N>` — find process→service→logs chain
- **"ports"**: Call `mcp__dotfiles__port_list` — all listening TCP ports
- **"tree <pid>"**: Call `mcp__dotfiles__ps_tree` with `pid=<pid>` — process hierarchy
- **"kill <pid>"**: Call `mcp__dotfiles__kill_process` with `pid=<pid>` — send SIGTERM
- **"service <name>"**: Call `mcp__dotfiles__investigate_service` — service→process→ports→logs chain
- **"gpu"**: Call `mcp__dotfiles__gpu_status` — NVIDIA GPU utilization and processes
