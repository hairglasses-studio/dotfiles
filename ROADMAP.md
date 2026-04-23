# Roadmap

## Completed Marathon Phases

- 2026-04-21 `ticker`: Phase 7 — Phase 7 — input device integrations + smoke harness (`4b6c7dc`)

## Current State

Manjaro Linux dotfiles with 90+ managed configs (chezmoi + install.sh), 139 DarkWindow GLSL shaders with shuffled playlists, Hairglasses Neon palette applied to 20+ tools, Hyprland-first compositor automation, an Ironbar menubar on the theme pipeline, and a full boot stack (rEFInd + Plymouth). Idempotent installer with chezmoi declarative management (134 managed entries, 6 lifecycle scripts).

Single consolidated MCP server (dotfiles-mcp) with ~400 tools across 30+ modules. Shader collection is one of the largest curated GLSL terminal shader sets publicly available. All configs MIT licensed.

## Recently Completed

### Palette Playlist + Kitty Theme Validator (2026-04-21)
Turns the desktop color pipeline into a rotatable playlist and hardens the kitty theme rotation against typos:

- **Palette playlist** (`scripts/palette-playlist.sh`): `theme/palette.env` becomes a symlink into `theme/palettes/<name>.env` — 9 curated palettes (hairglasses-neon, amber, deep-purple, forest, ice, matrix, rose-pine, sunset, synthwave) filling an identical 23-token schema. `list / current / status / set / next / prev / random / reset / preview`, atomic symlink swap, first-time migration backup, state recorded in `$XDG_STATE_HOME/palette/`, delegates to `palette-propagate.sh` for reload.
- **Palette rules doc**: `.claude/rules/snazzy-palette.md` rewritten to token-first model with semantic invariants (`THEME_PRIMARY` always headline, `THEME_DANGER` always destructive) palettes must honor.
- **Kitty theme playlists**: 5 new curated rotations (cool, cyberpunk, high-contrast, pastel, warm) consumed by `kitty-shader-playlist.sh`.
- **Kitty playlist validator** (`scripts/kitty-playlist-validate.sh`): millisecond-fast edit-time check that resolves every playlist entry against `kitty/themes/themes.json` with fuzzy-match suggestions. Exposed via `~/.local/bin/kitty-playlist-validate`; wired into `tests/repo_smoke.bats` so `ci-bats` + `ci-smoke` enforce catalog parity on every push.

### RetroArch Workstation Tranche (2026-04-21)
Shared RetroArch profile drives core + BIOS + runtime automation:

- **Shared profile**: `~/hairglasses-studio/romhub/configs/device-profiles/retroarch.yaml` now carries the authoritative `retroarch_playlist_map` and `retroarch_requirements` rows (19 systems + 17 BIOS/helper entries). `scripts/lib/retroarch_profile.py` keeps built-in defaults as a safety net.
- **Audit** (`retroarch-workstation-audit`): cores, BIOS/helper assets, display profile, runtime status → `$XDG_STATE_HOME/retroarch/workstation-audit.json`.
- **BIOS/helper apply** (`retroarch-bios-apply`): populates missing required `dc/` (Flycast) and `PPSSPP/` (PSP helper assets via sparse clone of `hrydgard/ppsspp#assets`). `--dry-run` + `--source-dir` for local BIOS drops.
- **Cores**: pacman-first installer with optional AUR pass (`libretro-beetle-vb-git` for Virtual Boy, yay/paru auto-detected, `--skip-aur` for CI). `retroarch-build-libretro-cores` source-builds `race` (NGP) and `beetle-wswan` (WonderSwan) into `/usr/lib/libretro/` for the two cores with no package.
- **Runtime OSD** (`retroarch-apply-network-cmd`): atomic `retroarch.cfg` writer library with timestamped backup, UDP `VERSION` probe, `--revert`, `--dry-run`. Enables `retroarch-archive-homebrew-sync --notify-runtime` to deliver `SHOW_MSG` notifications.
- **Orchestrator** (`retroarch-archive-homebrew-sync`): chains manifest → fetch → import → playlists → audit with env-overridable tool paths for testing.
- **Tests** (`tests/retroarch_workstation.bats`): six cases covering profile-driven playlist mapping, zip-backed ROM counts, conditional BIOS, orchestrator wiring, atomic cfg writer with `--revert`, bios-apply dry-run, and source-build dry-run.
- **Docs**: `docs/retroarch-workstation.md` operator-facing reference with first-run workflow and state file map.

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

### Wayland Graphics Pipeline Consolidation (2026-04-16)

Cross-stack NVIDIA/compositor tuning, unified color propagation, template fan-out:

- **Phase 1 (NVIDIA + frame pacing)**: G-Sync/VRR enabled for 240Hz panel,
  `misc:vrr=2` fullscreen-only, per-monitor vrr flag in monitors.conf,
  `cursor:no_hardware_cursors=true` (NVIDIA fix), `render:explicit_sync=2`,
  `debug:damage_tracking=2`, `GBM_BACKEND=nvidia-drm`, `AQ_NO_MODIFIERS=1`,
  `ELECTRON_OZONE_PLATFORM_HINT=auto`. New `scripts/hypr-perf-mode.sh` for
  quality↔performance runtime toggle ($mod CTRL ALT Q).
- **Phase 2 (color pipeline)**: New `matugen/` directory with 9 envsubst
  templates (gtk-colors, kitty, hyprland, hyprlock, btop, yazi, zsh-fzf,
  cava). Rewrote `palette-propagate.sh` from stub to authoritative renderer —
  now actually updates 12 consumers instead of just 5. `theme-sync.sh`
  simplified to delegation. `rice-reload.sh` now restarts ticker.
- **Phase 5 (cleanup)**: Removed `borders/bordersrc` (macOS JankyBorders
  leftover), `hyprland/pyprland.toml` (stale duplicate of `pypr/config.toml`),
  `cava/shaders/` (unreachable under `method=ncurses`). Fixed duplicate
  `swaync` in Pacfile. Built `color_pipeline` + `perf_profile` skills (novel —
  no public Claude Code rice skills exist). New `.claude/rules/nvidia-wayland.md`
  documenting 2026 NVIDIA + Hyprland best practice. Updated `shaders.md` with
  `misc:vfr` decision rationale.
- **Phase 3 (shader consolidation)**: Tier playlists via
  `kitty/shaders/bin/shader-tier.sh` (45 cheap / 46 mid / 48 heavy, static
  size+loops heuristic). Deduplicated wallpaper renderer — `papertoy` removed
  as dead code (wasn't installed), `shaderbg` is canonical. Focus-driven
  hyprshade daemon at `scripts/hypr-shader-focus-daemon.sh` (opt-in service
  `dotfiles-hypr-shader-focus.service`). `misc:vfr` tension documented in
  `.claude/rules/shaders.md`. **Deferred**: runtime GPU-delta benchmarking in
  `shader_benchmark` MCP tool (requires live IPC + nvtop sampling).
- **Phase 4 (event-driven reload)**: `dotfiles-rice-watch.service` runs
  `inotifywait` on `palette.env` + `matugen/templates/` → auto-fires
  `palette-propagate.sh` on close_write. Ticker restart added to palette-
  propagate reload block. Ticker's in-process `Gio.FileMonitor` replaced by
  the simpler systemd-restart pattern (the 1553-line ticker has palette hex
  values baked in at code level; restart is ~500ms and unnoticeable).
- **Phase 6 (new MCP tools)**: `mod_wayland_perf.go` module with 5 tools —
  `hypr_perf_mode`, `hypr_vrr_status`, `hypr_frame_overlay`,
  `color_pipeline_apply`, `shader_tier`. Registered in discovery, contract
  snapshot regenerated (410→415 tools).

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

### Test Infra
- [x] [P2][S] `tests/repo_smoke.bats` / `hg mcp mirror parity check` — trimmed the parity manifest to the three still-mirrored modules; the consolidated record lives in the new `mcp/mirror-parity.json` `consolidated` array (commit `dbcddab`).
- [x] [P2][S] `tests/repo_smoke.bats ok 4` — trimmed the kitty-wrapper pinning grep to the three live consumer files (`44014e8`). The stale `hyprland/pyprland.toml` and makima Xbox controller paths are gone.
- [x] [P2][S] `tests/repo_smoke.bats ok 7` — replaced the retired `hg input --help` test with an equivalent `hg gamepad --help` assertion (`44014e8`). All 12 smoke tests now pass.
- [x] [P2][S] `tests/repo_smoke.bats ok 13` — walk install.sh `--print-link-specs`, filter to `retroarch-*` sources, assert each exists and has +x (`1753da3`). Landing the gate caught three scripts missing +x (sync/audit/install-workstation-cores) from the 2026-04-21 workstation commit.
- [x] [P2][S] `scripts/audit-orphan-scripts.sh` + `tests/repo_smoke.bats ok 13 (rev)` — walks `scripts/*.{sh,py}`, asserts each is referenced by at least one in-repo consumer (install.sh / Makefile / CI / systemd / configs / other scripts / docs / skills / bats); allowlisted categories for user-global Claude hooks + hardware helpers + manual-invoke utilities. Wired into ci-smoke in `--strict` mode (`e92670b`). Current state: orphans=0 scripts_scanned=176 consumers=903.
- [x] [P2][S] `tests/repo_smoke.bats ok 15` — repo-wide `bash -n` parse gate for `scripts/*.sh` + `scripts/lib/*.sh`, skipping the two non-bash shebangs (zsh, python3). Completes the parse-time triad alongside config-syntax (ok 13) and py_compile (ok 14). Landing the gate caught a real regression — `scripts/lib/agent-post-tool-audit.sh` had `; do` stranded on its own line inside a `for` continuation block (`c9297e7`).
- [x] [P2][S] `tests/repo_smoke.bats ok 18` — skill surface drift gate: loads `.agents/skills/surface.yaml`, diffs against `.agents/skills/` directory listing, fails on either side. Landing the gate caught the `ticker` skill dir that had been present since Phase 7 but never registered in surface.yaml, so chezmoi was silently skipping its publication (`62e5276`). Follow-up (`8f1d81b`) extended the gate to require each canonical SKILL.md frontmatter `name:` to match its directory, catching the one drift (`dotfiles_audit` dir with `name: dotfiles-audit` hyphen) that would have quietly broken skill routing.
- [x] [P2][S] CI trigger paths — `ci-bats.yml` + `ci-smoke.yml` now trigger on `systemd/**`, `.agents/skills/**`, `.claude/skills/**` so the systemd-analyze gate (ok 13) and skill-surface gate (ok 18) actually run when those surfaces change instead of riding on unrelated edits (`a0df984`).
- [x] [P2][S] `tests/repo_smoke.bats ok 20` — every unit in `install.sh --list-services` resolves to an on-disk file (`systemd/`, template `@.`, manjaro-only, or chezmoi `symlink_*` placeholder). Catches the silent-enable-failure class where a unit is added to the installer array but the file is missing or misnamed (`670121c`).
- [x] [P2][S] `tests/repo_smoke.bats ok 24` — every entry in `kitty/shaders/playlists/*.txt` resolves to a `.glsl` under `kitty/shaders/darkwindow/`. Parallel to the theme-playlist gate (ok 12); catches the class where a shader gets renamed or deleted but the playlist drops the entry silently (`7d78c25`).
- [x] [P2][S] `tests/repo_smoke.bats ok 26` — every `[shaders.<name>.presets.*]` block in `kitty/shaders/presets.toml` resolves to `<name>.glsl` under `kitty/shaders/darkwindow/`. Catches the drift where a shader rename leaves the MCP `shader_preset_apply` tool pointing at a ghost path (`4da9aa5`).
- [x] [P2][S] `tests/repo_smoke.bats ok 28` — compositor/bar configs that bypass the `~/.local/bin` symlink indirection and point straight at `$HOME/hairglasses-studio/dotfiles/<path>` must resolve to a tracked file. ok 15 only covers the wrapper path; direct refs in hyprland binds, ironbar cmd= lines, and pypr command= lines slip through and silently no-op on rename. Same iteration landed `shader_forge` in `surface.yaml` after ok 23 flagged the missing registration (`418988b`).
- [x] [P2][S] `tests/migrate_repo_crontab_paths.bats` — added the missing `scripts/migrate-repo-crontab-paths.sh` so the three orphaned bats cases go green (`e2a3a2a`). `--check` / `--apply` modes rewrite `chromecast4k/scripts/` → `hg-android/scripts/` in the user crontab.
- [x] [P2][M] `TestDesktopSessionListWindows_UnsupportedWithoutDBus` — root cause was a bug in `desktopSessionEnv`: it early-returned on empty record fields, so the OS-level `DBUS_SESSION_BUS_ADDRESS` from the live shell leaked into test environments that explicitly cleared it. The test's "no bus" scenario ended up running pyatspi anyway and hitting the Python 3.14 `__getattr__` regression. Fixed by filtering the key out when value is empty (`d43c021`). Full dotfiles-mcp test package is green.
- [x] [P2][S] `TestDotfilesWorkstationDiagnosticsDegradedWhenDesktopReadinessFails` — updated assertion from the stale `desktop.hyprland` component to `rice.hyprland`, matching the diagnostic namespace split that moved Hyprland-readiness into the `rice.*` bucket (`47f5649`).

### Blocked (needs external infrastructure)
- [ ] [BLOCKED: needs headless Hyprland] Shader: preview gallery with static renders
- [ ] [BLOCKED: needs headless Hyprland] Automated rice screenshot CI comparison

## Future Considerations

- **Shell evolution**: Quickshell (C++/QML, native GLSL ShaderEffect) is now staged as the bar/ticker/notification-history pilot. `dotfiles-quickshell.service` runs in parallel, `ticker-bridge.py` exposes existing ticker streams to QML with rotation/manual advance, `notification-bridge.py` exposes local notification history without taking D-Bus ownership, and `hg shell` controls pilot/cutover/rollback modes with JSON status for agents. Ironbar, swaync, and the Python ticker stay live until each replacement surface is verified.
- **Cross-repo semantic search**: evaluate CodeMCP for SCIP-based indexing across all 20 active repos

---

## Gap Research: Hook Infrastructure (2026-04-16)

Identified from GitHub research across 25+ Claude Code repos (60K+ combined stars). See `docs/research/agents/claude-code-skill-gap-research-2026-04-16.md` for full citations.

### Tier 1 — High Priority Hooks

- [x] [P1][M] Post-compaction re-anchor hook — scripts/claude-post-compact.sh (52-line re-anchor, needs settings.json hook registration)
- [x] [P1][M] File protection system — scripts/claude-file-protect.sh (PreToolUse hook blocking go.mod, pipeline.mk, .well-known, snapshots)
- [x] [P1][M] Circuit breaker for overnight loops — scripts/lib/circuit-breaker.sh (N-failure stop, no-ship streak, budget ceiling, rate limit detection)

### Tier 2 — Medium Priority Hooks

- [x] [P2][M] YAML ledger handoff hook — `scripts/claude-session-ledger.sh {write|read}` pairs Stop+SessionStart hooks. Write mode captures branch, HEAD, dirty files, recent commits, and activity (source/test writes, test runs) to ~/.cache/claude-session-ledger/project-$PROJECT/latest.yaml. Read mode injects it as additionalContext so the next session starts with the prior session's state. Keyed by basename($PWD) so repos don't collide. Ref: Continuous-Claude-v3
- [x] [P2][S] TDD enforcement hook — `scripts/claude-tdd-reminder.sh` is a PreToolUse advisory hook: on Go source writes, injects a systemMessage reminder unless a test file was written in-session or committed in the last hour. Generated-code and test-file writes stay silent. Opt-in via settings.json. Ref: nizos/tdd-guard, obra/superpowers
- [x] [P2][S] Verify-before-complete gate — `scripts/claude-verify-gate.sh` (Stop hook) + `scripts/claude-verify-track.sh` (PreToolUse tracker). Emits a systemMessage at stop time if the session wrote source without running tests. Tracks 6 test-runner patterns (go/cargo/pytest/npm/yarn/make). Idempotent — reminds once per session. Ref: obra/superpowers
- [ ] [P2][M] PostToolUse hook wiring — marathon completion events sync to docs-mcp roadmap state. Ref: autonomy gap analysis

### Tier 3 — Low Priority / Exploratory

- [ ] [P3][M] Skill auto-activation hook — PreToolUse detects project context (Go files, MCP config, shader GLSL) and injects relevant skill automatically without manual slash command. Ref: diet103 showcase, obra/superpowers

## Gap Research: New Skills (2026-04-16)

### Tier 1 — High Priority Skills

- [x] [P1][L] Security audit skill — SAST, supply-chain audit, spec-to-code compliance, second-opinion pattern for pre-publish security review. Ref: trailofbits/skills (professional security firm). Deployed to `.agents/skills/security_audit/SKILL.md`
- [x] [P1][S] Canary monitoring skill — post-deploy watch loop for MCP server health after git push / release. Ref: garrytan/gstack `/canary`. Deployed to `~/.claude/commands/canary.md` + `~/.agents/skills/canary/SKILL.md`

### Tier 2 — Medium Priority Skills

- [ ] [P2][M] Phase-gated pipeline — hard enforcement of plan -> human review -> implement -> verify phases in dev-loop; agents cannot skip steps. Ref: avifenesh/agentsys
- [x] [P2][S] Hidden assumption surfacer — `.agents/skills/common_ground/SKILL.md` deployed. Surfaces Claude's implicit priors about a repo (language, build system, test framework, CI, etc.), verifies each in parallel via fast file checks, reports confirmed/rebutted/unknown deltas, and prompts the user for redirect before code changes start. Read-only, 2-3 minute budget. Ref: jeffallan/claude-skills
- [x] [P2][S] Decision journal skill — `.agents/skills/decision_journal/SKILL.md` deployed. Appends ADR-lite entries (context, decision, rationale, alternatives, consequences) to `docs/decisions.md`, dedup by title + date, supports `--export` for stakeholder markdown tables. Ref: pcatattacks/solopreneur-plugin + Michael Nygard ADR template.
