# RetroArch Workstation

Operator-facing reference for the RetroArch workstation tooling in this repo:
a one-shot Archive.org homebrew/public-domain sync plus a set of standalone
helpers for core install, BIOS/helper provisioning, and runtime control.

Shared RetroArch device profile:
`~/hairglasses-studio/romhub/configs/device-profiles/retroarch.yaml` — carries
the authoritative `retroarch_playlist_map` and `retroarch_requirements` rows.
`scripts/lib/retroarch_profile.py` falls back to built-in defaults when the
shared profile is unreachable.

## One-shot sync

```bash
retroarch-archive-homebrew-sync --dry-run
retroarch-archive-homebrew-sync --notify-runtime
```

Chains `manifest → fetch → import → playlists → audit`. With `--notify-runtime`
and RetroArch running with network commands enabled, the sync sends a
`SHOW_MSG` OSD notification on success.

Summary JSON: `$XDG_STATE_HOME/retroarch-archive/sync-summary.json`.

## Individual commands

| Command | Purpose |
|---|---|
| `retroarch-workstation-audit` | Audit cores, BIOS/asset dirs, display, runtime. Writes JSON to `$XDG_STATE_HOME/retroarch/workstation-audit.json`. |
| `retroarch-bios-apply` | Populate missing required BIOS/helper dirs (Dreamcast `dc/`, PSP `PPSSPP/`). Supports `--dry-run` and `--source-dir` for local BIOS drops. |
| `retroarch-install-workstation-cores` | Install pacman-packaged libretro cores; optional AUR pass for `libretro-beetle-vb-git` when `yay`/`paru` is on `PATH`. `--skip-aur` disables the AUR pass. |
| `retroarch-build-libretro-cores` | Source-build `race` (NGP) and `beetle-wswan` (WonderSwan) cores into `/usr/lib/libretro/`. `--dry-run` prints the steps. |
| `retroarch-apply-network-cmd` | Flip `network_cmd_enable` + `network_cmd_port` in `retroarch.cfg` atomically with a timestamped `retroarch.cfg.bak.*` backup. `--dry-run` and `--revert` supported. |

## Suggested workflow (first-run)

```bash
retroarch-workstation-audit                     # surface gaps
retroarch-bios-apply                            # fill dreamcast + psp dirs
retroarch-install-workstation-cores             # pacman + optional AUR
sudo retroarch-build-libretro-cores             # source build race + beetle-wswan
retroarch-apply-network-cmd                     # flip cfg + probe
# restart RetroArch so it binds UDP 55355
retroarch-archive-homebrew-sync --notify-runtime
retroarch-workstation-audit                     # confirm clean
```

## State files

- `$XDG_STATE_HOME/retroarch/workstation-audit.json` — audit snapshot.
- `$XDG_STATE_HOME/retroarch-archive/homebrew-manifest.json` — curated Archive.org manifest.
- `$XDG_STATE_HOME/retroarch-archive/sync-summary.json` — last orchestrator run.
- `$XDG_STATE_HOME/retroarch/bios-apply.json` — last BIOS helper run.

## Runtime OSD prerequisite

`retroarch-apply-network-cmd` writes the cfg but RetroArch only binds UDP 55355
at startup. Restart RetroArch after applying so the socket is live. The audit
report's `runtime.version_probe` confirms the socket is answering.

## Tests

```bash
bats tests/retroarch_workstation.bats
```

Covers the profile-driven playlist mapping, audit summarisation with
zip-backed ROM counts and conditional BIOS, orchestrator end-to-end wiring,
the atomic-with-backup cfg writer, bios-apply dry-run planning, and
build-libretro-cores dry-run planning.

## Deferred

- Playlist hot-reload — RetroArch's network command API has no verb for
  reloading all playlists. `SHOW_MSG` remains the only runtime signal.
- Audit perf (~19 s on the current ROM tree) — acceptable for batch runs;
  revisit only if the sync orchestrator starts calling it in a tight loop.
