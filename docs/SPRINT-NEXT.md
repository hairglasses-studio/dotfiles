# Sprint 2 — Mac Branch Polish & Parity

Post-migration sprint closing gaps identified during the Sprint 1 audit (6-workstream cherry-pick from main).

## Completed

- [x] **AeroSpace F13/F14 encoder keybinds** — `alt-f13`/`alt-f14` for focus left/right, matching Hyprland + Keychron V1 Ultra encoder firmware
- [x] **SketchyBar uptime widget** — `uptime.sh` plugin, green icon, 60s refresh, displays `Xd Yh` / `Xh Ym` / `Xm`
- [x] **SketchyBar network widget** — `network.sh` plugin, cyan icon, 5s refresh, shows `↓ XM ↑ YK` throughput
- [x] **Brewfile completeness** — added `lavat`; documented `neo-matrix` (cargo) and `tte` (pipx) as comments
- [x] **Source lib scripts** — `compositor.sh` + `config.sh` now sourced in zshrc (was orphaned after Sprint 1 migration)
- [x] **CLAUDE.md updates** — shader count 132 -> 137, documented repo-consumed files (.claude/, .mcp.json, keyboard/, docs/)

## Decided: No Action Needed

- **Install.sh symlinks for .claude/, .mcp.json, keyboard/, docs/** — These are repo-consumed (Claude Code reads from working directory) or reference-only. No runtime symlink needed.

## Future Candidates (not in this sprint)

- GPU temperature widget for SketchyBar (macOS lacks standard sysfs, needs `powermetrics` or IOKit)
- Docker/container status widget
- Wallpaper shader cycling keybinds (macOS equivalent of `$mod+Shift+W`)
- RetroVisor version pinning in install.sh
- `unimatrix` in tmux screensaver rotation
