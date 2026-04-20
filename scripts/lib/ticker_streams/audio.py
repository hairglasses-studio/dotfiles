"""audio — default sink / volume / mute via wpctl."""
from __future__ import annotations

import re
import subprocess
from html import escape

import ticker_render as tr

META = {"name": "audio", "preset": "minimal", "refresh": 10}

_LABEL = " AUDIO"


def build():
    parts = [tr.badge(_LABEL, "#fb7185")]
    segments = []
    nick = None
    try:
        inspect = subprocess.run(
            ["wpctl", "inspect", "@DEFAULT_AUDIO_SINK@"],
            capture_output=True, text=True, timeout=3,
        ).stdout
        m = re.search(r'node\.nick\s*=\s*"([^"]+)"', inspect)
        if m:
            nick = m.group(1)
        if not nick:
            m = re.search(r'node\.description\s*=\s*"([^"]+)"', inspect)
            if m:
                nick = m.group(1)
    except Exception:
        pass
    volume_pct = None
    muted = False
    try:
        vol_out = subprocess.run(
            ["wpctl", "get-volume", "@DEFAULT_AUDIO_SINK@"],
            capture_output=True, text=True, timeout=3,
        ).stdout.strip()
        m = re.search(r"([0-9]+\.[0-9]+)", vol_out)
        if m:
            volume_pct = int(round(float(m.group(1)) * 100))
        if "MUTED" in vol_out.upper():
            muted = True
    except Exception:
        pass
    if nick is None and volume_pct is None:
        return tr.empty(_LABEL, "#fb7185", "no default sink")
    icon = "\U000f0581" if muted else "\U000f057e"
    vol_color = "#ff5c8a" if muted else "#f7fbff"
    nick_s = escape(nick or "default sink")
    vol_s = f"{volume_pct}%" if volume_pct is not None else "?"
    mute_s = "  [MUTED]" if muted else ""
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 11" foreground="{vol_color}">'
        f'  {icon} {nick_s} \u00b7 {vol_s}{escape(mute_s)}  \u00b7</span>'
    )
    if nick:
        segments.append(nick)
    return tr.dup("".join(parts)), segments
