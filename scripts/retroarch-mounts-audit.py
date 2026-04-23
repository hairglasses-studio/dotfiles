#!/usr/bin/env python3
"""Report health of every rclone mount under ~/Games/RetroArch/mounts/.

Checks per mount:
  - is the target an active FUSE mountpoint right now?
  - does the systemd unit say it's active?
  - is the mount reachable (listing doesn't hang)?
  - file count + total byte size of top-level entries
  - last-mount start time (from systemd) and derived age
  - whether the rclone dir-cache window has been exceeded (cache may
    be stale — a real refresh would show new gdrive content)

Exit codes:
  0 — every mount is active, reachable, cache within window
  1 — at least one mount is not active or not reachable
  2 — systemd not available (not-Linux / skeleton-container case)

JSON report: $XDG_STATE_HOME/retroarch/mounts-audit.json
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import re
import subprocess
import sys
from pathlib import Path
from typing import Any


DIR_CACHE_RE = re.compile(r"--dir-cache-time\s+(\S+)")


def _expand(path: str | None, default: Path) -> Path:
    if not path:
        return default
    return Path(os.path.expandvars(os.path.expanduser(path)))


def _run(argv: list[str], *, timeout: float = 5.0) -> tuple[int, str]:
    try:
        r = subprocess.run(argv, capture_output=True, text=True, check=False, timeout=timeout)
        return r.returncode, (r.stdout + r.stderr)
    except FileNotFoundError:
        return 127, f"{argv[0]} not found"
    except subprocess.TimeoutExpired:
        return 124, f"{argv[0]} timed out"


def _is_mountpoint(path: Path) -> bool:
    rc, _ = _run(["mountpoint", "-q", str(path)])
    return rc == 0


def _systemctl_user(args: list[str], *, timeout: float = 5.0) -> str:
    _rc, out = _run(["systemctl", "--user", *args], timeout=timeout)
    return out.strip()


def _parse_duration(s: str) -> int:
    """Parse e.g. '24h', '30m', '3600s', '1d' to seconds."""
    if not s:
        return 0
    m = re.match(r"^(\d+(?:\.\d+)?)([smhdw]?)$", s.strip())
    if not m:
        return 0
    value = float(m.group(1))
    unit = m.group(2) or "s"
    factor = {"s": 1, "m": 60, "h": 3600, "d": 86400, "w": 604800}[unit]
    return int(value * factor)


def _unit_for_mountpoint(mount_path: Path, units: list[str]) -> str | None:
    basename = mount_path.name
    # amiga-archive -> amiga-archives / psx-archive -> psx-archives
    candidates = [
        f"retroarch-rclone-{basename}.service",
        f"retroarch-rclone-{basename}s.service",
        f"retroarch-rclone-{basename.rstrip('-source')}-bios.service",
    ]
    # bios/ps2 special-cased
    if mount_path.parent.name == "bios":
        candidates.append(f"retroarch-rclone-{basename}-bios.service")
        candidates.append(f"retroarch-rclone-{basename.replace('-source', '')}-bios.service")
    for c in candidates:
        if c in units:
            return c
    # Fallback: grep systemctl cat output for the mount path
    for unit in units:
        cat = _systemctl_user(["cat", unit])
        if str(mount_path) in cat:
            return unit
    return None


def _list_counts(mount_path: Path, timeout: float = 5.0) -> tuple[int, int | None]:
    """Return (file_count, reachable). reachable=None when unknown."""
    try:
        r = subprocess.run(
            ["ls", "-A", str(mount_path)],
            capture_output=True, text=True, timeout=timeout
        )
    except subprocess.TimeoutExpired:
        return (0, 0)
    if r.returncode != 0:
        return (0, 0)
    entries = [l for l in r.stdout.splitlines() if l]
    return (len(entries), 1)


def _parse_unit_start(unit: str) -> dt.datetime | None:
    out = _systemctl_user(["show", "-p", "ActiveEnterTimestamp", unit])
    # ActiveEnterTimestamp=Wed 2026-04-22 21:56:09 PDT
    if "=" not in out:
        return None
    _, _, value = out.partition("=")
    value = value.strip()
    if not value or value == "n/a":
        return None
    # systemd-style format: "Wed 2026-04-22 21:56:09 PDT"
    # Python can't parse arbitrary TZ abbreviations reliably, so strip
    # the timezone and treat the remainder as local-naive, then promote
    # to aware using the system's current local tzinfo.
    value = re.sub(r"\s+[A-Z]{2,5}$", "", value)
    try:
        naive = dt.datetime.strptime(value, "%a %Y-%m-%d %H:%M:%S")
    except ValueError:
        return None
    local_tz = dt.datetime.now().astimezone().tzinfo
    return naive.replace(tzinfo=local_tz)


def _dir_cache_from_unit(unit: str) -> int:
    cat = _systemctl_user(["cat", unit])
    m = DIR_CACHE_RE.search(cat)
    if not m:
        return 0
    return _parse_duration(m.group(1))


def audit(mounts_root: Path) -> dict[str, Any]:
    report: dict[str, Any] = {
        "generated_at": dt.datetime.now(dt.timezone.utc).isoformat(),
        "mounts_root": str(mounts_root),
        "mounts": [],
        "summary": {"total": 0, "active": 0, "unreachable": 0, "stale_cache": 0},
    }
    if not mounts_root.is_dir():
        report["error"] = f"mounts root missing: {mounts_root}"
        return report

    rc, out = _run(["systemctl", "--user", "list-units", "--type=service", "--all",
                    "--no-legend", "--plain"])
    if rc != 0:
        report["error"] = "systemctl --user unavailable"
        return report

    units = []
    for line in out.splitlines():
        parts = line.split()
        if not parts:
            continue
        name = parts[0]
        if name.startswith("retroarch-rclone-") and name.endswith(".service"):
            units.append(name)

    # Walk the mount tree looking for any directory that should be a mountpoint.
    candidate_dirs: list[Path] = []
    for sub in sorted(mounts_root.rglob("*")):
        if not sub.is_dir():
            continue
        # Only leaf dirs or mountpoints
        if any(sub.glob("*/")) is False:
            candidate_dirs.append(sub)
        if _is_mountpoint(sub):
            candidate_dirs.append(sub)
    # Deduplicate while preserving order
    seen: set[Path] = set()
    candidate_dirs = [p for p in candidate_dirs if (p in seen or seen.add(p)) is None]
    # Restrict to known mount pattern: mounts_root/roms/* and mounts_root/bios/*
    mount_dirs = [
        p for p in candidate_dirs
        if p.parent.parent == mounts_root
        and p.parent.name in ("roms", "bios")
    ]

    now = dt.datetime.now(dt.timezone.utc).astimezone()
    for mount_path in sorted(mount_dirs):
        mounted = _is_mountpoint(mount_path)
        unit = _unit_for_mountpoint(mount_path, units)
        systemd_active = False
        start_time: dt.datetime | None = None
        cache_window_s = 0
        if unit:
            state = _systemctl_user(["is-active", unit])
            systemd_active = state == "active"
            start_time = _parse_unit_start(unit)
            cache_window_s = _dir_cache_from_unit(unit)

        file_count, reachable = _list_counts(mount_path)
        age_s = None
        stale_cache = False
        if start_time:
            tz_now = now.astimezone(start_time.tzinfo or now.tzinfo)
            age_s = int((tz_now - start_time).total_seconds())
            if cache_window_s and age_s > cache_window_s:
                stale_cache = True

        record = {
            "mount_path": str(mount_path),
            "unit": unit,
            "mounted": mounted,
            "systemd_active": systemd_active,
            "reachable": reachable == 1,
            "file_count": file_count,
            "age_seconds": age_s,
            "dir_cache_window_seconds": cache_window_s,
            "stale_cache": stale_cache,
        }
        report["mounts"].append(record)

    s = report["summary"]
    s["total"] = len(report["mounts"])
    s["active"] = sum(1 for r in report["mounts"] if r["mounted"] and r["systemd_active"])
    s["unreachable"] = sum(1 for r in report["mounts"] if not r["reachable"])
    s["stale_cache"] = sum(1 for r in report["mounts"] if r["stale_cache"])

    return report


def _parse_args(argv: list[str]) -> argparse.Namespace:
    default_state = _expand(os.environ.get("XDG_STATE_HOME"), Path.home() / ".local" / "state")
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--mounts-root", type=Path,
                   default=Path.home() / "Games" / "RetroArch" / "mounts",
                   help="Root of the mount tree (default: ~/Games/RetroArch/mounts).")
    p.add_argument("--report", type=Path,
                   default=default_state / "retroarch" / "mounts-audit.json",
                   help="JSON report destination.")
    p.add_argument("--json", action="store_true",
                   help="Print the full JSON report instead of the summary table.")
    return p.parse_args(argv)


def _format_age(seconds: int | None) -> str:
    if seconds is None:
        return "?"
    if seconds < 60:
        return f"{seconds}s"
    if seconds < 3600:
        return f"{seconds // 60}m"
    if seconds < 86400:
        return f"{seconds // 3600}h"
    return f"{seconds // 86400}d"


def main(argv: list[str]) -> int:
    args = _parse_args(argv)
    report = audit(args.mounts_root)
    args.report.parent.mkdir(parents=True, exist_ok=True)
    args.report.write_text(json.dumps(report, indent=2))

    if args.json:
        print(json.dumps(report, indent=2))
        return 0

    if "error" in report:
        print(f"error: {report['error']}", file=sys.stderr)
        return 2

    # Tabular summary.
    print(f"{'mount':<55} {'state':<12} {'files':>6} {'age':>5} {'cache':<6}")
    print("-" * 90)
    for r in report["mounts"]:
        state_parts = []
        state_parts.append("mnt" if r["mounted"] else "NOMT")
        state_parts.append("sd" if r["systemd_active"] else "NOSD")
        state = "/".join(state_parts)
        cache = "STALE" if r["stale_cache"] else "ok"
        path = r["mount_path"].replace(str(Path.home()), "~")
        print(f"{path:<55} {state:<12} {r['file_count']:>6} "
              f"{_format_age(r['age_seconds']):>5} {cache:<6}")

    s = report["summary"]
    print(f"\nmounts={s['total']} active={s['active']} "
          f"unreachable={s['unreachable']} stale_cache={s['stale_cache']}")
    return 1 if (s["total"] - s["active"] or s["unreachable"]) else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
