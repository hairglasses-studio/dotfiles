# Decision Log

Append-only ADR-lite record of architectural calls that shaped this repo. Managed by the `decision_journal` skill. Newest entries at the top.

## 2026-04-21 — Profile-driven RetroArch playlist and BIOS mapping

**Context**: The RetroArch archive-homebrew playlist generator hardcoded a per-system dict (cores, extensions, filenames) inside `scripts/retroarch-archive-homebrew-playlists.py`. The workstation audit needed the same system ↔ core ↔ BIOS data but from a different entrypoint, risking drift between the two.

**Decision**: Lift the playlist map and BIOS/helper requirements into `~/hairglasses-studio/romhub/configs/device-profiles/retroarch.yaml` as two new keys, `retroarch_playlist_map` and `retroarch_requirements`. Keep the built-in defaults in `scripts/lib/retroarch_profile.py` as a fallback when the shared profile is unreachable.

**Rationale**: The shared profile already carries the generic RetroArch device contract; adding the mapping rows there makes it the single source of truth for any downstream consumer (audit, sync orchestrator, playlist generator, future MCP tools). The hardcoded defaults stay because the dotfiles scripts must keep working on a machine where romhub is not checked out.

**Alternatives considered**:
- Leave the dict hardcoded, accept drift — rejected; the audit and the playlist generator already disagreed on what "ngp" meant.
- Put the mapping in a dotfiles-local YAML — rejected; it would fork from the romhub profile and multiply sources of truth.

**Consequences**: `retroarch_profile.py` owns the schema and fallback defaults; romhub owns the data. Any new consumer reads via `load_playlist_map()` / `load_requirement_catalog()`.

## 2026-04-21 — Sparse-clone for PPSSPP BIOS helper assets

**Context**: The PSP system dir requires `PPSSPP/` to be non-empty (assets like shader presets, flash files, UI images). There is no pacman package that drops this content. Shipping the assets in dotfiles would grow the repo and duplicate upstream.

**Decision**: `retroarch-bios-apply.py --system psp` uses a sparse `git clone --depth=1` of `hrydgard/ppsspp#assets` into `$RETROARCH_SYSTEM_DIR/PPSSPP/`, matching the existing Dolphin-Sys pattern in `scripts/retroarch-dolphin-sync-sys.sh`.

**Rationale**: Same precedent already solved for Dolphin — minimal disk use, always current, no license entanglement, and the one subtree we need is isolated. Users can still `--source-dir` a local drop when offline.

**Alternatives considered**:
- Download a release tarball — rejected; hrydgard doesn't publish assets as a separate artifact.
- Vendor the assets in dotfiles — rejected; size + license ambiguity + drift.

## 2026-04-21 — Atomic `retroarch.cfg` writer with timestamped backup

**Context**: The workstation tranche needed to flip `network_cmd_enable = true` programmatically. The repo previously only *read* `retroarch.cfg`; there was no precedent for rewriting it, and the file is irreplaceable user state (hotkeys, input bindings, quality settings accumulated over years).

**Decision**: New `scripts/lib/retroarch_cfg_writer.py` provides `apply_settings(cfg_path, updates, *, backup, dry_run)`. Preserves original line order, replaces existing keys in place, appends unknown keys, writes via `mktemp`+`os.replace` for atomicity, and copies `retroarch.cfg` → `retroarch.cfg.bak.<ISO-utc>` before the first mutation of each invocation.

**Rationale**: Any retroarch.cfg edit is a destructive operation until we have a backup and an atomic swap. The timestamped suffix prevents a second run from clobbering the previous backup. `--dry-run` + `--revert` let callers validate and roll back without shell glue.

**Alternatives considered**:
- Edit via `sed -i` — rejected; not atomic, trivial to corrupt on SIGINT.
- Full file rewrite from parsed dict — rejected; loses any keys we don't know about (RetroArch's cfg has hundreds), violates the "don't surprise the user" principle.

**Consequences**: `retroarch-apply-network-cmd.py` is now a thin wrapper; any future retroarch.cfg mutator (e.g. per-monitor overrides, driver swaps) should go through the same library.
