#!/usr/bin/env bash
# Install vim-plug for Neovim (runs once per machine)
set -euo pipefail

plug_path="${XDG_DATA_HOME:-$HOME/.local/share}/nvim/site/autoload/plug.vim"
if [[ ! -f "$plug_path" ]]; then
    echo "[chezmoi] Installing vim-plug..."
    curl -fLo "$plug_path" --create-dirs \
        https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim
fi
