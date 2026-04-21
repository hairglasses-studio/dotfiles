#!/usr/bin/env python3
"""Minimal RetroArch runtime helpers for network command integration."""

from __future__ import annotations

import socket


def send_udp_command(
    command: str,
    host: str = "127.0.0.1",
    port: int = 55355,
    expect_response: bool = False,
    timeout_seconds: float = 1.0,
) -> dict[str, object]:
    payload = command.encode("utf-8")
    response_text = None
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as sock:
            sock.settimeout(timeout_seconds)
            sock.sendto(payload, (host, port))
            if expect_response:
                response, _ = sock.recvfrom(4096)
                response_text = response.decode("utf-8", errors="ignore").strip()
    except OSError as exc:
        return {
            "ok": False,
            "command": command,
            "host": host,
            "port": port,
            "error": str(exc),
            "response": None,
        }

    return {
        "ok": True,
        "command": command,
        "host": host,
        "port": port,
        "error": None,
        "response": response_text,
    }
