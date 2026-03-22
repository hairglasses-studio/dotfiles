# dotfiles

Full macOS rice — Snazzy-on-black palette from the desktop layer down to every TUI.

AeroSpace + SketchyBar + JankyBorders + Ghostty + Starship + Oh My Zsh + Neovim + k9s + tmux + lazygit + btop + yazi + bat + delta.

## Quick Install

```bash
git clone git@github.com:hairglasses-studio/dotfiles.git ~/dotfiles
cd ~/dotfiles && bash install.sh
```

The installer is idempotent — safe to run multiple times. Existing files are backed up to `~/.dotfiles-backup-*/`.

What it does:

- Installs Homebrew + 65 packages from `Brewfile`
- Installs Oh My Zsh + 13 plugins + Powerlevel10k
- Installs vim-plug for Neovim
- Symlinks all configs to their expected locations

Post-install:

```bash
bash ~/dotfiles/scripts/macos-defaults.sh    # Dock autohide, fast keys, Finder tweaks
bash install.sh --check                       # Validate everything is linked
```

## What's Inside

| Config | Description |
|--------|-------------|
| `aerospace/` | i3-style tiling WM — alt+hjkl focus/move, 9 workspaces, 8px gaps |
| `sketchybar/` | Custom menu bar — workspaces, front app, k8s context, clock, battery, CPU, Wi-Fi |
| `borders/` | Window borders — blue active (`#57c7ff`), gray inactive (`#686868`) |
| `ghostty/` | Visor terminal, 52 GLSL shaders with random rotation, Snazzy palette, P3 wide gamut |
| `zsh/` | Oh My Zsh, transient Starship prompt, 13 plugins, 490+ lines of aliases |
| `starship/` | Fill-based right alignment, git metrics, cloud context, helm/container modules |
| `nvim/` | vim-plug, 23 plugins, Snazzy colorscheme, transparent background, CoC LSP |
| `btop/` | System monitor with custom Snazzy theme, braille graphs, transparent bg |
| `yazi/` | Terminal file manager with full Snazzy theme, file type colors, preview support |
| `k9s/` | Snazzy skin, 7 plugins (stern, debug, dive, port-forward, yaml, jq-logs, helm), 18 aliases |
| `tmux/` | Snazzy status bar, C-a prefix, vi copy mode, hjkl pane navigation |
| `lazygit/` | Full Snazzy theme with nerd font icons |
| `bat/` | 1337 theme, Go/Terraform/proto syntax mappings |
| `fastfetch/` | System splash — OS, kernel, WM, GPU, display, battery, packages |
| `git/` | Snazzy delta diffs, 1Password SSH signing, URL shortcuts, PR aliases |
| `cava/` | Audio visualizer — 4-color Snazzy gradient |
| `glow/` | Markdown renderer, dark style |
| `gh/` | GitHub CLI (SSH protocol) |
| `ssh/` | 1Password SSH agent |

## Window Management

AeroSpace + SketchyBar + JankyBorders work together:

```
┌─────────────────────────────────────────────────────────┐
│  SketchyBar: [1 2 3 4 5]  Ghostty      k8s:prod  12:30 │
├────────────────────────────┬────────────────────────────┤
│                            │                            │
│  Ghostty (focused)         │  Ghostty                   │
│  ┌─ blue border ──────┐   │  ┌─ gray border ──────┐   │
│  │                     │   │  │                     │   │
│  │  k9s / lazygit      │   │  │  btop / yazi        │   │
│  │                     │   │  │                     │   │
│  └─────────────────────┘   │  └─────────────────────┘   │
│                            │                            │
└────────────────────────────┴────────────────────────────┘
```

Key bindings (alt as modifier):

| Key | Action |
|-----|--------|
| `alt-h/j/k/l` | Focus left/down/up/right |
| `alt-shift-h/j/k/l` | Move window |
| `alt-1..9` | Switch workspace |
| `alt-shift-1..9` | Move window to workspace |
| `alt-f` | Fullscreen |
| `alt-shift-space` | Toggle float/tile |
| `alt-shift-;` | Resize mode (then h/j/k/l) |

## Shaders

52 GLSL shaders from 4 community repos, organized in 6 categories:

| Category | Count | Examples |
|----------|-------|---------|
| CRT | 8 | amber-crt, green-crt, bettercrt, retro-terminal |
| Post-FX | 10 | bloom, dither, glitchy, glow-rgbsplit, spotlight |
| Background | 18 | galaxy, matrix-hallway, starfield, underwater, fireworks |
| Cursor | 8 | cursor_blaze, cursor_tail, ripple_cursor, sonic_boom |
| Watercolor | 9 | flat-wash, glazing, salt, splatter, wet-on-wet |

```bash
shader-random          # random shader (runs automatically per terminal)
shader-pick            # fzf picker with categories
shader-bloom           # set bloom.glsl
shader-crt             # set green-crt.glsl
shader-none            # disable shader
```

## Palette — Snazzy on Black

| Color | Hex | Usage |
|-------|-----|-------|
| Blue | `#57c7ff` | Primary, active borders, focused states |
| Magenta | `#ff6ac1` | Accents, highlights |
| Green | `#5af78e` | Success, additions |
| Yellow | `#f3f99d` | Warnings, search |
| Red | `#ff5c57` | Errors, deletions |
| Gray | `#686868` | Muted text, inactive borders |
| Dark | `#1a1a1a` | Secondary background, panels |
| Bright | `#f1f1f0` | Foreground text |
| BG | `#000000` | Terminal/window background |

Applied to: AeroSpace (SketchyBar), JankyBorders, Ghostty, Neovim, k9s, tmux, lazygit, btop, yazi, cava, Starship, bat, delta, fastfetch.

## Key Aliases

```bash
# Window management
aero-reload    # reload AeroSpace config
bar-reload     # reload SketchyBar

# New tools
top            # btop (replaces macOS top)
yy             # yazi file manager
viz            # cava audio visualizer
md             # glow markdown renderer

# Go
gobr           # go build -race ./...
gotr           # go test -race ./...
golc           # golangci-lint run

# Kubernetes
kx / kn        # kubectx / kubens
kgp / klf      # kubectl get pods / logs -f

# Containers
dk / dkc       # docker / docker compose
lzd            # lazydocker

# Git
gst / gco      # status / checkout
gcm / gp / gl  # commit / push / pull
```

## Brewfile Highlights

```
# Desktop rice
aerospace, sketchybar, borders              # Window management
btop, yazi, cava, glow                      # TUI tools

# CLI essentials
bat, eza, fd, fzf, ripgrep, jq, yq         # Search & display
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
