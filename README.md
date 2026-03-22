# dotfiles

macOS terminal rice: Ghostty + Starship + Oh My Zsh + neovim.

## What's inside

| Config | Description |
|---|---|
| `zsh/` | Oh My Zsh, Starship/p10k toggle, 20+ plugins, 392 lines of aliases |
| `ghostty/` | Visor terminal, bloom shader, Snazzy-on-black palette, P3 wide gamut |
| `starship/` | Prompt with git metrics, cloud context, language versions |
| `nvim/` | vim-plug, 23 plugins, LSP via CoC, LLM-friendly functions |
| `bat/` | 1337 dark theme |
| `fastfetch/` | Minimal system splash with Nerd Font icons |
| `git/` | Delta diff viewer with Catppuccin theme |
| `k9s/` | Kubernetes CLI transparent skin |
| `gh/` | GitHub CLI config (SSH protocol) |
| `ssh/` | 1Password SSH agent |

## Install

```bash
git clone git@github.com:hairglasses-studio/dotfiles.git ~/dotfiles
cd ~/dotfiles && ./install.sh
```

The installer will:
- Install Homebrew packages from `Brewfile`
- Install Oh My Zsh + custom plugins + Powerlevel10k
- Install vim-plug for neovim
- Symlink all configs to their expected locations
- Back up any existing files to `~/.dotfiles-backup-*/`

## Uninstall

```bash
cd ~/dotfiles && ./uninstall.sh
```

## Font

[JetBrainsMono Nerd Font Mono](https://www.nerdfonts.com/) — installed via `brew install --cask font-jetbrains-mono-nerd-font`.
