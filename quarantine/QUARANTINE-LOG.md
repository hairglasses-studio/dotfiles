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

## config/waybar/ (Phase 2)
**Replaced by:** eww (active bar, exec-once in hyprland.conf)
**Why:** Waybar has zero references in hyprland.conf, systemd, or scripts. Eww is primary.
**Restore:** `git mv quarantine/config/waybar waybar && ln -sf ~/hairglasses-studio/dotfiles/waybar ~/.config/waybar`

## scripts/hypr-vertical-columns.sh (Phase 2)
**Why:** Orphaned layout script with no keybind and zero references from any config or alias.

## scripts/font-mix.sh (Phase 1)
**Replaced by:** Maple Mono NF CN standard (no mixing needed)
**Why:** Monaspace multi-font mixer is dead after font standardization.
