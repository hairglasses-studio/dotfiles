"""pomodoro — reads ~/.local/state/keybind-ticker/pomodoro.json (written by `hg pomo`).

Phase 4 of the marathon adds a completion hook: on the first `remaining
== 0` zero-crossing of a `running` session, the plugin fires a
the ticker-control wrapper and sets urgent-mode so the scroll flashes. The hook
guards against re-firing every frame by tracking the last-seen
(started_at, session_kind) tuple in module scope; the same session
only triggers once.
"""
from __future__ import annotations

import json
import os
import subprocess
import time
from html import escape

import ticker_render as tr

META = {"name": "pomodoro", "preset": "cyberpunk", "refresh": 1}

_LABEL = "\U000f050d POMODORO"
_STATE_PATH = os.path.expanduser("~/.local/state/keybind-ticker/pomodoro.json")
_TICKER_CONTROL = os.path.expanduser("~/hairglasses-studio/dotfiles/scripts/ticker-control.sh")

# Track the last session that already fired its completion hook so we
# don't re-trigger ShowBanner every frame while the state file sits on
# `remaining == 0` until the user ack's it.
_last_fired: tuple | None = None


def _state():
    try:
        with open(_STATE_PATH) as f:
            return json.load(f)
    except (OSError, ValueError):
        return None


def _fire_completion_hook(kind: str, started: float):
    """Notify ticker-control banner + set urgent mode.

    Uses subprocess calls so the plugin stays dependency-free. The banner text
    and urgent signal are independent calls; both are best-effort.
    """
    label = "BREAK" if kind.upper() in ("BREAK", "REST", "PAUSE") else "WORK"
    color = "#3dffb5" if label == "WORK" else "#4aa8ff"
    text = f"Pomodoro: {label} done"
    for args in (
        [_TICKER_CONTROL, "banner", text, color],
        [_TICKER_CONTROL, "urgent", "true"],
    ):
        try:
            subprocess.Popen(
                args, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
                start_new_session=True,
            )
        except Exception:
            pass


def build():
    global _last_fired
    st = _state()
    if not st:
        return tr.empty(_LABEL, "#a3e635", "idle \u2014 hg pomo start")
    state = st.get("state", "idle")
    started = float(st.get("started_at", 0))
    duration = int(st.get("duration_s", 0))
    kind = st.get("session_kind", state).upper()
    now = time.time()
    if state == "running":
        remaining = max(int(started + duration - now), 0)
        mm, ss = divmod(remaining, 60)
        color = "#a3e635" if kind == "WORK" else "#60a5fa"
        status = f"{kind} {mm:02d}:{ss:02d}"
        if remaining == 0:
            color = "#ff5c8a"
            status = f"{kind} done"
            # First-zero-crossing guard: fire exactly once per session
            # (keyed on the started_at timestamp so a new session resets
            # the fired-flag automatically).
            fire_key = (started, kind)
            if _last_fired != fire_key:
                _last_fired = fire_key
                _fire_completion_hook(kind, started)
    elif state == "paused":
        remaining = int(st.get("remaining_s", duration))
        mm, ss = divmod(remaining, 60)
        color = "#fbbf24"
        status = f"{kind} \u23f8 {mm:02d}:{ss:02d}"
    else:
        color = "#66708f"
        status = "idle"
    parts = [tr.badge(_LABEL, color)]
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 15" foreground="{color}">'
        f'  {escape(status)}  \u00b7</span>'
    )
    return tr.dup("".join(parts)), []
