"""github — /notifications feed via `gh api`, clickable segments open the item.

Slow: each refresh spawns `gh api /notifications --paginate` which can
block on network I/O for seconds. Runs on the shared background thread
via META["slow"] = True.
"""
from __future__ import annotations

import json
import subprocess
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "github", "preset": None, "refresh": 120, "slow": True}

_LABEL = " GITHUB"
_TYPE_ICONS = {
    "PullRequest": "",
    "Issue":       "",
    "Release":     "",
    "Discussion":  "\U000f0361",
    "CheckSuite":  "",
}


def build():
    parts = [tr.badge(_LABEL, "#3dffb5")]
    try:
        raw = subprocess.run(
            ["gh", "api", "/notifications", "--paginate", "--jq",
             '.[] | {type: .subject.type, title: .subject.title, '
             'repo: .repository.name, reason: .reason, url: .subject.url}'],
            capture_output=True, text=True, timeout=15,
        ).stdout.strip()
    except Exception:
        return tr.empty(_LABEL, "#3dffb5", "github unavailable")
    if not raw:
        return tr.empty(_LABEL, "#3dffb5", "no notifications")

    seen = 0
    fc = len(FONTS)
    segments: list[str] = []
    for line in raw.splitlines():
        if seen >= 20:
            break
        try:
            n = json.loads(line)
        except Exception:
            continue
        icon = _TYPE_ICONS.get(n.get("type", ""), "")
        title = escape(str(n.get("title", ""))[:60])
        repo = escape(str(n.get("repo", "")))
        font = FONTS[seen % fc]
        parts.append(
            f'<span font_desc="{font}">  {icon} {repo}: {title}  \u00b7</span>'
        )
        # Convert API URL to HTML URL for xdg-open click handler
        api_url = str(n.get("url", ""))
        html_url = api_url.replace("api.github.com/repos/", "github.com/")
        segments.append(html_url or "")
        seen += 1
    if seen == 0:
        return tr.empty(_LABEL, "#3dffb5", "no notifications")
    return tr.dup("".join(parts)), segments
