# Roadmap

## Current State

Manjaro Linux dotfiles with 90+ managed configs (chezmoi + install.sh), 139 DarkWindow GLSL shaders with shuffled playlists, Hairglasses Neon palette applied to 20+ tools, Hyprland-first compositor automation, an Ironbar menubar on the theme pipeline, and a full boot stack (rEFInd + Plymouth). Idempotent installer with chezmoi declarative management (134 managed entries, 6 lifecycle scripts).

Single consolidated MCP server (dotfiles-mcp) with ~400 tools across 30+ modules. Shader collection is one of the largest curated GLSL terminal shader sets publicly available. All configs MIT licensed.

## Recently Completed

### Dotfiles Cleanup & Wiring (2026-04-16)
Major cleanup removing -27k lines of accumulated config debt:

- **Removed**: ghostty terminal (122 shaders), juhradial input device, makima gamepad remapper, CRTty shader catalog (131 files), p10k prompt engine (816 lines)
- **Unified palette**: "Hairglasses Neon" replaces dual Snazzy/Voltage After Dark palettes across all 20+ consumers (FZF, cava, btop, yazi, kitty, hyprland, ironbar)
- **Chezmoi migration**: declarative dotfile management with symlink_ entries, palette data in `.chezmoidata.toml`, 6 lifecycle scripts (run_once_ for OMZ/vim-plug/TPM, run_onchange_ for theme/bat/systemd)
- **MCP consolidation**: tmux-mcp/systemd-mcp/process-mcp merged into dotfiles-mcp; eww/juhradial/makima tool handlers removed; CLAUDE.md updated to ~400 tools
- **Boot ordering**: swww-daemon.service for supervised wallpaper daemon, readiness polls replace sleep workarounds, ironbar ExecStartPre hyprctl poll
- **Ironbar theme pipeline**: wired into theme-sync.sh via `@import theme.generated.css`
- **Contract snapshot**: regenerated .well-known/mcp.json (removed 14 stale tools, added 15 new)
- **Stale reference sweep**: 50+ files cleaned across scripts, Go code, CI, docs, templates

### GitHub Stars Integration (2026-04-16)
Audited ~1,900 GitHub stars for dotfiles-relevant tools. Implemented:

- **hyprshade**: Config schedule + 5 MCP tools
- **wluma**: Adaptive brightness via ddcutil
- **cliphist**: Clipboard history + 4 MCP tools
- **zsh-auto-notify**: Long-running command notifications
- **kanshi**: Declarative display profiles
- **kitty-scrollback.nvim**: Neovim scrollback integration
- **glshell**: GLSL shader layershell overlay
- **MCP resources**: `shader://current`, `dotfiles://palette`, `validate-rice` prompt

## Planned

### Phase 1 — Chezmoi Hardening
- CI gate: add `chezmoi verify --source home/` to GitHub Actions
- Palette templates: template kitty/cava/btop/yazi configs from `.chezmoidata.toml` so palette changes auto-propagate
- `make lint` should include chezmoi verify

### Phase 2 — Ironbar Menubar Polish
- Cache-fed fleet widgets stay visible without blocking the GTK layer
- Workspace and focused-window modules stay legible on 5120x1440 ultrawide
- Keybind ticker stability and scroll performance
- Power, weather, and update affordances remain theme-aligned

### Phase 3 — Shader Pipeline
- DarkWindow shader performance benchmarks in CI (flag regressions above GPU budget)
- Shader preview gallery (static renders for README/docs)
- Wallpaper shader expansion (more procgen options via papertoy/shaderbg)
- Shader parameter presets (expose uniforms as config)

## Future Considerations

- **Status bar research**: evaluate GPU-capable alternatives to ironbar (ags, fabric, custom Wayland layer-shell bar)
- **hg-mcp extraction**: move `mcp/hg-mcp/` (319MB) to its own `hairglasses-studio/hg-mcp` repo
- **Docker MCP adoption**: install mcp-server-docker for container management
- **Automated rice screenshots**: Hyprland headless + screenshot comparison in CI
- **Cross-repo semantic search**: evaluate CodeMCP for indexing all 20 active repos
