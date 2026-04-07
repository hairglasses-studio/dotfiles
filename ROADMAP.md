# Roadmap

## Current State

Manjaro Linux dotfiles with 60+ symlinked configs, 138 GLSL shaders with shuffled playlists, Snazzy-on-black palette applied to 15+ tools, Hyprland-first compositor automation, eww bar widgets, and a full boot stack (rEFInd + Plymouth). Idempotent installer with backup/restore support.

Shader collection is one of the largest curated GLSL terminal shader sets publicly available. All configs MIT licensed.

## Planned

### Phase 1 — Linux Installer Hardening
- Keep `install.sh` Linux-only and catch non-Manjaro drift early
- Tighten package validation for pacman/yay/metapac
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
