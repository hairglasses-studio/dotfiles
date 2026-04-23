"""fleet — ralphglasses /tmp/rg-status.json summary (running / loops / cost / models).

Also fires a ticker banner on cost-threshold crossings so the
fleet cost growing past $10 / $50 / $100 / $500 / $1000 surfaces as
a banner even when the fleet stream isn't currently on screen. Uses
the ticker-control wrapper so Quickshell cutover and legacy DBus both work.
"""
from __future__ import annotations

import json
import os
import subprocess
from html import escape

import ticker_render as tr

META = {"name": "fleet", "preset": "cyberpunk", "refresh": 30}

_LABEL = "\U000f0168 FLEET"
_TICKER_CONTROL = os.path.expanduser("~/hairglasses-studio/dotfiles/scripts/ticker-control.sh")

# Cost milestones — banners fire when total_spend_usd crosses any of
# these for the first time since the plugin started. The last-seen
# value prevents re-firing on the same crossing across rebuilds.
_COST_MILESTONES = (10.0, 50.0, 100.0, 500.0, 1000.0)
_last_cost: float = 0.0


def _maybe_fire_cost_banner(cost_now: float):
    global _last_cost
    try:
        prev = _last_cost
        _last_cost = cost_now
        for threshold in _COST_MILESTONES:
            if prev < threshold <= cost_now:
                color = "#ff5c8a" if threshold >= 100 else "#ffe45e"
                msg = f"Fleet cost past ${threshold:.0f}  (now ${cost_now:.2f})"
                subprocess.Popen(
                    [_TICKER_CONTROL, "banner", msg, color],
                    stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
                    start_new_session=True,
                )
                break  # one banner per build() call
    except Exception:
        pass


def build():
    parts = [tr.badge(_LABEL, "#ff47d1")]
    try:
        with open("/tmp/rg-status.json") as f:
            data = json.load(f)
    except Exception:
        return tr.empty(_LABEL, "#ff47d1", "no fleet data")
    fl = data.get("fleet", {})
    cost = data.get("cost", {})
    loops = data.get("loops", {})
    try:
        _maybe_fire_cost_banner(float(cost.get("total_spend_usd", 0) or 0))
    except (TypeError, ValueError):
        pass
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 15">'
        f'  {int(fl.get("running", 0))} running  '
        f'{int(fl.get("completed", 0))} done  '
        f'{int(fl.get("failed", 0))} failed  '
        f'{int(fl.get("pending", 0))} pending  \u00b7</span>'
    )
    parts.append(
        f'<span font_desc="Maple Mono NF CN SemiBold 15">'
        f'  {int(loops.get("total_runs", 0))} loops  '
        f'${escape(str(cost.get("total_spend_usd", 0)))}  \u00b7</span>'
    )
    for m in data.get("models", [])[:3]:
        model_name = escape(str(m.get("model", "")))
        count = int(m.get("count", 0))
        parts.append(
            f'<span font_desc="Maple Mono NF CN Italic 15">'
            f'  {model_name} \u00d7{count}  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
