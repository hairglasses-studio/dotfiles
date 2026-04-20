#!/usr/bin/env python3
"""Shared RetroArch widescreen audit/apply helpers for Flycast and Dolphin."""

from __future__ import annotations

import argparse
import json
import os
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


_TRUE_VALUES = {"1", "true", "yes", "on", "enabled"}
_FALSE_VALUES = {"0", "false", "no", "off", "disabled"}


PROFILES: dict[str, dict[str, Any]] = {
    "flycast": {
        "core_name": "Flycast",
        "core_filename": "flycast_libretro.so",
        "option_dir": "Flycast",
        "option_filename": "Flycast.opt",
        "extensions": {".cdi", ".chd", ".cue", ".gdi", ".m3u"},
        "rom_aliases": ["dreamcast", "dc"],
        "status_core_missing": "core_missing",
        "status_no_content": "no_content",
    },
    "dolphin": {
        "core_name": "Dolphin",
        "core_filename": "dolphin_libretro.so",
        "option_dir": "Dolphin",
        "option_filename": "Dolphin.opt",
        "extensions": {".ciso", ".dol", ".elf", ".gcm", ".gcz", ".iso", ".rvz", ".wad", ".wbfs"},
        "rom_aliases": ["gamecube", "gc", "wii"],
        "status_core_missing": "core_missing",
        "status_no_content": "no_content",
    },
}


def _expand_path(value: str | None, default: Path) -> Path:
    if not value:
        return default
    return Path(os.path.expandvars(os.path.expanduser(value)))


def _parse_ra_kv_file(path: Path) -> dict[str, str]:
    data: dict[str, str] = {}
    if not path.exists():
        return data
    for raw_line in path.read_text(errors="ignore").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        value = value.strip()
        if value.startswith('"') and value.endswith('"') and len(value) >= 2:
            value = value[1:-1]
        data[key.strip()] = value
    return data


def _boolish(value: str | None) -> bool | None:
    if value is None:
        return None
    lowered = value.strip().lower()
    if lowered in _TRUE_VALUES:
        return True
    if lowered in _FALSE_VALUES:
        return False
    return None


def _first_existing(candidates: list[Path]) -> Path | None:
    for candidate in candidates:
        if candidate.exists():
            return candidate
    return None


def _discover_core_path(core_filename: str, retroarch_settings: dict[str, str], config_dir: Path, explicit_core_dir: Path | None) -> tuple[Path, bool]:
    candidates: list[Path] = []
    if explicit_core_dir is not None:
        candidates.append(explicit_core_dir / core_filename)

    libretro_dir = retroarch_settings.get("libretro_directory")
    if libretro_dir:
        candidates.append(_expand_path(libretro_dir, config_dir / "cores") / core_filename)

    candidates.extend(
        [
            config_dir / "cores" / core_filename,
            Path("/usr/lib/libretro") / core_filename,
            Path("/usr/local/lib/libretro") / core_filename,
        ]
    )
    existing = _first_existing(candidates)
    if existing is not None:
        return existing, True
    return candidates[0], False


def _discover_roms(roms_dir: Path, aliases: list[str], extensions: set[str]) -> list[Path]:
    matches: list[Path] = []
    seen: set[Path] = set()
    for alias in aliases:
        candidate_dir = roms_dir / alias
        if not candidate_dir.is_dir():
            continue
        for path in sorted(candidate_dir.rglob("*")):
            if not path.is_file():
                continue
            if path.suffix.lower() not in extensions:
                continue
            resolved = path.resolve()
            if resolved in seen:
                continue
            seen.add(resolved)
            matches.append(resolved)
    return matches


def _candidate_content_option_paths(config_dir: Path, option_dir: str, rom_path: Path) -> list[Path]:
    base = config_dir / "config" / option_dir
    return [
        base / f"{rom_path.stem}.opt",
        base / f"{rom_path.name}.opt",
    ]


def _existing_option_file(paths: list[Path]) -> Path | None:
    return _first_existing(paths)


def _dedupe(items: list[str]) -> list[str]:
    seen: set[str] = set()
    out: list[str] = []
    for item in items:
        if item in seen:
            continue
        seen.add(item)
        out.append(item)
    return out


def _int_or_none(value: str | None) -> int | None:
    if value is None:
        return None
    try:
        return int(value)
    except ValueError:
        return None


def _aspect_class(value: float | None) -> str | None:
    if value is None:
        return None
    if value >= 3.2:
        return "32:9"
    if value >= 2.2:
        return "21:9"
    if value >= 1.68:
        return "16:9"
    return "4:3-or-less"


def _display_profile(retroarch_settings: dict[str, str]) -> dict[str, Any]:
    width = _int_or_none(retroarch_settings.get("video_fullscreen_x"))
    height = _int_or_none(retroarch_settings.get("video_fullscreen_y"))
    aspect_value = None
    if width and height and height != 0:
        aspect_value = round(width / height, 6)
    return {
        "fullscreen_width": width,
        "fullscreen_height": height,
        "aspect_value": aspect_value,
        "aspect_class": _aspect_class(aspect_value),
        "monitor_index": _int_or_none(retroarch_settings.get("video_monitor_index")),
    }


def _dolphin_content_type(path: Path) -> str:
    parts = {part.lower() for part in path.parts}
    ext = path.suffix.lower()
    if "wii" in parts or ext in {".wbfs", ".wad"}:
        return "wii"
    if "gamecube" in parts or "gc" in parts:
        return "gamecube"
    return "gamecube"


def _flycast_effective_mode(cheats: bool, hack: bool) -> str:
    if cheats and hack:
        return "mixed"
    if cheats:
        return "cheats"
    if hack:
        return "hack"
    return "disabled"


def _dolphin_effective_mode(content_type: str, native_wii: bool, hack: bool) -> str:
    if content_type == "wii" and native_wii and hack:
        return "native_plus_hack"
    if content_type == "wii" and native_wii:
        return "native_wii"
    if hack:
        return "hack"
    return "disabled"


def _base_paths(system: str, args: argparse.Namespace) -> dict[str, Path]:
    config_dir = _expand_path(
        args.config_dir or os.environ.get("RETROARCH_CONFIG_DIR"),
        Path.home() / ".config" / "retroarch",
    )
    roms_dir = _expand_path(
        args.roms_dir or os.environ.get("RETROARCH_ROMS_DIR"),
        Path.home() / "Games" / "RetroArch" / "roms",
    )
    state_home = _expand_path(
        args.state_home or os.environ.get("RETROARCH_STATE_HOME"),
        Path.home() / ".local" / "state",
    )
    retroarch_settings = _parse_ra_kv_file(config_dir / "retroarch.cfg")
    system_dir = _expand_path(
        args.system_dir or os.environ.get("RETROARCH_SYSTEM_DIR") or retroarch_settings.get("system_directory"),
        config_dir / "system",
    )
    core_dir = None
    if args.core_dir or os.environ.get("RETROARCH_CORE_DIR"):
        core_dir = _expand_path(args.core_dir or os.environ.get("RETROARCH_CORE_DIR"), config_dir / "cores")
    output_path = _expand_path(
        args.output or os.environ.get(f"RETROARCH_{system.upper()}_AUDIT_PATH"),
        state_home / f"retroarch-{system}" / "widescreen-audit.json",
    )
    return {
        "config_dir": config_dir,
        "roms_dir": roms_dir,
        "state_home": state_home,
        "system_dir": system_dir,
        "output_path": output_path,
        "core_dir": core_dir,
        "retroarch_settings": Path(str(config_dir / "retroarch.cfg")),
        "retroarch_settings_values": retroarch_settings,
    }


def _write_text_if_changed(path: Path, content: str) -> bool:
    if path.exists() and path.read_text(errors="ignore") == content:
        return False
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content)
    return True


def _ensure_dir(path: Path) -> bool:
    if path.exists():
        return False
    path.mkdir(parents=True, exist_ok=True)
    return True


def _install_core_binary(core_path: Path, package_basename: str) -> bool:
    if core_path.exists():
        return False
    package_path = Path("/usr/lib/libretro") / f"{package_basename}_libretro.so"
    if not package_path.exists():
        return False
    core_path.parent.mkdir(parents=True, exist_ok=True)
    if core_path.exists() or core_path.is_symlink():
        core_path.unlink()
    try:
        core_path.symlink_to(package_path)
    except FileExistsError:
        pass
    return True


def _append_step(results: dict[str, list[str]], key: str, message: str) -> None:
    results.setdefault(key, []).append(message)


def _normalize_paths_for_apply(system: str, args: argparse.Namespace) -> dict[str, Any]:
    paths = _base_paths(system, args)
    profile = PROFILES[system]
    paths["profile"] = profile
    core_path, core_installed = _discover_core_path(
        profile["core_filename"],
        paths["retroarch_settings_values"],
        paths["config_dir"],
        paths["core_dir"],
    )
    paths["core_path"] = core_path
    paths["core_installed"] = core_installed
    paths["core_option_path"] = paths["config_dir"] / "config" / profile["option_dir"] / profile["option_filename"]
    paths["core_cfg_path"] = paths["config_dir"] / "config" / profile["option_dir"] / f"{profile['option_dir']}.cfg"
    return paths


def _apply_flycast_defaults(args: argparse.Namespace) -> dict[str, Any]:
    paths = _normalize_paths_for_apply("flycast", args)
    display = _display_profile(paths["retroarch_settings_values"])
    results: dict[str, list[str]] = {"created": [], "updated": [], "notes": []}

    created = _ensure_dir(paths["config_dir"] / "config" / "Flycast")
    if created:
        _append_step(results, "created", f"Created config directory: {paths['config_dir'] / 'config' / 'Flycast'}")

    core_linked = _install_core_binary(paths["config_dir"] / "cores" / "flycast_libretro.so", "flycast")
    if core_linked:
        _append_step(results, "updated", f"Installed local core link: {paths['config_dir'] / 'cores' / 'flycast_libretro.so'}")

    cfg_content = (
        'video_smooth = "true"\n'
        'run_ahead_enabled = "false"\n'
    )
    if _write_text_if_changed(paths["core_cfg_path"], cfg_content):
        _append_step(results, "updated", f"Wrote core config: {paths['core_cfg_path']}")

    opt_content = (
        'flycast_widescreen_cheats = "On"\n'
        'flycast_widescreen_hack = "Off"\n'
    )
    if _write_text_if_changed(paths["core_option_path"], opt_content):
        _append_step(results, "updated", f"Wrote core options: {paths['core_option_path']}")

    for alias in ("dreamcast", "dc"):
        alias_dir = paths["roms_dir"] / alias
        if _ensure_dir(alias_dir):
            _append_step(results, "created", f"Created ROM directory: {alias_dir}")

    if display["aspect_class"] == "32:9":
        _append_step(results, "notes", "Detected 5120x1440-class display. Flycast defaults were kept on the safe 16:9 path via widescreen cheats, with the generic hack left off.")
    else:
        _append_step(results, "notes", "Flycast defaults prefer widescreen cheats over the generic hack.")

    return {
        "system": "flycast",
        "applied_at": datetime.now(timezone.utc).isoformat(),
        "display": display,
        "paths": {
            "config_dir": str(paths["config_dir"]),
            "roms_dir": str(paths["roms_dir"]),
            "core_cfg_path": str(paths["core_cfg_path"]),
            "core_option_path": str(paths["core_option_path"]),
        },
        "results": results,
    }


def _apply_dolphin_defaults(args: argparse.Namespace) -> dict[str, Any]:
    paths = _normalize_paths_for_apply("dolphin", args)
    display = _display_profile(paths["retroarch_settings_values"])
    results: dict[str, list[str]] = {"created": [], "updated": [], "notes": []}

    config_dir = paths["config_dir"] / "config" / "Dolphin"
    if _ensure_dir(config_dir):
        _append_step(results, "created", f"Created config directory: {config_dir}")

    core_linked = _install_core_binary(paths["config_dir"] / "cores" / "dolphin_libretro.so", "dolphin")
    if core_linked:
        _append_step(results, "updated", f"Installed local core link: {paths['config_dir'] / 'cores' / 'dolphin_libretro.so'}")

    cfg_content = (
        'video_hard_sync = "false"\n'
        'run_ahead_enabled = "false"\n'
        'video_frame_delay_auto = "false"\n'
        'fastforward_ratio = "0.0"\n'
    )
    if _write_text_if_changed(paths["core_cfg_path"], cfg_content):
        _append_step(results, "updated", f"Wrote core config: {paths['core_cfg_path']}")

    opt_content = (
        'dolphin_widescreen = "ON"\n'
        'dolphin_widescreen_hack = "OFF"\n'
    )
    if _write_text_if_changed(paths["core_option_path"], opt_content):
        _append_step(results, "updated", f"Wrote core options: {paths['core_option_path']}")

    dolphin_sys = paths["system_dir"] / "dolphin-emu" / "Sys"
    if _ensure_dir(dolphin_sys):
        _append_step(results, "created", f"Created Dolphin Sys placeholder directory: {dolphin_sys}")
        _append_step(results, "notes", "The Dolphin Sys directory was created, but the upstream compatibility data still needs to be synced into it.")

    for alias in ("gamecube", "gc", "wii"):
        alias_dir = paths["roms_dir"] / alias
        if _ensure_dir(alias_dir):
            _append_step(results, "created", f"Created ROM directory: {alias_dir}")

    if display["aspect_class"] == "32:9":
        _append_step(results, "notes", "Detected 5120x1440-class display. Dolphin defaults were kept on the safe 16:9 path: native Wii widescreen on, generic widescreen hack off.")
    else:
        _append_step(results, "notes", "Dolphin defaults prefer native Wii widescreen and keep the generic widescreen hack off.")

    return {
        "system": "dolphin",
        "applied_at": datetime.now(timezone.utc).isoformat(),
        "display": display,
        "paths": {
            "config_dir": str(paths["config_dir"]),
            "roms_dir": str(paths["roms_dir"]),
            "core_cfg_path": str(paths["core_cfg_path"]),
            "core_option_path": str(paths["core_option_path"]),
            "dolphin_sys_path": str(dolphin_sys),
        },
        "results": results,
    }


def _build_flycast_report(args: argparse.Namespace) -> dict[str, Any]:
    profile = PROFILES["flycast"]
    paths = _base_paths("flycast", args)
    retroarch_settings = paths["retroarch_settings_values"]
    display = _display_profile(retroarch_settings)
    core_path, core_installed = _discover_core_path(
        profile["core_filename"],
        retroarch_settings,
        paths["config_dir"],
        paths["core_dir"],
    )
    option_path = paths["config_dir"] / "config" / profile["option_dir"] / profile["option_filename"]
    global_options = _parse_ra_kv_file(option_path)
    global_cheats = _boolish(global_options.get("flycast_widescreen_cheats")) is True
    global_hack = _boolish(global_options.get("flycast_widescreen_hack")) is True
    roms = _discover_roms(paths["roms_dir"], profile["rom_aliases"], profile["extensions"])

    entries: list[dict[str, Any]] = []
    counts = Counter()
    ultrawide_counts = Counter()
    warnings: list[str] = []
    next_steps: list[str] = []

    if not core_installed:
        warnings.append(f"Missing core binary: expected {core_path}")
        next_steps.append("Install the Flycast libretro core before running launch-based verification.")

    if not option_path.exists():
        warnings.append(f"Missing core option file: {option_path}")
        next_steps.append("Launch Flycast once in RetroArch and save core options to create config/Flycast/Flycast.opt.")

    if global_hack:
        warnings.append("Global Flycast widescreen hack is enabled; prefer cheats first because the hack can reveal broken geometry.")
    if global_cheats and global_hack:
        warnings.append("Both Flycast widescreen cheats and the generic hack are enabled globally; keep the hack off unless a title needs fallback coverage.")

    if not roms:
        next_steps.append("Mount or copy Dreamcast launch files into ~/Games/RetroArch/roms/dreamcast or ~/Games/RetroArch/roms/dc.")

    for rom_path in roms:
        content_opt_path = _existing_option_file(
            _candidate_content_option_paths(paths["config_dir"], profile["option_dir"], rom_path)
        )
        content_options = _parse_ra_kv_file(content_opt_path) if content_opt_path else {}
        effective_cheats = _boolish(content_options.get("flycast_widescreen_cheats"))
        if effective_cheats is None:
            effective_cheats = global_cheats
        effective_hack = _boolish(content_options.get("flycast_widescreen_hack"))
        if effective_hack is None:
            effective_hack = global_hack

        mode = _flycast_effective_mode(bool(effective_cheats), bool(effective_hack))
        counts[mode] += 1
        ultrawide_mode = "experimental_32:9_via_hack" if bool(effective_hack) else ("safe_16:9_only" if bool(effective_cheats) else "disabled")
        ultrawide_counts[ultrawide_mode] += 1
        recommendation = "Keep widescreen cheats enabled and leave the generic hack off." if mode in {"cheats", "mixed"} else (
            "Fallback hack is active; verify geometry and UI for this title." if mode == "hack" else "Enable flycast_widescreen_cheats before considering the generic hack."
        )
        if display["aspect_class"] == "32:9" and ultrawide_mode == "safe_16:9_only":
            recommendation += " On a 5120x1440-class display, keep the game at safe 16:9 and use borders or side-fill instead of forcing true 32:9."
        elif display["aspect_class"] == "32:9" and ultrawide_mode == "experimental_32:9_via_hack":
            recommendation += " The current setup can stretch into true 32:9 via the generic hack, but this should stay opt-in and verified per title."
        entries.append(
            {
                "title": rom_path.stem,
                "path": str(rom_path),
                "extension": rom_path.suffix.lower(),
                "content_opt_path": str(content_opt_path) if content_opt_path else None,
                "effective_widescreen_cheats": bool(effective_cheats),
                "effective_widescreen_hack": bool(effective_hack),
                "effective_mode": mode,
                "ultrawide_mode": ultrawide_mode,
                "recommendation": recommendation,
            }
        )

    status = "ready"
    if not core_installed:
        status = profile["status_core_missing"]
    elif not roms:
        status = profile["status_no_content"]

    summary = {
        "count": len(entries),
        "cheats_count": counts["cheats"],
        "hack_only_count": counts["hack"],
        "mixed_count": counts["mixed"],
        "disabled_count": counts["disabled"],
    }
    ultrawide_summary = {
        "display_aspect_class": display["aspect_class"],
        "safe_16_9_only_count": ultrawide_counts["safe_16:9_only"],
        "experimental_32_9_count": ultrawide_counts["experimental_32:9_via_hack"],
        "disabled_count": ultrawide_counts["disabled"],
    }

    next_steps.append("Prefer flycast_widescreen_cheats over flycast_widescreen_hack for default automation.")
    if display["aspect_class"] == "32:9":
        next_steps.append("For 5120x1440 displays, treat Flycast cheats as the safe 16:9 path and reserve true 32:9 for explicit hack-based verification.")

    return {
        "system": "flycast",
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "status": status,
        "summary": summary,
        "ultrawide_summary": ultrawide_summary,
        "core": {
            "name": profile["core_name"],
            "path": str(core_path),
            "installed": core_installed,
        },
        "display": display,
        "paths": {
            "retroarch_cfg": str(paths["retroarch_settings"]),
            "config_dir": str(paths["config_dir"]),
            "roms_dir": str(paths["roms_dir"]),
            "system_dir": str(paths["system_dir"]),
            "output_path": str(paths["output_path"]),
            "core_option_path": str(option_path),
        },
        "global_options": {
            "flycast_widescreen_cheats": global_options.get("flycast_widescreen_cheats"),
            "flycast_widescreen_hack": global_options.get("flycast_widescreen_hack"),
        },
        "warnings": _dedupe(warnings),
        "next_steps": _dedupe(next_steps),
        "entries": entries,
    }


def _build_dolphin_report(args: argparse.Namespace) -> dict[str, Any]:
    profile = PROFILES["dolphin"]
    paths = _base_paths("dolphin", args)
    retroarch_settings = paths["retroarch_settings_values"]
    display = _display_profile(retroarch_settings)
    core_path, core_installed = _discover_core_path(
        profile["core_filename"],
        retroarch_settings,
        paths["config_dir"],
        paths["core_dir"],
    )
    option_path = paths["config_dir"] / "config" / profile["option_dir"] / profile["option_filename"]
    global_options = _parse_ra_kv_file(option_path)
    global_native_wii = _boolish(global_options.get("dolphin_widescreen")) is True
    global_hack = _boolish(global_options.get("dolphin_widescreen_hack")) is True
    roms = _discover_roms(paths["roms_dir"], profile["rom_aliases"], profile["extensions"])
    dolphin_sys = paths["system_dir"] / "dolphin-emu" / "Sys"

    entries: list[dict[str, Any]] = []
    counts = Counter()
    ultrawide_counts = Counter()
    warnings: list[str] = []
    next_steps: list[str] = []

    if not core_installed:
        warnings.append(f"Missing core binary: expected {core_path}")
        next_steps.append("Install the Dolphin libretro core before running launch-based verification.")

    if not dolphin_sys.exists():
        warnings.append(f"Missing Dolphin Sys directory: {dolphin_sys}")
        next_steps.append("Install or sync the Dolphin Sys folder into the RetroArch system directory before relying on per-game compatibility data.")

    if not option_path.exists():
        warnings.append(f"Missing core option file: {option_path}")
        next_steps.append("Launch Dolphin once in RetroArch and save core options to create config/Dolphin/Dolphin.opt.")

    if global_hack:
        warnings.append("Global Dolphin widescreen hack is enabled; upstream warns that it often breaks graphics and should not be the default path.")
    if global_native_wii and global_hack:
        warnings.append("Both Wii native widescreen and the Dolphin widescreen hack are enabled globally; keep the hack off unless a specific title requires fallback testing.")

    if not roms:
        next_steps.append("Mount or copy GameCube/Wii launch files into ~/Games/RetroArch/roms/gamecube, ~/Games/RetroArch/roms/gc, or ~/Games/RetroArch/roms/wii.")

    for rom_path in roms:
        content_type = _dolphin_content_type(rom_path)
        content_opt_path = _existing_option_file(
            _candidate_content_option_paths(paths["config_dir"], profile["option_dir"], rom_path)
        )
        content_options = _parse_ra_kv_file(content_opt_path) if content_opt_path else {}
        native_wii = _boolish(content_options.get("dolphin_widescreen"))
        if native_wii is None:
            native_wii = global_native_wii
        hack = _boolish(content_options.get("dolphin_widescreen_hack"))
        if hack is None:
            hack = global_hack

        mode = _dolphin_effective_mode(content_type, bool(native_wii), bool(hack))
        counts[mode] += 1
        counts[content_type] += 1
        ultrawide_mode = "experimental_32:9_via_hack" if bool(hack) else ("safe_16:9_only" if content_type == "wii" and bool(native_wii) else "disabled")
        ultrawide_counts[ultrawide_mode] += 1

        if content_type == "wii":
            recommendation = "Prefer native Wii widescreen; only add the generic hack for targeted fallback testing." if mode in {"native_wii", "native_plus_hack"} else (
                "Enable dolphin_widescreen for Wii titles before using the generic hack." if mode == "disabled" else "Generic hack is active; verify HUD and post-processing artifacts for this Wii title."
            )
        else:
            recommendation = "Prefer per-game AR/Gecko patches for GameCube titles; do not default to the generic widescreen hack." if mode == "disabled" else (
                "Generic hack is active for a GameCube title; verify UI and screen effects carefully."
            )
        if display["aspect_class"] == "32:9" and ultrawide_mode == "safe_16:9_only":
            recommendation += " On a 5120x1440-class display, keep this title on the safe 16:9 path unless you intentionally validate hack behavior."
        elif display["aspect_class"] == "32:9" and ultrawide_mode == "experimental_32:9_via_hack":
            recommendation += " The current setup can expose more than 16:9 on a 32:9 display, but that remains experimental and should not be the default automation path."

        entries.append(
            {
                "title": rom_path.stem,
                "path": str(rom_path),
                "extension": rom_path.suffix.lower(),
                "content_type": content_type,
                "content_opt_path": str(content_opt_path) if content_opt_path else None,
                "effective_widescreen_wii": bool(native_wii),
                "effective_widescreen_hack": bool(hack),
                "effective_mode": mode,
                "ultrawide_mode": ultrawide_mode,
                "recommendation": recommendation,
            }
        )

    status = "ready"
    if not core_installed:
        status = profile["status_core_missing"]
    elif not roms:
        status = profile["status_no_content"]
    elif not dolphin_sys.exists():
        status = "needs_setup"

    summary = {
        "count": len(entries),
        "wii_count": counts["wii"],
        "gamecube_count": counts["gamecube"],
        "native_wii_count": counts["native_wii"],
        "native_plus_hack_count": counts["native_plus_hack"],
        "hack_count": counts["hack"],
        "disabled_count": counts["disabled"],
    }
    ultrawide_summary = {
        "display_aspect_class": display["aspect_class"],
        "safe_16_9_only_count": ultrawide_counts["safe_16:9_only"],
        "experimental_32_9_count": ultrawide_counts["experimental_32:9_via_hack"],
        "disabled_count": ultrawide_counts["disabled"],
    }

    next_steps.append("Prefer dolphin_widescreen for Wii titles and reserve dolphin_widescreen_hack for explicit fallback testing.")
    if display["aspect_class"] == "32:9":
        next_steps.append("For 5120x1440 displays, use Dolphin's hack path only for explicit per-title ultrawide validation; keep native Wii widescreen and GameCube patch workflows on the safe 16:9 path by default.")

    return {
        "system": "dolphin",
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "status": status,
        "summary": summary,
        "ultrawide_summary": ultrawide_summary,
        "core": {
            "name": profile["core_name"],
            "path": str(core_path),
            "installed": core_installed,
        },
        "display": display,
        "paths": {
            "retroarch_cfg": str(paths["retroarch_settings"]),
            "config_dir": str(paths["config_dir"]),
            "roms_dir": str(paths["roms_dir"]),
            "system_dir": str(paths["system_dir"]),
            "output_path": str(paths["output_path"]),
            "core_option_path": str(option_path),
            "dolphin_sys_path": str(dolphin_sys),
        },
        "global_options": {
            "dolphin_widescreen": global_options.get("dolphin_widescreen"),
            "dolphin_widescreen_hack": global_options.get("dolphin_widescreen_hack"),
        },
        "warnings": _dedupe(warnings),
        "next_steps": _dedupe(next_steps),
        "entries": entries,
    }


def build_report(system: str, args: argparse.Namespace) -> dict[str, Any]:
    if system == "flycast":
        return _build_flycast_report(args)
    if system == "dolphin":
        return _build_dolphin_report(args)
    raise ValueError(f"Unsupported system: {system}")


def apply_defaults(system: str, args: argparse.Namespace) -> dict[str, Any]:
    if system == "flycast":
        return _apply_flycast_defaults(args)
    if system == "dolphin":
        return _apply_dolphin_defaults(args)
    raise ValueError(f"Unsupported system: {system}")


def _parse_args(system: str, argv: list[str] | None, mode: str) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=f"{mode.title()} RetroArch widescreen settings for {system}.")
    parser.add_argument("--config-dir", help="Override RetroArch config directory.")
    parser.add_argument("--roms-dir", help="Override RetroArch ROM root directory.")
    parser.add_argument("--core-dir", help="Override RetroArch libretro core directory.")
    parser.add_argument("--system-dir", help="Override RetroArch system directory.")
    parser.add_argument("--state-home", help="Override state home directory.")
    parser.add_argument("--output", help="Write the report to a custom path.")
    return parser.parse_args(argv)


def main(system: str, argv: list[str] | None = None) -> int:
    args = _parse_args(system, argv, "audit")
    report = build_report(system, args)
    output_path = Path(report["paths"]["output_path"])
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(report, indent=2))

    summary = report["summary"]
    summary_bits = " ".join(f"{key}={value}" for key, value in summary.items())
    print(output_path)
    print(f"status={report['status']} {summary_bits}")
    return 0


def apply_main(system: str, argv: list[str] | None = None) -> int:
    args = _parse_args(system, argv, "apply")
    result = apply_defaults(system, args)
    output_path = _expand_path(
        args.output or os.environ.get(f"RETROARCH_{system.upper()}_APPLY_PATH"),
        Path.home() / ".local" / "state" / f"retroarch-{system}" / "widescreen-apply.json",
    )
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(result, indent=2))
    created = len(result["results"].get("created", []))
    updated = len(result["results"].get("updated", []))
    print(output_path)
    print(f"created={created} updated={updated}")
    return 0
