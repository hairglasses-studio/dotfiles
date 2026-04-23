#!/usr/bin/env python3
"""Auto-assign core_path on existing DETECT/empty playlist entries, and
scan ROM directories for unmapped files to add as new entries.

Runs in two phases:

  1. Walk every playlist under `--playlists-dir` (default
     ~/.config/retroarch/playlists), look up the system → core mapping
     in the shared RomHub retroarch profile + built-in fallbacks,
     and rewrite each entry whose `core_path` is empty or "DETECT"
     to point at the canonical installed core for that system (when
     it's available on disk).

  2. For each system listed in the profile whose ROM directory
     exists, enumerate files whose extension matches the system's
     preferred_extensions, compare against the current playlist
     item paths, and append missing files as new entries with
     core_path populated and label derived from the filename stem.

Both phases skip systems whose canonical core isn't installed.
Idempotent: re-runs do not duplicate entries or re-rewrite already
well-formed rows.

JSON report: $XDG_STATE_HOME/retroarch/map-roms.json
"""

from __future__ import annotations

import argparse
import json
import os
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

import retroarch_profile  # noqa: E402


DETECT_SENTINELS = {"", "DETECT"}
_ARCHIVE_EXTS = {".zip", ".7z", ".rar", ".chd", ".tar", ".gz", ".bz2", ".pbp"}


def _rom_file_part(path: str) -> str:
    # Match retroarch-playlist-audit semantics: only strip `#` when the
    # part before it is a known archive container extension. Directory
    # segments with a literal `#` (e.g. `#Unreleased/game.lha`) stay intact.
    idx = path.rfind("#")
    if idx <= 0:
        return path
    before = path[:idx]
    for ext in _ARCHIVE_EXTS:
        if before.lower().endswith(ext):
            return before
    return path

CORE_SEARCH_ROOTS = [
    Path("/usr/lib/libretro"),
    Path.home() / ".config" / "retroarch" / "cores",
    Path("/usr/local/lib/libretro"),
]


def _find_core(filename: str) -> Path | None:
    for root in CORE_SEARCH_ROOTS:
        candidate = root / filename
        if candidate.exists():
            return candidate
    return None


def _expand(path: str | None, default: Path) -> Path:
    if not path:
        return default
    return Path(os.path.expandvars(os.path.expanduser(path)))


def _load_system_map() -> dict[str, dict[str, Any]]:
    """Build `system → {playlist, core_filename, core_name, extensions}`."""
    # retroarch_profile exposes DEFAULT_PLAYLIST_MAP_ROWS as a list of
    # pipe-encoded strings: `system|playlist|core|name|exts|package`.
    rows = retroarch_profile.DEFAULT_PLAYLIST_MAP_ROWS
    result: dict[str, dict[str, Any]] = {}
    for row in rows:
        parts = row.split("|")
        if len(parts) < 5:
            continue
        system, playlist, core_fn, core_name, exts = parts[:5]
        result[system] = {
            "playlist": playlist,
            "core_filename": core_fn,
            "core_name": core_name,
            "extensions": {e.strip().lower() for e in exts.split(",") if e.strip()},
        }
    return result


def _load_playlist(path: Path) -> dict[str, Any]:
    if path.exists():
        return json.loads(path.read_text())
    return {
        "version": "1.5",
        "default_core_path": "",
        "default_core_name": "",
        "label_display_mode": 0,
        "right_thumbnail_mode": 0,
        "left_thumbnail_mode": 0,
        "thumbnail_match_mode": 0,
        "sort_mode": 0,
        "items": [],
    }


def _save_playlist(path: Path, data: dict[str, Any], *, dry_run: bool) -> None:
    if dry_run:
        return
    # Atomic write via tmp + rename.
    tmp = path.with_suffix(path.suffix + ".tmp")
    tmp.write_text(json.dumps(data, indent=2, ensure_ascii=False))
    tmp.replace(path)


def _reassign_detect_entries(
    playlists_dir: Path,
    system_map: dict[str, dict[str, Any]],
    *,
    dry_run: bool,
) -> list[dict[str, Any]]:
    changes: list[dict[str, Any]] = []
    # Invert system_map: playlist filename → system record
    playlist_to_system: dict[str, dict[str, Any]] = {}
    for rec in system_map.values():
        playlist_to_system[rec["playlist"]] = rec

    for lpl in sorted(playlists_dir.glob("*.lpl")):
        rec = playlist_to_system.get(lpl.name)
        if not rec:
            continue
        core_path = _find_core(rec["core_filename"])
        if not core_path:
            continue
        try:
            data = _load_playlist(lpl)
        except json.JSONDecodeError:
            continue
        touched = 0
        for item in data.get("items", []):
            cp_raw = item.get("core_path", "")
            cp = cp_raw.strip() if isinstance(cp_raw, str) else ""
            if cp in DETECT_SENTINELS:
                item["core_path"] = str(core_path)
                item["core_name"] = rec["core_name"]
                touched += 1
        if touched:
            _save_playlist(lpl, data, dry_run=dry_run)
            changes.append(
                {"playlist": lpl.name, "reassigned": touched, "core": str(core_path)}
            )
    return changes


def _scan_missing_roms(
    playlists_dir: Path,
    roms_root: Path,
    system_map: dict[str, dict[str, Any]],
    *,
    dry_run: bool,
) -> list[dict[str, Any]]:
    additions: list[dict[str, Any]] = []
    seen_systems: set[str] = set()
    for system, rec in system_map.items():
        if system in seen_systems:
            continue
        seen_systems.add(system)
        rom_dir = roms_root / system
        if not rom_dir.is_dir():
            # Some systems use aliases; use the dir only when it exists
            # for the canonical name.
            continue
        core_path = _find_core(rec["core_filename"])
        if not core_path:
            continue
        lpl = playlists_dir / rec["playlist"]
        data = _load_playlist(lpl)
        # Build the set of paths ALREADY covered across EVERY playlist
        # under playlists_dir, not just this system's playlist. The
        # same .lha file can legitimately live in both Amiga and CD32
        # playlists by design — but we must not duplicate an entry
        # into this playlist if another playlist already owns it,
        # otherwise a re-scan loop keeps appending.
        existing_paths: set[str] = set()
        for other_lpl in playlists_dir.glob("*.lpl"):
            try:
                other_data = _load_playlist(other_lpl)
            except json.JSONDecodeError:
                continue
            for other_item in other_data.get("items", []):
                raw = (other_item.get("path") or "").strip()
                existing_paths.add(_rom_file_part(raw))

        extensions = rec["extensions"]
        candidates: list[Path] = []
        for path in sorted(rom_dir.rglob("*")):
            if not path.is_file():
                continue
            if path.suffix.lower() not in extensions:
                continue
            # Skip core artifact files (e.g. puae_libretro.uae is the
            # core's default config, not a ROM). Any file whose stem
            # ends in `_libretro` is a core artifact.
            if path.stem.endswith("_libretro"):
                continue
            if str(path) in existing_paths:
                continue
            candidates.append(path)

        if not candidates:
            continue
        new_items = 0
        for rom_path in candidates:
            entry = {
                "path": str(rom_path),
                "label": rom_path.stem,
                "core_path": str(core_path),
                "core_name": rec["core_name"],
                "crc32": "",
                "db_name": rec["playlist"],
            }
            data.setdefault("items", []).append(entry)
            new_items += 1

        data["items"].sort(key=lambda item: item.get("label", "").lower())
        _save_playlist(lpl, data, dry_run=dry_run)
        additions.append(
            {
                "playlist": rec["playlist"],
                "system": system,
                "added": new_items,
                "rom_dir": str(rom_dir),
            }
        )
    return additions


def _parse_args(argv: list[str]) -> argparse.Namespace:
    default_state = _expand(os.environ.get("XDG_STATE_HOME"), Path.home() / ".local" / "state")
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--playlists-dir", type=Path,
                   default=Path.home() / ".config" / "retroarch" / "playlists")
    p.add_argument("--roms-root", type=Path,
                   default=Path.home() / "Games" / "RetroArch" / "roms")
    p.add_argument("--report", type=Path,
                   default=default_state / "retroarch" / "map-roms.json")
    p.add_argument("--dry-run", action="store_true",
                   help="Print planned changes without writing the playlist files.")
    p.add_argument("--json", action="store_true",
                   help="Emit the structured JSON report.")
    return p.parse_args(argv)


def main(argv: list[str]) -> int:
    args = _parse_args(argv)
    if not args.playlists_dir.is_dir():
        print(f"error: playlists dir not found: {args.playlists_dir}", file=sys.stderr)
        return 2
    system_map = _load_system_map()

    reassigned = _reassign_detect_entries(args.playlists_dir, system_map, dry_run=args.dry_run)
    scanned = _scan_missing_roms(args.playlists_dir, args.roms_root, system_map, dry_run=args.dry_run)

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "playlists_dir": str(args.playlists_dir),
        "roms_root": str(args.roms_root),
        "dry_run": args.dry_run,
        "reassigned": reassigned,
        "scanned": scanned,
        "summary": {
            "playlists_touched": len(reassigned) + len(scanned),
            "entries_reassigned": sum(r["reassigned"] for r in reassigned),
            "entries_added": sum(s["added"] for s in scanned),
        },
    }

    if not args.dry_run:
        args.report.parent.mkdir(parents=True, exist_ok=True)
        args.report.write_text(json.dumps(report, indent=2))

    if args.json:
        print(json.dumps(report, indent=2))
    else:
        s = report["summary"]
        print(f"playlists_touched={s['playlists_touched']} "
              f"entries_reassigned={s['entries_reassigned']} "
              f"entries_added={s['entries_added']} "
              f"dry_run={'yes' if args.dry_run else 'no'}")
        for r in reassigned:
            print(f"  REASSIGNED: {r['playlist']}: {r['reassigned']} entries -> {Path(r['core']).name}")
        for s in scanned:
            print(f"  SCANNED:    {s['playlist']}: +{s['added']} entries from {s['rom_dir']}")

    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
