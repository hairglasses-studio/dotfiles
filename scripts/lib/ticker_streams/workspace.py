"""workspace — active workspace/window + per-monitor workspace summary via hyprctl."""
from __future__ import annotations

import json
import subprocess
from html import escape

import ticker_render as tr

META = {"name": "workspace", "preset": None, "refresh": 5}

_LABEL = "\U000f0708 WORKSPACE"


def _hyprctl_json(args, timeout=3):
    return json.loads(
        subprocess.run(
            ["hyprctl", *args, "-j"],
            capture_output=True, text=True, timeout=timeout,
        ).stdout
    )


def build():
    parts = [tr.badge(_LABEL, "#ff47d1")]
    try:
        aws = _hyprctl_json(["activeworkspace"])
        monitor = escape(str(aws.get("monitor", "")))
        name = escape(str(aws.get("name", "")))
        windows = int(aws.get("windows", 0))
        parts.append(
            f'<span font_desc="Maple Mono NF CN Bold 15">'
            f'  ws={name} on {monitor}  {windows} windows  \u00b7</span>'
        )

        aw = _hyprctl_json(["activewindow"])
        cls = escape(str(aw.get("class", ""))[:30])
        title = escape(str(aw.get("title", ""))[:50])
        if cls:
            parts.append(
                f'<span font_desc="Maple Mono NF CN Italic 15">'
                f'  {cls}: {title}  \u00b7</span>'
            )

        all_ws = _hyprctl_json(["workspaces"])
        total = len(all_ws)
        busy = sum(1 for w in all_ws if w.get("windows", 0) > 0)
        parts.append(
            f'<span font_desc="Maple Mono NF CN 15">'
            f'  {busy}/{total} workspaces active  \u00b7</span>'
        )
    except Exception:
        return tr.empty(_LABEL, "#ff47d1", "hyprctl unavailable")
    return tr.dup("".join(parts)), []
