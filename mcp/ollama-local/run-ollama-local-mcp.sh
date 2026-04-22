#!/usr/bin/env bash
# Launcher for ollama-local MCP server — routes coding tasks to local models.
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
exec "$DIR/.venv/bin/python" "$DIR/server.py"
