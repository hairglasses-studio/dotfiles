#!/bin/bash
# Network throughput — samples bytes over 1s interval on active interface
INTERFACE=$(route -n get default 2>/dev/null | awk '/interface:/{print $2}')
[ -z "$INTERFACE" ] && INTERFACE="en0"

# Sample bytes at two points 1s apart
read RX1 TX1 <<< $(netstat -ibn | awk -v iface="$INTERFACE" '$1 == iface && $3 ~ /\./ {print $7, $10; exit}')
sleep 1
read RX2 TX2 <<< $(netstat -ibn | awk -v iface="$INTERFACE" '$1 == iface && $3 ~ /\./ {print $7, $10; exit}')

# Calculate bytes/sec
DL=$((RX2 - RX1))
UL=$((TX2 - TX1))

# Human-readable format
fmt() {
    local b=$1
    if [ "$b" -gt 1048576 ]; then
        echo "$(( b / 1048576 ))M"
    elif [ "$b" -gt 1024 ]; then
        echo "$(( b / 1024 ))K"
    else
        echo "${b}B"
    fi
}

sketchybar --set $NAME label="$(fmt $DL) $(fmt $UL)"
