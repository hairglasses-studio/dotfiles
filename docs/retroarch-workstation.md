# RetroArch Workstation

Operator-facing reference for the RetroArch workstation tooling in this repo:
a one-shot Archive.org homebrew/public-domain sync plus a set of standalone
helpers for core install, BIOS/helper provisioning, and runtime control.

Shared RetroArch device profile:
`~/hairglasses-studio/romhub/configs/device-profiles/retroarch.yaml` — carries
the authoritative `retroarch_playlist_map` and `retroarch_requirements` rows.
`scripts/lib/retroarch_profile.py` falls back to built-in defaults when the
shared profile is unreachable.

## One-shot full setup

```bash
retroarch-complete --dry-run   # preview the plan
retroarch-complete             # chain every idempotent step
```

`retroarch-complete` runs `audit → bios-apply → install-workstation-cores
→ (conditional) apply-network-cmd → post-audit`, skipping anything
already in place. No `sudo` anywhere. If the workstation still needs
source-built cores (`race` + `beetle-wswan`), the orchestrator prints
the exact follow-up command and exits with a clear `note` — that's
the one external step sandboxed agents can't run on your behalf.

## Archive.org homebrew sync

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
| `retroarch-complete` | End-to-end orchestrator that chains every other step idempotently. `--dry-run` prints the plan; `--skip-build` suppresses the source-build nag. No `sudo`, no external clone. |
| `retroarch-workstation-audit` | Audit cores, BIOS/asset dirs, display, runtime. Writes JSON to `$XDG_STATE_HOME/retroarch/workstation-audit.json`. |
| `retroarch-bios-apply` | Populate missing required BIOS/helper dirs (Dreamcast `dc/`, PSP `PPSSPP/`). Supports `--dry-run` and `--source-dir` for local BIOS drops. |
| `retroarch-install-workstation-cores` | Install pacman-packaged libretro cores; optional AUR pass for `libretro-beetle-vb-git` when `yay`/`paru` is on `PATH`. `--skip-aur` disables the AUR pass. |
| `retroarch-build-libretro-cores` | Source-build `race` (NGP) and `beetle-wswan` (WonderSwan) cores. Defaults to `/usr/lib/libretro/` (requires `sudo`); pass `--install-dir ~/.config/retroarch/cores` to drop into the user-local cores dir without `sudo`. `--dry-run` prints the steps; `--only race\|beetle-wswan\|beetle-vb` restricts the build set. |
| `retroarch-apply-network-cmd` | Flip `network_cmd_enable` + `network_cmd_port` in `retroarch.cfg` atomically with a timestamped `retroarch.cfg.bak.*` backup. `--dry-run` and `--revert` supported. |
| `retroarch-command` | Send a UDP network command to a running RetroArch. `--list` prints the known taxonomy, `--osd <text>` is a `SHOW_MSG` shortcut, `--json` emits the structured result, `--wait-for-ready` polls `VERSION` until the socket answers (useful post-restart). Requires `network_cmd_enable = "true"` and a running RetroArch. |
| `retroarch-playlist-audit` | Walk `~/.config/retroarch/playlists/*.lpl`, flag entries with a missing `core_path` (silent no-op at launch) or a missing `path` (ROM file moved/renamed). `DETECT` and empty `core_path` are treated as intentional ("pick at launch"). Archive-inner paths (`file.zip#inner`) check the archive, not the inner path. Writes JSON to `$XDG_STATE_HOME/retroarch/playlist-audit.json`. Called automatically at the end of `retroarch-complete`. |
| `retroarch-mounts-audit` | Report health of every rclone mount under `~/Games/RetroArch/mounts/`. Checks mountpoint status, systemd service state, reachability (no-hang ls), file count, age since last unit start, and whether the `--dir-cache-time` window has been exceeded (stale cache). Exit 0 = all healthy, 1 = unreachable/inactive, 2 = no systemd / no mounts root. JSON report at `$XDG_STATE_HOME/retroarch/mounts-audit.json`. Called automatically at the end of `retroarch-complete`. |
| `retroarch-bios-link` | Walk the requirement catalog (`retroarch_profile.DEFAULT_REQUIREMENT_ROWS`), find matching BIOS files under `~/Games/RetroArch/mounts/bios/**` (case-insensitive match on filename, md5 check when declared), and symlink them into `~/.config/retroarch/system/`. Creates symlinks, not copies — gdrive stays the source of truth. `--dry-run` previews; `--bios-mount` adds extra roots. Report at `$XDG_STATE_HOME/retroarch/bios-link.json`. |
| `retroarch-map-roms` | Two-phase playlist mapper. (1) Reassigns `core_path` on every `DETECT`/empty entry whose system has its canonical core installed. (2) Walks `~/Games/RetroArch/roms/<system>/` for each profile-declared system and appends entries for files not already in any playlist (cross-playlist dedup prevents adding the same ROM to Amiga and CD32 twice). `--dry-run` previews. Uses the shared RomHub profile + built-in fallbacks. Writes JSON to `$XDG_STATE_HOME/retroarch/map-roms.json`. |

## Suggested workflow (first-run)

```bash
retroarch-workstation-audit                     # surface gaps
retroarch-bios-apply                            # fill dreamcast + psp dirs
retroarch-install-workstation-cores             # pacman + optional AUR
retroarch-build-libretro-cores \
  --install-dir "$HOME/.config/retroarch/cores" # source build race + beetle-wswan, no sudo
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
- `$XDG_STATE_HOME/retroarch/playlist-audit.json` — playlist core/ROM integrity.
- `$XDG_STATE_HOME/retroarch/mounts-audit.json` — rclone mount health.
- `$XDG_STATE_HOME/retroarch/thumbnail-audit.json` — thumbnail fill rate per system.
- `$XDG_STATE_HOME/retroarch/thumbnail-fill.json` — last fetch run (first 500 actions capped).
- `/tmp/bar-retroarch.txt` — ticker cache (refreshed every 30s by `bar-retroarch.timer`).

## Ticker widget

`bar-retroarch.timer` runs `scripts/bar-retroarch-cache.sh` every 30 s.
The script reads the first entry of `~/.config/retroarch/playlists/builtin/
content_history.lpl` plus `mounts-audit.json` summary, and writes a
composite label to `/tmp/bar-retroarch.txt`:

```
󰊗 Shadow of the Colossus [LRPS2] ⎈5/5
```

Format: `<icon> <title with (region) stripped> [<core short-name>] ⎈<active>/<total>`.
The mount badge appends `!` when `active < total` (stale cache or
unreachable mount). Empty output when no history + no mount info → the
Ironbar widget hides.

The Ironbar widget is defined in `ironbar/config.toml` under the
`[[monitors."DP-3".end]]` block with `class = "bar-retroarch"`.

## Save/state backup

`retroarch-saves-backup.timer` fires `OnCalendar=daily` + `OnBootSec=15min`
with `Persistent=true`. Pushes `~/.config/retroarch/{saves,states}` to
`gdrive:Gaming & Emulation/RetroArch/{saves,states}` via `rclone sync`.

Replaced or removed files get archived to
`gdrive:Gaming & Emulation/RetroArch/.backup/<iso-utc>/{saves,states}`
via `--backup-dir`, so an accidental local `rm` doesn't erase the remote
copy — recovery is a single `rclone copy` back into
`~/.config/retroarch/saves`.

Manual trigger: `systemctl --user start retroarch-saves-backup.service`.

## Runtime OSD prerequisite

`retroarch-apply-network-cmd` writes the cfg but RetroArch only binds UDP 55355
at startup. Restart RetroArch after applying so the socket is live. Confirm
readiness with:

```bash
retroarch-command --wait-for-ready --wait-timeout 30
# prints the RetroArch version string on success, exits 1 on timeout
```

The audit report's `runtime.version_probe` covers the same signal on every
audit run.

## Tests

```bash
bats tests/retroarch_workstation.bats
```

Covers the profile-driven playlist mapping, audit summarisation with
zip-backed ROM counts and conditional BIOS, orchestrator end-to-end wiring,
the atomic-with-backup cfg writer, bios-apply dry-run planning, and
build-libretro-cores dry-run planning.

## Runtime network commands

`retroarch-apply-network-cmd` flips `network_cmd_enable` in `retroarch.cfg`;
once RetroArch binds UDP 55355, commands can be sent via
`retroarch --command '<CMD>'` or `scripts/lib/retroarch_runtime.send_udp_command()`.

Commands we rely on today:

- `SHOW_MSG <text>` — OSD notification (used by `--notify-runtime`).
- `VERSION` — socket health probe, surfaced as `runtime.version_probe` in the audit JSON.

Commands available but currently unused:

- `SET_SHADER <path>` — swap shader preset.
- `LOAD_CORE <path>` — load a core by filesystem path.
- `LOAD_STATE_SLOT`, `SAVE_FILES`, `SCREENSHOT`, `PAUSE_TOGGLE`, `RESET`, `QUIT`.

Not available — drive these from the menu or a fresh `retroarch` invocation:

- Playlist reload / rescan. No UDP verb exists; see `Deferred` below.
- Content library rebuild; use `retroarch --scan=PATH` in a separate process.
- Config reload without restart.

Full canonical list: <https://docs.libretro.com/development/retroarch/network-control-interface/>.

## Dreamcast BIOS

`retroarch-bios-apply` guarantees the Flycast system subdirectory exists —
the audit requirement row `dreamcast|dir|required|dc|` clears once the
empty directory is created. The boot/flash ROMs themselves are
proprietary Sega firmware and are not shipped with this repo.

Expected location, filename, and md5:

```
~/.config/retroarch/system/dc/dc_boot.bin    md5 e10c53c2f8b90bab96ead2d368858623
~/.config/retroarch/system/dc/dc_flash.bin   md5 (region-specific; check Flycast docs)
```

Both are marked `optional` in the shared requirement catalog
(`scripts/lib/retroarch_profile.py`): Flycast boots HLE when the files
are absent but defaults to LLE when `dc_boot.bin` is present. Audit
will not flag them as missing unless the catalog is edited — they are
intentionally optional rows.

Supply your own BIOS copies (typically from a Dreamcast you own) and
use `--source-dir` to install them:

```bash
retroarch-bios-apply --system dreamcast --source-dir /path/to/your/bios
```

The script copies any matching `dc_boot.bin` / `dc_flash.bin` out of
`--source-dir` into `~/.config/retroarch/system/dc/`, skips files that
are already present with the correct hash, and writes an entry to
`$XDG_STATE_HOME/retroarch/bios-apply.json`. No network access; safe
to run under `--dry-run` first to confirm which files it would touch.

After dropping the BIOS, confirm via the audit:

```bash
retroarch-workstation-audit
python3 -c 'import json; d=json.load(open("'"$HOME"'/.local/state/retroarch/workstation-audit.json")); [print(r) for r in d["requirements"] if r["system"]=="dreamcast"]'
```

The optional rows for `dc_boot.bin` / `dc_flash.bin` will flip to
`status: present` with a populated `md5` field.

## Optional BIOS files the sandbox can't auto-grab

Two entries in the requirement catalog stay optional_missing on this
workstation because they don't exist on any gdrive mount, aren't
packaged for Arch/AUR, and can't be downloaded from inside a
sandboxed agent session:

| File | System | Expected md5 | Source |
|---|---|---|---|
| `PSXONPSP660.bin` | psx | `c53ca5908936d412331790f4426c6c33` | Extracted from PSP firmware (`popstation` / `pops660.PBP`); region-free PSX BIOS override used mainly to bypass regional lockout on some PSX discs. |
| `ps1_rom.bin` | psx | `81bbe60ba7a3d1cea1d48c14cbcc647b` | PS3 firmware's PS1 emulation BIOS; region-free. Obtained by extracting from a PS3's firmware update blob. |

Both are purely optional — the canonical SCPH5500/5501/5502 regional
BIOS set is already in place and handles the common case. Drop
either file at its expected filename into one of the gdrive BIOS
mounts (e.g. `gdrive:Gaming & Emulation/ROMs/bios/psx/`) and rerun
`retroarch-bios-link`; the tool will detect the md5 match and
symlink it into `~/.config/retroarch/system/` automatically.

## Deferred

- **Playlist hot-reload.** Verified 2026-04-22 against RetroArch 1.22.2:
  the UDP command interface has no verb for reloading playlists or
  rebuilding the content library. Options considered and rejected:
  (a) no network verb exists; (b) `retroarch --scan=PATH` always forks a
  new process and does not talk to a running instance; (c) RetroArch's
  own playlist refresh is user-driven via menu → Playlists → right-thumb.
  A filesystem watcher could fire inotify on `~/.config/retroarch/playlists/`
  but RetroArch itself does not poll for changes, so the user still has
  to trigger the menu refresh. `SHOW_MSG` remains the only runtime
  signal from `retroarch-archive-homebrew-sync --notify-runtime`.
- Audit perf (2–3 s warm-cache on the 71 GB / 91-file ROM tree as of
  2026-04-22; originally noted as ~19 s but that was a cold-cache or
  pre-prune number). No optimization planned — batch-run cost is
  already in the noise.
