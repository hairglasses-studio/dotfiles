from __future__ import annotations

from pathlib import Path
from subprocess import run
from typing import Any

from kitty.boss import Boss
from kitty.window import Window

SCRIPT = Path(__file__).resolve().parents[2] / "scripts" / "kitty-shader-playlist.sh"


def _theme_for_window(window_id: int) -> str | None:
    result = run(
        ["bash", str(SCRIPT), "theme-for-window", str(window_id)],
        capture_output=True,
        check=False,
        text=True,
    )
    if result.returncode != 0:
        return None

    parts = result.stdout.strip().split("\t")
    if len(parts) < 2:
        return None
    return parts[1]


def on_resize(boss: Boss, window: Window, data: dict[str, Any]) -> None:
    old_geometry = data.get("old_geometry")
    if old_geometry is None:
        return
    if getattr(old_geometry, "xnum", 0) != 0 or getattr(old_geometry, "ynum", 0) != 0:
        return

    theme_conf = _theme_for_window(window.id)
    if not theme_conf:
        return

    try:
        boss.call_remote_control(window, ("set-colors", f"--match=id:{window.id}", theme_conf))
    except Exception:
        return
