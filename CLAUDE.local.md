# dotfiles local RetroArch tranche

Temporary local handoff for the 2026-04-20 RetroArch workstation-first tranche. Keep this uncommitted.

## What changed

- Added shared profile/runtime helpers:
  - `scripts/lib/retroarch_profile.py`
  - `scripts/lib/retroarch_runtime.py`
- Added workstation entrypoints:
  - `scripts/retroarch-workstation-audit.py`
  - `scripts/retroarch-archive-homebrew-sync.py`
  - `scripts/retroarch-install-workstation-cores.sh`
- Updated `scripts/retroarch-archive-homebrew-playlists.py` to load playlist/core metadata from `romhub/configs/device-profiles/retroarch.yaml` when available, with local fallback rows.
- Added install exposure in `install.sh`.
- Added targeted tests in `tests/retroarch_workstation.bats`.

## Live workstation state

- Linked commands into `~/.local/bin`:
  - `retroarch-workstation-audit`
  - `retroarch-archive-homebrew-sync`
  - `retroarch-install-workstation-cores`
- Installed packaged cores with `retroarch-install-workstation-cores.sh`:
  - newly added on this workstation: `libretro-desmume`, `libretro-ppsspp`, `ppsspp-assets`
  - already present: `sameboy`, `mgba`, `mesen`, `beetle-pce-fast`, `flycast`, `dolphin`, `genesis-plus-gx`, `bsnes-hd`, `mupen64plus-next`, `kronos`
- Live audit report: `~/.local/state/retroarch/workstation-audit.json`

Current audit summary:
- `core_total=17`
- `core_missing=3`
- `required_assets_missing=2`
- `optional_assets_missing=9`
- `asset_hash_mismatches=0`
- display detected as `5120x1440`, `32:9`, monitor index `2`

Current hard gaps:
- missing cores: `ngp`, `vb`, `wonderswan`
- required assets missing: `dreamcast`, `psp`
- network command interface is currently off in RetroArch, so runtime OSD notifications are not active yet

## Verification

Run these after touching this tranche:

```bash
python3 -m py_compile \
  scripts/lib/retroarch_profile.py \
  scripts/lib/retroarch_runtime.py \
  scripts/retroarch-workstation-audit.py \
  scripts/retroarch-archive-homebrew-playlists.py \
  scripts/retroarch-archive-homebrew-sync.py

bash -n scripts/retroarch-install-workstation-cores.sh install.sh

bats tests/retroarch_workstation.bats

python3 scripts/retroarch-workstation-audit.py
```

## Known caveats

- The live workstation audit now avoids symlinked and foreign-device subtrees, but it still takes about 19 seconds on the current ROM tree.
- `retroarch-archive-homebrew-sync.py` only does runtime OSD notification via `SHOW_MSG`; it does not attempt playlist hot-reload.
- `scripts/lib/retroarch_archive_homebrew_verified.json` was already dirty before this tranche. Do not overwrite or revert it blindly.

## Follow-up tranche shipped 2026-04-21 (commits fb622f9 + b230980)

All three next-tranche items landed in a single session:

1. **BIOS/helper apply** — `retroarch-bios-apply` seeds `dc/` for Flycast
   (with optional `--source-dir` drop for dc_boot/dc_flash) and populates
   `PPSSPP/` via sparse clone of `hrydgard/ppsspp#assets`.
2. **Missing cores** — `retroarch-build-libretro-cores` source-builds
   `race` (NGP) and `beetle-wswan` (WonderSwan) into `/usr/lib/libretro`.
   `retroarch-install-workstation-cores --skip-aur` option added;
   `libretro-beetle-vb-git` is the AUR fallback for VB.
3. **Runtime OSD network cmd** — `retroarch-apply-network-cmd` with
   atomic `retroarch.cfg` write, timestamped `.bak`, UDP VERSION probe,
   and `--revert`. Backed by new `retroarch_cfg_writer` library.

Supporting work: `docs/retroarch-workstation.md` documents the chain,
`install.sh` symlinks the three new scripts, bats covers all three
new helpers under `--dry-run` and the cfg-writer round trip.

Re-run the documented verification (py_compile, bash -n, bats) after
any follow-up edit — it now covers eight Python targets and all three
shell scripts.

## Remaining workstation backlog

- Audit wall-clock (19s) on the current ROM tree is acceptable but not
  great. Candidate: cache the per-dir stat result keyed by dir mtime.
- `retroarch-archive-homebrew-sync.py` still does not hot-reload
  playlists; RetroArch OSD only.
- Dreamcast BIOS drop is manual — we don't ship dc_boot.bin/dc_flash.bin
  and can't (proprietary). Document-only; do not automate.
