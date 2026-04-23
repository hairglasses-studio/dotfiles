# MCP directory submission drafts

Prepared copy for submitting `dotfiles-mcp` to the three main MCP
discovery directories. **Not auto-submitted** — these files exist as
drafts for a human review-and-submit step. Each section below is
copy-paste-ready for the respective directory's form or PR.

## Server metadata (authoritative)

Pulled from `mcp/dotfiles-mcp/.well-known/mcp.json` and
`mcp/dotfiles-mcp/snapshots/contract/overview.json`:

| Field | Value |
|---|---|
| Name | `io.github.hairglasses-studio.dotfiles-mcp` |
| Version | 2.2.0 |
| Tool count | 434 |
| Module count | 41 |
| Resources | 25 |
| Prompts | 13 |
| Homepage | https://github.com/hairglasses-studio/dotfiles-mcp |
| Repo | https://github.com/hairglasses-studio/dotfiles-mcp |
| License | MIT |
| Language | Go |
| Categories | desktop, desktop_interact, discovery, github, hyprland, input_simulate, systemd, workflow |
| Tags | linux, desktop, hyprland, wayland, bluetooth, input, github-org, fleet-management, canonical-source |

## Short descriptions (copy-paste)

### 80-char hook (directory list headings)

> Hyprland + GitHub org + fleet-management tools for Linux workstation automation.

### 200-char pitch (card body)

> Canonical Linux workstation MCP server: 434 discovery-first tools across Hyprland IPC, desktop automation, Bluetooth/MIDI, Kitty visual pipeline, GitHub org lifecycle, fleet auditing, systemd control. MIT.

### Full description (long-form pages)

> `dotfiles-mcp` is the canonical MCP surface for a Linux workstation running Hyprland (Wayland). It ships 434 tools across 41 registered modules — all discovery-first, so the initial context load is ~85 % smaller than eagerly loading the full catalog. Categories include:
>
> - **Hyprland IPC**: 19 tools for screenshots, windowrules, monitor config, layers, keybinds, hotreload.
> - **Desktop automation**: atomic config writes with auto-backup, compositor abstraction, session orchestration.
> - **Bluetooth + MIDI**: pairing, trust, connection state, audio routing, MIDI device discovery.
> - **Kitty visual pipeline**: theme rotation, font switching, shader cycling, window tiling.
> - **GitHub org lifecycle**: bulk clone, settings sync, transfer, archive, star-lists taxonomy.
> - **Fleet auditing**: ralphglasses (multi-LLM orchestrator) status, roadmap snapshots.
> - **Systemd control**: list, status, restart, logs, timer management (user units).
>
> Built on [mcpkit](https://github.com/hairglasses-studio/mcpkit), the same Go MCP toolkit used in production server deployments. All tools bounded, permissioned, and observable via OpenTelemetry.

## Per-directory submission templates

### PulseMCP

Directory URL: https://www.pulsemcp.com/submit

Current submission note, verified 2026-04-23: PulseMCP asks for a server/client URL and says it ingests the Official MCP Registry daily, processes entries weekly, and accepts a GitHub repo, subfolder URL, or standalone website URL. Use the standalone publish mirror:

```text
https://github.com/hairglasses-studio/dotfiles-mcp
```

Form fields:

- **Server Name**: `dotfiles-mcp`
- **Homepage**: https://github.com/hairglasses-studio/dotfiles-mcp
- **Author**: hairglasses-studio
- **Language**: Go
- **License**: MIT
- **Tags**: `hyprland`, `wayland`, `linux`, `desktop-automation`, `github`, `fleet`, `systemd`
- **Short description**: (80-char hook above)
- **Long description**: (full description above)
- **Install snippet**:
  ```sh
  go install github.com/hairglasses-studio/dotfiles-mcp/cmd/dotfiles-mcp@latest
  ```
- **Transport**: stdio (default)
- **Config example**:
  ```json
  {
    "mcpServers": {
      "dotfiles": { "command": "dotfiles-mcp" }
    }
  }
  ```

### Glama

Directory URL: https://glama.ai/

Current submission note, verified 2026-04-23: Glama indexes open-source MCP servers from GitHub repositories and recommends adding `glama.json` metadata to control display name, description, category, environment variables, and build details. Use the standalone publish mirror:

```text
https://github.com/hairglasses-studio/dotfiles-mcp
```

Committed `glama.json` content in the standalone mirror:

```json
{
  "displayName": "dotfiles-mcp",
  "description": "Hyprland + GitHub org + fleet-management tools for Linux workstation automation.",
  "homepage": "https://github.com/hairglasses-studio/dotfiles-mcp",
  "repository": "https://github.com/hairglasses-studio/dotfiles-mcp",
  "language": "Go",
  "license": "MIT",
  "tags": ["hyprland", "wayland", "linux", "desktop-automation", "github", "fleet", "systemd", "bluetooth"],
  "categories": ["desktop", "automation", "linux", "github"],
  "transport": "stdio",
  "installCommand": "go install github.com/hairglasses-studio/dotfiles-mcp/cmd/dotfiles-mcp@latest",
  "author": {
    "name": "hairglasses-studio",
    "url": "https://github.com/hairglasses-studio"
  }
}
```

### MCP Market

Directory URL: https://mcpmarket.com/submit

Current submission note, verified 2026-04-23: MCP Market's submit page asks for the full GitHub repository URL for the MCP server and reviews it for inclusion. Submit the standalone publish mirror:

```text
https://github.com/hairglasses-studio/dotfiles-mcp
```

If it asks for a category, use `Desktop & System`.

## Checklist before submitting

- [x] Verified `hairglasses-studio/dotfiles` is public and non-archived on 2026-04-23
- [x] Verified `hairglasses-studio/dotfiles-mcp` is public, non-archived, transferred under `hairglasses-studio`, and installable on 2026-04-23
- [x] Tagged valid Go module release `v1.1.0`; `v2.2.0` is the server contract version, but the Go module path is still v1
- [x] Cleared standalone projection drift: clean `HEAD:mcp/dotfiles-mcp` export reports `status=in_sync`
- [x] Verified `go install github.com/hairglasses-studio/dotfiles-mcp/cmd/dotfiles-mcp@latest` resolves `v1.1.0`
- [x] `README.md` of `mcp/dotfiles-mcp` has a "What is MCP?" intro for discovery traffic
- [x] Confirm `.well-known/mcp.json` is externally crawlable: `https://raw.githubusercontent.com/hairglasses-studio/dotfiles/main/mcp/dotfiles-mcp/.well-known/mcp.json` returns name `io.github.hairglasses-studio.dotfiles-mcp`, version `2.2.0`, tool count `434`
- [x] Directory submit URLs verified on 2026-04-23
- [x] Screenshots or demo GIF ready (README has `docs/assets/ticker-demo.gif`)
- [ ] Roadmap item in ROADMAP.md marked done once listings appear
