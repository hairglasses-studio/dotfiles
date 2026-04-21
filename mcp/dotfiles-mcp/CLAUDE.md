# dotfiles-mcp

MCP server for dotfiles configuration management, GitHub org lifecycle, fleet auditing, and desktop service orchestration. Built with [mcpkit](https://github.com/hairglasses-studio/mcpkit).

## Build & Test
```bash
GOWORK=off go build ./...
GOWORK=off go vet ./...
GOWORK=off go test ./... -count=1
GOWORK=off go install .
```

## Tools (~420)

This server has grown to 30+ modules with ~420 tools (see `.well-known/mcp.json` for the exact count). Use the discovery tools (`dotfiles_tool_search`, `dotfiles_tool_catalog`, `dotfiles_tool_schema`) to browse the full live catalog. The sections below cover the core tools.

### Config Management (4)
- `dotfiles_list_configs` — List dotfiles config directories with symlink health and format
- `dotfiles_validate_config` — Validate TOML or JSON config syntax
- `dotfiles_reload_service` — Reload desktop service (hyprland, ironbar, mako, sway, tmux)
- `dotfiles_check_symlinks` — Check health of all expected dotfiles symlinks

### GitHub Org Lifecycle (12)
- `dotfiles_gh_list_personal_repos` — List personal repos with fork/visibility metadata
- `dotfiles_gh_list_org_repos` — List org repos with local clone sync status (supports language/archived/missing filters)
- `dotfiles_gh_transfer_repos` — Bulk transfer non-fork repos to org
- `dotfiles_gh_recreate_forks` — Squash forks into fresh org repos
- `dotfiles_gh_onboard_repos` — Fork public repos, squash history, clone locally (batch)
- `dotfiles_gh_local_sync_audit` — Audit local dirs vs org repos (orphaned/missing/mismatched)
- `dotfiles_gh_bulk_clone` — Clone all missing org repos locally
- `dotfiles_gh_pull_all` — Fetch/pull all local repos (detects dirty/detached)
- `dotfiles_gh_clean_stale` — Remove orphaned local clones (safety: checks uncommitted/unpushed)
- `dotfiles_gh_full_sync` — One-command fleet sync (pull + audit + clone missing)
- `dotfiles_gh_bulk_archive` — Batch archive repos
- `dotfiles_gh_bulk_settings` — Batch apply repo settings with before/after reporting

### Fleet Auditing & CI (4)
- `dotfiles_fleet_audit` — Per-repo language, Go version, CI status, test count, commit age, CLAUDE.md presence
- `dotfiles_health_check` — Org-wide health dashboard
- `dotfiles_dep_audit` — Go dependency version skew across fleet
- `dotfiles_workflow_sync` — Sync CI workflows from canonical sources

### Build & Sync (5)
- `dotfiles_pipeline_run` — Run build+test pipeline on a repo (Go/Node/Python)
- `dotfiles_bulk_pipeline` — Run pipeline across N repos with language filtering
- `dotfiles_go_sync` — Sync Go version across all repos
- `dotfiles_mcpkit_version_sync` — Sync mcpkit dependency across all thin MCP servers
- `dotfiles_create_repo` — Scaffold new repo with standard files

### Desktop (3)
- `dotfiles_cascade_reload` — Ordered multi-service reload with health verification
- `dotfiles_rice_check` — Compositor/shader/wallpaper/service status + Hairglasses Neon palette compliance
- `dotfiles_onboard_repo` — Add standard files to any repo (.editorconfig, CI, LICENSE)

### Hyprland Desktop (19)
- `hypr_list_windows` — List all windows with address, title, class, workspace
- `hypr_list_workspaces` — List workspaces with window count, monitor, focused status
- `hypr_get_monitors` — List monitors with resolution, refresh rate, position, scale
- `hypr_screenshot` — Capture screenshot (single monitor or all)
- `hypr_screenshot_monitors` — Capture separate screenshots per monitor
- `hypr_screenshot_window` — Capture a specific window by address or class (scale-aware, resized for LLM vision)
- `hypr_focus_window` — Focus window by address or class name
- `hypr_switch_workspace` — Switch to workspace by ID
- `hypr_reload_config` — Reload Hyprland config and check for errors
- `hypr_click` — Click at coordinates using ydotool
- `hypr_type_text` — Type text at cursor using wtype
- `hypr_key` — Send key events using ydotool
- `hypr_set_monitor` — Configure monitor resolution, position, or scale
- `hypr_move_window` — Move a window to exact pixel coordinates
- `hypr_resize_window` — Resize a window to exact pixel dimensions
- `hypr_close_window` — Close a window by address or class
- `hypr_toggle_floating` — Toggle floating state of a window
- `hypr_minimize_window` — Minimize a window to special:minimized workspace
- `hypr_fullscreen_window` — Toggle fullscreen/maximize for a window

### Shader Pipeline (13)
- `shader_list` — List GLSL shaders, optionally filter by category
- `shader_set` — Apply a DarkWindow shader and keep the paired Kitty theme state in sync
- `shader_cycle` — Advance the active Kitty visual playlist (next/prev)
- `shader_random` — Pick and apply a random paired Kitty visual from the active playlist
- `shader_status` — Current shader, Kitty theme, visual label, animation state, playlist position, auto-rotate
- `shader_meta` — Full manifest metadata (category, cost, source, playlists)
- `shader_test` — Compile-test shaders via glslangValidator
- `shader_build` — Preprocess and validate shaders
- `shader_playlist` — List playlists or pick a random paired visual from one
- `shader_get_state` — Read current Kitty shader and theme state
- `wallpaper_set` — Set a live wallpaper shader via shaderbg
- `wallpaper_random` — Set random wallpaper shader
- `wallpaper_list` — List available wallpaper shaders

### Bluetooth (9)
- `bt_list_devices` — List BT devices with connection status and battery levels
- `bt_device_info` — Detailed device info (battery, profiles, trust, UUIDs)
- `bt_scan` — Scan for nearby devices with configurable timeout (default 8s)
- `bt_pair` — Pair with interactive agent (BLE-safe, handles auth handshake). `remove_first` clears stale bonds
- `bt_connect` — Connect with BLE retry logic, resolves names against all known devices
- `bt_disconnect` — Disconnect a device
- `bt_remove` — Forget a paired device
- `bt_trust` — Trust or untrust a device
- `bt_power` — Toggle BT adapter power

### Input Devices (3)
- `input_detect_controllers` — Scan for gamepads with brand detection and makima profile status
- `input_generate_controller_profile` — Generate makima profile from template (desktop/gaming/media/macropad)
- `input_controller_test` — Detect controllers, generate missing profiles, optionally restart makima

### Input Services (2)
- `input_status` — Show running state of input services and MX Master battery status
- `input_restart_services` — Restart `mouse`, `controller`, or `all` input service groups

### MIDI (4)
- `midi_list_devices` — Detect connected USB MIDI controllers via ALSA
- `midi_generate_mapping` — Generate MIDI controller mapping config from template
- `midi_get_mapping` — Read existing MIDI controller mapping config
- `midi_set_mapping` — Create or update MIDI mapping (validates TOML)

### Composed Workflows (3)
- `bt_discover_and_connect` — **Composed**: scan→find→remove stale→pair (with agent)→trust→connect (with retry)
- `input_auto_setup_controller` — **Composed**: detect controllers→generate missing profiles→restart makima
- `dotfiles_repo_git_hygiene` — **Composed**: dry-run-first scan or cleanup for merged branches, extra worktrees, and managed worktree residue

### Open-Source Readiness (2)
- `dotfiles_oss_score` — Score a repo's open-source readiness (0-100) across 8 categories: community files, README quality, Go module, testing, CI/CD, security, release, maintenance. Returns structured report with per-check pass/fail and top action items.
- `dotfiles_oss_check` — Run checks for a single category with detailed suggestions

### SDLC Operations (21)
- `ops_build` — Build project (Go/Node/Python), parse compile errors into structured JSON
- `ops_test_smart` — Run tests on changed packages only (Go: go test -json, Node: jest --json, Python: pytest -v)
- `ops_changed_files` — List changed files with diff stats and Go package mapping
- `ops_analyze_failures` — Categorize build/test failures (type_error, missing_dep, timeout, etc.) with fix suggestions
- `ops_auto_fix` — Auto-fix mechanical failures: missing deps (go mod tidy), missing imports (goimports), unused vars. Dry-run by default
- `ops_branch_create` — Create feature branch with conventional naming (dry-run by default)
- `ops_commit` — Stage + commit with conventional message validation (dry-run by default)
- `ops_pr_create` — Push branch + create PR via gh CLI (dry-run by default)
- `ops_ci_status` — Poll GitHub Actions checks with optional wait (up to 5min)
- `ops_pre_push` — Gate: vet → lint → build → test (language-aware, short-circuits on failure)
- `ops_iterate` — **Core loop**: build → test → analyze → track iteration. Returns structured NextActions with file:line
- `ops_ship` — **Composed**: pre-push gate → commit → push → create PR (dry-run by default)
- `ops_revert` — Safely undo last commit (soft reset if unpushed, revert if pushed)
- `ops_session_create` — Create SDLC iteration tracking session (persisted to ~/.local/state/ops/)
- `ops_session_status` — Session stats: iterations, error trend, convergence detection
- `ops_session_list` — List active sessions (auto-cleans >7 days old)
- `ops_session_handoff` — Generate Agent Handoff Protocol document from session + git state
- `ops_fleet_diff` — Fleet-wide changes since a date: per-repo commits, churn, commit types, authors
- `ops_tech_debt` — Score tech debt 0-100 across 6 dimensions with fleet mode and trend tracking
- `ops_research_check` — Search docs knowledge base for existing research with gap detection
- `ops_iteration_patterns` — Mine historical sessions for common failures, convergence rates, hot files

### Sandbox Testing (13)
- `sandbox_create` — Create Docker container with GPU (nvidia-container-toolkit)
- `sandbox_start` — Start container, wait for Hyprland ready, auto-resize to 2560x1440
- `sandbox_stop` — Stop a running sandbox
- `sandbox_destroy` — Remove container and config dir
- `sandbox_list` — List sandboxes with status
- `sandbox_status` — GPU utilization, memory, CPU stats
- `sandbox_sync` — Deploy dotfile symlinks + reload Hyprland inside container
- `sandbox_test` — Run test suite (bats/selftest/symlinks/shaders/config)
- `sandbox_exec` — Execute command inside sandbox
- `sandbox_diff` — Compare symlink health inside container
- `sandbox_screenshot` — Capture Hyprland display via grim, return base64 PNG
- `sandbox_visual_diff` — Compare screenshot against reference via ImageMagick
- `sandbox_validate` — **Composed**: create → sync → test → screenshot → destroy

### Discovery (10)
- `dotfiles_tool_search` — Keyword search across the tool catalog
- `dotfiles_tool_catalog` — Browse tools by category with deferred hints
- `dotfiles_tool_schema` — Inspect one tool's full descriptor/schemas
- `dotfiles_tool_stats` — Counts: total tools, modules, deferred, resources, prompts
- `dotfiles_server_health` — Contract shape: profile, version, tool counts
- `dotfiles_desktop_status` — Full desktop control readiness
- `dotfiles_launcher_audit` — Launcher and switcher diagnostics
- `dotfiles_bar_audit` — Status bar diagnostics
- `dotfiles_workstation_diagnostics` — Workstation health check
- `dotfiles_workspace_scene` — Live workspace state snapshot

### Audio (5)
- `audio_status`, `audio_volume`, `audio_mute`, `audio_devices`, `audio_device_switch`

### Network (6)
- `network_status`, `network_connections`, `network_dns`, `network_wifi_list`, `network_wifi_connect`, `network_vpn_toggle`

### System (7)
- `system_info`, `system_memory`, `system_disk`, `system_gpu`, `system_temps`, `system_uptime`, `system_health_check`

### Systemd (11)
- `systemd_status`, `systemd_start`, `systemd_stop`, `systemd_restart`, `systemd_enable`, `systemd_disable`, `systemd_logs`, `systemd_list_units`, `systemd_list_timers`, `systemd_failed`, `systemd_restart_verify`

### Tmux (11)
- `tmux_list_sessions`, `tmux_list_windows`, `tmux_list_panes`, `tmux_new_session`, `tmux_new_window`, `tmux_send_keys`, `tmux_capture_pane`, `tmux_kill_session`, `tmux_search_panes`, `tmux_wait_for_text`, `tmux_workspace`

### Process (8)
- `ps_list`, `ps_tree`, `port_list`, `investigate_port`, `investigate_service`, `kill_process`, `gpu_status`, `system_updates`

### Claude Sessions (16)
- Session scan, search, detail, logs, health, replay, compare, tag, crash detection, fleet recovery, workspace snapshots, repo status/diff/roadmap

### Desktop Interaction (18)
- Screenshot, OCR, click, find, type, form fill, accessibility tree, window management

### Desktop Sessions (30)
- AT-SPI automation: connect, start, stop, launch app, find/click/focus elements, type text, set values, read values, clipboard, screenshots, D-Bus calls, wait for elements

### Arch Linux (13)
- Package search/info, AUR search, updates, orphans, mirror status, file owner, pacman log, PKGBUILD audit, ArchWiki search/page, news

### Prompt Registry (8)
- `prompt_capture`, `prompt_search`, `prompt_get`, `prompt_improve`, `prompt_score`, `prompt_export`, `prompt_tag`, `prompt_stats`

### Kitty Terminal (23)
- Window/tab management, text send/get, layout, font size, theme, opacity, image display, remote commands, config

### GitHub Stars (16)
- Star list management, membership, audit, sync, taxonomy, cleanup, bootstrap

### Mapping Engine (11)
- Profile management, template generation, import/export, validation, diff, migration, daemon control, monitoring

## Key Patterns
- All batch/write tools use dry-run by default (`execute: true` for live mode)
- `bulk_settings` reports previous state before applying changes
- `clean_stale` checks for uncommitted/unpushed work before deletion
- `pull_all` detects dirty repos and detached HEAD, skips safely
- Composed "tool-of-tools" (full_sync, fleet_audit, cascade_reload, rice_check, bulk_pipeline) eliminate multi-step token waste
