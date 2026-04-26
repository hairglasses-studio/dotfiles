---
name: heal
description: Triage and auto-apply pending remediations queued by the dotfiles-event-bus daemon. Reads ~/.local/state/dotfiles/events.jsonl, looks up each event's structured Remediation via remediation_lookup, presents the pending fixes to the user, and applies the ones they approve. Use when the user says "heal", "auto-fix pending issues", "what's broken", "apply remediations", or after a swaync notification shows an event-bus alert fired. Also use it before running a /canary to clear known-safe events first.
allowed-tools:
  - Bash
  - Read
  - mcp__dotfiles__events_tail
  - mcp__dotfiles__remediation_lookup
  - mcp__dotfiles__remediation_list
  - mcp__dotfiles__hypr_config_list
  - mcp__dotfiles__hypr_config_rollback
  - mcp__dotfiles__hypr_reload_config
  - mcp__dotfiles__ops_auto_fix
  - mcp__dotfiles__ops_lint_fix
  - mcp__dotfiles__ops_test_smart
  - mcp__dotfiles__keybinds_refresh_ticker
  - mcp__dotfiles__systemd_restart
  - mcp__dotfiles__systemd_restart_verify
  - mcp__dotfiles__systemd_reset_failed
  - mcp__dotfiles__dotfiles_reload_service
  - mcp__dotfiles__dotfiles_mcpkit_version_sync
  - mcp__dotfiles__color_pipeline_apply
  - mcp__dotfiles__audio_device_switch
  - mcp__dotfiles__shader_random
  - mcp__dotfiles__notify_send
---

# /heal

Triage pending remediations from the event bus and apply the approved ones.

`systemd_reset_failed` is not a remediation registry target; it is kept as a
human-approved escape hatch when a failed unit needs clearing before another
restart attempt.

## The flow

1. **Tail the event bus** — call `events_tail` with `since_minutes=30, limit=50`. If the log doesn't exist yet, report "event bus not running" and stop; ask the user whether to `systemctl --user enable --now dotfiles-event-bus.service`.

2. **Group events by error_code** — many events with the same code usually share one remediation; surface them together with a count.

3. **Look up remediations** — for each unique `error_code`, call `remediation_lookup {code}`. Events without a registered remediation go in a "no auto-fix" bucket the user can review manually.

4. **Present the plan** — concise per-group summary:
   ```
   hypr_reload_induced_drm  (3 events, last 00:02 UTC)
     → hypr_config_rollback {name: "latest"}    [risk=reload]
     Why: A kernel DRM error appeared within 5s of a Hyprland reload.

   ticker_stale  (1 event, 00:01 UTC)
     → keybinds_refresh_ticker                  [risk=reload]
   ```

5. **Ask per-group** unless the user said "auto" — then auto-apply every remediation where `risk == "safe"` without asking, and still prompt for `risk == "reload"` and `risk == "destructive"`.

6. **Apply approved fixes** — call the referenced tool with the default `args` merged with any user overrides. Record each outcome back to the event log by appending `{"type": "heal_applied", "at": ..., "code": ..., "result": ...}` to `~/.local/state/dotfiles/events.jsonl`.

6b. **Verify the fix worked** — for each code you just remediated, wait 30s (or one dedup window, whichever is shorter), then call `events_tail {type: <code>, since_minutes: 2}`. If the same error_code re-fires within that window, the remediation didn't take — append `{"type":"heal_failed", "at":..., "code":<code>, "prior_attempts":N}` to the events log and flag this group prominently in step 7.

   **Cap auto-retry at 1.** A second failure is a signal that a human needs to look, not a cue to loop. Chronic `heal_failed` entries are the exact failure mode the registry exists to prevent; re-running the same remediation harder is the wrong answer.

7. **Summarize** — what was applied, what verified clean, which fixes failed and need human attention (`heal_failed` entries), what the user skipped, and anything that still has no remediation mapping (candidates to add to the registry next time).

## Arguments

- `/heal` — interactive, prompt per group
- `/heal --auto` — auto-apply every `risk=safe` remediation silently, still prompt on `reload`/`destructive`
- `/heal --since 2h` — widen the lookback window (default 30 min)
- `/heal --type <code>` — heal only a specific error_code

## Staying honest

- Never apply a remediation whose `risk` is `destructive` without explicit user approval on *this* invocation — even with `--auto`.
- If a remediation's `Tool` is not in `allowed-tools` above, surface the tool name + args and ask the user to run it manually rather than bypass the permission boundary.
- When a remediation fails, record that outcome but do not retry automatically. Let the user see the failure and decide — blind retries are exactly the failure mode the registry exists to prevent.

## Related

- Add new error_code→fix mappings in `mcp/dotfiles-mcp/internal/remediation/registry.go`.
- `/pre-risky` is the pair: it snapshots BEFORE a change; `/heal` rolls back AFTER one goes bad.
