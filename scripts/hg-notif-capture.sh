#!/usr/bin/env bash
# hg-notif-capture.sh — Capture and analyze desktop notifications via D-Bus
# Usage: hg-notif-capture.sh [duration_minutes]  (default: 60)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/hg-core.sh
source "$SCRIPT_DIR/lib/hg-core.sh"

DURATION="${1:-60}"
LOGFILE="/tmp/notif-capture-$(date +%Y%m%d-%H%M%S).log"

hg_info "Capturing notifications for ${DURATION} minute(s)..."
hg_info "Log: $LOGFILE"
hg_info "Press Ctrl+C to stop early and see analysis."

# Trap to show analysis on exit (normal or Ctrl+C)
cleanup() {
    if [[ -s "$LOGFILE" ]]; then
        echo ""
        hg_info "═══ NOTIFICATION ANALYSIS ═══"
        echo ""

        local total
        total=$(wc -l < "$LOGFILE")
        hg_info "Total notifications captured: $total"
        echo ""

        hg_info "── By App ──"
        awk -F'|' '{gsub(/^ *app=| *$/, "", $2); print $2}' "$LOGFILE" | sort | uniq -c | sort -rn
        echo ""

        hg_info "── By Urgency ──"
        awk -F'|' '{gsub(/^ *urgency=| *$/, "", $4); print $4}' "$LOGFILE" | sort | uniq -c | sort -rn
        echo ""

        hg_info "── By Summary (top 20) ──"
        awk -F'|' '{gsub(/^ *summary=| *$/, "", $3); print $3}' "$LOGFILE" | sort | uniq -c | sort -rn | head -20
        echo ""

        hg_info "── Unique App+Summary Pairs ──"
        awk -F'|' '{gsub(/^ */, "", $2); gsub(/^ */, "", $3); print $2 " :: " $3}' "$LOGFILE" | sort | uniq -c | sort -rn | head -30
        echo ""

        hg_ok "Full log: $LOGFILE"
    else
        hg_info "No notifications captured."
    fi
}
trap cleanup EXIT

# Parse D-Bus monitor output for Notify method calls.
# D-Bus Notify signature: STRING app_name, UINT32 replaces_id, STRING app_icon,
#                         STRING summary, STRING body, ARRAY actions,
#                         DICT hints, INT32 expire_timeout
#
# We use gdbus monitor which gives cleaner structured output than dbus-monitor.
# Fall back to dbus-monitor if gdbus isn't available.

if ! command -v dbus-monitor &>/dev/null; then
    hg_error "dbus-monitor not found. Install dbus."
    exit 1
fi

# Initialize state before the parse loop
str_count=0
app="" summary="" body="" urgency="normal"
in_urgency=false

timeout "${DURATION}m" dbus-monitor --session \
    "type='method_call',interface='org.freedesktop.Notifications',member='Notify'" \
    2>/dev/null | while IFS= read -r line; do

    # D-Bus Notify args (string positions after skipping uint32 replaces_id):
    #   1=app_name, 2=app_icon, 3=summary, 4=body
    # Urgency is in hints dict: string "urgency" → variant byte N
    if [[ "$line" == *"method call"* ]]; then
        str_count=0
        app="" summary="" body="" urgency="normal" in_urgency=false
        continue
    fi

    if [[ "$line" =~ ^[[:space:]]*string[[:space:]]+\"(.*)\" ]]; then
        local_str="${BASH_REMATCH[1]}"
        # Track if this string is the "urgency" dict key
        if [[ "$local_str" == "urgency" ]]; then
            in_urgency=true
            continue
        fi
        in_urgency=false
        str_count=$(( str_count + 1 ))
        case $str_count in
            1) app="$local_str" ;;
            3) summary="$local_str" ;;
            4) body="$local_str" ;;
        esac
    fi

    # Urgency byte value follows the "urgency" dict key
    if [[ "$in_urgency" == true && "$line" =~ byte[[:space:]]+([0-2]) ]]; then
        case "${BASH_REMATCH[1]}" in
            0) urgency="low" ;;
            1) urgency="normal" ;;
            2) urgency="critical" ;;
        esac
        in_urgency=false
    fi

    # Emit log line on int32 (expire_timeout — last arg in Notify call)
    if [[ "$line" =~ ^[[:space:]]*int32 && -n "${app:-}" && -n "${summary:-}" ]]; then
        printf '%s | app=%s | summary=%s | urgency=%s | body=%s\n' \
            "$(date '+%Y-%m-%d %H:%M:%S')" "$app" "$summary" "$urgency" "${body:0:200}" \
            >> "$LOGFILE"
        app="" summary="" body="" urgency="normal" str_count=0 in_urgency=false
    fi
done || true
