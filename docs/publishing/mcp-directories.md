# MCP directory submission drafts

Prepared copy for submitting `dotfiles-mcp` to the main MCP discovery
directories. **Not all surfaces are safely auto-submittable** — browser-gated
forms and account-backed registry publication remain human review-and-submit
steps. Each section below is copy-paste-ready for the respective directory,
form, email, or registry package follow-up.

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

## Submission status (verified 2026-04-23)

| Surface | Status | Next action |
|---|---|---|
| Standalone mirror | Ready: public, non-archived, projection-clean (`status=in_sync`), valid Go module tag `v1.1.0`, GitHub Release published, and MCPB registry artifact released | Use https://github.com/hairglasses-studio/dotfiles-mcp as the public submit URL |
| Official MCP Registry | Ready for authenticated publish: standalone `server.json` validates with `mcp-publisher validate` and points at the public MCPB release artifact | Run `mcp-publisher login github`, then `mcp-publisher publish` |
| PulseMCP | Server submission path points to the Official MCP Registry; no stable direct server POST endpoint is exposed | Publish to the Official MCP Registry first, then wait for PulseMCP ingestion or use PulseMCP contact for manual adjustments |
| Glama | Repo-side metadata complete: `glama.json` is committed in the standalone mirror | Wait for GitHub indexing or submit the repo URL through Glama's browser flow |
| MCP Market | Ready for browser submission; CLI requests hit a Vercel security checkpoint | Submit https://github.com/hairglasses-studio/dotfiles-mcp through the browser form |

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

### Official MCP Registry

Registry docs: https://modelcontextprotocol.io/registry/quickstart

Current publication note, verified 2026-04-23: the official registry is the
upstream path PulseMCP prefers for server submissions. The standalone mirror now
contains `server.json` and a public Linux/amd64 MCPB package release:

```text
https://github.com/hairglasses-studio/dotfiles-mcp/releases/download/mcpb-2.2.0/dotfiles-mcp_2.2.0_linux_amd64.mcpb
```

`server.json` uses the official registry name
`io.github.hairglasses-studio/dotfiles-mcp` because the registry schema and
GitHub auth namespace require exactly one slash. The checked-in
`.well-known/mcp.json` still carries the canonical contract name
`io.github.hairglasses-studio.dotfiles-mcp`.

Validation already passed:

```text
mcp-publisher validate
✅ server.json is valid
```

Remaining manual command:

```bash
mcp-publisher login github
mcp-publisher publish
```

### PulseMCP

Directory URL: https://www.pulsemcp.com/submit

Current submission note, verified 2026-04-23: selecting **MCP Server** on the
PulseMCP submit page shows the Official MCP Registry path rather than a direct
server form. PulseMCP says it ingests the Official MCP Registry daily and
processes entries weekly. The `/submit` URL form is under the alternate client
path, so do not use it for this server.

Once the Official MCP Registry package blocker is cleared, use the standalone
publish mirror:

```text
https://github.com/hairglasses-studio/dotfiles-mcp
```

Manual adjustment payload if PulseMCP asks for details:

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

Current submission note, verified 2026-04-23: MCP Market's submit page asks
for the full GitHub repository URL for the MCP server and reviews it for
inclusion. CLI requests currently hit a Vercel security checkpoint, so this is
a browser submission step. Submit the standalone publish mirror:

```text
https://github.com/hairglasses-studio/dotfiles-mcp
```

If it asks for a category, use `Desktop & System`.

## Checklist before submitting

- [x] Verified `hairglasses-studio/dotfiles` is public and non-archived on 2026-04-23
- [x] Verified `hairglasses-studio/dotfiles-mcp` is public, non-archived, transferred under `hairglasses-studio`, and installable on 2026-04-23
- [x] Tagged valid Go module release `v1.1.0`; `v2.2.0` is the server contract version, but the Go module path is still v1
- [x] Published GitHub Release `v1.1.0`: https://github.com/hairglasses-studio/dotfiles-mcp/releases/tag/v1.1.0
- [x] Published public MCPB registry artifact: https://github.com/hairglasses-studio/dotfiles-mcp/releases/tag/mcpb-2.2.0
- [x] Validated standalone `server.json` with `mcp-publisher validate`
- [x] Updated standalone repo description, homepage, and GitHub topics for directory discovery
- [x] Cleared standalone projection drift: clean `HEAD:mcp/dotfiles-mcp` export reports `status=in_sync`
- [x] Verified `go install github.com/hairglasses-studio/dotfiles-mcp/cmd/dotfiles-mcp@latest` resolves `v1.1.0`
- [x] Committed standalone `glama.json` metadata for Glama indexing
- [x] `README.md` of `mcp/dotfiles-mcp` has a "What is MCP?" intro for discovery traffic
- [x] Confirm `.well-known/mcp.json` is externally crawlable: `https://raw.githubusercontent.com/hairglasses-studio/dotfiles/main/mcp/dotfiles-mcp/.well-known/mcp.json` returns name `io.github.hairglasses-studio.dotfiles-mcp`, version `2.2.0`, tool count `434`
- [x] Directory submit URLs verified on 2026-04-23
- [x] Screenshots or demo GIF ready (README has `docs/assets/ticker-demo.gif`)
- [ ] Browser-submit MCP Market and optional Glama listing if crawler pickup is slow
- [ ] Run authenticated Official MCP Registry publish with `mcp-publisher login github` and `mcp-publisher publish`
- [ ] Roadmap item in ROADMAP.md marked done once listings appear
