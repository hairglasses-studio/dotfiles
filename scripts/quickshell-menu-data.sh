#!/usr/bin/env bash
# quickshell-menu-data.sh - JSON data and actions for Quickshell menu overlays.

set -euo pipefail

action="${1:-list}"
mode="${2:-apps}"
item_id="${3:-}"

python3 - "$action" "$mode" "$item_id" <<'PY'
from __future__ import annotations

import configparser
import hashlib
import json
import os
import pathlib
import re
import subprocess
import sys
from typing import Any

action, mode, item_id = sys.argv[1:4]
home = pathlib.Path.home()


def run(argv: list[str]) -> None:
    subprocess.Popen(argv, start_new_session=True)


def clean(value: Any, limit: int = 220) -> str:
    text = re.sub(r"\s+", " ", str(value or "")).strip()
    return text[: limit - 1] + "..." if len(text) > limit else text


def entry(
    ident: str,
    title: str,
    subtitle: str = "",
    badge: str = "",
    search: str = "",
    danger: bool = False,
    payload: str = "",
) -> dict[str, Any]:
    return {
        "id": ident,
        "title": clean(title, 140),
        "subtitle": clean(subtitle, 260),
        "badge": clean(badge, 24),
        "search": clean(search, 600),
        "danger": danger,
        "payload": payload,
    }


def emit(entries: list[dict[str, Any]]) -> None:
    print(json.dumps({"entries": entries}, ensure_ascii=False, separators=(",", ":")))


def desktop_files() -> list[pathlib.Path]:
    roots: list[pathlib.Path] = []
    data_home = pathlib.Path(os.environ.get("XDG_DATA_HOME", home / ".local/share"))
    roots.append(data_home / "applications")
    for root in os.environ.get("XDG_DATA_DIRS", "/usr/local/share:/usr/share").split(":"):
        if root:
            roots.append(pathlib.Path(root) / "applications")

    seen: set[str] = set()
    files: list[pathlib.Path] = []
    for root in roots:
        if not root.is_dir():
            continue
        for path in root.rglob("*.desktop"):
            key = path.name
            if key in seen:
                continue
            seen.add(key)
            files.append(path)
    return files


def apps() -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    for path in desktop_files():
        parser = configparser.ConfigParser(interpolation=None, strict=False)
        try:
            parser.read(path, encoding="utf-8")
            section = parser["Desktop Entry"]
        except Exception:
            continue

        if section.get("Type", "Application") != "Application":
            continue
        if section.getboolean("NoDisplay", fallback=False) or section.getboolean("Hidden", fallback=False):
            continue
        name = section.get("Name", path.stem)
        generic = section.get("GenericName", "")
        comment = section.get("Comment", "")
        categories = section.get("Categories", "").replace(";", " ")
        keywords = section.get("Keywords", "").replace(";", " ")
        badge = (categories.split() or ["APP"])[0].upper()[:10]
        rows.append(entry(str(path), name, generic or comment, badge, f"{name} {generic} {comment} {categories} {keywords}"))
    rows.sort(key=lambda e: e["title"].lower())
    return rows


def clients(agent_only: bool = False) -> list[dict[str, Any]]:
    try:
        raw = subprocess.check_output(["hyprctl", "clients", "-j"], text=True, stderr=subprocess.DEVNULL)
        data = json.loads(raw)
    except Exception:
        return []
    rows: list[dict[str, Any]] = []
    for client in data:
        address = client.get("address") or ""
        title = client.get("title") or ""
        klass = client.get("class") or "app"
        workspace = client.get("workspace") or {}
        if not address:
            continue
        is_agent = klass == "kitty" and (title.startswith("────") or "claude" in title.lower() or "codex" in title.lower())
        if agent_only and not is_agent:
            continue
        subtitle = f"{klass} on workspace {workspace.get('name') or workspace.get('id') or '?'}"
        badge = str(workspace.get("name") or workspace.get("id") or "WIN").upper()[:10]
        rows.append(entry(address, title or klass, subtitle, badge, f"{title} {klass} {subtitle}"))
    rows.sort(key=lambda e: e["title"].lower())
    return rows


def emoji_rows() -> list[dict[str, Any]]:
    path = pathlib.Path("/usr/share/oh-my-zsh/plugins/emoji/gemoji_db.json")
    if not path.exists():
        return [entry("missing", "Emoji database missing", str(path), "ERR", danger=True)]
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except Exception as exc:
        return [entry("error", "Emoji database error", str(exc), "ERR", danger=True)]
    rows: list[dict[str, Any]] = []
    for item in data:
        emoji = item.get("emoji") or ""
        desc = item.get("description") or ""
        aliases = " ".join(item.get("aliases") or [])
        tags = " ".join(item.get("tags") or [])
        category = item.get("category") or "emoji"
        if emoji:
            rows.append(entry(emoji, f"{emoji}  {desc}", aliases or category, category.upper()[:10], f"{desc} {aliases} {tags} {category}"))
    return rows


def clipboard_rows() -> list[dict[str, Any]]:
    try:
        raw = subprocess.check_output(["clipse", "-output-all", "raw"], text=True, stderr=subprocess.DEVNULL, timeout=1.5)
    except Exception:
        return [entry("missing", "Clipboard history unavailable", "clipse -listen may not be running", "CLIP", danger=True)]
    rows: list[dict[str, Any]] = []
    seen: set[str] = set()
    for line in raw.splitlines():
        text = line.strip()
        if not text or text in seen:
            continue
        seen.add(text)
        ident = hashlib.sha256(text.encode("utf-8")).hexdigest()
        rows.append(entry(ident, text, f"{len(text)} chars", "CLIP", text, payload=text))
        if len(rows) >= 120:
            break
    return rows


def power_rows() -> list[dict[str, Any]]:
    return [
        entry("logout", "Log out of Hyprland", "hyprctl dispatch exit", "EXIT", "logout exit session", True),
        entry("lock", "Lock session", "hyprlock", "LOCK", "lock screen"),
        entry("reboot", "Reboot", "systemctl reboot", "REBOOT", "restart reboot", True),
        entry("poweroff", "Power off", "systemctl poweroff", "POWER", "shutdown poweroff", True),
    ]


def list_mode(name: str) -> list[dict[str, Any]]:
    if name == "apps":
        return apps()
    if name == "windows":
        return clients(False)
    if name == "agents":
        rows = clients(True)
        return rows or [entry("none", "No agent sessions found", "No matching kitty agent windows", "AGENT")]
    if name == "emoji":
        return emoji_rows()
    if name == "clipboard":
        return clipboard_rows()
    if name == "power":
        return power_rows()
    return [entry("unknown", f"Unknown menu: {name}", "", "ERR", danger=True)]


def find_row(name: str, ident: str) -> dict[str, Any] | None:
    for row in list_mode(name):
        if row["id"] == ident:
            return row
    return None


def exec_mode(name: str, ident: str) -> int:
    row = find_row(name, ident)
    if row is None or ident in {"none", "missing", "error", "unknown"}:
        return 0

    if name == "apps":
        run(["gio", "launch", ident])
    elif name in {"windows", "agents"}:
        run(["hyprctl", "dispatch", "focuswindow", f"address:{ident}"])
    elif name == "emoji":
        subprocess.run(["wl-copy"], input=ident, text=True, check=False)
        run(["wtype", "-s", "80", ident])
    elif name == "clipboard":
        text = row.get("payload") or row["title"]
        subprocess.run(["wl-copy"], input=text, text=True, check=False)
        run(["wtype", "-s", "80", "-M", "ctrl", "-k", "v", "-m", "ctrl"])
    elif name == "power":
        if ident == "logout":
            run(["hyprctl", "dispatch", "exit"])
        elif ident == "lock":
            run(["hyprlock"])
        elif ident == "reboot":
            run(["systemctl", "reboot"])
        elif ident == "poweroff":
            run(["systemctl", "poweroff"])
    return 0


if action == "list":
    emit(list_mode(mode))
elif action == "exec":
    raise SystemExit(exec_mode(mode, item_id))
else:
    print(f"unknown action: {action}", file=sys.stderr)
    raise SystemExit(2)
PY
