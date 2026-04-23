#!/usr/bin/env python3
"""Fetch missing RetroArch playlist thumbnails from the libretro-thumbnails GitHub repos.

Consumes (or recomputes) the output of `retroarch-thumbnail-audit.py`
and downloads the missing PNG files from the canonical upstream:

    https://github.com/libretro-thumbnails/<System>/raw/master/<Category>/<Label>.png

Each RetroArch system has its own repo with the same name as the
playlist stem (e.g. `Nintendo - Game Boy`, `Sony - PlayStation 2`).

Exit codes:
  0 — every previously-missing thumbnail is now present (or --dry-run)
  1 — at least one fetch failed
  2 — audit could not be computed

JSON report at `$XDG_STATE_HOME/retroarch/thumbnail-fill.json`.

Sandbox note: external GitHub fetches are blocked inside most agent
sandboxes. Run this from a user shell when fill is needed. `--dry-run`
works anywhere — it just prints the URLs.
"""

from __future__ import annotations

import argparse
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


_SCRIPT_DIR = Path(os.path.realpath(__file__)).parent
sys.path.insert(0, str(_SCRIPT_DIR))

# Import the audit module to recompute on the fly.
# The audit script has a hyphen in its name; Python can't import that
# directly, so we load it via importlib.
import importlib.util

_AUDIT_PATH = _SCRIPT_DIR / "retroarch-thumbnail-audit.py"
_spec = importlib.util.spec_from_file_location("retroarch_thumbnail_audit", _AUDIT_PATH)
_audit_module = importlib.util.module_from_spec(_spec)
_spec.loader.exec_module(_audit_module)  # type: ignore[union-attr]


BASE_URL = "https://github.com/libretro-thumbnails"


def _expand(value: str | None, default: Path) -> Path:
    if not value:
        return default
    return Path(os.path.expandvars(os.path.expanduser(value)))


def _build_url(system: str, category: str, label: str) -> str:
    # libretro-thumbnails repo names replace spaces with underscores:
    # "Nintendo - Nintendo Entertainment System" → "Nintendo_-_Nintendo_Entertainment_System"
    # File paths inside the repo keep the original label verbatim, so
    # spaces there get URL-encoded as %20 the normal way.
    repo_name = system.replace(" ", "_")
    encoded_repo = urllib.parse.quote(repo_name)
    encoded_label = urllib.parse.quote(label)
    return f"{BASE_URL}/{encoded_repo}/raw/master/{category}/{encoded_label}.png"


def _fetch(url: str, dest: Path, *, timeout: float) -> tuple[bool, str]:
    dest.parent.mkdir(parents=True, exist_ok=True)
    req = urllib.request.Request(url, headers={"User-Agent": "retroarch-thumbnail-fill/1.0"})
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            data = resp.read()
    except urllib.error.HTTPError as exc:
        return False, f"http {exc.code}"
    except urllib.error.URLError as exc:
        reason = getattr(exc, "reason", str(exc))
        return False, f"urlerror: {reason}"
    except OSError as exc:
        return False, f"oserror: {exc}"
    tmp = dest.with_suffix(dest.suffix + ".tmp")
    try:
        tmp.write_bytes(data)
        tmp.replace(dest)
    except OSError as exc:
        return False, f"write: {exc}"
    return True, "ok"


def _parse_args(argv: list[str]) -> argparse.Namespace:
    default_state = _expand(os.environ.get("XDG_STATE_HOME"), Path.home() / ".local" / "state")
    default_cfg = Path.home() / ".config" / "retroarch"
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--playlists-dir", type=Path, default=default_cfg / "playlists")
    p.add_argument("--thumbnails-dir", type=Path, default=default_cfg / "thumbnails")
    p.add_argument("--report", type=Path,
                   default=default_state / "retroarch" / "thumbnail-fill.json")
    p.add_argument("--dry-run", action="store_true",
                   help="Print planned URLs without fetching.")
    p.add_argument("--limit", type=int, default=0,
                   help="Stop after N fetch attempts (0 = unlimited).")
    p.add_argument("--timeout", type=float, default=10.0,
                   help="Per-request timeout in seconds.")
    p.add_argument("--sleep", type=float, default=0.05,
                   help="Sleep between fetches in seconds (rate-limit courtesy).")
    p.add_argument("--system", action="append", default=[],
                   help="Restrict to named systems (repeatable). Default: all.")
    return p.parse_args(argv)


def main(argv: list[str]) -> int:
    args = _parse_args(argv)

    audit = _audit_module.audit(args.playlists_dir, args.thumbnails_dir)
    if "error" in audit:
        print(f"error: {audit['error']}", file=sys.stderr)
        return 2

    missing = audit["missing"]
    if args.system:
        allowed = set(args.system)
        missing = [m for m in missing if m["system"] in allowed]

    # Rebuild the full missing list from systems (audit caps the
    # top-level `missing` at 200 rows; we need them all for fill).
    full_missing: list[dict[str, str]] = []
    for sys_row in audit.get("systems", []):
        if args.system and sys_row["system"] not in set(args.system):
            continue
        system = sys_row["system"]
        # Re-walk the playlist to recover all missing entries (the
        # audit's capped list isn't enough for fill).
        lpl = args.playlists_dir / sys_row["playlist"]
        try:
            data = json.loads(lpl.read_text())
        except (OSError, json.JSONDecodeError):
            continue
        for item in data.get("items") or []:
            label = _audit_module._thumbnail_label(item)
            if not label:
                continue
            for category in _audit_module.CATEGORIES:
                png = args.thumbnails_dir / system / category / f"{label}.png"
                if png.is_file():
                    continue
                full_missing.append({
                    "system": system,
                    "category": category,
                    "label": label,
                    "dest": str(png),
                })

    total = len(full_missing)
    if args.limit and total > args.limit:
        full_missing = full_missing[: args.limit]

    report: dict[str, Any] = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "dry_run": args.dry_run,
        "requested": total,
        "attempted": len(full_missing),
        "fetched": 0,
        "failed": 0,
        "actions": [],
    }

    for i, row in enumerate(full_missing, 1):
        url = _build_url(row["system"], row["category"], row["label"])
        if args.dry_run:
            report["actions"].append({**row, "url": url, "result": "planned"})
            continue
        ok, msg = _fetch(url, Path(row["dest"]), timeout=args.timeout)
        report["actions"].append({**row, "url": url, "result": msg})
        if ok:
            report["fetched"] += 1
        else:
            report["failed"] += 1
        if i % 50 == 0:
            print(f"  progress: {i}/{len(full_missing)} "
                  f"(fetched={report['fetched']} failed={report['failed']})",
                  file=sys.stderr)
        if args.sleep > 0:
            time.sleep(args.sleep)

    if not args.dry_run:
        args.report.parent.mkdir(parents=True, exist_ok=True)
        # Slim the actions list for the on-disk report (else it can hit MB).
        slim_report = {**report, "actions": report["actions"][:500]}
        args.report.write_text(json.dumps(slim_report, indent=2))

    if args.dry_run:
        # In dry-run, print a summary + the first 20 URLs so the user
        # can verify the expected shape before a live run.
        print(f"dry_run=yes requested={total} attempted={len(full_missing)}")
        for row in report["actions"][:20]:
            print(f"  {row['system']} / {row['category']} / {row['label']}")
            print(f"    {row['url']}")
        return 0

    print(f"requested={total} fetched={report['fetched']} failed={report['failed']}")
    return 1 if report["failed"] else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
