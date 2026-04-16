# Roadmap

## Current State

Manjaro Linux dotfiles with 60+ symlinked configs, 138 GLSL shaders with shuffled playlists, Snazzy-on-black palette applied to 15+ tools, Hyprland-first compositor automation, an Ironbar-first menubar, and a full boot stack (rEFInd + Plymouth). Idempotent installer with backup/restore support.

Shader collection is one of the largest curated GLSL terminal shader sets publicly available. All configs MIT licensed.

## Recently Completed

### GitHub Stars Integration (2026-04-16)
Audited ~1,900 GitHub stars for dotfiles-relevant tools. Identified 10 already-integrated, 14 actionable, 10 deferred. Implemented:

- **hyprshade**: Config schedule + 5 MCP tools (`hyprshade_list/set/toggle/off/status`)
- **wluma**: Adaptive brightness via ddcutil with time-of-day ALS config
- **cliphist**: Clipboard history backend + 4 MCP tools (`cliphist_list/get/delete/clear`)
- **zsh-auto-notify**: Long-running command notifications via swaync
- **kanshi**: Declarative display config with home-dual/single/portable profiles
- **kitty-scrollback.nvim**: Open kitty scrollback in Neovim (lazy.nvim plugin)
- **hyprlax**: Parallax wallpaper mode in wallpaper-mode.sh
- **papertoy**: Shadertoy-compatible animated wallpaper renderer
- **glshell**: GLSL shader layershell overlay (cyberpunk rain effect)
- **Hyprchroma**: Chroma key window transparency plugin
- **MCP resources**: `shader://current`, `dotfiles://palette`, `validate-rice` prompt, `hypr_screenshot_region` tool

## Planned

### Phase 1 — Linux Installer Hardening
- Keep `install.sh` Linux-only and catch non-Manjaro drift early
- Tighten package validation for pacman/yay/metapac
- Hyprland-specific installer steps (ironbar, mako, wofi, wlogout setup)
- Automated symlink validation in CI

### Phase 2 — Ironbar Menubar Polish
- Menubar restart and recovery path stays reliable at login and hot reload time
- Cache-fed fleet widgets stay visible without blocking the GTK layer
- Workspace and focused-window modules stay legible on mixed-density monitors
- Power, weather, and update affordances remain theme-aligned without writable bar config drift

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

<!-- whiteclaw-rollout:start -->
## Whiteclaw-Derived Improvements (2026-04-08)

Recommendations seeded from the restored whiteclaw snapshot research, prompt-pack audit, MCP explorer packaging review, and current fleet gap scan.

### Recommended Work
- [x] Audit AGENTS.md against the actual build/test/release loop and keep `CLAUDE.md`, `GEMINI.md`, and Copilot instructions as thin mirrors of the canonical guidance.
- [x] Write or refresh a searchable architecture/provenance note so future cross-repo research does not depend on raw code spelunking alone.
- [x] Audit the existing `.agents/skills/` surface for stale workflows, missing references, and opportunities to split generic steps into sharper skills.
- [x] Bootstrap a minimal `.ralph` loop with verification gates, cost observations, and improvement journaling so the repo can participate in controlled autonomous sweeps.
- [x] Expand handler/CLI/MCP integration coverage around the most user-facing surfaces and add runnable examples for the public entrypoints.
- [x] Prefer typed contracts for tools, commands, and workflow inputs at system boundaries instead of ad ad hoc maps or implicit structs.
- [x] Add lightweight validation (shell lint, config checks, link checks, JSON/schema validation, or snapshot verification) that matches the actual artifact types in the repo.
- [x] Harden public-facing docs, examples, and release notes so outside consumers can discover the intended workflow without org-private context.
- [x] Add search-oriented architecture notes for installer flow, shader pipeline, and desktop control-plane integration so the repo is navigable without spelunking.

### Rationale Snapshot
- Tier: `tier-1`
- Lifecycle: `active`
- Language profile: `Go/Shell/Config`
- Visibility / sensitivity: `PUBLIC` / `public`
- Surface gaps: skills=`no`, codex=`no`, ralph=`yes`, roadmap=`no`

<!-- whiteclaw-rollout:end -->
