# Sprint 4 — Mac Branch Audit & Comprehensive Fixes

Full-branch audit after 3 sprints of migration and museum-inspired enhancements. Identified and fixed critical bugs, closed parity gaps with main branch, added quality-of-life improvements.

## Completed

- [x] **Remove duplicate functions in aliases.zsh** — ssh(), command_not_found_handler(), yay() were defined twice (lines 562-594 and 861-895). Removed duplicates.
- [x] **Fix nvimconfig alias** — `init.vim` → `init.lua`
- [x] **Migrate 7 cyberpunk shaders from main** — synthwave-horizon, holo-display, neon-hex-grid, circuit-trace, rain-on-glass, cyber-glitch-holo, lib/hex.glsl. Fixes broken shader quick-switch aliases.
- [x] **Fix hardcoded username in plist** — `com.dotfiles.shader-rotate.plist` referenced `/Users/mitchnotmitchell/`, corrected to `/Users/mitch/`
- [x] **Fix hardcoded username in zshrc** — ralph alias used absolute path, changed to `$HOME`
- [x] **Fix tmux GPU temp for macOS** — Replaced Linux-only `nvidia-smi`/`/sys/class/drm/` with `system_profiler SPDisplaysDataType` chipset query
- [x] **Port Claude Code rice-developer agent** — Adapted from Hyprland/eww to AeroSpace/SketchyBar/Ghostty
- [x] **Port rice-check skill** — macOS service checks (AeroSpace, SketchyBar, borders), shader pipeline validation
- [x] **Port screenshot-review skill** — `screencapture` instead of `grim`, SketchyBar instead of eww
- [x] **Add direnv hook** — `eval "$(direnv hook zsh)"` in zshrc (was on main, missing from mac)
- [x] **Auto-LS on cd** — `chpwd() { ls }` in aliases.zsh (museum pattern from alba4k)

## Previous Sprints

### Sprint 3 — Museum-Inspired Enhancements
- SketchyBar widgets: memory, weather, disk, docker, power menu, screen recorder
- Custom fastfetch ASCII art (cyberpunk HG-STUDIO logo)
- Analyzed 28 r/unixporn repos in dotfile-museum

### Sprint 2 — Polish & Parity
- AeroSpace F13/F14 encoder keybinds
- SketchyBar uptime + network widgets
- Brewfile completeness (lavat, neo-matrix/tte documented)
- Source compositor.sh + config.sh in zshrc
- CLAUDE.md updates (shader count 137, repo-consumed files)

### Sprint 1 — Main Branch Migration
- 6 workstreams cherry-picking cross-platform features from main to mac
- dotfiles.toml, lib scripts, Claude Code integration, keyboard firmware, TUI configs

## Future Candidates

- GPU temperature widget (needs `powermetrics` or IOKit — requires sudo)
- Wallpaper shader cycling keybinds (macOS equivalent of `$mod+Shift+W`)
- RetroVisor version pinning in install.sh
- Smart bar auto-hide (static/dynamic/hover modes — nekorosys-inspired)
- `unimatrix` in tmux screensaver rotation
- Dark mode toggle function (nekorosys-inspired)
- Theme state management / persistence
