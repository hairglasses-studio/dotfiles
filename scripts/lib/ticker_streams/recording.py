"""recording — live screen-recording indicator.

Cache written by ticker-recordwatch.sh (single tab-separated line:
`tool\tpid\tstarted_epoch`). If missing or empty the stream shows
'not recording' so it stays legal on the recording.txt playlist.
"""
from __future__ import annotations

import os
import subprocess
import time
from html import escape

import ticker_render as tr

META = {"name": "recording", "preset": "cyberpunk", "refresh": 1}

_LABEL = "\U000f04d0 REC"


def build():
    try:
        with open("/tmp/bar-recording.txt") as f:
            line = f.readline().strip()
    except OSError:
        return tr.empty(_LABEL, "#dc2626", "not recording")
    if not line:
        return tr.empty(_LABEL, "#dc2626", "not recording")
    fields = line.split("\t")
    tool = fields[0] if fields else "recorder"
    try:
        started = float(fields[2]) if len(fields) >= 3 else 0.0
    except ValueError:
        started = 0.0
    dur = max(0, int(time.time() - started)) if started else 0
    mm, ss = divmod(dur, 60)
    hh, mm = divmod(mm, 60)
    clock = f"{hh:02d}:{mm:02d}:{ss:02d}" if hh else f"{mm:02d}:{ss:02d}"

    try:
        free = subprocess.run(
            ["df", "-h", "--output=avail", os.path.expanduser("~")],
            capture_output=True, text=True, timeout=1,
        ).stdout.splitlines()
        free_str = free[1].strip() if len(free) >= 2 else "?"
    except Exception:
        free_str = "?"

    try:
        mic = subprocess.run(
            ["pactl", "get-source-mute", "@DEFAULT_SOURCE@"],
            capture_output=True, text=True, timeout=1,
        ).stdout.strip()
        mic_on = "off" not in mic.lower()
    except Exception:
        mic_on = False
    mic_icon = "\U000f036c" if mic_on else "\U000f036d"
    mic_color = "#34d399" if mic_on else "#fca5a5"

    parts = [tr.badge(_LABEL, "#dc2626")]
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 12" foreground="#f87171">'
        f'  {escape(clock)}  \u00b7</span>'
    )
    parts.append(
        f'<span font_desc="Maple Mono NF CN 11" foreground="#cbd5f5">'
        f'  {escape(tool)}  \u00b7</span>'
    )
    parts.append(
        f'<span font_desc="Maple Mono NF CN 11" foreground="{mic_color}">'
        f'  {mic_icon} mic  \u00b7</span>'
    )
    parts.append(
        f'<span font_desc="Maple Mono NF CN 11" foreground="#fbbf24">'
        f'  disk {escape(free_str)}  \u00b7</span>'
    )
    return tr.dup("".join(parts)), []
