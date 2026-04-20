"""hacker — rotating hacker-spirited quotes from a plaintext corpus.

Format: one line per quote, optional `\t<attribution>` suffix. Comment
lines start with `#`. Source file is resolved to
``$HG_DOTFILES/ticker/quotes/hacker.txt`` unless ``HACKER_QUOTES_PATH``
is set in the environment.
"""
from __future__ import annotations

import os
import random
from html import escape

import ticker_render as tr

META = {"preset": "cyberpunk", "refresh": 45}

_DEFAULT_PATH = os.path.expanduser(
    "~/hairglasses-studio/dotfiles/ticker/quotes/hacker.txt")

_cache: tuple[float, list[str]] | None = None


def _load() -> list[str]:
    global _cache
    path = os.environ.get("HACKER_QUOTES_PATH", _DEFAULT_PATH)
    try:
        mtime = os.path.getmtime(path)
    except OSError:
        return []
    if _cache and _cache[0] == mtime:
        return _cache[1]
    try:
        with open(path) as f:
            lines = [l.rstrip("\n") for l in f
                     if l.strip() and not l.lstrip().startswith("#")]
    except OSError:
        return _cache[1] if _cache else []
    _cache = (mtime, lines)
    return lines


def build():
    quotes = _load()
    if not quotes:
        return tr.empty("\U000f05f3 HACKER", "#34d399", "no quotes file")
    line = random.choice(quotes)
    parts = [tr.badge("\U000f05f3 HACKER", "#34d399")]
    if "\t" in line:
        quote, attribution = line.split("\t", 1)
        parts.append(
            f'<span font_desc="Maple Mono NF CN Italic 11" foreground="#a3e635">  {escape(quote)}  </span>'
        )
        parts.append(
            f'<span font_desc="Maple Mono NF CN Bold 10" foreground="#6ee7b7">\u2014 {escape(attribution)}  \u00b7</span>'
        )
    else:
        parts.append(
            f'<span font_desc="Maple Mono NF CN Italic 11" foreground="#a3e635">  {escape(line)}  \u00b7</span>'
        )
    return tr.dup("".join(parts)), []
