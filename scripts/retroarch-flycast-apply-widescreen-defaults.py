#!/usr/bin/env python3
"""Apply safe RetroArch Flycast widescreen defaults."""

from __future__ import annotations

import os
import sys


sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

import retroarch_widescreen_audit


if __name__ == "__main__":
    raise SystemExit(retroarch_widescreen_audit.apply_main("flycast", sys.argv[1:]))
