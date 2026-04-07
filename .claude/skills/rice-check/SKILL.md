---
name: rice-check
description: Validate the entire Snazzy cyberpunk rice — symlinks, services, palette, configs, fonts
allowed-tools: Bash, Read, Grep, Glob, mcp__dotfiles__dotfiles_check_symlinks, mcp__dotfiles__dotfiles_validate_config
---

Run a comprehensive health check on the dotfiles rice. Execute these checks in parallel:

1. **Symlink health** — Use `dotfiles_check_symlinks` MCP tool or run `install.sh --check`
2. **Service status** — Check these are running: `pgrep -la 'AeroSpace|sketchybar|borders'`
3. **AeroSpace config** — Run `aerospace reload-config --dry-run 2>&1` or validate TOML syntax
4. **Snazzy palette** — Grep all config files for hex colors not in the Snazzy palette: `#57c7ff, #ff6ac1, #5af78e, #f3f99d, #ff5c57, #686868, #f1f1f0, #000000, #1a1a1a, #1a1b26, #9aedfe, #eff0eb`
5. **Font availability** — Check `system_profiler SPFontsDataType 2>/dev/null | grep -i 'JetBrainsMono\|Monoid\|Maple Mono'`
6. **SketchyBar** — Run `bash -n sketchybar/sketchybarrc && bash -n sketchybar/plugins/*.sh` to validate syntax
7. **MCP servers** — Verify servers in `.mcp.json` are configured correctly
8. **Shader pipeline** — Run `bash ghostty/shaders/bin/shader-meta.sh validate` to check manifest consistency

Report results as a table: Component | Status | Details
