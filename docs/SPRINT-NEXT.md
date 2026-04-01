# Sprint 3 — Museum-Inspired Mac Branch Enhancements

Inspired by analyzing 28 r/unixporn repos in `hairglasses-studio/dotfile-museum`. Patterns that appeared across the community's best rices (AwesomeWM, i3, Hyprland, Openbox) adapted to our macOS stack.

## Completed

- [x] **SketchyBar RAM/memory widget** — `memory.sh`, green 󰍛 icon, 10s refresh (inspired by crylia, eromatiya, elenapan)
- [x] **SketchyBar weather widget** — `weather.sh`, yellow  icon, 10min refresh via wttr.in, hides offline (inspired by elenapan, eromatiya)
- [x] **SketchyBar disk usage widget** — `disk.sh`, gray 󰋊 icon, 5min refresh, escalates to red >90% (inspired by barbaross93, elenapan)
- [x] **SketchyBar Docker container count** — `docker.sh`, blue 󰡨 icon, hides when Docker not running (inspired by codeheister)
- [x] **SketchyBar power/session menu** — `power.sh`, red ⏻ icon with popup: Lock, Sleep, Logout, Shutdown (inspired by evankoe STFU mode, eromatiya, nekorosys wlogout)
- [x] **SketchyBar screen recorder toggle** — `recorder.sh`, click to start/stop `screencapture -v`, red REC indicator (inspired by eromatiya)
- [x] **Custom fastfetch ASCII art** — cyberpunk HG-STUDIO logo in cyan/magenta, replaces blank `"type": "none"` (inspired by jinpots, alba4k, syndrizzle)

## Previous Sprints

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
