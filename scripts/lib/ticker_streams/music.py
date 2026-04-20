"""music — playerctl status / metadata / position.

"No media playing" is a normal idle state, not an error — return a real
markup with an "idle" span so the ticker error-backoff path doesn't
short-circuit the 10s refresh into 30s during quiet periods, and
`hg ticker health` doesn't mark the stream as faulting.
"""
from __future__ import annotations

import subprocess
from html import escape

import ticker_render as tr

META = {"name": "music", "preset": "minimal", "refresh": 10, "slow": True}

_LABEL = " MUSIC"


def build():
    parts = [tr.badge(_LABEL, "#ff47d1")]
    try:
        status = subprocess.run(
            ["playerctl", "status"], capture_output=True, text=True, timeout=3,
        ).stdout.strip()
    except Exception:
        status = ""
    if status not in ("Playing", "Paused"):
        parts.append(
            '<span font_desc="Maple Mono NF CN 15" foreground="#66708f">'
            '   idle  \u00b7</span>'
        )
        return tr.dup("".join(parts)), []
    try:
        icon = "" if status == "Playing" else ""
        meta = subprocess.run(
            ["playerctl", "metadata", "--format",
             "{{artist}} \u2014 {{title}} [{{album}}]"],
            capture_output=True, text=True, timeout=3,
        ).stdout.strip()
        pos = subprocess.run(
            ["playerctl", "position", "--format", "{{duration(position)}}"],
            capture_output=True, text=True, timeout=3,
        ).stdout.strip()
        dur = subprocess.run(
            ["playerctl", "metadata", "--format",
             "{{duration(mpris:length)}}"],
            capture_output=True, text=True, timeout=3,
        ).stdout.strip()
        parts.append(
            f'<span font_desc="Maple Mono NF CN Bold 15">'
            f'  {icon} {escape(meta)}  {escape(pos)}/{escape(dur)}  \u00b7</span>'
        )
    except Exception:
        return tr.empty(_LABEL, "#ff47d1", "playerctl unavailable")
    return tr.dup("".join(parts)), []
