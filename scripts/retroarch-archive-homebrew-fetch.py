#!/usr/bin/env python3
"""Fetch selected entries from a previously generated Archive.org manifest."""

from __future__ import annotations

import argparse
import json
import os
import urllib.request
from pathlib import Path


def _expand(path: str | None, default: Path) -> Path:
    if not path:
        return default
    return Path(os.path.expandvars(os.path.expanduser(path)))


def _safe_relative(entry: dict) -> Path:
    system = entry["system"]
    source = entry["source_identifier"]
    name = Path(entry["archive_path"]).name
    return Path(system) / source / name


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Fetch selected Archive.org homebrew/public-domain entries.")
    parser.add_argument(
        "--manifest",
        help="Path to the manifest JSON.",
    )
    parser.add_argument(
        "--tier",
        action="append",
        choices=["public_domain", "homebrew_unverified", "utility_unverified"],
        help="Download additional tiers. Default is public_domain only.",
    )
    parser.add_argument(
        "--system",
        action="append",
        choices=["dreamcast", "wii", "n64", "snes", "genesis"],
        help="Limit downloads to specific systems.",
    )
    parser.add_argument(
        "--download-root",
        help="Root directory for mirrored downloads.",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Print what would be downloaded without writing files.",
    )
    args = parser.parse_args(argv)

    manifest_path = _expand(
        args.manifest or os.environ.get("RETROARCH_ARCHIVE_HOMEBREW_MANIFEST"),
        Path.home() / ".local" / "state" / "retroarch-archive" / "homebrew-manifest.json",
    )
    download_root = _expand(
        args.download_root or os.environ.get("RETROARCH_ARCHIVE_DOWNLOAD_ROOT"),
        Path.home() / "Games" / "RetroArch" / "archives" / "internet-archive",
    )

    manifest = json.loads(manifest_path.read_text())
    allowed_tiers = set(args.tier or ["public_domain"])
    allowed_systems = set(args.system or [])

    selected = []
    for entry in manifest.get("entries", []):
        if entry["tier"] not in allowed_tiers:
            continue
        if allowed_systems and entry["system"] not in allowed_systems:
            continue
        if entry["tier"] == "public_domain" or entry.get("default_selected"):
            selected.append(entry)
        elif args.tier:
            selected.append(entry)

    downloaded = 0
    skipped = 0
    for entry in selected:
        rel = _safe_relative(entry)
        dest = download_root / rel
        if dest.exists():
            skipped += 1
            print(f"SKIP {dest}")
            continue
        print(f"GET  {entry['download_url']} -> {dest}")
        if args.dry_run:
            continue
        dest.parent.mkdir(parents=True, exist_ok=True)
        urllib.request.urlretrieve(entry["download_url"], dest)
        downloaded += 1

    print(f"selected={len(selected)} downloaded={downloaded} skipped={skipped} dry_run={int(args.dry_run)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
