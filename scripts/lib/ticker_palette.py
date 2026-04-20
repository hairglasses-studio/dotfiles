"""ticker_palette — single source of truth for the Hairglasses Neon palette.

Six surface scripts were independently copy-pasting `#29f0ff`, `#ff47d1`,
`#3dffb5`, etc.; a palette change meant editing six files. This module
centralises the hex values so `rsvp-ticker.py`, `lyrics-ticker.py`,
`toast-ticker.py`, `subtitle-ticker.py`, `window-label.py`, and
`fleet-sparkline.py` can all import from one place.

Import from a sibling script (since the scripts dir is not a package):

    import sys, os
    sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))
    from ticker_palette import NEON

The palette mirrors `.claude/rules/snazzy-palette.md`. ``ticker_render.py``
already has a ``HAIRGLASSES_NEON`` dict — this module re-exports it so
there's a stable `ticker_palette.NEON` name without ripping out the
existing ticker_render constant (keybind-ticker.py uses it by that name).
"""
from __future__ import annotations

NEON: dict[str, str] = {
    "primary":        "#29f0ff",  # cyan    — color6
    "secondary":      "#ff47d1",  # magenta — color5
    "tertiary":       "#3dffb5",  # green   — color2
    "warning":        "#ffe45e",  # yellow  — color3
    "danger":         "#ff5c8a",  # red     — color1
    "blue":           "#4aa8ff",  # blue    — color4
    "muted":          "#66708f",  # — color8
    "fg":             "#f7fbff",  # — color7
    "bg":             "#05070d",  # — color0
    "surface":        "#0f1219",
    "surface_alt":    "#161c2b",
    "panel":          "#0d1018",
    "border":         "#2a3246",
    "border_strong":  "#46506d",
}

# Stream-oriented semantic aliases (common ticker use cases)
STREAM_COLORS: dict[str, str] = {
    "ok":        NEON["tertiary"],   # green   — stream healthy / no issues
    "warn":      NEON["warning"],    # yellow  — degraded
    "error":     NEON["danger"],     # red     — stream failing
    "stale":     "#f59e0b",          # amber   — cache beyond refresh window
    "info":      NEON["blue"],       # blue    — neutral info
    "emphasis":  NEON["primary"],    # cyan    — highlight
    "accent":    NEON["secondary"],  # magenta — secondary accent
    "idle":      NEON["muted"],      # grey    — idle state
}


def rgb(name: str) -> tuple[float, float, float]:
    """Return (r, g, b) floats in [0, 1] for the named palette entry.

    Falls back to ``NEON["tertiary"]`` (green) if the key is unknown so
    callers never crash on a typo.
    """
    h = NEON.get(name, NEON["tertiary"]).lstrip("#")
    return (
        int(h[0:2], 16) / 255.0,
        int(h[2:4], 16) / 255.0,
        int(h[4:6], 16) / 255.0,
    )
