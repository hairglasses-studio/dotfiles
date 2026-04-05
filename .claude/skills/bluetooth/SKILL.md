---
description: "Manage Bluetooth devices. $ARGUMENTS can be: (empty)=list devices, 'scan'=discover nearby, 'connect <name>'=connect device, 'disconnect <name>'=disconnect, 'pair <name>'=full pairing flow, 'battery'=show battery levels"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)**: Call `mcp__dotfiles__bt_list_devices` — show all paired devices with status and battery
- **"scan"**: Call `mcp__dotfiles__bt_scan` — discover nearby devices (8s timeout)
- **"connect <name>"**: Call `mcp__dotfiles__bt_connect` with `device=<name>` — connect with BLE retry
- **"disconnect <name>"**: Call `mcp__dotfiles__bt_disconnect` with `device=<name>`
- **"pair <name>"**: Call `mcp__dotfiles__bt_discover_and_connect` with `device=<name>` — full flow: scan→pair→trust→connect
- **"battery"**: Call `mcp__dotfiles__bt_list_devices` and display only battery levels
- **"power on/off"**: Call `mcp__dotfiles__bt_power` with `action=on/off`
