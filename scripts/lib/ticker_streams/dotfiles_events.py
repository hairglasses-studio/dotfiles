"""dotfiles_events — pending events from the dotfiles-event-bus daemon.

Reads ~/.local/state/dotfiles/events.jsonl and surfaces high-severity
entries from the last 24h as a persistent glanceable badge. swaync
already pops transient notifications on emit; this stream gives the
user a steady ambient count so an unresolved condition doesn't drop
off the radar after the pop-up fades.

Emits `__URGENT__` in segments when any high-severity event is unresolved
so the ticker main loop enters its escalated visual state.
"""
from __future__ import annotations

import json
import os
import time
from collections import Counter
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "dotfiles_events", "preset": None, "refresh": 60}

_LABEL = "\U000f0026 EVENTS"
_COLOR = "#ff5c8a"  # THEME_DANGER shade
_LOG = os.path.expanduser("~/.local/state/dotfiles/events.jsonl")
_WINDOW_S = 24 * 3600  # last 24h
_SCAN_LIMIT = 500      # read at most this many recent lines


def _recent_high(log_path: str) -> list[dict]:
    """Return the high-severity events emitted within the last WINDOW.

    Reads at most the last SCAN_LIMIT lines so the stream stays cheap
    even when the log grows. Silently skips corrupt or non-JSON lines.
    """
    try:
        with open(log_path) as f:
            lines = f.readlines()
    except (FileNotFoundError, OSError):
        return []
    if len(lines) > _SCAN_LIMIT:
        lines = lines[-_SCAN_LIMIT:]
    cutoff = time.time() - _WINDOW_S
    out: list[dict] = []
    for line in lines:
        try:
            rec = json.loads(line)
        except Exception:
            continue
        if rec.get("severity") != "high":
            continue
        at = rec.get("at", "")
        # Parse RFC3339 UTC — "YYYY-MM-DDTHH:MM:SSZ"
        try:
            ts = time.mktime(time.strptime(at, "%Y-%m-%dT%H:%M:%SZ")) - time.timezone
        except Exception:
            continue
        if ts < cutoff:
            continue
        out.append(rec)
    return out


def build():
    events = _recent_high(_LOG)
    if not events:
        return tr.empty(_LABEL, _COLOR, "no pending events")

    # Group by error_code so recurring conditions render as one line with a
    # count rather than N scrolling entries.
    codes = Counter(e.get("error_code", "unknown") for e in events)
    total = sum(codes.values())

    parts = [tr.badge(f"{_LABEL} ×{total}", _COLOR)]
    fc = len(FONTS)
    for i, (code, count) in enumerate(codes.most_common(6)):
        font = FONTS[i % fc]
        text = f"{escape(code)} ×{count}" if count > 1 else escape(code)
        parts.append(
            f'<span font_desc="{font}">  {text}  ·</span>'
        )
    markup = tr.dup("".join(parts))
    return markup, ["__URGENT__"]
