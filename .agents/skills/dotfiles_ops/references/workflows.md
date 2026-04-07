# Dotfiles Ops Workflow Reference

This skill replaces the previous narrow catalog around config management, workstation health, repo onboarding, CI sync, and fleet maintenance.

## Workstation And Config Ops

- Use the config and service helpers for validation, reloads, and symlink checks before editing live config.
- Health checks should summarize system state, failed services, and the first safe remediation step.
- Keep changes reproducible through scripts whenever the same task recurs.

## Repo Tooling And Fleet Work

- Use the onboarding, workflow-sync, and org/fleet scripts instead of one-off shell loops.
- Shared scripts should be dry-run friendly and explicit about scope.
- When syncing across repos, separate audit output from mutation output.

## CI And Shipping

- Fix the smallest failing surface first, then verify repo-wide checks.
- For release-oriented work, capture the gate result, pending risks, and the exact command path used.

## Legacy Compression

The previous Claude skills now collapse into this one canonical surface:

- `config`, `service`, `desktop-health`, `syscheck`, `process`, `network`, `audio`, `bluetooth`
- `onboard`, `workflow-sync`, `fleet-ci`, `fleet-health`, `fleet-iterate`, `fleet-push`, `gh-org`
- `fix-ci`, `review`, `ship`, `ops-status`, `workspace-snapshot`
