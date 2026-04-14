#!/usr/bin/env bash
# greetd-deploy.sh — Deploy greetd config and /etc/environment
# Requires: sudo
set -euo pipefail

DOTFILES="$(cd "$(dirname "$0")/.." && pwd)"
GREETER_HOME="/var/lib/greetd"

ensure_greeter_state() {
    if ! getent passwd greeter >/dev/null 2>&1; then
        echo "greeter user not present; skipping greeter home fix"
        return 0
    fi

    local current_home
    current_home="$(getent passwd greeter | cut -d: -f6)"

    echo "Ensuring greeter home and keyring dirs..."
    sudo install -d -m700 -o greeter -g greeter "$GREETER_HOME"
    sudo install -d -m700 -o greeter -g greeter "$GREETER_HOME/.local"
    sudo install -d -m700 -o greeter -g greeter "$GREETER_HOME/.local/share"
    sudo install -d -m700 -o greeter -g greeter "$GREETER_HOME/.local/share/keyrings"
    sudo chown -R greeter:greeter "$GREETER_HOME"

    if [[ "$current_home" != "$GREETER_HOME" ]]; then
        sudo usermod -d "$GREETER_HOME" greeter
        echo "  Updated greeter home: $current_home -> $GREETER_HOME"
    else
        echo "  Greeter home already set: $GREETER_HOME"
    fi
}

echo "Deploying greetd config..."
sudo install -Dm644 "$DOTFILES/greetd/config.toml" /etc/greetd/config.toml
echo "  Copied greetd/config.toml -> /etc/greetd/config.toml"

echo "Deploying /etc/environment..."
sudo install -Dm644 "$DOTFILES/greetd/environment" /etc/environment
echo "  Copied greetd/environment -> /etc/environment"

if [[ -d "$DOTFILES/greetd/pam" ]]; then
    echo "Deploying greetd PAM configs..."
    for pam_file in "$DOTFILES/greetd/pam/"*; do
        [[ -f "$pam_file" ]] || continue
        sudo install -Dm644 "$pam_file" "/etc/pam.d/$(basename "$pam_file")"
        echo "  Copied greetd/pam/$(basename "$pam_file") -> /etc/pam.d/$(basename "$pam_file")"
    done
fi

ensure_greeter_state

echo "greetd config deployed. Hyprland is now the default session."
echo "Next greetd login will use the updated greeter home."
