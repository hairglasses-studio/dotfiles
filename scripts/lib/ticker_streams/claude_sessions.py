"""claude-sessions — recent Claude Code session counts per project (last 24h).

Clicking a segment opens the project path in the file manager (handled
by the generic segment click-handler). Slow because scanning many
JSONL files touches a lot of inodes on a busy workstation.
"""
from __future__ import annotations

import os
import time
from html import escape

import ticker_render as tr
from ticker_streams import FONTS

META = {"name": "claude-sessions", "preset": None, "refresh": 120, "slow": True}

_LABEL = " CLAUDE"
_STUDIO_PREFIX = "-home-hg-hairglasses-studio-"
_HOME_PREFIX = "-home-hg-"


def build():
    parts = [tr.badge(_LABEL, "#c084fc")]
    segments: list[str] = []
    try:
        projects_dir = os.path.expanduser("~/.claude/projects")
        if not os.path.isdir(projects_dir):
            return tr.empty(_LABEL, "#c084fc", "no ~/.claude/projects")
        now = time.time()
        cutoff = now - 86400  # last 24h
        project_stats = []
        for entry in os.scandir(projects_dir):
            if not entry.is_dir():
                continue
            session_mtimes = []
            try:
                for f in os.scandir(entry.path):
                    if f.name.endswith(".jsonl") and f.is_file():
                        st = f.stat()
                        if st.st_mtime >= cutoff:
                            session_mtimes.append(st.st_mtime)
            except OSError:
                continue
            if session_mtimes:
                project_stats.append(
                    (entry.name, len(session_mtimes), max(session_mtimes))
                )
        if not project_stats:
            return tr.empty(_LABEL, "#c084fc", "no recent sessions")
        project_stats.sort(key=lambda t: t[2], reverse=True)
        total_projects = len(project_stats)
        total_sessions = sum(s for _, s, _ in project_stats)
        parts.append(
            f'<span font_desc="Maple Mono NF CN Bold 15">'
            f'  {total_projects} projects \u00b7 {total_sessions} sessions  \u00b7</span>'
        )
        fc = len(FONTS)
        for i, (encoded, count, mtime) in enumerate(project_stats[:8]):
            if encoded.startswith(_STUDIO_PREFIX):
                short = encoded[len(_STUDIO_PREFIX):]
                display_path = f"~/hairglasses-studio/{short}"
            elif encoded.startswith(_HOME_PREFIX):
                short = encoded[len(_HOME_PREFIX):]
                display_path = f"~/{short}"
            else:
                short = encoded.lstrip("-")
                display_path = encoded
            age_s = now - mtime
            if age_s < 60:
                age = f"{int(age_s)}s"
            elif age_s < 3600:
                age = f"{int(age_s / 60)}m"
            else:
                age = f"{int(age_s / 3600)}h"
            font = FONTS[i % fc]
            parts.append(
                f'<span font_desc="{font}">'
                f'  {escape(short)} ({count}\u00d7, {age} ago)  \u00b7</span>'
            )
            segments.append(display_path)
    except Exception:
        return tr.empty(_LABEL, "#c084fc", "session scan failed")
    return tr.dup("".join(parts)), segments
