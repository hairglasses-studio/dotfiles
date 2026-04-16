#!/usr/bin/env bash
# Don't use -e — some commands legitimately fail (e.g. dbus-uuidgen as non-root)
set -uo pipefail

# ── Sandbox entrypoint ────────────────────────────────────────────────────
# Starts Hyprland as a nested Wayland compositor inside the host's session.
# GPU-accelerated via nvidia-container-toolkit (shared RTX 3080).
# The host's dotfiles are bind-mounted at /dotfiles (read-only).

DOTFILES_SRC="/dotfiles"
READY_SIGNAL="/tmp/sandbox-ready"
HYPR_LOG="/tmp/hyprland.log"

# ── Phase 0: Fix container quirks ────────────────────────────────────────
dbus-uuidgen > /etc/machine-id 2>/dev/null || true

# ── Minimal profile: skip Hyprland entirely ──────────────────────────────
if [[ "${SANDBOX_SKIP_HYPRLAND:-0}" == "1" ]]; then
    echo "[sandbox] Minimal profile — skipping Hyprland"
    touch "$READY_SIGNAL"
    exec sleep infinity
fi

# ── Phase 1: Create config symlinks ──────────────────────────────────────
echo "[sandbox] Creating config symlinks..."
mkdir -p "$HOME/.config"
for dir in hypr ironbar swaync kitty; do
    if [[ -d "$DOTFILES_SRC/$dir" ]]; then
        ln -sfn "$DOTFILES_SRC/$dir" "$HOME/.config/$dir"
    fi
done
[[ -f "$DOTFILES_SRC/starship/starship.toml" ]] && \
    ln -sfn "$DOTFILES_SRC/starship/starship.toml" "$HOME/.config/starship.toml" || true

# ── Phase 2: Copy sandbox Hyprland config ─────────────────────────────────
mkdir -p "$HOME/.config/hypr"
if [[ -f "/sandbox/hyprland-headless.conf" ]]; then
    cp /sandbox/hyprland-headless.conf "$HOME/.config/hypr/hyprland-sandbox.conf"
fi

# ── Phase 3: Start Hyprland (nested Wayland) ─────────────────────────────
echo "[sandbox] Starting Hyprland as nested compositor..."

export XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/1000}"
export XDG_SESSION_TYPE=wayland
export XDG_CURRENT_DESKTOP=Hyprland
export LIBVA_DRIVER_NAME=nvidia
export __GLX_VENDOR_LIBRARY_NAME=nvidia
export NVD_BACKEND=direct

if [[ -z "${WAYLAND_DISPLAY:-}" ]]; then
    echo "[sandbox] ERROR: WAYLAND_DISPLAY not set. Mount the host's wayland socket."
    exit 1
fi

# Validate the socket file exists
WAYLAND_SOCKET="$XDG_RUNTIME_DIR/$WAYLAND_DISPLAY"
if [[ ! -S "$WAYLAND_SOCKET" ]]; then
    echo "[sandbox] ERROR: Wayland socket not found at $WAYLAND_SOCKET"
    echo "[sandbox] Ensure the host's wayland socket is bind-mounted into the container."
    exit 1
fi

Hyprland -c "$HOME/.config/hypr/hyprland-sandbox.conf" >"$HYPR_LOG" 2>&1 &
HYPR_PID=$!

# ── Phase 4: Wait for Hyprland ───────────────────────────────────────────
echo "[sandbox] Waiting for Hyprland to start (PID: $HYPR_PID)..."
TIMEOUT=30
ELAPSED=0
while [[ $ELAPSED -lt $TIMEOUT ]]; do
    # Check if a nested wayland socket appeared (different from the host's)
    NESTED=$(ls "$XDG_RUNTIME_DIR"/wayland-*.lock 2>/dev/null | grep -v "${WAYLAND_DISPLAY}.lock" | head -1 || true)
    if [[ -n "$NESTED" ]]; then
        # Extract the display name (e.g., wayland-2)
        NESTED_DISPLAY=$(basename "$NESTED" .lock)
        INST=$(ls "$XDG_RUNTIME_DIR/hypr/" 2>/dev/null | head -1 || true)

        echo "$NESTED_DISPLAY" > /tmp/sandbox-wayland-display
        echo "$INST" > /tmp/sandbox-instance

        echo "[sandbox] Hyprland is ready. Display: $NESTED_DISPLAY, Instance: $INST"
        touch "$READY_SIGNAL"
        break
    fi
    # Check if Hyprland died
    if ! kill -0 "$HYPR_PID" 2>/dev/null; then
        echo "[sandbox] Hyprland exited unexpectedly. Log:"
        cat "$HYPR_LOG" 2>/dev/null || true
        exit 1
    fi
    sleep 1
    ELAPSED=$((ELAPSED + 1))
done

if [[ ! -f "$READY_SIGNAL" ]]; then
    echo "[sandbox] Timeout waiting for Hyprland. Log:"
    cat "$HYPR_LOG" 2>/dev/null || true
    exit 1
fi

# ── Phase 5: Keep alive ──────────────────────────────────────────────────
echo "[sandbox] Sandbox running. Waiting for Hyprland..."
wait "$HYPR_PID"
