# Roadmap

## Current State

Cross-platform dotfiles (macOS + Manjaro Linux) with 60+ symlinked configs, 138 GLSL shaders with shuffled playlists, Snazzy-on-black palette applied to 15+ tools, compositor abstraction layer (Hyprland/Sway/AeroSpace), eww bar with custom widgets, and a full boot stack (rEFInd + Plymouth). Idempotent installer with backup/restore support.

Shader collection is the largest curated Ghostty shader set publicly available. All configs MIT licensed.

## Planned

### Phase 1 — Linux Parity & Installer
- Linux installer path (`install-linux.sh`) — currently macOS-focused
- Package list for pacman/yay equivalent to Brewfile
- Hyprland-specific installer steps (eww, mako, wofi, wlogout setup)
- Automated symlink validation in CI

### Phase 2 — Eww Bar Polish
- Sidebar with system stats, calendar, and notifications
- Revealer-on-hover patterns for dense info display
- Workspace indicator with empty-workspace placeholders
- Theme generation from wallpaper colors via matugen integration

### Phase 3 — Shader Pipeline
- Shader performance benchmarks in CI (flag regressions above GPU budget)
- Shader preview gallery (static renders for README/docs)
- Wallpaper shader expansion (more procgen options)
- Shader parameter presets (expose uniforms as config)

## Future Considerations
- NixOS or home-manager alternative for declarative config management
- Neovim migration from vim-plug to lazy.nvim
- Shared dotfiles module system (pick configs a la carte instead of all-or-nothing)
- Automated rice screenshot CI (Hyprland headless + screenshot comparison)
