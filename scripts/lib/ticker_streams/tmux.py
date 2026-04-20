"""tmux — list-sessions with attached markers and age."""
from __future__ import annotations

import subprocess
import time
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "tmux", "preset": None, "refresh": 60}

_LABEL = "\U000f06a0 TMUX"


def build():
    parts = [tr.badge(_LABEL, "#a8ff60")]
    try:
        raw = subprocess.run(
            ["tmux", "list-sessions", "-F",
             "#{session_name}|#{session_windows}|#{session_attached}|#{session_created}"],
            capture_output=True, text=True, timeout=3,
        )
    except FileNotFoundError:
        return tr.empty(_LABEL, "#a8ff60", "tmux not installed")
    except Exception:
        return tr.empty(_LABEL, "#a8ff60", "tmux unavailable")
    if raw.returncode != 0 or not raw.stdout.strip():
        return tr.empty(_LABEL, "#a8ff60", "no tmux sessions")
    now = time.time()
    fc = len(FONTS)
    for i, line in enumerate(raw.stdout.strip().splitlines()):
        fields = line.split("|")
        if len(fields) < 4:
            continue
        name, windows, attached, created = fields
        try:
            age = int(now - int(created))
            hours = age // 3600
            age_s = f"{hours}h" if hours else f"{age // 60}m"
        except ValueError:
            age_s = "?"
        attached_i = "\u25cf" if attached == "1" else "\u25cb"
        color = "#a8ff60" if attached == "1" else "#66708f"
        font = FONTS[i % fc]
        parts.append(
            f'<span font_desc="{font}" foreground="{color}">'
            f'  {attached_i} {escape(name)} ({escape(windows)}w, {age_s})  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
