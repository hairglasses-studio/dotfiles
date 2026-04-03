#!/usr/bin/env bash
# gcalcli-auth.sh — Authenticate gcalcli using credentials from org .env
set -euo pipefail

ENV_FILE="$HOME/hairglasses-studio/.env"
source "$ENV_FILE"

CLIENT_ID="${GOOGLE_CALENDAR_CLIENT_ID:?GOOGLE_CALENDAR_CLIENT_ID not set in $ENV_FILE}"
CLIENT_SECRET="${GOOGLE_CALENDAR_CLIENT_SECRET:?GOOGLE_CALENDAR_CLIENT_SECRET not set in $ENV_FILE}"

echo "Authenticating gcalcli with client $CLIENT_ID ..."
exec gcalcli --client-id "$CLIENT_ID" --client-secret "$CLIENT_SECRET" list
