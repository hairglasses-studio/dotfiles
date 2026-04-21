#!/usr/bin/env python3
"""Fetch selected entries from a previously generated Archive.org manifest."""

from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import urllib.error
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


def _ia_nested_dest(entry: dict, dest: Path) -> Path:
    return dest.parent / Path(entry["archive_path"])


def _cleanup_empty_parents(path: Path, stop_at: Path) -> None:
    current = path
    while current != stop_at and current.exists():
        try:
            current.rmdir()
        except OSError:
            break
        current = current.parent


def _normalize_ia_download(entry: dict, dest: Path) -> None:
    nested = _ia_nested_dest(entry, dest)
    if dest.exists() or not nested.exists():
        return
    dest.parent.mkdir(parents=True, exist_ok=True)
    shutil.move(str(nested), str(dest))
    _cleanup_empty_parents(nested.parent, dest.parent)


def _ia_authenticated() -> bool:
    if shutil.which("ia") is None:
        return False
    result = subprocess.run(
        ["ia", "configure", "--check"],
        capture_output=True,
        text=True,
        check=False,
    )
    return result.returncode == 0


def _download_via_ia(entry: dict, dest: Path, dry_run: bool) -> None:
    cmd = [
        "ia",
        "download",
        entry["source_identifier"],
        entry["archive_path"],
        "--destdir",
        str(dest.parent),
        "--no-directories",
    ]
    if dry_run:
        return
    dest.parent.mkdir(parents=True, exist_ok=True)
    result = subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        if dest.exists():
            dest.unlink()
        detail = (result.stderr or result.stdout).strip() or f"exit {result.returncode}"
        raise RuntimeError(detail)
    _normalize_ia_download(entry, dest)


def _download_direct(entry: dict, dest: Path, dry_run: bool) -> None:
    print(f"GET  {entry['download_url']} -> {dest}")
    if dry_run:
        return
    dest.parent.mkdir(parents=True, exist_ok=True)
    try:
        urllib.request.urlretrieve(entry["download_url"], dest)
    except (urllib.error.HTTPError, urllib.error.URLError):
        if dest.exists():
            dest.unlink()
        raise


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Fetch selected Archive.org homebrew/public-domain entries.")
    parser.add_argument(
        "--manifest",
        help="Path to the manifest JSON.",
    )
    parser.add_argument(
        "--tier",
        action="append",
        choices=["public_domain", "verified_redistributable", "homebrew_unverified", "utility_unverified"],
        help="Download additional tiers. Default includes public_domain and verified_redistributable.",
    )
    parser.add_argument(
        "--system",
        action="append",
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
    parser.add_argument(
        "--transport",
        choices=["auto", "direct", "ia"],
        default="auto",
        help="Download transport. Default prefers authenticated ia when available.",
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
    allowed_tiers = set(args.tier or ["public_domain", "verified_redistributable"])
    allowed_systems = set(args.system or [])
    ia_ready = _ia_authenticated()
    use_ia = args.transport == "ia" or (args.transport == "auto" and ia_ready)
    if args.transport == "ia" and not ia_ready:
        print("ERR  requested ia transport, but `ia configure --check` failed")
        return 1

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

    downloaded = 0
    skipped = 0
    failed = 0
    for entry in selected:
        rel = _safe_relative(entry)
        dest = download_root / rel
        _normalize_ia_download(entry, dest)
        if dest.exists():
            skipped += 1
            print(f"SKIP {dest}")
            continue
        try:
            if use_ia:
                print(f"IA   {entry['source_identifier']}:{entry['archive_path']} -> {dest}")
                _download_via_ia(entry, dest, args.dry_run)
            else:
                _download_direct(entry, dest, args.dry_run)
            if not args.dry_run:
                downloaded += 1
        except (urllib.error.HTTPError, urllib.error.URLError, RuntimeError) as exc:
            failed += 1
            if dest.exists():
                dest.unlink()
            if use_ia:
                print(f"ERR  {entry['source_identifier']}:{entry['archive_path']} -> {dest} :: {exc}")
            else:
                print(f"ERR  {entry['download_url']} -> {dest} :: {exc}")

    print(
        f"selected={len(selected)} downloaded={downloaded} skipped={skipped} "
        f"failed={failed} dry_run={int(args.dry_run)} transport={'ia' if use_ia else 'direct'}"
    )
    return 0 if failed == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
