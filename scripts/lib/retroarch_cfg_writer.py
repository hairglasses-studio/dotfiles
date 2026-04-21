#!/usr/bin/env python3
"""Minimal in-place key/value writer for RetroArch's retroarch.cfg."""

from __future__ import annotations

import os
import shutil
import tempfile
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


def _format_line(key: str, value: str) -> str:
    return f'{key} = "{value}"\n'


def _parse_cfg_key(raw_line: str) -> str | None:
    line = raw_line.strip()
    if not line or line.startswith("#") or "=" not in line:
        return None
    key, _ = line.split("=", 1)
    return key.strip()


def _backup_path(cfg_path: Path) -> Path:
    stamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    return cfg_path.with_suffix(cfg_path.suffix + f".bak.{stamp}")


def apply_settings(
    cfg_path: Path,
    updates: dict[str, str],
    *,
    backup: bool = True,
    dry_run: bool = False,
) -> dict[str, Any]:
    """Set `updates` keys in `cfg_path`, preserving order and appending unknown keys.

    Returns {"set": {...}, "added": {...}, "unchanged": {...}, "backup_path": str|None, "applied": bool}.
    - `set`: keys that were present with a different value, now rewritten.
    - `added`: keys that were not present and got appended.
    - `unchanged`: keys that already had the desired value.
    - `backup_path`: populated only when a backup was actually written.
    - `applied`: False when dry_run is True or when no change was needed.
    """
    if not cfg_path.is_file():
        raise FileNotFoundError(f"retroarch.cfg not found: {cfg_path}")

    original_lines = cfg_path.read_text().splitlines(keepends=True)
    remaining = dict(updates)
    rewritten: list[str] = []
    set_keys: dict[str, dict[str, str]] = {}
    unchanged_keys: dict[str, str] = {}

    for raw_line in original_lines:
        key = _parse_cfg_key(raw_line)
        if key is not None and key in remaining:
            new_value = remaining.pop(key)
            current_value = raw_line.split("=", 1)[1].strip()
            if current_value.startswith('"') and current_value.endswith('"') and len(current_value) >= 2:
                current_value = current_value[1:-1]
            if current_value == new_value:
                rewritten.append(raw_line)
                unchanged_keys[key] = new_value
            else:
                rewritten.append(_format_line(key, new_value))
                set_keys[key] = {"from": current_value, "to": new_value}
            continue
        rewritten.append(raw_line)

    added_keys: dict[str, str] = {}
    if remaining:
        if rewritten and not rewritten[-1].endswith("\n"):
            rewritten[-1] = rewritten[-1] + "\n"
        for key, value in remaining.items():
            rewritten.append(_format_line(key, value))
            added_keys[key] = value

    no_change = not set_keys and not added_keys
    result: dict[str, Any] = {
        "set": set_keys,
        "added": added_keys,
        "unchanged": unchanged_keys,
        "backup_path": None,
        "applied": False,
    }

    if no_change or dry_run:
        return result

    backup_target = None
    if backup:
        backup_target = _backup_path(cfg_path)
        shutil.copy2(cfg_path, backup_target)
        result["backup_path"] = str(backup_target)

    tmp_fd, tmp_name = tempfile.mkstemp(
        prefix=cfg_path.name + ".",
        suffix=".tmp",
        dir=str(cfg_path.parent),
    )
    try:
        with os.fdopen(tmp_fd, "w") as handle:
            handle.writelines(rewritten)
        os.replace(tmp_name, cfg_path)
    except Exception:
        try:
            os.unlink(tmp_name)
        except FileNotFoundError:
            pass
        raise

    result["applied"] = True
    return result
