"""failed-units — summary of failed user + system systemd units.

Red badge when any are failed, green sentinel when all clean. Scope is
prefixed on each listed unit so `[user]` and `[sys]` are visually
distinct.
"""
from __future__ import annotations

import subprocess
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "failed-units", "preset": None, "refresh": 60}


def _scope_list(args):
    try:
        return subprocess.run(args, capture_output=True, text=True,
                              timeout=3).stdout
    except Exception:
        return None


def build():
    user_out = _scope_list(["systemctl", "--user", "--failed",
                            "--no-legend", "--plain"])
    sys_out = _scope_list(["systemctl", "--failed", "--no-legend", "--plain"])
    if user_out is None or sys_out is None:
        return tr.empty("\U000f0028 FAILED", "#ef4444", "systemctl unavailable")
    entries = []
    for scope, out in (("user", user_out), ("sys", sys_out)):
        for line in out.splitlines():
            tok = line.split()
            if tok:
                entries.append((scope, tok[0]))
    if not entries:
        return tr.empty("\U000f0028 FAILED", "#34d399", "no failed units")
    parts = [tr.badge("\U000f0028 FAILED", "#ef4444")]
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 11" foreground="#ef4444">'
        f'  {len(entries)} failed  \u00b7</span>'
    )
    fc = len(FONTS)
    for i, (scope, name) in enumerate(entries[:10]):
        font = FONTS[i % fc]
        parts.append(
            f'<span font_desc="{font}" foreground="#fbbf24">'
            f'  [{scope}] {escape(name)}  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
