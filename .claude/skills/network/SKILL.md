---
description: "Network diagnostics and WiFi management. $ARGUMENTS can be: (empty)=show status, 'wifi'=list networks, 'fix'=diagnose+reconnect, 'dns'=show DNS, 'connections'=list saved profiles"
user_invocable: true
---

Parse `$ARGUMENTS` to determine the action:

- **(empty)**: Call `mcp__dotfiles__network_status` and display connectivity, WiFi state, active connections
- **"wifi"**: Call `mcp__dotfiles__network_wifi_list` with `rescan=true` and display sorted by signal
- **"fix"**: Run diagnostic chain:
  1. `mcp__dotfiles__network_status` — check current state
  2. If disconnected: `mcp__dotfiles__network_wifi_list` with `rescan=true` — find available networks
  3. Suggest strongest known network to reconnect
  4. Report findings and recommended fix
- **"dns"**: Call `mcp__dotfiles__network_dns` and display per-interface DNS servers
- **"connections"**: Call `mcp__dotfiles__network_connections` and display all saved profiles
- **"ports"**: Call `mcp__dotfiles__port_list` and display listening TCP ports

Present results in a clean dashboard format.
