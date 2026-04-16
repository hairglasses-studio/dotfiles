#!/usr/bin/env bash
# Install Oh My Zsh and community plugins (runs once per machine)
set -euo pipefail

ZSH="${ZSH:-$HOME/.oh-my-zsh}"
ZSH_CUSTOM="${ZSH_CUSTOM:-$ZSH/custom}"

if [[ ! -d "$ZSH" ]]; then
    echo "[chezmoi] Installing Oh My Zsh..."
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended --keep-zshrc
fi

declare -A plugins=(
    [fast-syntax-highlighting]="https://github.com/zdharma-continuum/fast-syntax-highlighting"
    [zsh-autosuggestions]="https://github.com/zsh-users/zsh-autosuggestions"
    [zsh-completions]="https://github.com/zsh-users/zsh-completions"
    [you-should-use]="https://github.com/MichaelAquilina/zsh-you-should-use"
    [fzf-tab]="https://github.com/Aloxaf/fzf-tab"
)

for name in "${!plugins[@]}"; do
    dir="$ZSH_CUSTOM/plugins/$name"
    if [[ ! -d "$dir" ]]; then
        echo "[chezmoi] Installing OMZ plugin: $name"
        git clone --depth=1 "${plugins[$name]}" "$dir" 2>/dev/null
    fi
done
