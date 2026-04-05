---
name: rice-check
description: Validate the entire Snazzy cyberpunk rice — symlinks, services, palette, configs, fonts
allowed-tools: Bash, Read, Grep, Glob, mcp__dotfiles__dotfiles_check_symlinks, mcp__dotfiles__dotfiles_validate_config, mcp__hyprland__hypr_reload_config
---

Run a comprehensive health check on the dotfiles rice. Execute these checks in parallel:

1. **Symlink health** — Use `dotfiles_check_symlinks` MCP tool or run `install.sh --check`
2. **Service status** — Check these are running: `pgrep -la 'Hyprland|eww|mako|swww|hypridle|pypr'`
3. **Hyprland config** — Run `hyprctl configerrors` (should return empty)
4. **Snazzy palette** — Grep all config files for hex colors not in the Snazzy palette: `#57c7ff, #ff6ac1, #5af78e, #f3f99d, #ff5c57, #686868, #f1f1f0, #000000, #1a1a1a, #1a1b26, #9aedfe, #eff0eb`
5. **Font availability** — Check `fc-list | grep -i 'JetBrainsMono\|Monoid\|Maple Mono'`
6. **eww bar** — Run `eww state 2>&1 | head -10` to verify variables are populating
7. **MCP servers** — Verify all 4 servers in `.mcp.json` have compiled binaries

Report results as a table: Component | Status | Details
