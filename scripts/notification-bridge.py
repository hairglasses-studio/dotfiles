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
import uuid
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


def _swaync_call(*args: str) -> bool:
    try:
        subprocess.run(
            ["swaync-client", *args],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            timeout=2,
            check=True,
        )
        return True
    except Exception:
        return False


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


def _set_dnd(value: str) -> dict[str, Any]:
    if value == "toggle":
        current = _swaync_dnd()
        target = not current if current is not None else True
    else:
        target = value == "true"
    return {
        "ok": _swaync_call("-dn", "true" if target else "false"),
        "action": "dnd",
        "dnd": target,
    }


def _close_all() -> dict[str, Any]:
    close_ok = _swaync_call("--close-all")
    hide_ok = _swaync_call("--hide-all")
    return {"ok": close_ok or hide_ok, "action": "close-all"}


def _clear_history(path: Path) -> dict[str, Any]:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text("", encoding="utf-8")
    return {"ok": True, "action": "clear-history", "log_path": str(path)}


def _normalise_entry(raw: dict[str, Any]) -> dict[str, Any]:
    entry = dict(raw)
    now = int(time.time())
    entry.setdefault("id", str(uuid.uuid4()))
    entry.setdefault("ts", now)
    entry.setdefault("time", now)
    entry.setdefault("app", entry.get("appName", ""))
    entry.setdefault("summary", "")
    entry.setdefault("body", "")
    entry.setdefault("urgency", "normal")
    entry.setdefault("visible", True)
    entry.setdefault("dismissed", False)
    entry["text"] = _entry_text(entry)
    return entry


def _append_entry(path: Path, entry_json: str) -> dict[str, Any]:
    try:
        raw = json.loads(entry_json)
    except json.JSONDecodeError as exc:
        return {"ok": False, "action": "append-entry", "error": str(exc)}
    if not isinstance(raw, dict):
        return {"ok": False, "action": "append-entry", "error": "entry must be a JSON object"}
    entry = _normalise_entry(raw)
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(entry, ensure_ascii=True, separators=(",", ":")) + "\n")
    return {"ok": True, "action": "append-entry", "entry": entry, "log_path": str(path)}


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--history", type=Path, default=DEFAULT_HISTORY)
    parser.add_argument("--limit", type=int, default=12)
    parser.add_argument("--include-dismissed", action="store_true")
    parser.add_argument("--watch", action="store_true")
    parser.add_argument("--interval", type=float, default=15.0)
    parser.add_argument("--dnd", choices=["true", "false", "toggle"], help="Set or toggle swaync DND")
    parser.add_argument("--close-all", action="store_true", help="Close and hide all swaync notifications")
    parser.add_argument("--clear-history", action="store_true", help="Truncate the local notification history")
    parser.add_argument("--append-entry-json", help="Append one notification-history JSON object")
    args = parser.parse_args()

    if args.dnd:
        _emit(_set_dnd(args.dnd))
        return 0
    if args.close_all:
        _emit(_close_all())
        return 0
    if args.clear_history:
        _emit(_clear_history(args.history))
        return 0
    if args.append_entry_json:
        _emit(_append_entry(args.history, args.append_entry_json))
        return 0

    interval = max(1.0, args.interval)
    while True:
        _emit(_payload(args.history, max(1, args.limit), args.include_dismissed))
        if not args.watch:
            return 0
        time.sleep(interval)


if __name__ == "__main__":
    raise SystemExit(main())
