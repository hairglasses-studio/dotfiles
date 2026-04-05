#!/bin/bash
# Setlist Manager - Create and manage DJ setlists
# Usage: setlist.sh [command] [args]

SETLIST_DIR="${SETLIST_DIR:-$HOME/.setlists}"
mkdir -p "$SETLIST_DIR"

case "${1:-help}" in
    create)
        name="${2:-$(date +%Y-%m-%d)-set}"
        file="$SETLIST_DIR/$name.txt"
        echo "# Setlist: $name" > "$file"
        echo "# Created: $(date)" >> "$file"
        echo "" >> "$file"
        echo "Created setlist: $file"
        ;;
    add)
        name="${2:-current}"
        shift 2
        file="$SETLIST_DIR/$name.txt"
        [ ! -f "$file" ] && { echo "Setlist not found: $name"; exit 1; }
        for track in "$@"; do
            echo "$track" >> "$file"
            echo "Added: $track"
        done
        ;;
    show)
        name="${2:-current}"
        file="$SETLIST_DIR/$name.txt"
        [ -f "$file" ] && cat "$file" || echo "Setlist not found: $name"
        ;;
    list)
        ls -1 "$SETLIST_DIR"/*.txt 2>/dev/null | xargs -I{} basename {} .txt || echo "No setlists"
        ;;
    delete)
        name="$2"
        [ -z "$name" ] && { echo "Specify setlist name"; exit 1; }
        rm -i "$SETLIST_DIR/$name.txt"
        ;;
    *)
        echo "Setlist Manager"
        echo "Usage: setlist.sh <command> [args]"
        echo ""
        echo "Commands:"
        echo "  create [name]     Create new setlist"
        echo "  add <name> <tracks...>  Add tracks to setlist"
        echo "  show [name]       Show setlist contents"
        echo "  list              List all setlists"
        echo "  delete <name>     Delete a setlist"
        ;;
esac
