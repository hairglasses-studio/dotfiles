"""shader — hyprshade current + hyprsunset temperature."""
from __future__ import annotations

import subprocess
from html import escape

import ticker_render as tr

META = {"name": "shader", "preset": "cyberpunk", "refresh": 60}

_LABEL = " SHADER"


def build():
    parts = [tr.badge(_LABEL, "#f97316")]
    segments = []
    shader_name = None
    try:
        out = subprocess.run(
            ["hyprshade", "current"],
            capture_output=True, text=True, timeout=3,
        )
        if out.returncode == 0:
            shader_name = out.stdout.strip() or None
    except Exception:
        pass
    sunset_state = None
    try:
        out = subprocess.run(
            ["hyprctl", "hyprsunset", "temperature"],
            capture_output=True, text=True, timeout=2,
        )
        if out.returncode == 0:
            val = out.stdout.strip()
            if val and val != "0":
                sunset_state = f"{val}K"
    except Exception:
        pass
    if not shader_name and not sunset_state:
        return tr.empty(_LABEL, "#f97316", "no shader active")
    if shader_name:
        parts.append(
            f'<span font_desc="Maple Mono NF CN Bold 11" foreground="#f97316">'
            f'  {escape(shader_name)}  \u00b7</span>'
        )
        segments.append(shader_name)
    else:
        parts.append(
            '<span font_desc="Maple Mono NF CN Bold 11" foreground="#888888">'
            '  no shader  \u00b7</span>'
        )
    if sunset_state:
        parts.append(
            f'<span font_desc="Maple Mono NF CN Italic 11" foreground="#ffe45e">'
            f'  hyprsunset: {escape(sunset_state)}  \u00b7</span>'
        )
        segments.append(sunset_state)
    return tr.dup("".join(parts)), segments
