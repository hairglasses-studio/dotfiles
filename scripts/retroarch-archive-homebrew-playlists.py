#!/usr/bin/env python3
"""Generate RetroArch playlist entries for imported Archive homebrew content."""

from __future__ import annotations

import argparse
import json
import os
import sys
import zipfile
from pathlib import Path


sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

import retroarch_profile


SAFE_DEFAULT_TIERS = {"public_domain", "verified_redistributable"}
DEFAULT_SUBDIR = "archive-homebrew"
SYSTEM_PLAYLISTS: dict[str, dict] = {}


def _expand(path: str | None, default: Path) -> Path:
    if not path:
        return default
    return Path(os.path.expandvars(os.path.expanduser(path)))


def _load_system_playlists(profile_path: str | None = None) -> dict[str, dict]:
    raw = retroarch_profile.load_playlist_map(profile_path)
    systems: dict[str, dict] = {}
    for system, entry in raw.items():
        core_filename = str(entry["core_filename"])
        systems[system] = {
            "playlist": entry["playlist"],
            "core_path": str(Path("/usr/lib/libretro") / core_filename),
            "core_name": entry["core_name"],
            "extensions": set(entry["extensions"]),
        }
    return systems


def _load_playlist(path: Path) -> dict:
    if not path.exists():
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
    return json.loads(path.read_text())


def _select_inner_member(source: Path, extensions: set[str]) -> str | None:
    if source.suffix.lower() != ".zip":
        if source.suffix.lower() in extensions:
            return source.name
        return None
    with zipfile.ZipFile(source) as archive:
        candidates = []
        for member in archive.namelist():
            member_path = Path(member)
            if member_path.name.startswith("."):
                continue
            if member_path.suffix.lower() in extensions:
                candidates.append(member)
        if not candidates:
            return None
        candidates.sort(key=lambda item: ("/" in item, item.lower()))
        return candidates[0]


def _playlist_path(playlist_root: Path, system: str) -> Path | None:
    config = SYSTEM_PLAYLISTS.get(system)
    if not config:
        return None
    return playlist_root / config["playlist"]


def _item_path(target: Path, inner_member: str | None) -> str:
    if not inner_member or target.suffix.lower() != ".zip":
        return str(target)
    return f"{target}#{inner_member}"


def _item_label(entry: dict, inner_member: str | None) -> str:
    source = inner_member if inner_member else entry["file_name"]
    stem = Path(source).stem
    stem = stem.replace("_", " ")
    stem = stem.replace(" (PD)", "")
    stem = stem.replace("(PD)", "")
    stem = " ".join(stem.split())
    return stem


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Generate RetroArch playlist entries for imported Archive homebrew.")
    parser.add_argument("--manifest", help="Path to the manifest JSON.")
    parser.add_argument("--rom-root", help="RetroArch ROM root directory.")
    parser.add_argument("--playlist-root", help="RetroArch playlist directory.")
    parser.add_argument("--subdir", default=DEFAULT_SUBDIR, help="Imported ROM subdirectory under each system dir.")
    parser.add_argument(
        "--retroarch-profile",
        help="Optional path to the shared RomHub RetroArch profile YAML.",
    )
    parser.add_argument(
        "--tier",
        action="append",
        choices=["public_domain", "verified_redistributable", "homebrew_unverified", "utility_unverified"],
        help="Include one or more tiers. Default includes public_domain and verified_redistributable.",
    )
    parser.add_argument("--system", action="append", help="Limit generation to specific system ids.")
    parser.add_argument("--dry-run", action="store_true", help="Print planned playlist additions without writing files.")
    args = parser.parse_args(argv)

    manifest_path = _expand(
        args.manifest or os.environ.get("RETROARCH_ARCHIVE_HOMEBREW_MANIFEST"),
        Path.home() / ".local" / "state" / "retroarch-archive" / "homebrew-manifest.json",
    )
    rom_root = _expand(
        args.rom_root or os.environ.get("RETROARCH_ROM_ROOT"),
        Path.home() / "Games" / "RetroArch" / "roms",
    )
    playlist_root = _expand(
        args.playlist_root or os.environ.get("RETROARCH_PLAYLIST_ROOT"),
        Path.home() / ".config" / "retroarch" / "playlists",
    )

    manifest = json.loads(manifest_path.read_text())
    allowed_tiers = set(args.tier or sorted(SAFE_DEFAULT_TIERS))
    allowed_systems = set(args.system or [])
    global SYSTEM_PLAYLISTS
    SYSTEM_PLAYLISTS = _load_system_playlists(args.retroarch_profile)

    playlist_cache: dict[Path, dict] = {}
    existing_paths: dict[Path, set[str]] = {}
    added = 0
    skipped = 0
    missing = 0

    for entry in manifest.get("entries", []):
        if entry["tier"] not in allowed_tiers:
            continue
        if allowed_systems and entry["system"] not in allowed_systems:
            continue
        if entry["tier"] not in SAFE_DEFAULT_TIERS and not args.tier:
            continue

        system = entry["system"]
        config = SYSTEM_PLAYLISTS.get(system)
        playlist_path = _playlist_path(playlist_root, system)
        if not config or playlist_path is None:
            skipped += 1
            print(f"SKIP unsupported-system {system} {entry['file_name']}")
            continue

        target = rom_root / system / args.subdir / entry["file_name"]
        if not target.exists() and not target.is_symlink():
            missing += 1
            print(f"MISS {target}")
            continue

        inner_member = _select_inner_member(target, config["extensions"])
        if inner_member is None:
            skipped += 1
            print(f"SKIP no-launch-member {target}")
            continue

        item_path = _item_path(target, inner_member)
        playlist = playlist_cache.setdefault(playlist_path, _load_playlist(playlist_path))
        seen = existing_paths.setdefault(
            playlist_path,
            {item.get("path", "") for item in playlist.get("items", [])},
        )
        if item_path in seen:
            skipped += 1
            print(f"SKIP {item_path}")
            continue

        core_path = config["core_path"] if Path(config["core_path"]).exists() else "DETECT"
        core_name = config["core_name"] if core_path != "DETECT" else "DETECT"
        item = {
            "path": item_path,
            "label": _item_label(entry, inner_member),
            "core_path": core_path,
            "core_name": core_name,
            "crc32": "",
            "db_name": config["playlist"],
        }
        print(f"ADD  {config['playlist']} :: {item_path}")
        if not args.dry_run:
            playlist.setdefault("items", []).append(item)
            seen.add(item_path)
            added += 1

    if not args.dry_run:
        playlist_root.mkdir(parents=True, exist_ok=True)
        for playlist_path, playlist in playlist_cache.items():
            items = playlist.get("items", [])
            items.sort(key=lambda item: item.get("label", "").lower())
            playlist["items"] = items
            playlist_path.write_text(json.dumps(playlist, indent=2) + "\n")

    print(f"added={added} skipped={skipped} missing={missing} dry_run={int(args.dry_run)}")
    return 0 if missing == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
