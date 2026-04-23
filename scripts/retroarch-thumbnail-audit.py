#!/usr/bin/env python3
"""Audit RetroArch playlist thumbnails (boxart, snap, title) per system.

Walks every `~/.config/retroarch/playlists/*.lpl` and, for each item,
checks whether the expected thumbnail files exist under
`~/.config/retroarch/thumbnails/<System>/{Named_Boxarts,Named_Snaps,Named_Titles}/`.

Reports per-system fill rate (N present / M expected) across each of
the three thumbnail categories and an overall total. Writes the full
report (including the list of missing entries) to
`$XDG_STATE_HOME/retroarch/thumbnail-audit.json`.

Exit 0 if every expected thumbnail is present; 1 if any are missing.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
from collections import defaultdict
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


CATEGORIES = ("Named_Boxarts", "Named_Snaps", "Named_Titles")


def _expand(value: str | None, default: Path) -> Path:
    if not value:
        return default
    return Path(os.path.expandvars(os.path.expanduser(value)))


def _playlist_system_name(lpl_path: Path) -> str:
    # RetroArch's thumbnail convention: the system folder name is the
    # playlist stem (filename without .lpl).
    return lpl_path.stem


def _thumbnail_label(item: dict[str, Any]) -> str | None:
    """Sanitize a playlist item label into the canonical thumbnail filename.

    RetroArch replaces filesystem-unsafe characters in the label when
    looking up thumbnails. Upstream libretro-thumbnails follows the
    same rule: `& / * ? " < > | :` → `_`.
    """
    raw = (item.get("label") or "").strip()
    if not raw:
        return None
    # Match RetroArch's sanitize rule: replace reserved chars with underscore.
    sanitized = re.sub(r'[&*/?"<>|:]', "_", raw)
    # Collapse trailing whitespace.
    return sanitized.strip()


def audit(playlists_dir: Path, thumbnails_dir: Path) -> dict[str, Any]:
    report: dict[str, Any] = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "playlists_dir": str(playlists_dir),
        "thumbnails_dir": str(thumbnails_dir),
        "systems": [],
        "summary": {
            "playlists": 0,
            "entries": 0,
            "expected": 0,
            "present": 0,
            "missing": 0,
        },
        "missing": [],  # {system, label, category} rows, capped at 200 for output
    }
    if not playlists_dir.is_dir():
        report["error"] = f"playlists dir missing: {playlists_dir}"
        return report

    # Walk every .lpl that isn't a history playlist.
    lpl_files = sorted(
        p for p in playlists_dir.glob("*.lpl")
        if not p.name.startswith("content_")
        and not p.name.startswith("favorites")
    )

    missing_rows: list[dict[str, str]] = []

    for lpl in lpl_files:
        report["summary"]["playlists"] += 1
        system = _playlist_system_name(lpl)
        try:
            data = json.loads(lpl.read_text())
        except (OSError, json.JSONDecodeError):
            continue
        items = data.get("items") or []
        system_row: dict[str, Any] = {
            "system": system,
            "playlist": lpl.name,
            "entries": len(items),
            "expected": 0,
            "present": defaultdict(int),
            "missing": defaultdict(int),
        }
        for item in items:
            report["summary"]["entries"] += 1
            label = _thumbnail_label(item)
            if not label:
                continue
            for category in CATEGORIES:
                system_row["expected"] += 1
                report["summary"]["expected"] += 1
                png = thumbnails_dir / system / category / f"{label}.png"
                if png.is_file():
                    system_row["present"][category] += 1
                    report["summary"]["present"] += 1
                else:
                    system_row["missing"][category] += 1
                    report["summary"]["missing"] += 1
                    missing_rows.append({
                        "system": system,
                        "category": category,
                        "label": label,
                    })
        # Convert defaultdicts to plain dicts for JSON serialization.
        system_row["present"] = dict(system_row["present"])
        system_row["missing"] = dict(system_row["missing"])
        report["systems"].append(system_row)

    # Cap the missing list so the JSON stays manageable; summary counter
    # preserves the full count.
    report["missing"] = missing_rows[:200]
    return report


def _parse_args(argv: list[str]) -> argparse.Namespace:
    default_state = _expand(os.environ.get("XDG_STATE_HOME"), Path.home() / ".local" / "state")
    default_cfg = Path.home() / ".config" / "retroarch"
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--playlists-dir", type=Path, default=default_cfg / "playlists")
    p.add_argument("--thumbnails-dir", type=Path, default=default_cfg / "thumbnails")
    p.add_argument("--report", type=Path,
                   default=default_state / "retroarch" / "thumbnail-audit.json")
    p.add_argument("--json", action="store_true",
                   help="Emit the full JSON report instead of the summary line.")
    return p.parse_args(argv)


def main(argv: list[str]) -> int:
    args = _parse_args(argv)
    report = audit(args.playlists_dir, args.thumbnails_dir)
    args.report.parent.mkdir(parents=True, exist_ok=True)
    args.report.write_text(json.dumps(report, indent=2))

    if "error" in report:
        print(f"error: {report['error']}", file=sys.stderr)
        return 2

    if args.json:
        print(json.dumps(report, indent=2))
    else:
        s = report["summary"]
        print(
            f"playlists={s['playlists']} entries={s['entries']} "
            f"expected={s['expected']} present={s['present']} "
            f"missing={s['missing']}"
        )
        for row in report["systems"][:14]:
            fill = sum(row["present"].values())
            total = row["expected"]
            pct = (fill * 100 // total) if total else 100
            print(f"  {row['system']:<55} {fill:>5}/{total:<5} ({pct}%)")

    return 1 if report["summary"]["missing"] else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
