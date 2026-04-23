#!/usr/bin/env python3
"""Auto-link BIOS files from ~/Games/RetroArch/mounts/bios/ sources
into ~/.config/retroarch/system/ when requirements are missing locally.

Uses the audit's own requirement catalog to decide what to look for.
For each `file`-kind requirement whose status is `missing` or
`optional_missing`, scan the BIOS mount trees for a candidate file
matching the requirement's basename (case-insensitive) and, when the
requirement carries an expected md5, verify the content before linking.

Creates symlinks (not copies) so the gdrive backing file is the source
of truth. RetroArch resolves symlinks so this is transparent.

Exit codes:
  0 — no missing requirements OR every missing one got linked
  1 — at least one requirement still missing after the scan
  2 — system_directory not found

JSON report: $XDG_STATE_HOME/retroarch/bios-link.json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

import retroarch_profile  # noqa: E402


BIOS_MOUNT_ROOTS = [
    Path.home() / "Games" / "RetroArch" / "mounts" / "bios",
]


def _md5(path: Path, *, chunk: int = 1 << 20) -> str | None:
    try:
        h = hashlib.md5()
        with path.open("rb") as fh:
            for block in iter(lambda: fh.read(chunk), b""):
                h.update(block)
        return h.hexdigest()
    except OSError:
        return None


def _expand(path: str | None, default: Path) -> Path:
    if not path:
        return default
    return Path(os.path.expandvars(os.path.expanduser(path)))


def _parse_requirement_rows() -> list[dict[str, str]]:
    """Extract `file`-kind requirement rows from the shared profile."""
    rows: list[dict[str, str]] = []
    for raw in retroarch_profile.DEFAULT_REQUIREMENT_ROWS:
        parts = raw.split("|")
        if len(parts) < 5:
            continue
        system, kind, req, rel_path, expected_md5 = parts[0], parts[1], parts[2], parts[3], parts[4]
        if kind != "file":
            continue
        rows.append({
            "system": system,
            "requirement": req,  # required, optional, conditional
            "relative_path": rel_path,
            "expected_md5": expected_md5 or "",
        })
    return rows


def _candidate_iter(target_basename: str, expected_md5: str) -> list[tuple[Path, str]]:
    """Yield (path, md5) pairs that could satisfy the requirement."""
    wanted = target_basename.lower()
    found: list[tuple[Path, str]] = []
    for root in BIOS_MOUNT_ROOTS:
        if not root.is_dir():
            continue
        for candidate in root.rglob("*"):
            if not candidate.is_file():
                continue
            if candidate.name.lower() != wanted:
                continue
            actual = _md5(candidate) or ""
            if expected_md5 and actual and actual != expected_md5:
                # Name matches but content doesn't — skip; file on gdrive
                # is a different dump, and we only want exact matches.
                continue
            found.append((candidate, actual))
    return found


def link_requirements(
    system_dir: Path, *, dry_run: bool,
) -> dict[str, Any]:
    report: dict[str, Any] = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "system_dir": str(system_dir),
        "dry_run": dry_run,
        "actions": [],
        "summary": {
            "scanned": 0, "already_present": 0, "linked": 0,
            "skipped_no_candidate": 0, "skipped_md5_mismatch": 0,
            "still_missing": 0,
        },
    }

    for row in _parse_requirement_rows():
        report["summary"]["scanned"] += 1
        rel_path = row["relative_path"]
        dest = system_dir / rel_path
        dest.parent.mkdir(parents=True, exist_ok=True)

        if dest.exists():
            report["summary"]["already_present"] += 1
            continue

        candidates = _candidate_iter(Path(rel_path).name, row["expected_md5"])
        if not candidates:
            # Distinguish "no file with that basename" from "name matched
            # but md5 didn't match anywhere".
            any_name = False
            for root in BIOS_MOUNT_ROOTS:
                if not root.is_dir():
                    continue
                if any(p.name.lower() == Path(rel_path).name.lower()
                       for p in root.rglob("*")
                       if p.is_file()):
                    any_name = True
                    break
            if any_name and row["expected_md5"]:
                report["summary"]["skipped_md5_mismatch"] += 1
                report["actions"].append({
                    "system": row["system"], "relative_path": rel_path,
                    "result": "md5_mismatch",
                })
            else:
                report["summary"]["skipped_no_candidate"] += 1
                if row["requirement"] == "required":
                    report["summary"]["still_missing"] += 1
                report["actions"].append({
                    "system": row["system"], "relative_path": rel_path,
                    "result": "no_candidate",
                })
            continue

        src, _md5sum = candidates[0]
        if dry_run:
            report["actions"].append({
                "system": row["system"], "relative_path": rel_path,
                "result": "would_link", "source": str(src),
            })
            report["summary"]["linked"] += 1
            continue

        try:
            dest.symlink_to(src)
        except OSError as exc:
            report["actions"].append({
                "system": row["system"], "relative_path": rel_path,
                "result": f"error: {exc}", "source": str(src),
            })
            report["summary"]["still_missing"] += 1
            continue

        report["actions"].append({
            "system": row["system"], "relative_path": rel_path,
            "result": "linked", "source": str(src),
        })
        report["summary"]["linked"] += 1

    return report


def _parse_args(argv: list[str]) -> argparse.Namespace:
    default_state = _expand(os.environ.get("XDG_STATE_HOME"), Path.home() / ".local" / "state")
    default_system_dir = Path.home() / ".config" / "retroarch" / "system"
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--system-dir", type=Path, default=default_system_dir,
                   help="RetroArch system_directory target.")
    p.add_argument("--bios-mount", type=Path, action="append", default=[],
                   help="Additional BIOS mount root(s) to scan. Repeatable.")
    p.add_argument("--report", type=Path,
                   default=default_state / "retroarch" / "bios-link.json",
                   help="JSON report destination.")
    p.add_argument("--dry-run", action="store_true",
                   help="Print the plan without creating symlinks.")
    p.add_argument("--json", action="store_true",
                   help="Emit the full JSON report instead of the summary line.")
    return p.parse_args(argv)


def main(argv: list[str]) -> int:
    args = _parse_args(argv)
    if args.bios_mount:
        BIOS_MOUNT_ROOTS[:] = list(args.bios_mount) + BIOS_MOUNT_ROOTS

    if not args.system_dir.is_dir():
        print(f"error: system_directory missing: {args.system_dir}", file=sys.stderr)
        return 2

    report = link_requirements(args.system_dir, dry_run=args.dry_run)

    if not args.dry_run:
        args.report.parent.mkdir(parents=True, exist_ok=True)
        args.report.write_text(json.dumps(report, indent=2))

    if args.json:
        print(json.dumps(report, indent=2))
    else:
        s = report["summary"]
        print(f"scanned={s['scanned']} already_present={s['already_present']} "
              f"linked={s['linked']} still_missing={s['still_missing']} "
              f"md5_mismatch={s['skipped_md5_mismatch']} "
              f"dry_run={'yes' if args.dry_run else 'no'}")
        for a in report["actions"]:
            result = a["result"]
            if result in ("linked", "would_link"):
                print(f"  {result.upper()}: {a['system']}/{a['relative_path']} -> {a['source']}")
            elif result == "md5_mismatch":
                print(f"  MD5_MISMATCH: {a['system']}/{a['relative_path']} (candidate content did not match expected hash)")
            elif result != "no_candidate":
                print(f"  {result}: {a['system']}/{a['relative_path']}")

    return 1 if report["summary"]["still_missing"] else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
