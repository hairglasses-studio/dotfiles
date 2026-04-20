"""pomodoro — reads ~/.local/state/keybind-ticker/pomodoro.json (written by `hg pomo`)."""
from __future__ import annotations

import json
import os
import time
from html import escape

import ticker_render as tr

META = {"name": "pomodoro", "preset": "cyberpunk", "refresh": 1}

_LABEL = "\U000f050d POMODORO"
_STATE_PATH = os.path.expanduser("~/.local/state/keybind-ticker/pomodoro.json")


def _state():
    try:
        with open(_STATE_PATH) as f:
            return json.load(f)
    except (OSError, ValueError):
        return None


def build():
    st = _state()
    if not st:
        return tr.empty(_LABEL, "#a3e635", "idle \u2014 hg pomo start")
    state = st.get("state", "idle")
    started = float(st.get("started_at", 0))
    duration = int(st.get("duration_s", 0))
    kind = st.get("session_kind", state).upper()
    now = time.time()
    if state == "running":
        remaining = max(int(started + duration - now), 0)
        mm, ss = divmod(remaining, 60)
        color = "#a3e635" if kind == "WORK" else "#60a5fa"
        status = f"{kind} {mm:02d}:{ss:02d}"
        if remaining == 0:
            color = "#ff5c8a"
            status = f"{kind} done"
    elif state == "paused":
        remaining = int(st.get("remaining_s", duration))
        mm, ss = divmod(remaining, 60)
        color = "#fbbf24"
        status = f"{kind} \u23f8 {mm:02d}:{ss:02d}"
    else:
        color = "#66708f"
        status = "idle"
    parts = [tr.badge(_LABEL, color)]
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 11" foreground="{color}">'
        f'  {escape(status)}  \u00b7</span>'
    )
    return tr.dup("".join(parts)), []
