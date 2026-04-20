"""ticker_streams — declarative stream catalogue loader for keybind-ticker.

Loads a TOML file describing cache-fed streams and synthesises a
``build_<name>()`` callable for each one so the main ticker can consume
them through the same STREAMS / STREAM_META / FALLBACK_ORDER contract it
already uses for inline Python builders.

Supports two simple shapes today — both covering the bulk of bar-*.txt
cached content without per-stream Python:

``type = "cache_single"``  — one line of content shown in a bold span.
``type = "cache_list"``    — first line as a bold summary, subsequent
                             lines as list entries rendered in the
                             rotating FONTS cycle.

Complex streams (keybinds, github, music, …) stay as Python plugins in
``scripts/lib/ticker_streams/<name>.py`` with the ``META`` + ``build()``
contract — see the loader below.
"""
from __future__ import annotations

import importlib
import os
import tomllib
from html import escape
from pathlib import Path

import ticker_render as tr

_badge = tr.badge
_empty = tr.empty
_dup = tr.dup

FONTS = [
    "Maple Mono NF CN Bold 11",
    "Maple Mono NF CN Italic 11",
    "Maple Mono NF CN 11",
    "Maple Mono NF CN SemiBold 11",
]


def _read_cache_lines(path: str) -> list[str] | None:
    """Return cache lines or ``None`` if the file is missing.

    An empty file returns ``[]`` so callers can distinguish "cache not
    produced yet" from "cache says there's nothing to report."
    """
    try:
        with open(path) as f:
            return [l.rstrip("\n") for l in f if l.strip()]
    except FileNotFoundError:
        return None
    except OSError:
        return None


# ── Factories ─────────────────────────────────────────────────────────────

def _make_cache_single(cfg: dict):
    label = cfg["label"]
    color = cfg["color"]
    path = cfg["path"]
    text_font = cfg.get("text_font", "Maple Mono NF CN Bold 11")
    text_color = cfg.get("text_color")
    empty_msg = cfg.get("empty_message", "no data")
    missing_msg = cfg.get("missing_message", "cache missing")

    fg_attr = f' foreground="{text_color}"' if text_color else ""

    def build():
        try:
            with open(path) as f:
                raw = f.read().strip()
        except FileNotFoundError:
            return _empty(label, color, missing_msg)
        except OSError:
            return _empty(label, color, missing_msg)
        if not raw:
            return _empty(label, color, empty_msg)
        parts = [_badge(label, color)]
        parts.append(
            f'<span font_desc="{text_font}"{fg_attr}>  {escape(raw)}  \u00b7</span>'
        )
        return _dup("".join(parts)), []

    return build


def _make_cache_list(cfg: dict):
    label = cfg["label"]
    color = cfg["color"]
    path = cfg["path"]
    list_limit = int(cfg.get("list_limit", 6))
    has_summary = bool(cfg.get("has_summary", True))
    summary_font = cfg.get("summary_font", "Maple Mono NF CN Bold 11")
    summary_color = cfg.get("summary_color")
    item_color = cfg.get("item_color")
    fail_keywords = tuple(k.upper() for k in cfg.get("fail_keywords", []))
    fail_color = cfg.get("fail_color", "#ef4444")
    empty_msg = cfg.get("empty_message", "no data")
    missing_msg = cfg.get("missing_message", "cache missing")
    empty_is_success = bool(cfg.get("empty_is_success", False))
    empty_success_color = cfg.get("empty_success_color", "#34d399")

    summary_fg_attr = (
        f' foreground="{summary_color}"' if summary_color else ""
    )
    item_fg_attr = f' foreground="{item_color}"' if item_color else ""

    def build():
        lines = _read_cache_lines(path)
        if lines is None:
            return _empty(label, color, missing_msg)
        if not lines:
            empty_badge_color = empty_success_color if empty_is_success else color
            return _empty(label, empty_badge_color, empty_msg)

        effective_color = color
        if fail_keywords and any(
            any(kw in l.upper() for kw in fail_keywords) for l in lines
        ):
            effective_color = fail_color

        parts = [_badge(label, effective_color)]
        items = lines
        if has_summary:
            parts.append(
                f'<span font_desc="{summary_font}"{summary_fg_attr}>'
                f'  {escape(lines[0])}  \u00b7</span>'
            )
            items = lines[1:1 + list_limit]
        else:
            items = lines[:list_limit]
        fc = len(FONTS)
        for i, line in enumerate(items):
            font = FONTS[i % fc]
            parts.append(
                f'<span font_desc="{font}"{item_fg_attr}>  {escape(line)}  \u00b7</span>'
            )
        return _dup("".join(parts)), []

    return build


FACTORIES = {
    "cache_single": _make_cache_single,
    "cache_list":   _make_cache_list,
}


# ── Public entry points ───────────────────────────────────────────────────

def load_toml_streams(toml_path: str):
    """Load ``toml_path`` and return ``(builders, meta, order)``.

    - ``builders`` maps stream name → callable returning (markup, segments).
    - ``meta`` maps stream name → {"preset": str|None, "refresh": int, "dwell": int?}.
    - ``order`` is the declaration order so the caller can extend FALLBACK_ORDER.
    """
    if not os.path.exists(toml_path):
        return {}, {}, []
    with open(toml_path, "rb") as f:
        data = tomllib.load(f)
    builders: dict[str, callable] = {}
    meta: dict[str, dict] = {}
    order: list[str] = []
    for name, cfg in data.items():
        if not isinstance(cfg, dict):
            continue
        t = cfg.get("type")
        factory = FACTORIES.get(t)
        if factory is None:
            # Unknown type: skip rather than crash, keep the ticker rotating.
            continue
        builders[name] = factory(cfg)
        m = {"preset": cfg.get("preset"), "refresh": int(cfg.get("refresh", 300))}
        if "dwell" in cfg:
            m["dwell"] = int(cfg["dwell"])
        meta[name] = m
        order.append(name)
    return builders, meta, order


def load_plugin_streams(package_dir: str):
    """Discover ``<name>.py`` modules in ``package_dir`` that expose ``META``
    (dict) and ``build`` (callable). Returns ``(builders, meta, order, slow)``.

    ``slow`` is a set of stream names whose plugin declared
    ``META["slow"] = True`` — those get backgrounded the same way inline
    slow builders do today.
    """
    builders: dict[str, callable] = {}
    meta: dict[str, dict] = {}
    order: list[str] = []
    slow: set[str] = set()
    base = Path(package_dir)
    if not base.is_dir():
        return builders, meta, order, slow
    for f in sorted(base.glob("*.py")):
        if f.name == "__init__.py":
            continue
        name = f.stem
        mod_name = f"ticker_streams.{name}"
        try:
            mod = importlib.import_module(mod_name)
        except Exception:
            continue
        build = getattr(mod, "build", None)
        module_meta = getattr(mod, "META", None)
        if not callable(build) or not isinstance(module_meta, dict):
            continue
        # Allow the module to override its registered name via META["name"].
        reg_name = str(module_meta.get("name", name))
        builders[reg_name] = build
        entry = {
            "preset": module_meta.get("preset"),
            "refresh": int(module_meta.get("refresh", 300)),
        }
        if "dwell" in module_meta:
            entry["dwell"] = int(module_meta["dwell"])
        meta[reg_name] = entry
        order.append(reg_name)
        if module_meta.get("slow"):
            slow.add(reg_name)
    return builders, meta, order, slow
