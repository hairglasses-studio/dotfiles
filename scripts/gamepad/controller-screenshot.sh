#!/usr/bin/env bash
# controller-screenshot.sh — Gamepad screenshot (delegates to hg-screenshot.sh)
exec "$(dirname "$0")/../hg-screenshot.sh" full --clipboard --notify
