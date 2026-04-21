#!/usr/bin/env python3
"""Stage mirrored Archive.org homebrew/public-domain files into RetroArch ROM dirs."""

from __future__ import annotations

import argparse
import json
import os
import shutil
from pathlib import Path


def _expand(path: str | None, default: Path) -> Path:
    if not path:
        return default
    return Path(os.path.expandvars(os.path.expanduser(path)))


def _archive_source(archive_root: Path, entry: dict) -> Path:
    return archive_root / entry["system"] / entry["source_identifier"] / entry["file_name"]


def _rom_target(rom_root: Path, entry: dict, subdir: str) -> Path:
    return rom_root / entry["system"] / subdir / entry["file_name"]


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Import mirrored Archive.org homebrew files into RetroArch ROM dirs.")
    parser.add_argument("--manifest", help="Path to the manifest JSON.")
    parser.add_argument(
        "--archive-root",
        help="Root directory containing mirrored Archive.org downloads.",
    )
    parser.add_argument(
        "--rom-root",
        help="RetroArch ROM root directory.",
    )
    parser.add_argument(
        "--tier",
        action="append",
        choices=["public_domain", "verified_redistributable", "homebrew_unverified", "utility_unverified"],
        help="Import additional tiers. Default includes public_domain and verified_redistributable.",
    )
    parser.add_argument("--system", action="append", help="Limit imports to specific system ids.")
    parser.add_argument(
        "--mode",
        choices=["symlink", "copy"],
        default="symlink",
        help="How to stage files into the ROM tree.",
    )
    parser.add_argument(
        "--subdir",
        default="archive-homebrew",
        help="Subdirectory under each system ROM dir for imported files.",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print planned imports without modifying the ROM tree.",
    )
    args = parser.parse_args(argv)

    manifest_path = _expand(
        args.manifest or os.environ.get("RETROARCH_ARCHIVE_HOMEBREW_MANIFEST"),
        Path.home() / ".local" / "state" / "retroarch-archive" / "homebrew-manifest.json",
    )
    archive_root = _expand(
        args.archive_root or os.environ.get("RETROARCH_ARCHIVE_DOWNLOAD_ROOT"),
        Path.home() / "Games" / "RetroArch" / "archives" / "internet-archive",
    )
    rom_root = _expand(
        args.rom_root or os.environ.get("RETROARCH_ROM_ROOT"),
        Path.home() / "Games" / "RetroArch" / "roms",
    )

    manifest = json.loads(manifest_path.read_text())
    allowed_tiers = set(args.tier or ["public_domain", "verified_redistributable"])
    allowed_systems = set(args.system or [])

    selected = []
    for entry in manifest.get("entries", []):
        if entry["tier"] not in allowed_tiers:
            continue
        if allowed_systems and entry["system"] not in allowed_systems:
            continue
        if entry["tier"] in {"public_domain", "verified_redistributable"} or entry.get("default_selected"):
            selected.append(entry)
        elif args.tier:
            selected.append(entry)

    imported = 0
    skipped = 0
    missing = 0
    for entry in selected:
        source = _archive_source(archive_root, entry)
        target = _rom_target(rom_root, entry, args.subdir)
        if not source.exists():
            missing += 1
            print(f"MISS {source}")
            continue
        if target.exists() or target.is_symlink():
            if target.is_symlink() and target.resolve() == source.resolve():
                skipped += 1
                print(f"SKIP {target}")
                continue
            skipped += 1
            print(f"SKIP {target}")
            continue
        print(f"{'LINK' if args.mode == 'symlink' else 'COPY'} {source} -> {target}")
        if args.dry_run:
            continue
        target.parent.mkdir(parents=True, exist_ok=True)
        if args.mode == "copy":
            shutil.copy2(source, target)
        else:
            target.symlink_to(source)
        imported += 1

    print(
        f"selected={len(selected)} imported={imported} skipped={skipped} "
        f"missing={missing} dry_run={int(args.dry_run)} mode={args.mode}"
    )
    return 0 if missing == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
