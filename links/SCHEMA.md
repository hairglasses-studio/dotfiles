# Links Index Schema

How to read and add entries to this community dotfiles index.

## Directory Structure

```
links/
├── _registry.json        # Dedup registry, keyed by GitHub URL
├── INDEX.md              # Stats, tables, priority reference
├── SCHEMA.md             # This file
├── RESEARCH-LINKS.md     # Awesome-list URLs, Reddit threads, research cross-refs
├── hyprland/             # Full dotfiles with Hyprland as primary WM
├── sway/                 # Full dotfiles with Sway or i3
├── bars/                 # Standalone bar configs (eww, waybar, AGS, polybar)
├── terminals/            # Terminal configs, themes, plugins
├── shaders/              # GLSL shader collections, terminal visual effects
├── themes/               # Color scheme frameworks and ports
├── nixos/                # NixOS/home-manager declarative configs
├── components/           # Lockscreen, launcher, notification, SDDM, GRUB, Plymouth, cursor themes
├── tools/                # Ricing tools, dotfile managers, install frameworks
└── collections/          # Awesome-lists and curated indexes
```

## Entry Format

Each repo gets one `.md` file: `links/{category}/{github_username}.md` (or `{username}-{repo-suffix}.md` for multiple repos per user).

### YAML Frontmatter

```yaml
---
name: "repo-name"
github: "https://github.com/owner/repo"
author: "owner"
stars: 1234
last_updated: "2026-03-15"
discovered_in_round: 1
also_found_in_rounds: [3, 7]
is_fork: false
type: "dotfiles"            # dotfiles | tool | collection | theme | shader | framework
distro: ["arch", "nixos"]
wm: "hyprland"              # hyprland | sway | i3 | bspwm | awesome | dwm | qtile | xmonad | null
bar: "eww"                  # eww | waybar | polybar | ags | lemonbar | null
terminal: "foot"            # ghostty | kitty | alacritty | foot | wezterm | st | null
shell: "zsh"                # zsh | fish | bash | nushell | null
editor: "neovim"            # neovim | vim | emacs | helix | null
launcher: "wofi"            # rofi | wofi | fuzzel | dmenu | tofi | null
notifications: "mako"       # mako | dunst | swaync | fnott | null
lockscreen: null            # hyprlock | swaylock | i3lock | null
color_scheme: "catppuccin-mocha"  # catppuccin-* | nord | gruvbox | dracula | tokyo-night | rose-pine | custom | null
has_screenshots: true
has_install_script: true
notable_techniques: ["Material You theming"]
tags: ["polished", "feature-rich"]
relevance: "high"           # high | medium | low
category_in_awesome: "Rices"
source_awesome_list: "awesome-hyprland"
---
```

### Body

```markdown
Brief description of what makes this notable.

**Stack**: Hyprland + eww + foot + zsh | Custom palette, Material You dynamic theming

**Notable patterns**: Material You color extraction, custom widget framework

**Links**: [GitHub](url) | [Reddit](url) | [awesome-hyprland](url)
```

## Adding Entries Manually

1. Create `links/{category}/{owner}.md` with the frontmatter + body above
2. Add the GitHub URL to `links/_registry.json` under `entries`
3. Increment `total_entries` in the registry
4. Run validation (check frontmatter parses, URL is unique)

## Relevance Scoring

- **HIGH**: Uses our stack tools (Hyprland, eww, Ghostty, foot), cyberpunk/neon aesthetics, Arch Hyprland rices, desktop-as-MCP
- **MEDIUM**: General Wayland ricing, different tools but interesting patterns, NixOS Hyprland
- **LOW**: X11-only, macOS-only, Windows, no transferable techniques

## Classification Rules

- Classify by primary purpose — a NixOS config that uses Hyprland goes in `nixos/`, not `hyprland/`
- Standalone tool repos go in `tools/` even if they're commonly used in ricing
- Theme framework repos (catppuccin org, nord org) go in `themes/`
- Awesome-lists go in `collections/` with full category trees preserved in the body
