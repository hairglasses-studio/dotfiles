# Next Sprint Prompt — Dotfiles Development

## Session Summary (Previous Sprint)

**90 commits across 3 sessions** transformed a macOS-only dotfiles repo into a cross-platform cyberpunk rice with full Wayland support.

### What Was Built

**Infrastructure:**
- Hyprland 0.54 config (320 lines, 89 keybinds, 0.54 block windowrule syntax, NVIDIA optimized)
- Sway config consolidated to full parity with Hyprland (75 keybinds)
- Compositor abstraction library (`scripts/lib/compositor.sh` — 7 functions) + config library (`scripts/lib/config.sh` — 4 functions)
- Feature-flag system (`dotfiles.toml` — 42 toggleable components, 4 profiles: minimal/dev/full/cyberpunk)

**Visual Stack:**
- 5 procgen wallpaper shaders (cyber-rain, neon-grid, plasma-flow, fractal-pulse, particle-aurora) via shaderbg
- wlogout power menu (5 neon-colored buttons with per-button glow)
- Mako notifications (tighter, bar-aware, urgency-colored)
- Wofi launcher (hover accents, selection glow)
- GTK/Qt system theming (Adwaita/Breeze Dark + Papirus icons + Bibata cursor + Monaspace/Maple split)

**Boot:**
- rEFInd bootloader with Matrix cyberpunk theme (replaces GRUB)
- Plymouth animated boot splash (proxzima theme)
- Silent boot (`quiet splash loglevel=3`)
- NVIDIA kernel params (modeset + PreserveVideoMemoryAllocations)

**MCP Servers (27 tools):**
- `hyprland-mcp` (Go, 9 tools) — screenshot, windows, workspaces, input, config reload
- `dotfiles-mcp` (Go, 4 tools) — list configs, validate TOML/JSON, reload service, check symlinks
- `shader-mcp` (Go, 5 tools) — list/set/random/test/get_state for 132 shaders
- `sway-mcp` (Node.js, 9 tools) — screenshot, windows, input, clipboard

**Claude Code Integration:**
- 3 skills: `/rice-check`, `/screenshot-review`, `/shader-browse`
- 2 agents: `rice-developer` (opus, autonomous iteration), `config-validator` (haiku, syntax check)
- 3 rules: `hyprland.md` (0.54 syntax), `shaders.md` (self-contained GLSL), `snazzy-palette.md` (color enforcement)
- PostToolUse hook (best-effort Go formatting and live reload for Hyprland/swaync/eww/sway/bat config writes)
- Notification hook (alert on task completion)

**Keyboards:**
- Keychron V1 Ultra 8K encoder firmware (F13/F14 for focus cycling)
- Drop CTRL v1/v2 firmware (PrtSc/ScrLk/Pause trio → F13/Enter/F14)
- Both integrated into Sway + Hyprland keybinds

**Documentation:**
- `docs/RESOURCES.md` — 40+ awesome lists indexed by component
- `CLAUDE.md` — comprehensive architecture docs
- `keyboard/README.md` + `keyboard/drop-ctrl.md` — firmware flashing guides

---

## Next Sprint Prompt

Paste this into a new Claude Code session:

---

```
You are continuing development on my cross-platform cyberpunk dotfiles rice at ~/hairglasses-studio/dotfiles. This is a 90-commit, 35-component Manjaro Linux + macOS setup with Hyprland (primary) on Wayland, 139+ DarkWindow GLSL shaders, the Hairglasses Neon palette, and Go MCP servers (~400 tools total).

Read CLAUDE.md for architecture. Read docs/RESOURCES.md for community reference links. Read docs/SPRINT-NEXT.md for full session history and context.

The dotfiles use a feature-flag system (dotfiles.toml) with 42 toggleable components and 4 profiles (minimal/dev/full/cyberpunk). The install script (manjaro/install.sh) has is_enabled() guards on every symlink.

Key files:
- hyprland/hyprland.conf — Hyprland compositor (0.54 block windowrule syntax)
- scripts/lib/compositor.sh — Shared compositor detection library
- .mcp.json — MCP servers (hyprland, dotfiles, shader, input)
- .claude/skills/ — 3 skills (rice-check, screenshot-review, shader-browse)
- .claude/agents/ — 2 agents (rice-developer, config-validator)
- .claude/rules/ — 3 rules (hyprland, shaders, snazzy-palette)
- dotfiles.toml — Feature flags for all 42 components
- docs/RESOURCES.md — 40+ awesome-list community indexes

MCP servers are Go binaries in sibling repos:
- ~/hairglasses-studio/hyprland-mcp/ (12 tools: screenshot, windows, workspaces, input, config)
- ~/hairglasses-studio/dotfiles-mcp/ (~400 tools: list, validate, reload, symlinks, and more)
- ~/hairglasses-studio/shader-mcp/ (11 tools: list, set, random, test, state)

Hardware: AMD Ryzen 9 7950X, RTX 3090, 96GB RAM, dual monitors (Samsung 5120x1440 ultrawide + XEC 2560x1440), 7.3TB NVMe.

Remaining work tracked in memory (mcp_tool_ideas.md):
- MCP server enhancements: add resources (dotfiles://palette, shader://current), prompts (validate-rice, check-palette), new tools (shader_cycle, shader_benchmark, hypr_screenshot_region)
- Hyprland plugins to install: borders-plus-plus (config ready), hyprexpo (workspace overview), hyprls (LSP for config validation)
- Game-changing community projects to evaluate: pywal, telescope.nvim, zinit, zellij
- Install script unification (macOS install.sh still separate from manjaro/install.sh)
- Git pre-commit hooks for config validation

Start by: pulling latest changes, taking a screenshot to assess the current visual state, then entering plan mode to determine priorities.
```

---

## Key Patterns to Preserve

- **Atomic config writes**: `mktemp + mv` pattern (scripts/lib/config.sh)
- **Compositor detection**: `compositor_type` function returns hyprland/aerospace
- **Hyprland 0.54 syntax**: Block windowrules with `name =` field, `windowrulev2` is deprecated
- **Kitty terminal**: Primary terminal; use `kitty` class in all compositor window rules
- **GPG signing**: Currently disabled (`git -c commit.gpgsign=false`) due to key migration — re-enable when new key is configured
- **MCP env vars**: Servers need `WAYLAND_DISPLAY` and `HYPRLAND_INSTANCE_SIGNATURE` injected via .mcp.json bash wrappers
- **Hairglasses Neon palette**: deep near-black surfaces, cyan/magenta/green neon accents — source of truth at theme/palette.env
