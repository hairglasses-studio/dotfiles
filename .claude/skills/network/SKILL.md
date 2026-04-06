---
description: "Network diagnostics, WiFi, and VPN management. $ARGUMENTS: (empty)=show status, 'wifi'=list networks, 'connect <ssid>'=connect to WiFi, 'fix'=diagnose+reconnect, 'dns'=show DNS, 'connections'=list saved profiles, 'vpn'=toggle VPN, 'ports'=list listening ports"
user_invocable: true
allowed-tools: mcp__dotfiles__network_status, mcp__dotfiles__network_wifi_list, mcp__dotfiles__network_wifi_connect, mcp__dotfiles__network_dns, mcp__dotfiles__network_connections, mcp__dotfiles__network_vpn_toggle, mcp__dotfiles__port_list
---

Parse `$ARGUMENTS`:

- **(empty)**: Call `mcp__dotfiles__network_status` — display connectivity, WiFi state, active connections
- **"wifi"**: Call `mcp__dotfiles__network_wifi_list` with `rescan=true` — display sorted by signal strength
- **"connect <ssid>"**: Call `mcp__dotfiles__network_wifi_connect` with `ssid=<ssid>` — connect to a WiFi network
- **"vpn"**: Call `mcp__dotfiles__network_vpn_toggle` — toggle VPN connection on/off
- **"fix"**: Run diagnostic chain:
  1. `mcp__dotfiles__network_status` — check current state
  2. If disconnected: `mcp__dotfiles__network_wifi_list` with `rescan=true` — find available networks
  3. Suggest strongest known network to reconnect
  4. Report findings and recommended fix
- **"dns"**: Call `mcp__dotfiles__network_dns` — display per-interface DNS servers
- **"connections"**: Call `mcp__dotfiles__network_connections` — display all saved connection profiles
- **"ports"**: Call `mcp__dotfiles__port_list` — display listening TCP ports with PID and service name

Present results in a clean dashboard format.
