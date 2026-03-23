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

- Installs Homebrew + 76 packages from `Brewfile`
- Installs Oh My Zsh + 13 plugins + Powerlevel10k
- Installs vim-plug for Neovim
- Installs TPM (Tmux Plugin Manager)
- Symlinks all configs to their expected locations
- Builds bat theme cache

Post-install:

```bash
bash ~/dotfiles/scripts/macos-defaults.sh    # Dock autohide, fast keys, Finder tweaks
bash install.sh --check                       # Validate everything is linked
nvim +PlugInstall +qall                       # Install Neovim plugins
tmux new "prefix + I"                         # Install tmux plugins via TPM
```

## What's Inside

| Config | Description |
|--------|-------------|
| `aerospace/` | i3-style tiling WM — alt+hjkl focus/move, 9 workspaces, 8px gaps |
| `sketchybar/` | Custom menu bar — workspaces, front app, now playing, k8s context, clock, battery, CPU, Wi-Fi |
| `borders/` | Window borders — blue active (`#57c7ff`), gray inactive (`#686868`) |
| `ghostty/` | Visor terminal, 120 GLSL shaders with random rotation, Snazzy palette, P3 wide gamut |
| `zsh/` | Oh My Zsh, transient Starship prompt, 13 plugins, 500+ lines of aliases, command notifications |
| `starship/` | Fill-based right alignment, git metrics, cloud context, helm/container modules |
| `nvim/` | vim-plug, 29 plugins, alpha-nvim dashboard, treesitter, indent guides, colorizer, CoC LSP |
| `btop/` | System monitor with custom Snazzy theme, braille graphs, transparent bg |
| `yazi/` | Terminal file manager with full Snazzy theme, file type colors, preview support |
| `k9s/` | Snazzy skin, 7 plugins (stern, debug, dive, port-forward, yaml, jq-logs, helm), 18 aliases |
| `tmux/` | TPM, 7 plugins (resurrect, continuum, battery, cpu), vim-tmux-navigator, Snazzy status bar |
| `lazygit/` | Full Snazzy theme with nerd font icons |
| `bat/` | Custom Snazzy tmTheme, Go/Terraform/proto syntax mappings |
| `fastfetch/` | Custom skull ASCII art, Snazzy colors — OS, kernel, WM, GPU, display, battery, packages |
| `git/` | Snazzy delta diffs (custom Snazzy theme), 1Password SSH signing, URL shortcuts, PR aliases |
| `cava/` | Audio visualizer — 4-color Snazzy gradient |
| `glow/` | Markdown renderer, dark style |
| `gh/` | GitHub CLI (SSH protocol) |
| `ssh/` | 1Password SSH agent |

## Window Management

AeroSpace + SketchyBar + JankyBorders work together:

```
┌─────────────────────────────────────────────────────────────────┐
│  SketchyBar: [1 2 3 4 5]  Ghostty   ♫ now playing   k8s  12:30 │
├──────────────────────────────┬──────────────────────────────────┤
│                              │                                  │
│  Ghostty (focused)           │  Ghostty                         │
│  ┌─ blue border ──────┐     │  ┌─ gray border ──────┐         │
│  │                     │     │  │                     │         │
│  │  k9s / lazygit      │     │  │  btop / yazi        │         │
│  │                     │     │  │                     │         │
│  └─────────────────────┘     │  └─────────────────────┘         │
│                              │                                  │
└──────────────────────────────┴──────────────────────────────────┘
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

## Neovim

29 plugins with Snazzy dashboard on startup:

- **alpha-nvim** — custom "RICED" ASCII art dashboard with Snazzy gradient
- **nvim-treesitter** — modern syntax highlighting for 18 languages
- **indent-blankline** — visual indent guides (blue scope highlights)
- **nvim-colorizer** — inline hex color preview
- **CoC.nvim** — LSP completion engine
- **fzf.vim** — fuzzy finder integration
- **vim-tmux-navigator** — seamless C-hjkl navigation between vim and tmux panes

## Tmux

TPM-managed with session persistence:

- **tmux-resurrect + continuum** — sessions survive reboots, auto-save every 15min
- **vim-tmux-navigator** — seamless C-hjkl pane switching with Neovim
- **tmux-battery + tmux-cpu** — live battery/CPU in Snazzy status bar

## Shaders

120 GLSL shaders from 25+ community repos, gists, and Shadertoy adaptations:

| Category | Count | Examples |
|----------|-------|---------|
| CRT | 15 | crt-chromatic, amber-monitor, vt320-amber-glow, bettercrt, retro-terminal |
| Post-FX | 37 | cyberpunk, holo-shimmer, old-film, vcr-distortion, vaporwave, bloom variants, focus-blur |
| Background | 30 | electric, fireworks, matrix, lava, smoke-and-ghost, starfield variants, galaxy |
| Cursor | 33 | cursor_explosion, cursor_viberation, cursor_smear variants, manga_slash |
| Watercolor | 5 | graded-wash, salt, splatter, variegated-wash, wet-on-wet |

```bash
shader-random          # random shader (runs automatically per terminal)
shader-pick            # fzf picker with categories
shader-crt             # set green-crt.glsl
shader-none            # disable shader
```

## Terminal Fun

Hacker aesthetic tools for the full unix den experience:

```bash
matrix             # cmatrix — Matrix rain (cyan)
pipes              # pipes-sh — pipe screensaver
bonsai             # cbonsai — growing bonsai tree
clock              # tty-clock — terminal clock (blue)
screensaver        # pipes-sh in screensaver mode
banner "text"      # figlet — ASCII art text
rainbow            # lolcat — rainbow text
gitfetch           # onefetch — git repo info display
weather            # current weather (wttr.in)
forecast           # full weather forecast
cheat cmd          # cheat.sh — cheat sheets
colortest          # display all 256 terminal colors
```

Shell startup randomly alternates between `fastfetch` (custom skull art) and `fortune | cowsay | lolcat`.

Auto-displays `onefetch` when cd'ing into a git repository root.

Long-running commands (>30s) trigger macOS notifications via `terminal-notifier`.

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

# Tools
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

# Hacker aesthetic
pipes-sh, cbonsai, tty-clock, cmatrix       # Terminal fun
lolcat, figlet, toilet, cowsay, fortune     # ASCII art & text effects
onefetch, terminal-notifier                  # Git info & notifications

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
