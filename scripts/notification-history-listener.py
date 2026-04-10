#!/usr/bin/env python3
"""Persist desktop notifications into a local JSONL history log."""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path


STATE_HOME = Path(os.environ.get("XDG_STATE_HOME", Path.home() / ".local" / "state"))
LOG_DIR = STATE_HOME / "dotfiles" / "desktop-control" / "notifications"
LOG_PATH = LOG_DIR / "history.jsonl"
DBUS_CMD = [
    "dbus-monitor",
    "--session",
    "type='method_call',interface='org.freedesktop.Notifications',member='Notify'",
]

STRING_RE = re.compile(r'^\s*string\s+(".*")\s*$')
URGENCY_RE = re.compile(r"\bbyte\s+(\d+)\b")
EXPIRES_RE = re.compile(r"^\s*int32\s+-?\d+\s*$")


def _decode_dbus_string(token: str) -> str:
    try:
        return json.loads(token)
    except Exception:
        return token.strip('"')


def _urgency_from_block(lines: list[str]) -> str:
    for idx, line in enumerate(lines):
        if 'string "urgency"' not in line:
            continue
        for candidate in lines[idx + 1 : idx + 4]:
            match = URGENCY_RE.search(candidate)
            if not match:
                continue
            value = int(match.group(1))
            return {0: "low", 1: "normal", 2: "critical"}.get(value, "normal")
    return "normal"


def _parse_notify_block(lines: list[str]) -> dict[str, object] | None:
    strings: list[str] = []
    for line in lines:
        match = STRING_RE.match(line)
        if match:
            strings.append(_decode_dbus_string(match.group(1)))

    if len(strings) < 4:
        return None

    app, _icon, summary, body = strings[:4]
    timestamp = datetime.now(timezone.utc).astimezone().isoformat()
    return {
        "id": f"dbus-{int(datetime.now(timezone.utc).timestamp() * 1_000_000)}",
        "app": app,
        "summary": summary,
        "body": body,
        "urgency": _urgency_from_block(lines),
        "timestamp": timestamp,
        "visible": True,
        "dismissed": False,
        "source": "dbus-monitor",
    }


def _append_entry(entry: dict[str, object]) -> None:
    LOG_DIR.mkdir(parents=True, exist_ok=True)
    with LOG_PATH.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(entry, ensure_ascii=True))
        handle.write("\n")


def main() -> int:
    try:
        proc = subprocess.Popen(
            DBUS_CMD,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            bufsize=1,
        )
    except FileNotFoundError:
        print("dbus-monitor not found on PATH", file=sys.stderr)
        return 1

    current: list[str] = []
    collecting = False

    assert proc.stdout is not None
    for raw_line in proc.stdout:
        line = raw_line.rstrip("\n")
        if "member=Notify" in line and "interface=org.freedesktop.Notifications" in line:
            if collecting and current:
                entry = _parse_notify_block(current)
                if entry is not None:
                    _append_entry(entry)
            current = [line]
            collecting = True
            continue

        if not collecting:
            continue

        if line.startswith("method call "):
            entry = _parse_notify_block(current)
            if entry is not None:
                _append_entry(entry)
            current = []
            collecting = False
            continue

        current.append(line)
        if EXPIRES_RE.match(line):
            entry = _parse_notify_block(current)
            if entry is not None:
                _append_entry(entry)
            current = []
            collecting = False

    if collecting and current:
        entry = _parse_notify_block(current)
        if entry is not None:
            _append_entry(entry)

    return proc.wait()


if __name__ == "__main__":
    raise SystemExit(main())
