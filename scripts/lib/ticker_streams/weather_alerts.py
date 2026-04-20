"""weather-alerts — NWS severity-graded alerts from bar-weather-alerts cache."""
from __future__ import annotations

from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "weather-alerts", "preset": "cyberpunk", "refresh": 900}

_LABEL = "\U000f0e6e ALERT"

_SEVERITY_COLORS = {
    "extreme":  "#dc2626",
    "severe":   "#f97316",
    "moderate": "#f59e0b",
    "minor":    "#3b82f6",
}


def _read_lines():
    try:
        with open("/tmp/bar-weather-alerts.txt") as f:
            return [l.rstrip("\n") for l in f if l.strip()]
    except FileNotFoundError:
        return None
    except OSError:
        return []


def build():
    lines = _read_lines()
    if lines is None:
        return tr.empty(_LABEL, "#94a3b8", "alerts cache missing")
    if not lines:
        return tr.empty(_LABEL, "#34d399", "no active alerts")
    summary = lines[0]
    severity = summary.split()[0].lower() if summary else ""
    color = _SEVERITY_COLORS.get(severity, "#f97316")
    parts = [tr.badge(_LABEL, color)]
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 15" foreground="#fca5a5">'
        f'  {escape(summary)}  \u00b7</span>'
    )
    fc = len(FONTS)
    for i, line in enumerate(lines[1:6]):
        font = FONTS[i % fc]
        parts.append(
            f'<span font_desc="{font}">  {escape(line)}  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
