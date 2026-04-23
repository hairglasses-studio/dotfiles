#!/usr/bin/env python3
"""Emit the current MPRIS now-playing or synced lyric line as JSON."""

from __future__ import annotations

import argparse
import bisect
import hashlib
import json
import os
import re
import subprocess
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Any

LRCLIB_URL = "https://lrclib.net/api/get"
LRC_RE = re.compile(r"\[(\d+):(\d+(?:\.\d+)?)\](.*)")


def playerctl_bin() -> str:
    return os.environ.get("PLAYERCTL_BIN", "playerctl")


def playerctl(fmt: str) -> str:
    try:
        result = subprocess.run(
            [playerctl_bin(), "metadata", "--format", fmt],
            capture_output=True,
            text=True,
            timeout=1,
            check=False,
        )
    except (OSError, subprocess.TimeoutExpired):
        return ""
    return result.stdout.strip()


def cache_dir() -> Path:
    return Path(os.environ.get("LYRICS_BRIDGE_CACHE_DIR", Path.home() / ".cache/lyrics-ticker"))


def track_key(title: str, artist: str, album: str) -> str:
    raw = f"{artist}|{album}|{title}".encode("utf-8", "replace")
    return hashlib.sha1(raw).hexdigest()


def parse_lrc(text: str) -> list[tuple[float, str]]:
    parsed: list[tuple[float, str]] = []
    for raw_line in text.splitlines():
        line = raw_line.strip()
        if not line:
            continue
        stamps: list[float] = []
        rest = line
        while True:
            match = LRC_RE.match(rest)
            if not match:
                break
            minutes, seconds, rest = match.group(1), match.group(2), match.group(3)
            try:
                stamps.append(int(minutes) * 60 + float(seconds))
            except ValueError:
                pass
        body = rest.strip()
        for stamp in stamps:
            parsed.append((stamp, body))
    parsed.sort(key=lambda item: item[0])
    return parsed


def read_cache(key: str) -> dict[str, Any] | None:
    try:
        with (cache_dir() / f"{key}.json").open() as handle:
            data = json.load(handle)
        return data if isinstance(data, dict) else None
    except (OSError, ValueError):
        return None


def write_cache(key: str, data: dict[str, Any]) -> None:
    try:
        cache_dir().mkdir(parents=True, exist_ok=True)
        path = cache_dir() / f"{key}.json"
        tmp = path.with_suffix(".tmp")
        with tmp.open("w") as handle:
            json.dump(data, handle, separators=(",", ":"))
        tmp.replace(path)
    except OSError:
        pass


def fetch_lrclib(title: str, artist: str, album: str, duration_s: int) -> dict[str, Any] | None:
    if os.environ.get("LYRICS_BRIDGE_NO_FETCH") == "1" or not title or not artist:
        return None
    params = {
        "track_name": title,
        "artist_name": artist,
        "album_name": album,
        "duration": str(max(0, int(duration_s))),
    }
    request = urllib.request.Request(
        f"{LRCLIB_URL}?{urllib.parse.urlencode(params)}",
        headers={"User-Agent": "lyrics-bridge/1 (+https://github.com/hairglasses-studio)"},
    )
    try:
        with urllib.request.urlopen(request, timeout=4) as response:
            if response.status != 200:
                return None
            data = json.loads(response.read().decode("utf-8", "replace"))
        return data if isinstance(data, dict) else None
    except Exception:
        return None


def active_line(lrc: list[tuple[float, str]], position_s: float, fallback: str) -> str:
    if not lrc:
        return fallback
    stamps = [stamp for stamp, _ in lrc]
    index = bisect.bisect_right(stamps, position_s) - 1
    if index < 0:
        index = 0
    return lrc[index][1] or fallback


def sample() -> dict[str, Any]:
    status = playerctl("{{status}}")
    if status != "Playing":
        return {
            "ok": True,
            "status": status or "Idle",
            "line": "IDLE",
            "color": "#66708f",
            "synced": False,
            "title": "",
            "artist": "",
        }

    title = playerctl("{{title}}")
    artist = playerctl("{{artist}}")
    album = playerctl("{{album}}")
    fallback = " -- ".join(part for part in (title, artist) if part) or "Playing"
    try:
        duration_s = int(playerctl("{{mpris:length}}") or "0") // 1_000_000
    except ValueError:
        duration_s = 0
    try:
        position_s = int(playerctl("{{position}}") or "0") / 1_000_000
    except ValueError:
        position_s = 0.0

    key = track_key(title, artist, album)
    data = read_cache(key)
    if data is None:
        data = fetch_lrclib(title, artist, album, duration_s) or {}
        if data:
            write_cache(key, data)

    synced_text = data.get("syncedLyrics") if isinstance(data, dict) else ""
    lrc = parse_lrc(synced_text or "")
    synced = bool(lrc)
    return {
        "ok": True,
        "status": "Playing",
        "line": active_line(lrc, position_s, fallback),
        "color": "#3dffb5" if synced else "#ff47d1",
        "synced": synced,
        "title": title,
        "artist": artist,
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--once", action="store_true", help="emit one sample and exit")
    parser.parse_args()
    print(json.dumps(sample(), separators=(",", ":")))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
