"""keybinds — hyprctl binds -j rolled out with human-readable mod strings.

Left-click on a segment copies the `<mods><key>` combo to the clipboard
via the existing segment click-handler — the builder returns each combo
in the `segments` list in the same order as the visible spans.
"""
from __future__ import annotations

import json
import subprocess
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "keybinds", "preset": None, "refresh": 300}

_LABEL = " KEYBINDS"


def _fmt_mods(mask: int) -> str:
    out = ""
    if mask & 64: out += "Super+"
    if mask & 1:  out += "Shift+"
    if mask & 4:  out += "Ctrl+"
    if mask & 8:  out += "Alt+"
    return out


def build():
    try:
        raw = subprocess.run(
            ["hyprctl", "binds", "-j"],
            capture_output=True, text=True, timeout=5,
        ).stdout
        binds = json.loads(raw)
    except Exception:
        return tr.empty(_LABEL, "#29f0ff", "no keybinds loaded")

    parts = [tr.badge(_LABEL, "#29f0ff")]
    segments = []
    fc = len(FONTS)
    i = 0
    for b in binds:
        if b.get("has_description") and not b.get("submap") and not b.get("mouse"):
            mods = _fmt_mods(b["modmask"])
            desc = escape(b["description"])
            key_text = f"{mods}{b['key']}"
            key = escape(key_text)
            font = FONTS[i % fc]
            parts.append(f'<span font_desc="{font}">  {desc}  {key}  \u00b7</span>')
            segments.append(key_text)
            i += 1
    return tr.dup("".join(parts)), segments
