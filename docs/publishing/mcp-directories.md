# MCP directory submission drafts

Prepared copy for submitting `dotfiles-mcp` to the three main MCP
discovery directories. **Not auto-submitted** — these files exist as
drafts for a human review-and-submit step. Each section below is
copy-paste-ready for the respective directory's form or PR.

## Server metadata (authoritative)

Pulled from `mcp/dotfiles-mcp/.well-known/mcp.json`:

| Field | Value |
|---|---|
| Name | `io.github.hairglasses-studio.dotfiles-mcp` |
| Version | 2.2.0 |
| Tool count | 419 |
| Homepage | https://github.com/hairglasses-studio/dotfiles/tree/main/mcp/dotfiles-mcp |
| Repo | https://github.com/hairglasses-studio/dotfiles |
| License | MIT |
| Language | Go |
| Categories | desktop, desktop_interact, discovery, github, hyprland, input_simulate, systemd, workflow |
| Tags | linux, desktop, hyprland, wayland, bluetooth, input, github-org, fleet-management, canonical-source |

## Short descriptions (copy-paste)

### 80-char hook (directory list headings)

> Hyprland + GitHub org + fleet-management tools for Linux workstation automation.

### 200-char pitch (card body)

> Canonical Linux workstation MCP server: 419 discovery-first tools across Hyprland IPC (19), desktop automation, Bluetooth/MIDI, Kitty visual pipeline, GitHub org lifecycle, fleet auditing, systemd control. MIT.

### Full description (long-form pages)

> `dotfiles-mcp` is the canonical MCP surface for a Linux workstation running Hyprland (Wayland). It ships 434 tools + 25 resources across 30+ modules — all discovery-first, so the initial context load is ~85 % smaller than the upstream equivalents. Categories include:
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

Directory URL: https://www.pulsemcp.com/servers/submit

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
  go install github.com/hairglasses-studio/dotfiles-mcp@latest
  # or clone the dotfiles repo + run make install-mcp
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

Directory URL: https://glama.ai/mcp/servers (PR against https://github.com/glama-ai/mcp-servers-registry)

JSON manifest entry (append to the registry):

```json
{
  "name": "dotfiles-mcp",
  "displayName": "dotfiles-mcp",
  "description": "Hyprland + GitHub org + fleet-management tools for Linux workstation automation.",
  "homepage": "https://github.com/hairglasses-studio/dotfiles/tree/main/mcp/dotfiles-mcp",
  "repository": "https://github.com/hairglasses-studio/dotfiles",
  "language": "Go",
  "license": "MIT",
  "tags": ["hyprland", "wayland", "linux", "desktop-automation", "github", "fleet", "systemd", "bluetooth"],
  "categories": ["desktop", "automation", "linux", "github"],
  "transport": "stdio",
  "installation": {
    "npm": null,
    "go": "go install github.com/hairglasses-studio/dotfiles-mcp@latest",
    "docker": null
  },
  "author": {
    "name": "hairglasses-studio",
    "url": "https://github.com/hairglasses-studio"
  }
}
```

### MCP Market

Directory URL: https://mcp.market (submission via web form)

Form fields mirror PulseMCP; paste the same description + tags. The
one MCP Market-specific field is "Category" — use `Desktop & System`.

## Checklist before submitting

- [ ] Verify `dotfiles-mcp` repo at `hairglasses-studio/dotfiles-mcp` is public + pinned
- [ ] Tag a fresh release (v2.2.0 or later) — directory crawlers follow latest release
- [ ] `README.md` of `dotfiles-mcp` has a "What is MCP?" intro for discovery traffic
- [ ] Confirm `.well-known/mcp.json` is served at the well-known URL via GitHub Pages or similar (some directories crawl it)
- [ ] Screenshots or demo GIF ready (README already has `docs/assets/ticker-demo.gif`)
- [ ] Roadmap item in ROADMAP.md marked done once listings appear
