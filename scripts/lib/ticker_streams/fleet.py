"""fleet — ralphglasses /tmp/rg-status.json summary (running / loops / cost / models)."""
from __future__ import annotations

import json
from html import escape

import ticker_render as tr

META = {"name": "fleet", "preset": "cyberpunk", "refresh": 30}

_LABEL = "\U000f0168 FLEET"


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
