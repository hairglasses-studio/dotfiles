#!/usr/bin/env python3
"""notification-bridge.py — JSON bridge over the local notification history.

This is intentionally non-owning: it reads the dbus-monitor history maintained
by notification-history-listener.py and does not register
org.freedesktop.Notifications. Quickshell can consume this while swaync remains
the active notification daemon.
"""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import time
from pathlib import Path
from typing import Any


STATE_HOME = Path(os.environ.get("XDG_STATE_HOME", Path.home() / ".local" / "state"))
DEFAULT_HISTORY = STATE_HOME / "dotfiles" / "desktop-control" / "notifications" / "history.jsonl"


def _entry_text(entry: dict[str, Any]) -> str:
    summary = str(entry.get("summary", "")).strip()
    body = str(entry.get("body", "")).strip()
    if body and body != summary:
        return f"{summary}: {body}"
    return summary or body


def _read_entries(path: Path, limit: int, include_dismissed: bool) -> list[dict[str, Any]]:
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except FileNotFoundError:
        return []

    entries: list[dict[str, Any]] = []
    for line in reversed(lines[-max(limit * 4, limit, 25) :]):
        try:
            entry = json.loads(line)
        except json.JSONDecodeError:
            continue
        if not include_dismissed and (entry.get("dismissed") or entry.get("visible") is False):
            continue
        entry["text"] = _entry_text(entry)
        entries.append(entry)
        if len(entries) >= limit:
            break
    return entries


def _swaync_text(*args: str) -> str | None:
    try:
        return subprocess.check_output(
            ["swaync-client", *args],
            stderr=subprocess.DEVNULL,
            text=True,
            timeout=2,
        ).strip()
    except Exception:
        return None


def _swaync_count() -> int | None:
    text = _swaync_text("-c")
    if text is None:
        return None
    try:
        return int(text)
    except ValueError:
        return None


def _swaync_dnd() -> bool | None:
    text = _swaync_text("-D")
    if text is None:
        return None
    return text.lower() == "true"


def _payload(path: Path, limit: int, include_dismissed: bool) -> dict[str, Any]:
    entries = _read_entries(path, limit, include_dismissed)
    critical = sum(1 for entry in entries if entry.get("urgency") == "critical")
    latest = entries[0] if entries else None
    return {
        "ok": True,
        "ts": int(time.time()),
        "log_path": str(path),
        "visible": len(entries),
        "daemon_count": _swaync_count(),
        "dnd": _swaync_dnd(),
        "critical": critical,
        "latest": latest,
        "entries": entries,
    }


def _emit(payload: dict[str, Any]) -> None:
    print(json.dumps(payload, ensure_ascii=True, separators=(",", ":")), flush=True)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--history", type=Path, default=DEFAULT_HISTORY)
    parser.add_argument("--limit", type=int, default=12)
    parser.add_argument("--include-dismissed", action="store_true")
    parser.add_argument("--watch", action="store_true")
    parser.add_argument("--interval", type=float, default=15.0)
    args = parser.parse_args()

    interval = max(1.0, args.interval)
    while True:
        _emit(_payload(args.history, max(1, args.limit), args.include_dismissed))
        if not args.watch:
            return 0
        time.sleep(interval)


if __name__ == "__main__":
    raise SystemExit(main())
