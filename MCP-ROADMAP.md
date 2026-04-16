# MCP Tool Roadmap

Research conducted 2026-04-02, updated 2026-04-16 after major consolidation.
Current stack: Hyprland, Kitty (139 DarkWindow shaders), ironbar, Go/Bash,
dotfiles-mcp (~400 tools), mcpkit framework.

## Current Fleet (2 modules, ~400 tools)

| Module | Lang | Tools | Domain |
|--------|------|-------|--------|
| dotfiles-mcp | Go | ~400 | Config, Hyprland, Kitty, shaders, Bluetooth, input, mapping, systemd, tmux, process, ops, audio, network, Arch Linux, system, GitHub stars, sandbox, desktop interaction, desktop sessions, Claude sessions, prompt registry, screen, clipboard, notifications, hyprshade, theme |
| mapitall | Go | — | Input device mapping daemon (udev rules, systemd service) |

All previously separate servers (hyprland-mcp, sway-mcp, shader-mcp, input-mcp, systemd-mcp, tmux-mcp, process-mcp) are now consolidated into dotfiles-mcp as modules.

---

## Tier 1 — Adopt (install from GitHub)

### ~~arch-mcp~~ — Built In-House
- **Status:** Completed as `ArchModule` in dotfiles-mcp (13 tools)
- Tools: `arch_package_search`, `arch_aur_search`, `arch_package_info`, `arch_updates_dry_run`, `arch_orphans`, `arch_mirror_status`, `arch_file_owner`, `arch_pacman_log`, `arch_pkgbuild_audit`, `arch_news_latest`, `arch_news_critical`, `archwiki_search`, `archwiki_page`

### mcp-server-docker
- **Repo:** https://github.com/ckreiling/mcp-server-docker (695 stars)
- **Lang:** Python (uvx)
- **Tools:** 16 — containers, images, networks, volumes, compose
- **Why:** Docker is part of the dev stack. Full container lifecycle as MCP tools.
- **Status:** Not adopted. Still valid.
- **Install:** `uvx mcp-server-docker`

### mcp-server-kubernetes
- **Repo:** https://github.com/Flux159/mcp-server-kubernetes (1,371 stars)
- **Lang:** TypeScript (npx)
- **Tools:** 20+ — kubectl CRUD, Helm, port-forward, rollout management
- **Status:** Not adopted. Valid if running K8s on homelab.
- **Install:** `npx mcp-server-kubernetes`

### ~~mcp-gopls~~ — Deprioritized
- Partially covered by Context7 MCP + built-in LSP tool.

### ~~godoc-mcp-server~~ — Deprioritized
- Covered by Context7 MCP with pre-resolved library IDs.

---

## Tier 2 — Evaluate

### CodeMCP
- **Repo:** https://github.com/SimplyLiz/CodeMCP (80 stars)
- **Lang:** Go
- **Tools:** 80+ — semantic code search, impact analysis, call graphs via SCIP
- **Why:** Could index the entire hairglasses-studio org (20 active repos) for cross-repo semantic search.
- **Status:** Not adopted. Evaluate when cross-repo refactoring becomes a bottleneck.

### terminator
- **Repo:** https://github.com/mediar-ai/terminator (1,384 stars)
- **Lang:** Rust
- **Tools:** Desktop GUI automation via accessibility APIs
- **Why:** Could automate GUI apps not controllable via Hyprland IPC.
- **Status:** Partially covered by dotfiles-mcp DesktopSemanticModule (AT-SPI automation).

---

## Tier 3 — Build In-House — ALL COMPLETED

| Server | Status | Module | Tools |
|--------|--------|--------|-------|
| ~~systemd-mcp~~ | Done | `SystemdModule` | 11 tools: status, start, stop, restart, enable, disable, logs, list_units, list_timers, failed, restart_verify |
| ~~tmux-mcp~~ | Done | `TmuxModule` | 11 tools: list_sessions, list_windows, list_panes, new_session, new_window, send_keys, capture_pane, kill_session, search_panes, wait_for_text, workspace |
| ~~process-mcp~~ | Done | `ProcessModule` | 8 tools: ps_list, ps_tree, port_list, investigate_port, investigate_service, kill_process, gpu_status, system_updates |

---

## Already Covered

| Category | Covered By | Notes |
|----------|-----------|-------|
| Screenshots | dotfiles-mcp (HyprlandModule, ScreenModule) | hypr_screenshot, hypr_screenshot_monitors, hypr_screenshot_window, hypr_screenshot_region, screen_screenshot, screen_ocr |
| Clipboard | dotfiles-mcp (ClipboardModule) | clipboard_read, clipboard_write, clipboard_read_image, cliphist_list/get/delete/clear |
| Bluetooth | dotfiles-mcp (BluetoothModule) | 9 bt_* tools + bt_discover_and_connect composed workflow |
| Shaders | dotfiles-mcp (ShaderModule) | 18 shader_*/wallpaper_* tools |
| Compositor | dotfiles-mcp (HyprlandModule + HyprlandExtModule) | 37 hypr_* tools |
| Input devices | dotfiles-mcp (InputModule, ControllerModule, MidiModule) | 20 input_*/midi_* tools |
| Config validation | dotfiles-mcp (DotfilesModule) | dotfiles_validate_config, dotfiles_check_symlinks |
| GitHub sync | dotfiles-mcp (DotfilesModule) | 12 dotfiles_gh_* tools |
| Arch Linux | dotfiles-mcp (ArchModule) | 13 arch_*/archwiki_* tools |
| Audio | dotfiles-mcp (AudioModule) | 5 audio_* tools |
| Network | dotfiles-mcp (NetworkModule) | 6 network_* tools |
| System info | dotfiles-mcp (SystemModule) | 7 system_* tools |
| Desktop automation | dotfiles-mcp (DesktopInteractModule, DesktopSemanticModule, DesktopSessionModule) | 48 desktop_*/session_* tools |
| SDLC ops | dotfiles-mcp (OpsModule) | 21 ops_* tools |

---

## Remaining Adoption Plan

### Docker MCP (when needed)
```bash
uvx mcp-server-docker
```
Add to `.mcp.json` when container management becomes a daily workflow.

### Cross-Repo Semantic Search (evaluate)
Clone and test CodeMCP against the hairglasses-studio org to see if SCIP indexing provides value beyond Grep/Glob.

---

## Discovery Resources

- **[awesome-mcp-servers](https://github.com/punkpeye/awesome-mcp-servers)** — curated MCP server list
- **[glama.ai/mcp/servers](https://glama.ai/mcp/servers)** — web directory
- **[github/github-mcp-server](https://github.com/github/github-mcp-server)** — official GitHub MCP
- **[modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)** — reference implementations
