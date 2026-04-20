"""notifications — recent desktop notification history (mako/swaync log).

Returns a `__URGENT__` sentinel in segments if any recent entry is
critical-urgency. The main ticker loop watches for that sentinel to
enter urgent-mode (escalated glitch + CA for ~10 s).
"""
from __future__ import annotations

import json
import os
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "notifications", "preset": None, "refresh": 60}

_LABEL = "\U000f009a NOTIFICATIONS"
_HISTORY = os.path.expanduser(
    "~/.local/state/dotfiles/desktop-control/notifications/history.jsonl"
)
_URGENCY_ICONS = {
    "critical": "\U000f0026",
    "normal":   "\U000f009a",
    "low":      "\U000f009e",
}


def build():
    parts = [tr.badge(_LABEL, "#ff5c8a")]
    has_critical = False
    try:
        with open(_HISTORY) as f:
            lines = f.readlines()
    except FileNotFoundError:
        return tr.empty(_LABEL, "#ff5c8a", "no notification history")
    except Exception:
        return tr.empty(_LABEL, "#ff5c8a", "notifications unavailable")

    recent = lines[-30:] if len(lines) > 30 else lines
    recent.reverse()
    if not recent:
        return tr.empty(_LABEL, "#ff5c8a", "no notification history")

    fc = len(FONTS)
    for i, line in enumerate(recent):
        try:
            n = json.loads(line)
        except Exception:
            continue
        urgency = str(n.get("urgency", ""))
        if urgency == "critical":
            has_critical = True
        icon = _URGENCY_ICONS.get(urgency, "\U000f009a")
        app = escape(str(n.get("app", ""))[:20])
        summary = escape(str(n.get("summary", ""))[:40])
        body = escape(str(n.get("body", ""))[:40])
        font = FONTS[i % fc]
        text = f"{summary}: {body}" if body and body != summary else summary
        parts.append(
            f'<span font_desc="{font}">  {icon} {app} {text}  \u00b7</span>'
        )
    markup = tr.dup("".join(parts))
    segments = ["__URGENT__"] if has_critical else []
    return markup, segments
