#!/usr/bin/env python3
"""Emit recent desktop notification history for eww."""

from __future__ import annotations

import json
import os
import subprocess
from datetime import datetime
from pathlib import Path


STATE_HOME = Path(os.environ.get("XDG_STATE_HOME", Path.home() / ".local" / "state"))
LOG_PATH = STATE_HOME / "dotfiles" / "desktop-control" / "notifications" / "history.jsonl"
MAX_ENTRIES = 5


def current_swaync_count() -> int | None:
    try:
        proc = subprocess.run(
            ["swaync-client", "-c"],
            check=False,
            capture_output=True,
            text=True,
            timeout=2,
        )
    except (FileNotFoundError, subprocess.TimeoutExpired):
        return None

    if proc.returncode != 0:
        return None

    try:
        return int(proc.stdout.strip())
    except ValueError:
        return None


def format_time(timestamp: str) -> str:
    if not timestamp:
        return ""
    try:
        return datetime.fromisoformat(timestamp).strftime("%H:%M")
    except ValueError:
        return timestamp


def load_entries() -> list[dict[str, object]]:
    if not LOG_PATH.exists():
        return []

    entries: list[dict[str, object]] = []
    for raw in LOG_PATH.read_text(encoding="utf-8").splitlines():
        if not raw.strip():
            continue
        try:
            data = json.loads(raw)
        except json.JSONDecodeError:
            continue
        if not isinstance(data, dict):
            continue
        entries.append(
            {
                "app": str(data.get("app", "")),
                "summary": str(data.get("summary", "")),
                "body": str(data.get("body", "")),
                "urgency": str(data.get("urgency", "normal")),
                "timestamp": str(data.get("timestamp", "")),
                "time": format_time(str(data.get("timestamp", ""))),
                "dismissed": bool(data.get("dismissed", False)),
            }
        )
    entries.reverse()
    return entries[:MAX_ENTRIES]


def main() -> int:
    entries = load_entries()
    fallback_count = sum(1 for entry in entries if not entry.get("dismissed", False))
    swaync_count = current_swaync_count()
    payload = {
        "count": swaync_count if swaync_count is not None else fallback_count,
        "entry_count": len(entries),
        "entries": entries,
    }
    print(json.dumps(payload, ensure_ascii=True))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
