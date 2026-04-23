#!/usr/bin/env python3
"""Audit RetroArch user playlists for entries pointing at missing cores.

Walks every `*.lpl` under RetroArch's playlists directory and flags
entries whose `core_path` references a non-existent shared library.
RetroArch silently returns the user to the playlist with no error
when they try to launch a broken entry — the UI just "doesn't
respond" to the click.

Sentinel values for "user picks at launch time" are treated as
intentional, not drift: `""`, `"DETECT"`.

Exit codes:
  0 — no broken references
  1 — at least one entry points at a missing core
  2 — playlists directory missing

JSON report location: $XDG_STATE_HOME/retroarch/playlist-audit.json
"""

from __future__ import annotations

import argparse
import json
import os
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


DETECT_SENTINELS = {"", "DETECT"}


def _expand(value: str | None, default: Path) -> Path:
    if not value:
        return default
    return Path(os.path.expandvars(os.path.expanduser(value)))


def audit(playlists_dir: Path) -> dict[str, Any]:
    report: dict[str, Any] = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "playlists_dir": str(playlists_dir),
        "summary": {"playlists": 0, "entries": 0, "unassigned": 0, "broken": 0},
        "broken": [],
    }
    if not playlists_dir.is_dir():
        report["error"] = f"{playlists_dir} is not a directory"
        return report

    for lpl in sorted(playlists_dir.glob("*.lpl")):
        try:
            data = json.loads(lpl.read_text())
        except (OSError, json.JSONDecodeError):
            continue
        report["summary"]["playlists"] += 1
        for item in data.get("items", []):
            report["summary"]["entries"] += 1
            cp_raw = item.get("core_path", "")
            cp = cp_raw.strip() if isinstance(cp_raw, str) else ""
            if cp in DETECT_SENTINELS:
                report["summary"]["unassigned"] += 1
                continue
            if Path(cp).exists():
                continue
            report["summary"]["broken"] += 1
            report["broken"].append(
                {
                    "playlist": lpl.name,
                    "label": item.get("label"),
                    "core_path": cp,
                    "rom_path": item.get("path"),
                }
            )
    return report


def _parse_args(argv: list[str]) -> argparse.Namespace:
    default_playlists = Path.home() / ".config" / "retroarch" / "playlists"
    default_state = Path(os.environ.get("XDG_STATE_HOME", str(Path.home() / ".local" / "state")))
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--playlists-dir", type=Path, default=default_playlists,
                   help="Path to RetroArch playlists directory.")
    p.add_argument("--report", type=Path,
                   default=default_state / "retroarch" / "playlist-audit.json",
                   help="Where to write the JSON report.")
    p.add_argument("--json", action="store_true",
                   help="Print the full JSON report instead of the summary line.")
    return p.parse_args(argv)


def main(argv: list[str]) -> int:
    args = _parse_args(argv)
    report = audit(args.playlists_dir)

    if "error" in report:
        print(f"error: {report['error']}", file=sys.stderr)
        return 2

    args.report.parent.mkdir(parents=True, exist_ok=True)
    args.report.write_text(json.dumps(report, indent=2))

    if args.json:
        print(json.dumps(report, indent=2))
    else:
        s = report["summary"]
        print(
            f"playlists={s['playlists']} entries={s['entries']} "
            f"unassigned={s['unassigned']} broken={s['broken']}"
        )
        for b in report["broken"][:10]:
            print(f"  BROKEN: {b['playlist']}: {b.get('label') or '?'} -> {b['core_path']}")
        if s["broken"] > 10:
            print(f"  (+ {s['broken'] - 10} more — see {args.report})")

    return 1 if report["summary"]["broken"] else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
