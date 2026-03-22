# dotfiles

macOS terminal rice — Snazzy-on-black palette across every tool.

Ghostty + Starship + Oh My Zsh + Neovim + k9s + tmux + lazygit + bat + delta.

## Quick Install

```bash
git clone git@github.com:hairglasses-studio/dotfiles.git ~/dotfiles
cd ~/dotfiles && bash install.sh
```

The installer is idempotent — safe to run multiple times. Existing files are backed up to `~/.dotfiles-backup-*/`.

What it does:

- Installs Homebrew + 55 packages from `Brewfile`
- Installs Oh My Zsh + 13 plugins + Powerlevel10k
- Installs vim-plug for Neovim
- Symlinks all configs to their expected locations

## What's Inside

| Config | Description |
|--------|-------------|
| `zsh/` | Oh My Zsh, transient Starship prompt, 13 plugins, 470+ lines of aliases (Go, AWS, K8s, Helm, containers, MCP, shaders) |
| `ghostty/` | Visor terminal, 52 GLSL shaders with random rotation, Snazzy palette, P3 wide gamut, splits/tabs/clipboard keybinds |
| `starship/` | Fill-based right alignment, git metrics, cloud context, helm/container/env_var modules |
| `nvim/` | vim-plug, 23 plugins, LSP via CoC |
| `bat/` | 1337 theme, Go/Terraform/proto syntax mappings |
| `fastfetch/` | System splash with Nerd Font icons |
| `git/` | Delta diff viewer, 1Password SSH signing, URL shortcuts, PR aliases |
| `k9s/` | Snazzy skin, 7 plugins (stern, debug, dive, port-forward, yaml, jq-logs, helm), 18 resource aliases |
| `tmux/` | Snazzy status bar, C-a prefix, vi copy mode, hjkl pane navigation |
| `lazygit/` | Snazzy border/selection colors |
| `gh/` | GitHub CLI (SSH protocol) |
| `ssh/` | 1Password SSH agent |

## Shaders

52 GLSL shaders from 4 community repos, organized in 6 categories:

| Category | Count | Examples |
|----------|-------|---------|
| CRT | 8 | amber-crt, green-crt, bettercrt, retro-terminal |
| Post-FX | 10 | bloom, dither, glitchy, glow-rgbsplit, spotlight |
| Background | 18 | galaxy, matrix-hallway, starfield, underwater, fireworks |
| Cursor | 8 | cursor_blaze, cursor_tail, ripple_cursor, sonic_boom |
| Watercolor | 9 | flat-wash, glazing, salt, splatter, wet-on-wet |

### Shader Commands

```bash
shader-random          # pick a random shader (runs automatically per terminal)
shader-pick            # fzf picker with categories and descriptions
shader-bloom           # set bloom.glsl
shader-crt             # set green-crt.glsl
shader-amber           # set amber-crt.glsl
shader-none            # disable shader
```

Each new terminal window automatically gets a random shader via `shader-random` in `.zshrc`.

Validate all shaders:

```bash
bash ~/.config/ghostty/shaders/test-shaders.sh           # static analysis
bash ~/.config/ghostty/shaders/test-shaders.sh --list     # print catalog
bash ~/.config/ghostty/shaders/test-shaders.sh --visual   # launch each in Ghostty
```

## Palette — Snazzy on Black

| Color | Hex | Usage |
|-------|-----|-------|
| Blue | `#57c7ff` | Primary, active states, prompt |
| Magenta | `#ff6ac1` | Accents, highlights |
| Green | `#5af78e` | Success, git additions |
| Yellow | `#f3f99d` | Warnings, search |
| Red | `#ff5c57` | Errors, deletions |
| Gray | `#686868` | Muted text, borders |
| Dark | `#1a1a1a` | Secondary background |
| Bright | `#f1f1f0` | Foreground text |
| Background | `#000000` | Terminal background |

Applied to: Ghostty, k9s, tmux, lazygit, Starship, bat/delta.

## Key Aliases

```bash
# Go
gobr       # go build -race ./...
gotr       # go test -race ./...
golc       # golangci-lint run

# Kubernetes
kx         # kubectx (switch cluster)
kn         # kubens (switch namespace)
kgp        # kubectl get pods
klf        # kubectl logs -f
kpf        # kubectl port-forward

# Containers
dk         # docker
dkps       # docker ps
dkc        # docker compose
lzd        # lazydocker

# Git
gst        # git status
gco        # git checkout
gcm        # git commit -m
gp         # git push
gl         # git pull
```

See `zsh/aliases.zsh` for the full 470+ line alias set.

## Brewfile Highlights

Go toolchain, K8s tools, containers, and utilities:

```
bat, eza, fd, fzf, ripgrep, jq, yq         # CLI essentials
go, golangci-lint, goreleaser, air          # Go development
k9s, helm, stern, kubectx                   # Kubernetes
docker, colima, dive, lazydocker            # Containers
neovim, tmux, lazygit, gh                   # Editors & tools
ghostty, font-jetbrains-mono-nerd-font      # Terminal
```

## Font

[JetBrainsMono Nerd Font Mono](https://www.nerdfonts.com/) — installed via `brew install --cask font-jetbrains-mono-nerd-font`.

## License

MIT
