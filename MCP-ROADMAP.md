# MCP Tool Roadmap

Research conducted 2026-04-02 across GitHub ecosystem. Evaluated against current
dotfiles stack: Hyprland/Sway, Ghostty (144 shaders), eww, Go/Python/Bash,
5 custom MCP servers (86+ tools), mcpkit framework.

## Current Fleet (5 servers, 86+ tools)

| Server | Lang | Tools | Domain |
|--------|------|-------|--------|
| hyprland-mcp | Go | 12 | Compositor, screenshots, input, windows |
| ~~sway-mcp~~ | ~~Node~~ | ~~12~~ | ~~Removed — Sway deprecated, Hyprland only~~ |
| dotfiles-mcp | Go | 30 | Config, GitHub sync, eww, health, pipelines |
| shader-mcp | Go | 11 | GLSL shaders, playlists, benchmarks |
| input-mcp | Go | 29 | Bluetooth, controllers, MIDI, mouse, keyboard |

---

## Tier 1 — Adopt Now (install from GitHub)

### arch-mcp
- **Repo:** https://github.com/nihalxkumar/arch-mcp (35 stars)
- **Lang:** Python (uvx)
- **Tools:** Arch Wiki search, AUR package lookup, official repo queries, update checking, PKGBUILD analysis
- **Why:** Only Arch Linux MCP server in existence. Replaces manual `yay -Ss`, `pacman -Qi`, and wiki browsing during rice development. Directly useful for package discovery when adding new tools.
- **Claude substitute:** Partial — `Bash` tool can run pacman/yay but can't search the Arch Wiki or analyze PKGBUILDs with structured output.
- **Install:** `uvx arch-mcp` or pip
- **Priority:** **High**

### mcp-server-docker
- **Repo:** https://github.com/ckreiling/mcp-server-docker (695 stars)
- **Lang:** Python (uvx)
- **Tools:** 16 tools — containers (list/create/run/stop/logs), images (pull/build/push), networks, volumes, compose via natural language
- **Why:** Docker is part of the dev stack (lazydocker, dive, colima on macOS). This exposes full container lifecycle as MCP tools. The `docker_compose` prompt lets Claude compose multi-container setups conversationally.
- **Claude substitute:** Partial — `Bash` tool can run docker commands but lacks structured container introspection, stats resources, and the compose prompt.
- **Install:** `uvx mcp-server-docker`
- **Priority:** **High**

### mcp-server-kubernetes
- **Repo:** https://github.com/Flux159/mcp-server-kubernetes (1,371 stars)
- **Lang:** TypeScript (npx)
- **Tools:** 20+ tools — kubectl CRUD, Helm install/upgrade/uninstall, port-forward, rollout management, pod cleanup, node management, k8s-diagnose prompt
- **Why:** k9s, helm, stern, kubectx all in the dotfiles. Waybar has a k8s context module. This replaces manual kubectl for agent-driven cluster management. Non-destructive mode and secrets masking are important safety features.
- **Claude substitute:** Partial — `Bash` can run kubectl but lacks structured output, secrets masking, and the diagnostic prompt.
- **Install:** `npx mcp-server-kubernetes`
- **Priority:** **High** (if running K8s on homelab)

### mcp-gopls
- **Repo:** https://github.com/hloiseaufcms/mcp-gopls (75 stars)
- **Lang:** Go
- **Tools:** Go code analysis via gopls LSP — find references, implementations, call hierarchy, test execution, coverage reporting
- **Why:** All custom MCP servers are Go. This wraps gopls to provide semantic Go analysis — find callers of a function, get coverage for a package, run specific tests. Complements `go vet`/`go test` with IDE-level intelligence.
- **Claude substitute:** Limited — Claude can read Go code and run `go test` but can't do call graph analysis, find all callers, or get test coverage maps without gopls.
- **Install:** `go install` from source
- **Priority:** **High**

### godoc-mcp-server
- **Repo:** https://github.com/yikakia/godoc-mcp-server (34 stars)
- **Lang:** Go
- **Tools:** Query Go package documentation from pkg.go.dev — search packages, get function signatures, read doc comments
- **Why:** Frequently need to look up mcpkit, mcp-go, BurntSushi/toml API docs when writing MCP server code. This provides structured Go doc access without leaving the conversation.
- **Claude substitute:** Partial — `WebFetch` can read pkg.go.dev pages but output is noisy. This gives clean structured doc access.
- **Install:** `go install` from source
- **Priority:** **Medium**

---

## Tier 2 — Evaluate (promising, may adopt)

### mcp_server_notify
- **Repo:** https://github.com/Cactusinhand/mcp_server_notify (51 stars)
- **Lang:** Python
- **Tools:** Desktop notifications with sound effects when agent tasks complete
- **Why:** With ralphglasses orchestrating parallel agent sessions across repos, getting notified when long builds/tests finish would be valuable. Would fire through mako on Hyprland.
- **Claude substitute:** Can `Bash` run `notify-send`, but this provides a clean MCP interface usable by any client.
- **Priority:** **Medium**

### mcp-server-desktop-notify
- **Repo:** https://github.com/gbrigandi/mcp-server-desktop-notify (2 stars)
- **Lang:** Rust
- **Tools:** 5 notification tools — basic, urgent, with-icon, with-timeout, rich (urgency, category, app name)
- **Why:** Richer API than mcp_server_notify. Linux gets the best feature set (XDG notifications). Rust binary = low overhead.
- **Alternative to:** mcp_server_notify. Pick one.
- **Priority:** **Medium**

### CodeMCP
- **Repo:** https://github.com/SimplyLiz/CodeMCP (80 stars)
- **Lang:** Go
- **Tools:** 80+ tools — semantic code search, impact analysis, call graphs, cross-language (Go, TS, Python, Rust, Java via SCIP indexing)
- **Why:** Could index the entire hairglasses-studio org for cross-repo semantic search. Find all callers of a mcpkit function across 44 repos.
- **Claude substitute:** Grep/Glob tools do text search but can't do semantic call graph analysis.
- **Priority:** **Medium**

### mcp-server-pacman (package registries)
- **Repo:** https://github.com/oborchers/mcp-server-pacman (11 stars)
- **Lang:** Python
- **Tools:** Search PyPI, npm, crates.io, Docker Hub, Terraform Registry for packages
- **Why:** Despite the name, this queries language-level package registries, not Arch pacman. Useful for discovering Go/Python/Node dependencies.
- **Claude substitute:** `WebSearch` and `WebFetch` can browse registries but with more noise.
- **Priority:** **Low**

### stt-mcp-server-linux
- **Repo:** https://github.com/marcindulak/stt-mcp-server-linux (22 stars)
- **Lang:** Python (Docker)
- **Tools:** Local speech-to-text via Whisper, push-to-talk (Right Ctrl), tmux text injection
- **Why:** Voice-driven coding sessions. Linux-only. Requires Docker + `/dev/input` + `/dev/snd` access. Experimental but aligns with the cyberpunk aesthetic.
- **Priority:** **Low**

### terminator
- **Repo:** https://github.com/mediar-ai/terminator (1,384 stars)
- **Lang:** Rust
- **Tools:** Desktop GUI automation via accessibility APIs — click, type, find elements by label/role, screenshot. No vision models needed.
- **Why:** Could automate GUI apps (Firefox, Chromium) that aren't controllable via Hyprland IPC. Accessibility tree gives structured UI element access.
- **Claude substitute:** hyprland-mcp provides ydotool/wtype for blind input but can't inspect GUI element trees.
- **Priority:** **Low**

---

## Tier 3 — Build In-House (zero competition, high value)

These don't exist in the MCP ecosystem. Build them using mcpkit.

### systemd-mcp
- **Tools to build:**
  - `systemd_status` — `systemctl status <unit>` with structured output
  - `systemd_start/stop/restart/enable/disable` — unit lifecycle
  - `systemd_logs` — `journalctl -u <unit> --since` with filtering
  - `systemd_list_timers` — active timer overview
  - `systemd_list_failed` — failed units
  - `systemd_user_services` — `--user` scope (shader-rotate, makima, eww-calendar-sync, etc.)
- **Why:** 8 systemd user services in the dotfiles. Zero MCP servers for systemd exist anywhere on GitHub. Unique competitive advantage.
- **Effort:** Small (wrapper around systemctl/journalctl)
- **Priority:** **High**

### tmux-mcp
- **Tools to build:**
  - `tmux_list_sessions` — active sessions with window counts
  - `tmux_send_keys` — send keystrokes to a pane
  - `tmux_capture_pane` — read pane contents (structured output)
  - `tmux_new_session/window/pane` — create layout
  - `tmux_attach` — list attachable sessions
- **Why:** tmux.service runs at boot. ralphglasses orchestrates agent sessions. No mature tmux MCP exists (only the STT server tangentially touches it).
- **Effort:** Small
- **Priority:** **Medium**

### process-mcp
- **Tools to build:**
  - `ps_list` — processes with structured output (PID, CPU%, MEM%, CMD)
  - `ps_tree` — process tree for a PID
  - `kill_process` — SIGTERM/SIGKILL with confirmation
  - `port_in_use` — what's listening on a port
  - `gpu_status` — nvidia-smi structured output (temp, utilization, VRAM)
- **Why:** Near-zero competition. `cuba-exec` (2 stars) and a Windows-only task manager are the only things that exist.
- **Effort:** Small
- **Priority:** **Low**

---

## Already Covered (skip these)

| Category | Covered By | GitHub alternatives |
|----------|-----------|-------------------|
| Screenshots | hyprland-mcp, sway-mcp | prosperity-solutions/mcp-server-screenshot (macOS only) |
| Clipboard | sway-mcp (`clipboard_read/write`) | mcp-wayland-clipboard (1 tool, redundant) |
| Bluetooth | input-mcp (10 bt_* tools) | Nothing exists |
| Shaders | shader-mcp (11 tools) | Nothing exists |
| Wayland/Compositor | hyprland-mcp, sway-mcp | wayland-mcp (kurojs), hyprmcp — both smaller |
| Input devices | input-mcp (29 tools) | Nothing exists |
| Config validation | dotfiles-mcp | Nothing comparable |
| GitHub sync | dotfiles-mcp (10 gh_* tools) | github/github-mcp-server (complementary, not redundant) |
| eww widgets | dotfiles-mcp (3 eww_* tools) | Nothing exists |

## GitHub MCP vs Built-in gh CLI

The official `github/github-mcp-server` (28,499 stars, Go) is the largest MCP server by stars. However, Claude Code already has deep `gh` CLI integration via the Bash tool, and your dotfiles-mcp has 10 `gh_*` tools for org sync, bulk operations, and fleet management. The official GitHub MCP would add:
- Structured PR review with diff parsing
- Code search across orgs (vs grep-based local search)
- Workflow run management

**Verdict:** Complementary but not essential. The `gh` CLI + dotfiles-mcp covers 90% of the use cases. Add only if you need cross-org code search or structured PR review beyond what `gh pr view` provides.

---

## Installation Plan

### Phase 1 — Quick wins (pip/npx install)
```bash
# arch-mcp
pip install arch-mcp  # or uvx

# Docker MCP
uvx mcp-server-docker

# K8s MCP (if running k8s)
npx mcp-server-kubernetes
```

Add to `.mcp.json`:
```json
{
  "mcpServers": {
    "arch": {
      "command": "uvx",
      "args": ["arch-mcp"]
    },
    "docker": {
      "command": "uvx",
      "args": ["mcp-server-docker"]
    },
    "kubernetes": {
      "command": "npx",
      "args": ["-y", "mcp-server-kubernetes"]
    }
  }
}
```

### Phase 2 — Go servers (build from source)
```bash
# mcp-gopls
cd ~/hairglasses-studio && git clone https://github.com/hloiseaufcms/mcp-gopls
cd mcp-gopls && go build ./...

# godoc-mcp-server
git clone https://github.com/yikakia/godoc-mcp-server
cd godoc-mcp-server && go build ./...
```

### Phase 3 — Build in-house
```bash
# systemd-mcp (new mcpkit server)
cd ~/hairglasses-studio
# Use dotfiles-mcp as template, add systemd module
```

---

## Discovery Resources

- **[awesome-mcp-servers](https://github.com/punkpeye/awesome-mcp-servers)** — 84,106 stars, the definitive curated list
- **[glama.ai/mcp/servers](https://glama.ai/mcp/servers)** — Web directory for MCP servers
- **[github/github-mcp-server](https://github.com/github/github-mcp-server)** — 28,499 stars, official GitHub MCP
- **[modelcontextprotocol/servers](https://github.com/modelcontextprotocol/servers)** — Official reference implementations
