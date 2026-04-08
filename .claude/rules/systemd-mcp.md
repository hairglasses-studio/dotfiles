---
paths:
  - "mcp/systemd-mcp/**"
---

systemd-mcp conventions:
- Preserve the D-Bus primary path with `systemctl` and `journalctl` fallback behavior.
- Keep service-management actions explicit about target scope and current unit state.
- When touching module registration or tool contracts, keep the single-module surface coherent rather than fragmenting it into ad hoc handlers.
