#!/usr/bin/env python3
"""Audit RetroArch Dolphin widescreen readiness and emit a JSON report."""

from __future__ import annotations

import os
import sys


sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

import retroarch_widescreen_audit


if __name__ == "__main__":
    raise SystemExit(retroarch_widescreen_audit.main("dolphin", sys.argv[1:]))
