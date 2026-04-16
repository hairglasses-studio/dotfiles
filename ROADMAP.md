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

### Chezmoi & Palette Pipeline (2026-04-16)
- Chezmoi CI gate added to ci-lint.yml (`chezmoi verify --source home/`)
- Palette propagation script (`scripts/palette-propagate.sh`) + chezmoi `run_onchange_` trigger
- Docker MCP adopted (mcp-server-docker via uvx in `.mcp.json`)
- hg-mcp extracted to standalone `hairglasses-studio/hg-mcp` repo (319MB removed from dotfiles)

## Planned

### Phase 1 — Ironbar Menubar Polish
- [x] [P1][M] Ironbar: cache-fed fleet widgets via systemd timers (weather, updates, MX battery)
- [x] [P1][M] Ironbar: workspace and focused-window modules legible on 5120x1440 ultrawide
- [x] [P1][S] Ironbar: keybind ticker min-width to prevent layout jumps
- [x] [P1][S] Ironbar: widget colors aligned to semantic palette conventions

### Phase 2 — Shader Pipeline
- [x] [P2][S] Shader CI: glslangValidator validation workflow for DarkWindow + wallpaper shaders
- [x] [P2][S] Shader CI: README badge count verification gate
- [x] [P2][M] Shader: 3 new wallpaper shaders (void-pulse, hex-matrix, nebula-drift)
- [x] [P2][M] Shader: parameter presets exposing uniforms as config (presets.toml + 2 MCP tools)
- [x] [P2][S] MCP: `dotfiles_write_config` tool — atomic write + validate + backup + reload
- [x] [P2][S] MCP: `shader_benchmark` tool — glslangValidator compile time + file size benchmarking
- [x] [P2][S] MCP: `shader://categories` resource — category breakdown

### Cyberpunk Ticker Bar (2026-04-16)
Standalone GTK4 PangoCairo 240Hz scrolling ticker replacing ironbar script-based version:

- [x] [P1][L] Pixel-smooth scrolling via DrawingArea + `add_tick_callback`
- [x] [P1][M] 10-layer visual effects: water caustic, neon glow, gradient, scanlines, text outline, wave distortion, glitch/CA, shadow
- [x] [P1][M] 7 content streams: keybinds, system, fleet, weather, github, notifications, music
- [x] [P1][S] Click-to-copy keybinds via wl-copy, scroll wheel speed control
- [x] [P1][S] Layer-shell production mode via gtk4-layer-shell (systemd service)
- [x] [P1][S] 4 effect presets: ambient, cyberpunk, minimal, clean
- [x] [P2][S] Playlist persistence across restarts
- [x] [P2][S] `/ticker` skill for management
- [x] [P2][S] `capture-window-gif.sh` helper with output-crop for layer-shell surfaces

### Phase 3 — Public Content
- [x] [P1][S] README: add use-case section with 5 concrete workflow examples
- [x] [P1][S] README: add "Install MCP Server Only" section with go install one-liner
- [x] [P1][S] README: add skills table listing all 14 skills
- [x] [P1][S] GitHub Topics: add hyprland mcp wayland dotfiles desktop-automation linux go
- [x] [P2][S] Submit PR to awesome-hyprland (IPC section) — hyprland-community/awesome-hyprland#178
- [x] [P2][S] Submit PR to awesome-mcp-servers — punkpeye/awesome-mcp-servers#4958
- [x] [P2][S] Update .well-known/mcp.json with categories and tags
- [ ] [P2][M] Record 30-sec demo GIF for README
- [ ] [P3][M] Blog post: "Controlling Hyprland with an AI Agent via MCP"
- [ ] [P3][S] Submit to PulseMCP, Glama, MCP Market directories

### Blocked (needs external infrastructure)
- [ ] [BLOCKED: needs headless Hyprland] Shader: preview gallery with static renders
- [ ] [BLOCKED: needs headless Hyprland] Automated rice screenshot CI comparison

## Future Considerations

- **Status bar evolution**: Quickshell (C++/QML, native GLSL ShaderEffect) is the best GPU-capable alternative to ironbar — see `docs/STATUS-BAR-RESEARCH.md` for full evaluation of 7 alternatives. Ironbar stays short-term; prototype Quickshell on secondary monitor.
- **Cross-repo semantic search**: evaluate CodeMCP for SCIP-based indexing across all 20 active repos

---

## Gap Research: Hook Infrastructure (2026-04-16)

Identified from GitHub research across 25+ Claude Code repos (60K+ combined stars). See `docs/research/agents/claude-code-skill-gap-research-2026-04-16.md` for full citations.

### Tier 1 — High Priority Hooks

- [x] [P1][M] Post-compaction re-anchor hook — scripts/claude-post-compact.sh (52-line re-anchor, needs settings.json hook registration)
- [x] [P1][M] File protection system — scripts/claude-file-protect.sh (PreToolUse hook blocking go.mod, pipeline.mk, .well-known, snapshots)
- [x] [P1][M] Circuit breaker for overnight loops — scripts/lib/circuit-breaker.sh (N-failure stop, no-ship streak, budget ceiling, rate limit detection)

### Tier 2 — Medium Priority Hooks

- [ ] [P2][M] YAML ledger handoff hook — Stop hook writes structured YAML to `.claude/sessions/{id}.yaml` capturing session state, decisions, progress; SessionStart hook reads and injects. Ref: Continuous-Claude-v3 (3.7K stars)
- [ ] [P2][S] TDD enforcement hook — PreToolUse file-write interception: if `*.go` (non-test) changes but no `*_test.go` changed in same session, inject TDD reminder context. Ref: nizos/tdd-guard
- [ ] [P2][S] Verify-before-complete gate — hook preventing task completion without running tests first. Ref: obra/superpowers
- [ ] [P2][M] PostToolUse hook wiring — marathon completion events sync to docs-mcp roadmap state. Ref: autonomy gap analysis

### Tier 3 — Low Priority / Exploratory

- [ ] [P3][M] Skill auto-activation hook — PreToolUse detects project context (Go files, MCP config, shader GLSL) and injects relevant skill automatically without manual slash command. Ref: diet103 showcase, obra/superpowers

## Gap Research: New Skills (2026-04-16)

### Tier 1 — High Priority Skills

- [x] [P1][L] Security audit skill — SAST, supply-chain audit, spec-to-code compliance, second-opinion pattern for pre-publish security review. Ref: trailofbits/skills (professional security firm). Deployed to `.agents/skills/security_audit/SKILL.md`
- [x] [P1][S] Canary monitoring skill — post-deploy watch loop for MCP server health after git push / release. Ref: garrytan/gstack `/canary`. Deployed to `~/.claude/commands/canary.md` + `~/.agents/skills/canary/SKILL.md`

### Tier 2 — Medium Priority Skills

- [ ] [P2][M] Phase-gated pipeline — hard enforcement of plan -> human review -> implement -> verify phases in dev-loop; agents cannot skip steps. Ref: avifenesh/agentsys
- [ ] [P2][S] Hidden assumption surfacer — `/common-ground` skill revealing Claude's implicit priors about the repo before starting work. Ref: jeffallan/claude-skills
- [ ] [P2][S] Decision journal skill — auto-record architectural decisions with rationale, exportable for docs. Ref: pcatattacks/solopreneur-plugin
