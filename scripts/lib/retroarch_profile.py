#!/usr/bin/env python3
"""Shared RetroArch profile helpers for local workstation automation."""

from __future__ import annotations

import os
from pathlib import Path


DEFAULT_ROMHUB_PROFILE = (
    Path.home() / "hairglasses-studio" / "romhub" / "configs" / "device-profiles" / "retroarch.yaml"
)

DEFAULT_PLAYLIST_MAP_ROWS = [
    "gb|Nintendo - Game Boy.lpl|sameboy_libretro.so|Nintendo - Game Boy / Color (SameBoy)|.gb|libretro-sameboy",
    "gba|Nintendo - Game Boy Advance.lpl|mgba_libretro.so|Nintendo - Game Boy Advance (mGBA)|.gba|libretro-mgba",
    "gbc|Nintendo - Game Boy Color.lpl|sameboy_libretro.so|Nintendo - Game Boy / Color (SameBoy)|.gbc,.gb|libretro-sameboy",
    "genesis|Sega - Mega Drive - Genesis.lpl|genesis_plus_gx_libretro.so|Sega - MS/GG/MD/CD (Genesis Plus GX)|.bin,.gen,.md|libretro-genesis-plus-gx",
    "nes|Nintendo - Nintendo Entertainment System.lpl|mesen_libretro.so|Nintendo - NES / Famicom (Mesen)|.nes|libretro-mesen",
    "ngp|SNK - Neo Geo Pocket Color.lpl|race_libretro.so|SNK - Neo Geo Pocket / Color (RACE)|.ngc,.ngp|",
    "pce|NEC - PC Engine - TurboGrafx 16.lpl|mednafen_pce_fast_libretro.so|NEC - PC Engine / CD (Beetle PCE FAST)|.pce|libretro-beetle-pce-fast",
    "snes|Nintendo - Super Nintendo Entertainment System.lpl|bsnes_hd_beta_libretro.so|Nintendo - SNES (bsnes-hd beta)|.sfc,.smc,.fig|libretro-bsnes-hd",
    "wonderswan|Bandai - WonderSwan.lpl|mednafen_wswan_libretro.so|Bandai - WonderSwan / Color (Beetle Cygne)|.ws,.wsc|",
    "psx|Sony - PlayStation.lpl|mednafen_psx_hw_libretro.so|Sony - PlayStation (Beetle PSX HW)|.cue,.ccd,.chd,.m3u,.pbp|libretro-beetle-psx-hw",
    "ps2|Sony - PlayStation 2.lpl|pcsx2_libretro.so|Sony - PlayStation 2 (LRPS2)|.chd,.cso,.cue,.iso|",
    "dreamcast|Sega - Dreamcast.lpl|flycast_libretro.so|Sega - Dreamcast (Flycast)|.cdi,.chd,.cue,.gdi,.m3u|libretro-flycast",
    "dc|Sega - Dreamcast.lpl|flycast_libretro.so|Sega - Dreamcast (Flycast)|.cdi,.chd,.cue,.gdi,.m3u|libretro-flycast",
    "gamecube|Nintendo - GameCube.lpl|dolphin_libretro.so|Nintendo - GameCube / Wii (Dolphin)|.ciso,.dol,.elf,.gcm,.gcz,.iso,.rvz|libretro-dolphin",
    "gc|Nintendo - GameCube.lpl|dolphin_libretro.so|Nintendo - GameCube / Wii (Dolphin)|.ciso,.dol,.elf,.gcm,.gcz,.iso,.rvz|libretro-dolphin",
    "wii|Nintendo - Wii.lpl|dolphin_libretro.so|Nintendo - GameCube / Wii (Dolphin)|.iso,.rvz,.wad,.wbfs|libretro-dolphin",
    "nds|Nintendo - Nintendo DS.lpl|desmume_libretro.so|Nintendo - DS (DeSmuME)|.bin,.nds|libretro-desmume",
    "psp|Sony - PlayStation Portable.lpl|ppsspp_libretro.so|Sony - PlayStation Portable (PPSSPP)|.chd,.cso,.elf,.iso,.pbp,.prx|libretro-ppsspp",
    "vb|Nintendo - Virtual Boy.lpl|beetle_vb_libretro.so|Nintendo - Virtual Boy (Beetle VB)|.bin,.vb,.vboy|",
]

DEFAULT_REQUIREMENT_ROWS = [
    "psx|file|required|scph5501.bin|490f666e1afb15b7362b406ed1cea246|Beetle PSX HW US BIOS",
    "psx|file|optional|scph5500.bin|8dd7d5296a650fac7319bce665a6a53c|Beetle PSX HW JP BIOS",
    "psx|file|optional|scph5502.bin|32736f17079d0b2b7024407c39bd3050|Beetle PSX HW EU BIOS",
    "psx|file|optional|PSXONPSP660.bin|c53ca5908936d412331790f4426c6c33|Region-free PSP BIOS override",
    "psx|file|optional|ps1_rom.bin|81bbe60ba7a3d1cea1d48c14cbcc647b|Region-free PS3 BIOS override",
    "ps2|dir|required|pcsx2/bios||LRPS2 BIOS directory",
    "ps2|file|required|pcsx2/resources/GameIndex.yaml||LRPS2 compatibility database",
    "dreamcast|dir|required|dc||Flycast system subdirectory",
    "dreamcast|file|optional|dc/dc_boot.bin|e10c53c2f8b90bab96ead2d368858623|Dreamcast boot ROM",
    "dreamcast|file|optional|dc/dc_flash.bin||Dreamcast flash data",
    "gamecube|dir_nonempty|required|dolphin-emu/Sys||Dolphin Sys compatibility assets",
    "wii|dir_nonempty|required|dolphin-emu/Sys||Dolphin Sys compatibility assets",
    "nds|file|conditional|bios7.bin|df692a80a5b1bc90728bc3dfc76cd948|DeSmuME ARM7 BIOS when external firmware is enabled",
    "nds|file|conditional|bios9.bin|a392174eb3e572fed6447e956bde4b25|DeSmuME ARM9 BIOS when external firmware is enabled",
    "nds|file|conditional|firmware.bin||DeSmuME firmware when external firmware is enabled",
    "psp|dir_nonempty|required|PPSSPP||PPSSPP helper asset directory",
]


def _parse_simple_yaml(path: Path) -> dict[str, object]:
    data: dict[str, object] = {}
    if not path.exists():
        return data

    current_array: str | None = None
    array_items: list[str] = []

    for raw_line in path.read_text(errors="ignore").splitlines():
        if raw_line.startswith("#"):
            continue
        line = raw_line.rstrip()
        if not line.strip():
            continue
        if line.lstrip().startswith("- "):
            if current_array is None:
                continue
            item = line.lstrip()[2:].strip()
            if item.startswith(("'", '"')) and item.endswith(("'", '"')) and len(item) >= 2:
                item = item[1:-1]
            array_items.append(item)
            continue

        if current_array is not None:
            data[current_array] = list(array_items)
            current_array = None
            array_items = []

        if ":" not in line:
            continue
        key, raw_value = line.split(":", 1)
        key = key.strip()
        value = raw_value.strip()
        if not value:
            current_array = key
            array_items = []
            continue
        if value.startswith("[") and value.endswith("]"):
            items = []
            inner = value[1:-1].strip()
            if inner:
                for part in inner.split(","):
                    cleaned = part.strip()
                    if cleaned.startswith(("'", '"')) and cleaned.endswith(("'", '"')) and len(cleaned) >= 2:
                        cleaned = cleaned[1:-1]
                    items.append(cleaned)
            data[key] = items
            continue
        if value.startswith(("'", '"')) and value.endswith(("'", '"')) and len(value) >= 2:
            value = value[1:-1]
        data[key] = value

    if current_array is not None:
        data[current_array] = list(array_items)

    return data


def _profile_path(explicit: str | None = None) -> Path | None:
    candidates = []
    if explicit:
        candidates.append(Path(os.path.expandvars(os.path.expanduser(explicit))))
    env_value = os.environ.get("ROMHUB_RETROARCH_PROFILE")
    if env_value:
        candidates.append(Path(os.path.expandvars(os.path.expanduser(env_value))))
    candidates.append(DEFAULT_ROMHUB_PROFILE)

    for candidate in candidates:
        if candidate.exists():
            return candidate
    return None


def _decode_playlist_row(row: str) -> dict[str, object]:
    system, playlist, core_filename, core_name, extensions_csv, package_name = (row.split("|", 5) + [""])[:6]
    return {
        "system": system,
        "playlist": playlist,
        "core_filename": core_filename,
        "core_name": core_name,
        "extensions": {ext.strip().lower() for ext in extensions_csv.split(",") if ext.strip()},
        "package_name": package_name or None,
    }


def _decode_requirement_row(row: str) -> dict[str, object]:
    system, kind, requirement, relative_path, md5sum, description = (row.split("|", 5) + [""])[:6]
    return {
        "system": system,
        "kind": kind,
        "requirement": requirement,
        "relative_path": relative_path,
        "md5": md5sum or None,
        "description": description,
    }


def load_retroarch_profile(profile_path: str | None = None) -> dict[str, object]:
    path = _profile_path(profile_path)
    data = _parse_simple_yaml(path) if path else {}
    return {
        "profile_path": str(path) if path else None,
        "data": data,
    }


def load_playlist_map(profile_path: str | None = None) -> dict[str, dict[str, object]]:
    profile = load_retroarch_profile(profile_path)
    rows = profile["data"].get("retroarch_playlist_map")
    source_rows = rows if isinstance(rows, list) and rows else DEFAULT_PLAYLIST_MAP_ROWS
    decoded = [_decode_playlist_row(str(row)) for row in source_rows]
    return {str(entry["system"]): entry for entry in decoded}


def load_requirement_catalog(profile_path: str | None = None) -> list[dict[str, object]]:
    profile = load_retroarch_profile(profile_path)
    rows = profile["data"].get("retroarch_requirements")
    source_rows = rows if isinstance(rows, list) and rows else DEFAULT_REQUIREMENT_ROWS
    return [_decode_requirement_row(str(row)) for row in source_rows]
