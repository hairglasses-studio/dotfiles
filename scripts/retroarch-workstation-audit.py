#!/usr/bin/env python3
"""Audit RetroArch workstation readiness for cores, BIOS/helper assets, and runtime control."""

from __future__ import annotations

import argparse
import hashlib
import json
import os
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

import retroarch_profile
import retroarch_runtime


_TRUE_VALUES = {"1", "true", "yes", "on", "enabled"}


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
    if lowered in {"0", "false", "no", "off", "disabled"}:
        return False
    return None


def _display_profile(retroarch_settings: dict[str, str]) -> dict[str, Any]:
    width = _int_or_none(retroarch_settings.get("video_fullscreen_x"))
    height = _int_or_none(retroarch_settings.get("video_fullscreen_y"))
    aspect_value = round(width / height, 6) if width and height else None
    if aspect_value is None:
        aspect_class = None
    elif aspect_value >= 3.2:
        aspect_class = "32:9"
    elif aspect_value >= 2.2:
        aspect_class = "21:9"
    elif aspect_value >= 1.68:
        aspect_class = "16:9"
    else:
        aspect_class = "4:3-or-less"
    return {
        "fullscreen_width": width,
        "fullscreen_height": height,
        "aspect_value": aspect_value,
        "aspect_class": aspect_class,
        "monitor_index": _int_or_none(retroarch_settings.get("video_monitor_index")),
    }


def _int_or_none(value: str | None) -> int | None:
    if value is None:
        return None
    try:
        return int(value)
    except ValueError:
        return None


def _first_existing(candidates: list[Path]) -> Path | None:
    for candidate in candidates:
        if candidate.exists():
            return candidate
    return None


def _core_candidates(core_filename: str, libretro_directory: str | None, config_dir: Path) -> list[Path]:
    candidates: list[Path] = []
    if libretro_directory:
        candidates.append(_expand_path(libretro_directory, config_dir / "cores") / core_filename)
    candidates.extend(
        [
            config_dir / "cores" / core_filename,
            Path("/usr/lib/libretro") / core_filename,
            Path("/usr/local/lib/libretro") / core_filename,
        ]
    )
    deduped: list[Path] = []
    seen: set[str] = set()
    for candidate in candidates:
        key = str(candidate)
        if key in seen:
            continue
        seen.add(key)
        deduped.append(candidate)
    return deduped


def _count_roms(path: Path, extensions: set[str]) -> int:
    if not path.is_dir():
        return 0
    allowed = set(extensions)
    allowed.update({".zip", ".7z"})
    try:
        root_device = path.stat().st_dev
    except OSError:
        return 0
    count = 0
    for current_root, dirnames, filenames in os.walk(path, followlinks=False):
        pruned: list[str] = []
        for dirname in dirnames:
            candidate = Path(current_root) / dirname
            try:
                if candidate.is_symlink() or candidate.stat().st_dev != root_device:
                    continue
            except OSError:
                continue
            pruned.append(dirname)
        dirnames[:] = pruned

        for filename in filenames:
            item = Path(current_root) / filename
            if item.suffix.lower() not in allowed:
                continue
            try:
                if item.is_file() or item.is_symlink():
                    count += 1
            except OSError:
                continue
    return count


def _playlist_item_count(path: Path) -> int:
    if not path.exists():
        return 0
    try:
        data = json.loads(path.read_text())
    except json.JSONDecodeError:
        return 0
    items = data.get("items", [])
    return len(items) if isinstance(items, list) else 0


def _file_md5(path: Path) -> str | None:
    if not path.is_file():
        return None
    digest = hashlib.md5()
    with path.open("rb") as handle:
        while True:
            chunk = handle.read(1024 * 1024)
            if not chunk:
                break
            digest.update(chunk)
    return digest.hexdigest()


def _dir_nonempty(path: Path) -> bool:
    if not path.is_dir():
        return False
    return any(path.iterdir())


def _run_command(argv: list[str]) -> tuple[int, str]:
    completed = subprocess.run(argv, capture_output=True, text=True, check=False)
    output = completed.stdout.strip() or completed.stderr.strip()
    return completed.returncode, output


def _package_status(package_name: str | None) -> dict[str, Any]:
    if not package_name:
        return {"name": None, "available": None, "installed": None}
    if not shutil_which("pacman"):
        return {"name": package_name, "available": None, "installed": None}
    installed_code, _ = _run_command(["pacman", "-Q", package_name])
    available_code, _ = _run_command(["pacman", "-Si", package_name])
    return {
        "name": package_name,
        "installed": installed_code == 0,
        "available": available_code == 0,
    }


def shutil_which(name: str) -> str | None:
    for directory in os.environ.get("PATH", "").split(os.pathsep):
        candidate = Path(directory) / name
        if candidate.is_file() and os.access(candidate, os.X_OK):
            return str(candidate)
    return None


def _runtime_status(settings: dict[str, str]) -> dict[str, Any]:
    enabled = _boolish(settings.get("network_cmd_enable")) is True
    port = _int_or_none(settings.get("network_cmd_port")) or 55355
    running_code, running_output = _run_command(["pgrep", "-x", "retroarch"])
    running = running_code == 0
    version_probe = None
    if enabled and running:
        version_probe = retroarch_runtime.send_udp_command(
            "VERSION",
            port=port,
            expect_response=True,
        )
    return {
        "process_running": running,
        "process_probe": running_output if running_output else None,
        "network_cmd_enable": enabled,
        "network_cmd_port": port,
        "version_probe": version_probe,
    }


def _requirement_rows(
    requirement_catalog: list[dict[str, object]],
    system_dir: Path,
    config_dir: Path,
) -> list[dict[str, Any]]:
    desmume_options = _parse_ra_kv_file(config_dir / "config" / "DeSmuME" / "DeSmuME.opt")
    desmume_external_bios = _boolish(desmume_options.get("desmume_use_external_bios")) is True

    rows: list[dict[str, Any]] = []
    for entry in requirement_catalog:
        system = str(entry["system"])
        relative_path = str(entry["relative_path"])
        absolute_path = system_dir / relative_path
        kind = str(entry["kind"])
        requirement = str(entry["requirement"])
        expected_md5 = entry.get("md5")

        effective_requirement = requirement
        if system == "nds" and requirement == "conditional":
            effective_requirement = "required" if desmume_external_bios else "optional"

        exists = False
        actual_md5 = None
        status = "missing"
        if kind == "file":
            exists = absolute_path.is_file()
            actual_md5 = _file_md5(absolute_path)
            if exists:
                status = "ok"
                if expected_md5 and actual_md5 and actual_md5 != expected_md5:
                    status = "mismatch"
        elif kind == "dir":
            exists = absolute_path.is_dir()
            status = "ok" if exists else "missing"
        elif kind == "dir_nonempty":
            exists = _dir_nonempty(absolute_path)
            status = "ok" if exists else "missing"

        if effective_requirement == "optional" and status == "missing":
            status = "optional_missing"

        rows.append(
            {
                "system": system,
                "kind": kind,
                "requirement": effective_requirement,
                "description": entry["description"],
                "relative_path": relative_path,
                "absolute_path": str(absolute_path),
                "exists": exists,
                "expected_md5": expected_md5,
                "actual_md5": actual_md5,
                "status": status,
            }
        )
    return rows


def build_report(args: argparse.Namespace) -> dict[str, Any]:
    config_dir = _expand_path(
        args.config_dir or os.environ.get("RETROARCH_CONFIG_DIR"),
        Path.home() / ".config" / "retroarch",
    )
    retroarch_cfg = config_dir / "retroarch.cfg"
    settings = _parse_ra_kv_file(retroarch_cfg)
    system_dir = _expand_path(
        args.system_dir or os.environ.get("RETROARCH_SYSTEM_DIR") or settings.get("system_directory"),
        config_dir / "system",
    )
    rom_root = _expand_path(
        args.roms_dir or os.environ.get("RETROARCH_ROMS_DIR"),
        Path.home() / "Games" / "RetroArch" / "roms",
    )
    playlist_root = _expand_path(
        args.playlist_dir or os.environ.get("RETROARCH_PLAYLIST_ROOT") or settings.get("playlist_directory"),
        config_dir / "playlists",
    )
    state_home = _expand_path(
        args.state_home or os.environ.get("RETROARCH_STATE_HOME"),
        Path.home() / ".local" / "state",
    )

    playlist_map = retroarch_profile.load_playlist_map(args.retroarch_profile)
    requirement_catalog = retroarch_profile.load_requirement_catalog(args.retroarch_profile)
    requirement_rows = _requirement_rows(requirement_catalog, system_dir, config_dir)

    skip_aliases = {"dc", "gc"}
    libretro_directory = settings.get("libretro_directory")
    core_rows: list[dict[str, Any]] = []
    for system, entry in sorted(playlist_map.items()):
        if system in skip_aliases:
            continue
        candidates = _core_candidates(str(entry["core_filename"]), libretro_directory, config_dir)
        installed_path = _first_existing(candidates)
        package_status = _package_status(entry.get("package_name"))
        rom_count = _count_roms(rom_root / system, set(entry["extensions"]))
        playlist_path = playlist_root / str(entry["playlist"])
        core_rows.append(
            {
                "system": system,
                "playlist": entry["playlist"],
                "playlist_path": str(playlist_path),
                "playlist_items": _playlist_item_count(playlist_path),
                "core_filename": entry["core_filename"],
                "core_name": entry["core_name"],
                "core_installed": installed_path is not None,
                "core_path": str(installed_path) if installed_path else str(candidates[0]),
                "candidate_paths": [str(candidate) for candidate in candidates],
                "package": package_status,
                "rom_dir": str(rom_root / system),
                "rom_count": rom_count,
            }
        )

    runtime = _runtime_status(settings)
    display = _display_profile(settings)

    core_missing = [row for row in core_rows if not row["core_installed"]]
    required_missing = [
        row for row in requirement_rows if row["requirement"] == "required" and row["status"] != "ok"
    ]
    warnings = [row for row in requirement_rows if row["status"] == "mismatch"]
    optional_missing = [row for row in requirement_rows if row["status"] == "optional_missing"]

    next_steps: list[str] = []
    if core_missing:
        systems = ", ".join(sorted(row["system"] for row in core_missing))
        next_steps.append(f"Install or link missing RetroArch cores for: {systems}.")
    if required_missing:
        systems = ", ".join(sorted({row["system"] for row in required_missing}))
        next_steps.append(f"Populate required BIOS/helper assets under RetroArch system dir for: {systems}.")
    if display["aspect_class"] == "32:9":
        next_steps.append("Keep safe 16:9 defaults on the 5120x1440 display; reserve true 32:9 for per-title validation.")
    if runtime["process_running"] and runtime["network_cmd_enable"] and not runtime["version_probe"]:
        next_steps.append("RetroArch network commands are enabled but not responding; check the running instance or UDP port binding.")
    if not runtime["network_cmd_enable"]:
        next_steps.append("Enable RetroArch network commands if you want post-sync runtime notifications without leaving the game.")
    if optional_missing:
        next_steps.append("Optional BIOS or helper assets are still missing for some systems; fill them when you need those cores.")

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "paths": {
            "config_dir": str(config_dir),
            "retroarch_cfg": str(retroarch_cfg),
            "system_dir": str(system_dir),
            "rom_root": str(rom_root),
            "playlist_root": str(playlist_root),
            "output_path": str(
                _expand_path(
                    args.output or os.environ.get("RETROARCH_WORKSTATION_AUDIT_PATH"),
                    state_home / "retroarch" / "workstation-audit.json",
                )
            ),
            "retroarch_profile": retroarch_profile.load_retroarch_profile(args.retroarch_profile)["profile_path"],
        },
        "display": display,
        "runtime": runtime,
        "summary": {
            "core_total": len(core_rows),
            "core_missing": len(core_missing),
            "required_assets_missing": len(required_missing),
            "optional_assets_missing": len(optional_missing),
            "asset_hash_mismatches": len(warnings),
        },
        "cores": core_rows,
        "requirements": requirement_rows,
        "next_steps": next_steps,
    }
    return report


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Audit RetroArch workstation cores, BIOS/helper assets, and runtime control.")
    parser.add_argument("--config-dir", help="RetroArch config directory.")
    parser.add_argument("--system-dir", help="RetroArch system/BIOS directory.")
    parser.add_argument("--roms-dir", help="RetroArch ROM root.")
    parser.add_argument("--playlist-dir", help="RetroArch playlist directory.")
    parser.add_argument("--state-home", help="State directory used for default output.")
    parser.add_argument("--retroarch-profile", help="Optional path to the shared RomHub RetroArch profile YAML.")
    parser.add_argument("--output", help="Path for the JSON audit output.")
    args = parser.parse_args(argv)

    report = build_report(args)
    output_path = Path(report["paths"]["output_path"])
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(report, indent=2) + "\n")
    print(output_path)
    print(
        "cores_missing={core_missing} required_assets_missing={required_missing} runtime_network={runtime_network}".format(
            core_missing=report["summary"]["core_missing"],
            required_missing=report["summary"]["required_assets_missing"],
            runtime_network="on" if report["runtime"]["network_cmd_enable"] else "off",
        )
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
