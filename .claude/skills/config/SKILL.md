---
description: "Manage dotfiles configs — list directories, validate syntax, check symlinks, reload services. $ARGUMENTS: (empty)=list, 'validate <file>'=check syntax, 'symlinks'=check health, 'reload <service>'=reload service"
user_invocable: true
allowed-tools: mcp__dotfiles__dotfiles_list_configs, mcp__dotfiles__dotfiles_validate_config, mcp__dotfiles__dotfiles_check_symlinks, mcp__dotfiles__dotfiles_reload_service
---

Parse `$ARGUMENTS`:

- **(empty)** or **"list"**: Call `mcp__dotfiles__dotfiles_list_configs` — list all config directories with symlink health and detected format (TOML/JSON/INI)
- **"validate <file>"**: Call `mcp__dotfiles__dotfiles_validate_config` with `path=<file>` — validate TOML or JSON syntax, report errors with line numbers
- **"symlinks"**: Call `mcp__dotfiles__dotfiles_check_symlinks` — check health of all expected dotfile symlinks (OK/broken/missing)
- **"reload <service>"**: Call `mcp__dotfiles__dotfiles_reload_service` with `service=<service>` — reload a desktop service (hyprland, mako, swaync, eww, waybar, tmux)

For `list`, display as a table with columns: Directory, Symlink Status, Format. For `validate`, show OK or error details with line:column. For `symlinks`, show a health dashboard with OK/broken/missing counts.
