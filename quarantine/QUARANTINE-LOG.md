# Quarantine Log

Items moved here during the 2026-04-05 dotfiles audit. These configs are dead,
replaced, or conflicting with the active setup. Review before deleting permanently.

## config/mako/
**Replaced by:** swaync (active notification daemon, exec-once in hyprland.conf)
**Why:** mako config was fully themed but never launched. swaync migration completed.
**Restore:** `git mv quarantine/config/mako mako && ln -sf ~/hairglasses-studio/dotfiles/mako ~/.config/mako`

## config/logiops/
**Replaced by:** Solaar (official Logitech tool, exec-once in hyprland.conf)
**Why:** logiops and Solaar both manage MX Master 4. Only Solaar is launched.
**Restore:** `git mv quarantine/config/logiops logiops`

## config/rofi/
**Replaced by:** wofi (active launcher, $menu variable in hyprland.conf)
**Why:** rofi was fully themed but has no keybinds or exec-once. wofi is primary.
**Restore:** `git mv quarantine/config/rofi rofi && ln -sf ~/hairglasses-studio/dotfiles/rofi ~/.config/rofi`

## config/init.vim.bak
**Replaced by:** nvim/init.lua (lazy.nvim-based config)
**Why:** Old vimscript config, superseded by lua migration.

## scripts/logiops-deploy.sh
**Replaced by:** Solaar (no deployment needed, runs from exec-once)

## scripts/macos-defaults.sh
**Why:** macOS-only script, irrelevant on Manjaro Linux.
