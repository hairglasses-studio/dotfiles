#!/usr/bin/env python3
"""ticker-smoke-test — load every ticker_streams plugin and call build().

Verifies the hybrid-architecture plugin contract end to end without
needing GTK or the layer-shell runtime:

    META     must be a dict with at least ``refresh`` (int).
    build()  must be callable and return ``(markup: str, segments: list)``.

Also loads the TOML catalogue at ``ticker/streams.toml`` so cache-fed
streams are exercised through the same factory dispatch used at runtime.

Exit code: 0 if every stream passes, 1 if any failed.
"""
from __future__ import annotations

import argparse
import importlib
import sys
import traceback
from pathlib import Path

HERE = Path(__file__).resolve().parent
LIB = HERE / "lib"
sys.path.insert(0, str(LIB))

try:
    import ticker_streams as ts  # noqa: E402
except Exception as e:
    print(f"FATAL: cannot import ticker_streams package: {e}")
    sys.exit(2)


def _pango_parse(markup: str) -> str | None:
    """Try to validate markup via Pango. Returns an error string or None.

    Imported lazily so the fast --fast path doesn't pull gi in. On
    headless systems without Pango, degrades gracefully by falling back
    to a simple bracket-balance heuristic.
    """
    try:
        import gi  # type: ignore
        gi.require_version("Pango", "1.0")
        from gi.repository import Pango  # type: ignore
    except Exception:
        # Balanced-tag heuristic as a last-resort: <span> count must
        # equal </span> count. Catches stray unclosed spans.
        open_n = markup.count("<span")
        close_n = markup.count("</span>")
        if open_n != close_n:
            return f"unbalanced <span>: {open_n} open vs {close_n} close"
        return None
    # The PyGObject binding expects `accel_marker` as a gunichar-str,
    # not int. Passing '\x00' disables accelerator-key extraction (which
    # the ticker doesn't use), leaving this as a pure well-formed-check.
    try:
        Pango.parse_markup(markup, -1, "\x00")
    except Exception as e:
        return f"pango parse failed: {e}"
    return None


def _check(markup, segments, *, validate_pango: bool = True) -> str | None:
    if not isinstance(markup, str):
        return f"markup is {type(markup).__name__}, expected str"
    if not isinstance(segments, list):
        return f"segments is {type(segments).__name__}, expected list"
    for i, seg in enumerate(segments):
        # Segments should be strings (or a sentinel like __URGENT__).
        if not isinstance(seg, str):
            return f"segments[{i}] is {type(seg).__name__}, expected str"
    if "<span" not in markup and "foreground" not in markup:
        return "markup missing <span> — likely raw/empty"
    if validate_pango:
        err = _pango_parse(markup)
        if err:
            return err
    return None


def check_plugin(name: str, *, skip_slow: bool = False,
                 validate_pango: bool = True) -> tuple[bool, str]:
    try:
        mod = importlib.import_module(f"ticker_streams.{name}")
    except Exception as e:
        return False, f"import failed: {e}"
    meta = getattr(mod, "META", None)
    build = getattr(mod, "build", None)
    if not isinstance(meta, dict):
        return False, "META missing or not a dict"
    if "refresh" not in meta:
        return False, "META.refresh missing"
    if not callable(build):
        return False, "build is not callable"
    if skip_slow and meta.get("slow"):
        return True, (f"refresh={meta['refresh']}s preset={meta.get('preset')!s}"
                      f"  (slow — skipped)")
    try:
        result = build()
    except Exception as e:
        return False, f"build() raised: {e}\n{traceback.format_exc()}"
    if not isinstance(result, tuple) or len(result) != 2:
        return False, f"build() returned {type(result).__name__}, expected 2-tuple"
    err = _check(*result, validate_pango=validate_pango)
    if err:
        return False, err
    return True, f"refresh={meta['refresh']}s preset={meta.get('preset')!s}"


def check_toml_stream(name: str, fn, *,
                      validate_pango: bool = True) -> tuple[bool, str]:
    try:
        result = fn()
    except Exception as e:
        return False, f"build() raised: {e}"
    if not isinstance(result, tuple) or len(result) != 2:
        return False, f"build() returned {type(result).__name__}, expected 2-tuple"
    err = _check(*result, validate_pango=validate_pango)
    if err:
        return False, err
    return True, "ok"


def main():
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument(
        "--fast", action="store_true",
        help="skip slow-threaded plugins (github/music/updates/claude-sessions); "
             "suitable for `ExecStartPre=` where startup latency matters",
    )
    ap.add_argument(
        "--no-pango", action="store_true",
        help="skip Pango.parse_markup validation (heuristic fallback only)",
    )
    args = ap.parse_args()

    plugin_dir = LIB / "ticker_streams"
    plugin_names = sorted(
        p.stem for p in plugin_dir.glob("*.py") if p.stem != "__init__"
    )

    toml_path = HERE.parent / "ticker" / "streams.toml"
    toml_builders, toml_meta, _toml_order = ts.load_toml_streams(str(toml_path))
    toml_names = sorted(toml_builders.keys())

    name_width = max(
        (len(n) for n in plugin_names + toml_names),
        default=10,
    ) + 2

    failures: list[tuple[str, str]] = []

    validate_pango = not args.no_pango
    print(f"─── plugin modules ({len(plugin_names)}) ──────────────────────────")
    for name in plugin_names:
        ok, detail = check_plugin(name,
                                  skip_slow=args.fast,
                                  validate_pango=validate_pango)
        mark = "PASS" if ok else "FAIL"
        print(f"  {mark}  {name:<{name_width}}{detail}")
        if not ok:
            failures.append((f"plugin:{name}", detail))

    print(f"\n─── TOML streams  ({len(toml_names)}) ──────────────────────────")
    for name in toml_names:
        ok, detail = check_toml_stream(name, toml_builders[name],
                                       validate_pango=validate_pango)
        mark = "PASS" if ok else "FAIL"
        meta = toml_meta.get(name, {})
        refresh = meta.get("refresh", "?")
        print(f"  {mark}  {name:<{name_width}}refresh={refresh}s  {detail}")
        if not ok:
            failures.append((f"toml:{name}", detail))

    total = len(plugin_names) + len(toml_names)
    passed = total - len(failures)
    print(f"\n{passed} / {total} streams passed")
    if failures:
        print("\nFAILURES:")
        for name, detail in failures:
            print(f"  {name}: {detail}")
        sys.exit(1)


if __name__ == "__main__":
    main()
