#!/usr/bin/env bash
# screenshot-crop.sh — Crop-select screenshot (delegates to hg-screenshot.sh)
exec "$(dirname "$0")/hg-screenshot.sh" region --both
