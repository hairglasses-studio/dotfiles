#!/usr/bin/env python3
"""Send UDP commands to a running RetroArch instance and print the response.

Thin wrapper over scripts/lib/retroarch_runtime.send_udp_command — adds
argument parsing, a --list verb that prints the known command taxonomy,
and convenience shortcuts like --osd that wrap SHOW_MSG with quoting.

Requires `network_cmd_enable = "true"` in retroarch.cfg and a running
RetroArch binding UDP on port 55355 (default). Use `retroarch-apply-
network-cmd` to flip the cfg; restart RetroArch to bind the socket.
"""

from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path


sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

import retroarch_runtime


# Verified 2026-04-22 against RetroArch 1.22.2 and
# https://docs.libretro.com/development/retroarch/network-control-interface/
KNOWN_COMMANDS: list[tuple[str, str, str]] = [
    # (name, args, description)
    ("VERSION", "", "Return the RetroArch version string. Expects a response."),
    ("GET_STATUS", "", "Return emulation status (paused/playing/contentless)."),
    ("GET_CONFIG_PARAM", "<param>", "Query a config value by name. Expects a response."),
    ("SHOW_MSG", "<text>", "Display an on-screen OSD message."),
    ("SET_SHADER", "<path>", "Load a shader preset by filesystem path."),
    ("LOAD_CORE", "<path>", "Load a libretro core by filesystem path."),
    ("LOAD_STATE_SLOT", "<slot>", "Load save state from the given numeric slot."),
    ("PLAY_REPLAY_SLOT", "<slot>", "Start replay playback from slot."),
    ("SEEK_REPLAY", "<frame>", "Seek to a specific frame in the active replay."),
    ("READ_CORE_RAM", "<addr> <bytes>", "Read core RAM via achievement addresses."),
    ("WRITE_CORE_RAM", "<addr> <bytes>", "Write core RAM via achievement addresses."),
    ("READ_CORE_MEMORY", "<addr> <bytes>", "Read from the system memory map."),
    ("WRITE_CORE_MEMORY", "<addr> <bytes>", "Write to the system memory map."),
    ("SAVE_FILES", "", "Flush all SRAM to disk."),
    ("LOAD_FILES", "", "Reload all SRAM from disk."),
    ("QUIT", "", "Close RetroArch."),
    ("CLOSE_CONTENT", "", "Unload the active content without exiting."),
    ("RESET", "", "Reset the running core."),
    ("PAUSE_TOGGLE", "", "Toggle pause."),
    ("FRAMEADVANCE", "", "Advance one frame while paused."),
    ("FAST_FORWARD", "", "Toggle fast-forward."),
    ("FAST_FORWARD_HOLD", "", "Hold fast-forward while the command is held."),
    ("SLOWMOTION", "", "Toggle slow motion."),
    ("REWIND", "", "Trigger rewind (if rewind is enabled)."),
    ("MUTE", "", "Toggle audio mute."),
    ("VOLUME_UP", "", "Nudge volume up by one step."),
    ("VOLUME_DOWN", "", "Nudge volume down by one step."),
    ("MENU_TOGGLE", "", "Toggle the RetroArch menu."),
    ("LOAD_STATE", "", "Load state from the current slot."),
    ("SAVE_STATE", "", "Save state to the current slot."),
    ("STATE_SLOT_PLUS", "", "Advance the active state slot."),
    ("STATE_SLOT_MINUS", "", "Rewind the active state slot."),
    ("SCREENSHOT", "", "Take a screenshot."),
    ("RECORDING_TOGGLE", "", "Toggle video recording."),
    ("STREAMING_TOGGLE", "", "Toggle live streaming."),
    ("SHADER_TOGGLE", "", "Toggle shader effect on/off."),
    ("CHEAT_TOGGLE", "", "Toggle cheat activation."),
    ("DISK_EJECT_TOGGLE", "", "Toggle disk tray."),
]


# Commands that wait for a response when sent.
RESPONSIVE_COMMANDS = {"VERSION", "GET_STATUS", "GET_CONFIG_PARAM"}


def _print_table() -> None:
    name_w = max(len(name) for name, _, _ in KNOWN_COMMANDS) + 2
    args_w = max(len(args) for _, args, _ in KNOWN_COMMANDS) + 2
    for name, args, desc in KNOWN_COMMANDS:
        print(f"{name.ljust(name_w)}{args.ljust(args_w)}{desc}")


def _parse_args(argv: list[str]) -> argparse.Namespace:
    p = argparse.ArgumentParser(
        description="Send a UDP network command to a running RetroArch instance.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=(
            "Examples:\n"
            "  retroarch-command --list\n"
            "  retroarch-command VERSION\n"
            "  retroarch-command --osd 'sync complete'\n"
            "  retroarch-command SET_SHADER /path/to/preset.glslp\n"
        ),
    )
    p.add_argument("command", nargs="?", help="Command name (see --list). Omit with --osd.")
    p.add_argument("args", nargs="*", help="Arguments appended to the command, space-joined.")
    p.add_argument("--list", action="store_true", help="Print the known command taxonomy and exit.")
    p.add_argument("--osd", metavar="TEXT", help="Shortcut for SHOW_MSG with auto-quoting.")
    p.add_argument("--host", default="127.0.0.1", help="Target host (default: 127.0.0.1).")
    p.add_argument("--port", type=int, default=55355, help="Target UDP port (default: 55355).")
    p.add_argument("--timeout", type=float, default=1.0, help="Response timeout in seconds (default: 1.0).")
    p.add_argument("--expect-response", action="store_true",
                   help="Force reading a response even for commands that don't normally return one.")
    p.add_argument("--json", action="store_true", help="Emit the result dict as JSON instead of plain text.")
    return p.parse_args(argv)


def main(argv: list[str]) -> int:
    args = _parse_args(argv)

    if args.list:
        _print_table()
        return 0

    if args.osd is not None:
        command_line = f'SHOW_MSG "{args.osd}"'
    elif args.command:
        tail = " ".join(args.args) if args.args else ""
        command_line = f"{args.command} {tail}".strip()
    else:
        print("error: command name required (use --list to see options)", file=sys.stderr)
        return 2

    verb = command_line.split(" ", 1)[0].upper()
    expect_response = args.expect_response or verb in RESPONSIVE_COMMANDS

    result = retroarch_runtime.send_udp_command(
        command_line,
        host=args.host,
        port=args.port,
        expect_response=expect_response,
        timeout_seconds=args.timeout,
    )

    if args.json:
        print(json.dumps(result, indent=2))
    else:
        if result["ok"]:
            if result["response"] is not None:
                print(result["response"])
        else:
            print(f"error: {result['error']}", file=sys.stderr)

    return 0 if result["ok"] else 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
