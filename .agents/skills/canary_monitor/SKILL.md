---
name: canary_monitor
description: Post-deploy canary watch loop for MCP server and desktop service health after git push or release. Polls build status, service health, and contract drift, then reports pass/fail. Use after shipping changes to dotfiles-mcp, after system updates, after rice-reload, after any push that touches MCP modules or systemd services, or when the user says "canary", "health check after push", "watch the deploy", or "is it still healthy".
allowed-tools:
  - Bash
  - Read
  - Grep
  - mcp__dotfiles__dotfiles_server_health
  - mcp__dotfiles__dotfiles_rice_check
  - mcp__dotfiles__dotfiles_desktop_status
  - mcp__dotfiles__dotfiles_check_symlinks
  - mcp__dotfiles__dotfiles_workstation_diagnostics
  - mcp__dotfiles__systemd_failed
  - mcp__dotfiles__systemd_status
  - mcp__dotfiles__shader_status
  - mcp__dotfiles__hypr_get_config_errors
  - mcp__dotfiles__ops_ci_status
  - mcp__dotfiles__events_tail
---

# Canary Monitor

Post-deploy health watch loop. Run after pushing changes to verify nothing broke. Polls progressively — fast checks first, slower checks if fast ones pass.

## When to Run

- After `git push` that touches MCP modules (`mcp/dotfiles-mcp/`)
- After `git push` that touches systemd services (`systemd/`)
- After `git push` that touches Hyprland config (`hyprland/`)
- After `git push` that touches shell configs (`zsh/`, `kitty/`)
- After `rice-reload` or `dotfiles_cascade_reload`
- After system updates (`topgrade`, `pacman -Syu`)
- On demand: user says "canary", "health check", "is it working"

## Check Sequence

Run checks in this order. Stop and report on first failure.

### Tier 1: Build (fast, ~5s)
```bash
cd mcp/dotfiles-mcp && GOWORK=off go build ./...
```
If build fails, the push broke compilation. Report immediately.

### Tier 2: MCP Server Health (~5s)
- `dotfiles_server_health` — contract shape, version, tool counts
- Compare tool count against `.well-known/mcp.json` expected count
- If tool count dropped, a module may have failed to register

### Tier 3: Desktop Services (~10s)
- `systemd_failed` — any user services in failed state
- `systemd_status` for critical services: `ironbar`, `dotfiles-keybind-ticker`, `dotfiles-cliphist`, `dotfiles-kanshi`, `swww-daemon`
- `hypr_get_config_errors` — Hyprland config parse errors after reload

### Tier 4: Rice Integrity (~10s)
- `dotfiles_rice_check` — compositor, shader, wallpaper, service status, palette compliance
- `shader_status` — current shader is active and valid
- `dotfiles_check_symlinks` — no broken symlinks

### Tier 5: CI Status (~30s, optional)
- `ops_ci_status` — GitHub Actions check status for latest push
- Only run if the push was recent (within last 10 minutes)

### Tier 6: Event-bus liveness + pending remediations (~3s)
- Check `~/.claude/recovery-events.jsonl` with `tail -n 5`. If it contains
  a line with `"type":"mcp_dead"` from within the last 5 minutes, the
  MCP watchdog is reporting the server crashed — surface this immediately.
- Check `stat -c %Y ~/.local/state/dotfiles/events.jsonl`. If the mtime
  is older than 2 minutes, the dotfiles-event-bus is stale or stopped.
- Call `events_tail {since_minutes: 10, severity: "high"}`. Any result
  is a pending, high-severity remediation — report count + suggest running
  `/heal` before declaring the canary green.

## Reporting

After all checks pass:
```
Canary: ALL CLEAR (5/5 tiers passed)
  Build: OK (410 tools compiled)
  MCP: OK (410 tools, 24 resources, 13 prompts)
  Services: OK (7/7 active)
  Rice: OK (palette compliant, shader active)
  CI: OK (3/3 checks passed) or SKIPPED (no recent push)
```

On failure:
```
Canary: FAILURE at Tier N
  [tier name]: FAILED — [specific error]
  Recommendation: [what to do]
```

## Watch Mode

If invoked with "watch" or as part of a loop:
1. Run the full check sequence
2. If all pass, report and schedule next check in 60s (3 checks total)
3. If any fail, report immediately and stop watching
4. After 3 consecutive passes, report "stable" and stop

## Integration with Dev-Loop

The canary skill complements `/dev-loop` ship phase. After the loop commits and pushes, it can invoke canary to verify the deploy landed cleanly before moving to the next task. This prevents shipping a broken change and then building more broken changes on top.
