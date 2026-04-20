"""ci — GitHub workflow run summary from bar-ci cache.

Cache format (written by bar-ci-cache.sh):
  line 1: `PASS=N FAIL=N RUN=N` token summary
  lines 2+: `\t`-separated `repo\toutcome[\trun_id]` per row
"""
from __future__ import annotations

from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "ci", "preset": "cyberpunk", "refresh": 300}

_LABEL = "\U000f03a0 CI"


def build():
    try:
        with open("/tmp/bar-ci.txt") as f:
            lines = [l.rstrip("\n") for l in f if l.strip()]
    except FileNotFoundError:
        return tr.empty(_LABEL, "#a3e635", "ci cache missing")
    except Exception:
        return tr.empty(_LABEL, "#a3e635", "ci unavailable")

    if not lines:
        return tr.empty(_LABEL, "#a3e635", "no ci data")
    if lines[0] == "CI=unavailable":
        return tr.empty(_LABEL, "#a3e635", "gh or jq missing")

    counts = {"PASS": 0, "FAIL": 0, "RUN": 0}
    for tok in lines[0].split():
        k, _, v = tok.partition("=")
        if k in counts and v.isdigit():
            counts[k] = int(v)

    if counts["FAIL"] > 0:
        badge_color = "#ff5c8a"
    elif counts["RUN"] > 0:
        badge_color = "#ffe45e"
    elif counts["PASS"] == 0:
        return tr.empty(_LABEL, "#a3e635", "no ci runs")
    else:
        badge_color = "#a3e635"

    parts = [tr.badge(_LABEL, badge_color)]
    summary = f"\u2713{counts['PASS']}  \u2717{counts['FAIL']}  \u23f3{counts['RUN']}"
    parts.append(
        f'<span font_desc="Maple Mono NF CN Bold 15">  {escape(summary)}  \u00b7</span>'
    )

    fc = len(FONTS)
    for i, line in enumerate(lines[1:15]):
        fields = line.split("\t")
        if len(fields) < 2:
            continue
        repo, outcome = fields[0], fields[1]
        mark = "\u23f3" if outcome.startswith("running") else "\u2717"
        name_color = "#fbbf24" if outcome.startswith("running") else "#fb7185"
        font = FONTS[i % fc]
        parts.append(
            f'<span font_desc="{font}" foreground="{name_color}">'
            f'  {mark} {escape(repo)}  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
