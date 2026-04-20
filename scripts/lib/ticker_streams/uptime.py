"""uptime — /proc/uptime formatted human + absolute boot timestamp."""
from __future__ import annotations

import time
from html import escape

import ticker_render as tr

META = {"name": "uptime", "preset": "ambient", "refresh": 300}

_LABEL = "\U000f0907 UPTIME"


def build():
    parts = [tr.badge(_LABEL, "#c6a0ff")]
    try:
        with open("/proc/uptime") as f:
            secs = float(f.read().split()[0])
    except Exception:
        return tr.empty(_LABEL, "#c6a0ff", "uptime unavailable")
    days, rem = divmod(int(secs), 86400)
    hours, rem = divmod(rem, 3600)
    mins, _ = divmod(rem, 60)
    if days:
        human = f"{days}d {hours}h {mins}m"
    elif hours:
        human = f"{hours}h {mins}m"
    else:
        human = f"{mins}m"
    boot_ts = time.time() - secs
    boot = time.strftime("%Y-%m-%d %H:%M", time.localtime(boot_ts))
    color = "#ff5c8a" if days >= 30 else ("#ffe45e" if days >= 7 else "#c6a0ff")
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 15" foreground="{color}">'
        f'  {escape(human)}  \u00b7</span>'
        f'<span font_desc="Maple Mono NF CN 15" foreground="#9fb2ff">'
        f'  booted {escape(boot)}  \u00b7</span>'
    )
    return tr.dup("".join(parts)), []
