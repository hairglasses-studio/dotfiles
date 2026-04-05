#!/bin/bash
# Session Manager - Track studio sessions
# Usage: session.sh [command] [args]

SESSION_DIR="${SESSION_DIR:-$HOME/.luke-toolkit/sessions}"
CURRENT="$SESSION_DIR/.current"
mkdir -p "$SESSION_DIR"

case "${1:-help}" in
    start)
        name="${2:-$(date +%Y-%m-%d)-session}"
        file="$SESSION_DIR/$name.log"
        echo "$file" > "$CURRENT"
        echo "=== Session: $name ===" > "$file"
        echo "Started: $(date)" >> "$file"
        echo "" >> "$file"
        echo "Session started: $name"
        ;;
    log)
        [ ! -f "$CURRENT" ] && { echo "No active session. Run: session.sh start"; exit 1; }
        file=$(cat "$CURRENT")
        shift
        echo "[$(date +%H:%M:%S)] $*" >> "$file"
        echo "Logged: $*"
        ;;
    end)
        [ ! -f "$CURRENT" ] && { echo "No active session"; exit 1; }
        file=$(cat "$CURRENT")
        echo "" >> "$file"
        echo "Ended: $(date)" >> "$file"
        echo "Session saved: $file"
        rm "$CURRENT"
        ;;
    show)
        [ -f "$CURRENT" ] && cat "$(cat "$CURRENT")" || echo "No active session"
        ;;
    list)
        ls -1 "$SESSION_DIR"/*.log 2>/dev/null | xargs -I{} basename {} .log || echo "No sessions"
        ;;
    *)
        echo "Session Manager"
        echo "Usage: session.sh <command> [args]"
        echo ""
        echo "Commands:"
        echo "  start [name]     Start new session"
        echo "  log <message>    Log an event"
        echo "  end              End current session"
        echo "  show             Show current session"
        echo "  list             List all sessions"
        ;;
esac
