#!/usr/bin/env bash
# Install Tmux Plugin Manager (runs once per machine)
set -euo pipefail

tpm_dir="$HOME/.tmux/plugins/tpm"
if [[ ! -d "$tpm_dir" ]]; then
    echo "[chezmoi] Installing TPM..."
    git clone --depth=1 https://github.com/tmux-plugins/tpm "$tpm_dir" 2>/dev/null
fi
