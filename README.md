# dotfiles

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Shaders](https://img.shields.io/badge/GLSL_Shaders-138-purple)](ghostty/shaders/)
[![WM](https://img.shields.io/badge/WM-Hyprland-cyan)](https://hyprland.org/)

Full macOS rice — Snazzy-on-black palette from the desktop layer down to every TUI.

AeroSpace + SketchyBar + JankyBorders + Ghostty + Starship + Oh My Zsh + Neovim + k9s + tmux + lazygit + btop + yazi + bat + delta + RetroVisor + Tattoy.

## Prerequisites

You need Xcode Command Line Tools (git, clang, etc.). If not already installed:

```bash
xcode-select --install
```

Everything else (Homebrew, Oh My Zsh, plugins) is handled by `install.sh`.

## Install

```bash
git clone git@github.com:hairglasses-studio/dotfiles.git ~/dotfiles
cd ~/dotfiles
bash install.sh
```

The installer is idempotent — safe to run multiple times. Existing files are backed up to `~/.dotfiles-backup-*/`.

### What the installer does

1. Installs Homebrew (if missing) + 76 packages from `Brewfile`
2. Installs Oh My Zsh + 5 community plugins + Powerlevel10k theme
3. Installs vim-plug for Neovim + creates undo/backup/swap directories
4. Installs TPM (Tmux Plugin Manager)
5. Downloads and installs RetroVisor (CRT shader overlay)
6. Symlinks all 32+ configs to their expected locations
7. Links Ghostty shader collection to Tattoy
8. Builds bat theme cache

### Post-install steps

These require interactive input and can't be automated:

```bash
# 1. Apply macOS system preferences (Dock autohide, fast key repeat, Finder tweaks)
bash ~/dotfiles/scripts/macos-defaults.sh

# 2. Validate all symlinks are correct
bash ~/dotfiles/install.sh --check

# 3. Install Neovim plugins (opens nvim, installs, quits)
nvim +PlugInstall +qall

# 4. Install tmux plugins — open tmux then press prefix + I (C-a + I)
tmux new-session
# Inside tmux: press C-a then shift-I, wait for install, then exit

# 5. Reload your shell
source ~/.zshrc
```

### Machine-specific config

Before pushing your own changes, update these files with your own values:

- `git/gitconfig` — name, email, GPG signing key
- `ssh/config` — 1Password SSH agent path (if using a different password manager)

### Uninstall

```bash
bash ~/dotfiles/uninstall.sh
```

Removes all symlinks created by the installer. Does not uninstall Homebrew packages, Oh My Zsh, or other tools. Prints the path to your most recent backup directory for restoration.

## What's Inside

| Config | Description |
|--------|-------------|
| `aerospace/` | i3-style tiling WM — alt+hjkl focus/move, 9 workspaces, 8px gaps |
| `sketchybar/` | Custom menu bar — workspaces, front app, now playing, k8s context, clock, battery, CPU |
| `borders/` | Window borders — blue active (`#57c7ff`), gray inactive (`#686868`) |
| `ghostty/` | Visor terminal, 138 GLSL shaders with shuffled playlists, Snazzy palette, P3 wide gamut |
| `zsh/` | Oh My Zsh, transient Starship prompt, 13 plugins, 650+ lines of aliases, command notifications |
| `starship/` | Fill-based right alignment, git metrics, cloud context, helm/container modules |
| `nvim/` | vim-plug, 29 plugins, alpha-nvim dashboard, treesitter, indent guides, colorizer, CoC LSP |
| `btop/` | System monitor with custom Snazzy theme, braille graphs, transparent bg |
| `yazi/` | Terminal file manager with full Snazzy theme, file type colors, preview support |
| `k9s/` | Snazzy skin, 7 plugins (stern, debug, dive, port-forward, yaml, jq-logs, helm), 18 aliases |
| `tmux/` | TPM, 7 plugins (resurrect, continuum, battery, cpu), vim-tmux-navigator, Snazzy status bar |
| `lazygit/` | Full Snazzy theme with nerd font icons |
| `bat/` | Custom Snazzy tmTheme, Go/Terraform/proto syntax mappings |
| `fastfetch/` | Custom skull ASCII art, Snazzy colors — OS, kernel, WM, GPU, display, battery, packages |
| `git/` | Snazzy delta diffs, 1Password SSH signing, URL shortcuts, PR aliases |
| `cava/` | Audio visualizer — 8-color Snazzy gradient with Monstercat smoothing |
| `glow/` | Markdown renderer, dark style |
| `gh/` | GitHub CLI (SSH protocol) |
| `ssh/` | 1Password SSH agent |
| `tattoy/` | Terminal shader compositor — layers cursor + background shaders simultaneously |
| `retrovisor/` | CRT shader overlay — auto-launches at login via LaunchAgent |

### Directory layout and symlink targets

```
dotfiles/
├── aerospace/          → ~/.aerospace.toml
├── bat/               → ~/.config/bat
├── borders/           → ~/.config/borders
├── btop/              → ~/.config/btop
├── cava/              → ~/.config/cava
├── fastfetch/         → ~/.config/fastfetch
├── gh/                → ~/.config/gh
├── ghostty/           → ~/.config/ghostty
│   └── shaders/       → 138 GLSL shaders + playlists + scripts
├── git/               → ~/.gitconfig + ~/.config/delta + ~/.config/git/ignore
├── glow/              → ~/.config/glow
├── k9s/               → ~/.config/k9s
├── lazygit/           → ~/.config/lazygit
├── nvim/              → ~/.config/nvim
├── retrovisor/        → ~/Library/LaunchAgents/ (plist)
├── scripts/           (not symlinked — run manually)
├── sketchybar/        → ~/.config/sketchybar
├── ssh/               → ~/.ssh/config
├── starship/          → ~/.config/starship.toml
├── tattoy/            → ~/Library/Application Support/tattoy/tattoy.toml
├── tmux/              → ~/.tmux.conf
├── yazi/              → ~/.config/yazi
├── zsh/               → ~/.zshrc + ~/.zshenv + ~/.p10k.zsh
├── Brewfile
├── install.sh
└── uninstall.sh
```

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
| `alt-shift-r` | Reload AeroSpace config |
| `alt-shift-s` | Random shader |
| `` ` `` (backtick) | Toggle quick terminal (visor drop-down) |

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

Prefix is `C-a`. Split panes with `C-a |` (horizontal) and `C-a -` (vertical).

## Shaders

138 GLSL shaders from 25+ community repos, gists, and Shadertoy adaptations:

| Category | Count | Examples |
|----------|-------|---------|
| CRT | 15 | crt-chromatic, amber-monitor, vt320-amber-glow, bettercrt, retro-terminal |
| Post-FX | 37 | cyberpunk, holo-shimmer, old-film, vcr-distortion, vaporwave, bloom variants, focus-blur |
| Background | 30 | electric, fireworks, matrix, lava, smoke-and-ghost, starfield variants, galaxy |
| Cursor | 33 | cursor_explosion, cursor_viberation, cursor_smear variants, manga_slash |
| Watercolor | 5 | graded-wash, salt, splatter, variegated-wash, wet-on-wet |

### Shader playlists

Shaders rotate automatically on each new shell via shuffled playlists with no-repeat playback:

| Context | Playlist | Shaders |
|---------|----------|---------|
| Normal Ghostty windows | Low-intensity | 59 — blooms, subtle cursors, gentle retro, watercolors, calm backgrounds |
| Quick terminal (backtick) | High-intensity | 62 — matrix rain, heavy CRT, flashy cursors, intense glitch |
| Tattoy background layer | All non-cursor | 88 — CRT, backgrounds, post-FX, watercolor |
| Tattoy cursor layer | All cursor | 33 — blaze, smear, sweep, ripple, sparks |

Playlists are managed by `ghostty/shaders/bin/shader-playlist.sh` with Fisher-Yates shuffle and state persistence in `~/.local/state/ghostty/`.

### Shader commands

```bash
shader-next            # advance to next shader (auto-runs per new shell)
shader-rotate          # manually rotate Ghostty + Tattoy shaders
shader-random          # pick random shader from all 121 (ignores playlists)
shader-pick            # fzf picker with categories and descriptions
shader-crt             # set green-crt.glsl
shader-none            # disable shader
shader-status          # show playlist positions for all 4 queues
shader-reshuffle       # reset all playlists for a fresh shuffle
shader-audit           # interactive one-by-one audition of all shaders
```

### Tattoy (shader compositor)

Tattoy layers two independent shaders on top of the terminal simultaneously:

- **Background layer** — CRT, post-FX, or animated background at 30% opacity behind text
- **Cursor layer** — cursor trail/effect at 80% opacity

Toggle with `ALT+t`. Cycle shaders with `ALT+9` / `ALT+0`. Both layers rotate automatically alongside Ghostty on each new shell.

### RetroVisor (CRT overlay)

macOS app that applies a hardware CRT scanline/distortion overlay at the window level. Auto-launches at login via LaunchAgent.

```bash
crt-on                 # launch RetroVisor
crt-off                # kill RetroVisor
crt-toggle             # toggle on/off
```

## Terminal Fun

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

Applied to: AeroSpace, SketchyBar, JankyBorders, Ghostty, Neovim, k9s, tmux, lazygit, btop, yazi, cava, Starship, bat, delta, fastfetch.

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

# Shader & VFX
tattoy                                      # Terminal shader compositor
glslang                                     # Shader validation

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
1password-cli                               # Secrets
```

## Troubleshooting

**Shaders don't animate:** Check that `custom-shader-animation = true` is set in `~/.config/ghostty/config`. The playlist system sets this automatically for shaders that need it, but manual overrides may disable it.

**RetroVisor fails to install:** The automated installer downloads from GitHub releases. If it fails, install manually from [github.com/dirkwhoffmann/RetroVisor/releases](https://github.com/dirkwhoffmann/RetroVisor/releases).

**Symlinks point to wrong place:** Run `bash install.sh --check` to validate. If broken, re-run `bash install.sh` to recreate them.

**Powerlevel10k prompt looks broken:** Ensure JetBrainsMono Nerd Font Mono is installed (`brew install --cask font-jetbrains-mono-nerd-font`) and set as your terminal font.

**tmux plugins not loading:** Inside tmux, press `C-a` then `shift-I` to trigger TPM install. Wait for the install to complete.

**Neovim shows errors on startup:** Run `nvim +PlugInstall +qall` to install missing plugins. If CoC errors appear, run `:CocInstall` for your language servers.

**Tattoy not picking up shader changes:** Toggle effects with `ALT+t` or restart your Tattoy session.

**GPG signing fails on commits:** Ensure your GPG agent is running: `gpgconf --kill gpg-agent && gpg-agent --daemon`.

## Font

[JetBrainsMono Nerd Font Mono](https://www.nerdfonts.com/) — installed via `brew install --cask font-jetbrains-mono-nerd-font`.

## License

MIT
