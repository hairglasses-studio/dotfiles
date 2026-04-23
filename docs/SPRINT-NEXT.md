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

**MCP Surface (current):**
- `dotfiles-mcp` — canonical Hyprland/workstation MCP server with 434 tools across 41 modules in the checked-in contract snapshot
- `mapitall` — controller/MIDI mapping MCP runtime
- `mapping` — shared Go package for mapping/profile resolution
- Retired standalone handlers (`systemd-mcp`, `tmux-mcp`, `process-mcp`) are consolidated into `dotfiles-mcp`

**Claude Code Integration:**
- 3 skills: `/rice-check`, `/screenshot-review`, `/shader-browse`
- 2 agents: `rice-developer` (opus, autonomous iteration), `config-validator` (haiku, syntax check)
- 3 rules: `hyprland.md` (0.54 syntax), `shaders.md` (self-contained GLSL), `snazzy-palette.md` (color enforcement)
- PostToolUse hook (auto-reload Hyprland/mako/waybar on config writes)
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
You are continuing development on my cross-platform cyberpunk dotfiles rice at ~/hairglasses-studio/dotfiles. This is a Manjaro Linux + macOS setup with Hyprland (primary) on Wayland, 326 DarkWindow GLSL shaders, the Hairglasses Neon palette, and a consolidated Go MCP control plane. The checked-in dotfiles-mcp contract currently exposes 434 tools across 41 modules.

Read AGENTS.md first; it is the canonical instruction file. Then read docs/ARCHITECTURE-PROVENANCE.md for ownership boundaries and docs/RESOURCES.md for community reference links.

The dotfiles use a feature-flag system (dotfiles.toml) with 42 toggleable components and 4 profiles (minimal/dev/full/cyberpunk). The install script (manjaro/install.sh) has is_enabled() guards on every symlink.

Key files:
- hyprland/hyprland.conf — Hyprland compositor (0.54 block windowrule syntax)
- hyprland/monitors.conf — static monitor layout layered with dynamic monitor state
- scripts/ and scripts/lib/ — shared operator scripts, validation gates, and automation libraries
- .mcp.json — MCP server launcher config
- .agents/skills/ — canonical workflow skill surface
- .claude/skills/ — generated compatibility mirror of workflow skills
- mcp/dotfiles-mcp/ — canonical desktop/workstation MCP server source
- dotfiles.toml — Feature flags for all 42 components
- docs/RESOURCES.md — 40+ awesome-list community indexes

MCP modules are developed in this repo under mcp/:
- mcp/dotfiles-mcp/ — desktop config, Hyprland, systemd, tmux, process, input, GitHub, Kitty, and fleet tooling
- mcp/mapitall/ — controller/MIDI mapping MCP runtime
- mcp/mapping/ — shared Go package for mapitall mapping semantics

The standalone dotfiles-mcp/mapitall/mapping repos are publish mirrors. Mirror policy lives in docs/MCP-MIRROR-PARITY.md and mcp/mirror-parity.json.

Hardware: AMD Ryzen 9 7950X, RTX 3090, 96GB RAM, dual monitors (Samsung 5120x1440 ultrawide + XEC 2560x1440), 7.3TB NVMe.

Remaining work is tracked in ROADMAP.md:
- Public launch follow-through: submit dotfiles-mcp to PulseMCP, Glama, and MCP Market using the verified standalone mirror
- Hyprland plugins to install: borders-plus-plus (config ready), hyprexpo (workspace overview), hyprls (LSP for config validation)
- Game-changing community projects to evaluate: pywal, telescope.nvim, zinit, zellij
- Blocked until headless Hyprland exists: shader preview gallery and automated rice screenshot CI comparison

Start by: checking git status, reading ROADMAP.md, running the narrow validator for your intended surface, then implementing the highest-value non-blocked item.
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
