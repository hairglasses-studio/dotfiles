"""hn-top — Hacker News front-page items with clickable segments.

Cache format (written by bar-hn-cache.sh): one item per line,
tab-separated as `item_id\tscore\ttitle`. The segment list returned
here includes the `item_id` prefix so the click handler registered in
keybind-ticker.py (_click_hn_top) can parse it and open news.ycombinator.com.
"""
from __future__ import annotations

from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "hn-top", "preset": "ambient", "refresh": 600}

_LABEL = "\U000f00cd HN"
_ORANGE = "#ff6600"


def _read_lines():
    try:
        with open("/tmp/bar-hn.txt") as f:
            return [l.rstrip("\n") for l in f if l.strip()]
    except FileNotFoundError:
        return None
    except OSError:
        return []


def build():
    lines = _read_lines()
    if lines is None:
        return tr.empty(_LABEL, _ORANGE, "hn cache missing")
    if not lines:
        return tr.empty(_LABEL, _ORANGE, "no hn items")
    parts = [tr.badge(_LABEL, _ORANGE)]
    segments: list[str] = []
    fc = len(FONTS)
    for i, line in enumerate(lines[:8]):
        cells = line.split("\t", 2)
        if len(cells) != 3:
            continue
        item_id, score, title = cells
        font = FONTS[i % fc]
        parts.append(
            f'<span font_desc="Maple Mono NF CN Bold 11" foreground="#fbbf24">'
            f'  {escape(score)}pts  </span>'
            f'<span font_desc="{font}">{escape(title)}  \u00b7</span>'
        )
        segments.append(f"{item_id} {title}")
    return tr.dup("".join(parts)), segments
