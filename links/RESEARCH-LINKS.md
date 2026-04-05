# Research Links

Reference for future Claude Code sessions. Points to all sources scraped and existing research.

## Awesome-Lists Scraped

| List | URL | Categories | Entries |
|------|-----|------------|---------|
| awesome-hyprland | https://github.com/hyprland-community/awesome-hyprland | 23 categories | 192 |
| awesome-wayland | https://github.com/rcalixte/awesome-wayland | Compositors, tools, libs | ~150 |
| awesome-nix | https://github.com/nix-community/awesome-nix | Dotfiles, dev, resources | ~200 |
| awesome-arch | https://github.com/PandaFoss/Awesome-Arch | Distros, tools, resources | ~100 |
| awesome-ricing | https://github.com/fosslife/awesome-ricing | WMs, bars, launchers, themes | ~150 |
| awesome-dotfiles | https://github.com/webpro/awesome-dotfiles | Tools, frameworks, guides | ~100 |
| awesome-rices | https://github.com/zemmsoares/awesome-rices | Community rices by WM | ~50 |
| Awesome-Ghostty | https://github.com/fearlessgeekmedia/Awesome-Ghostty | Shaders, themes, tools | ~80 |

## Quick Reference

### Looking for eww patterns?
- Start with `links/bars/` — 39 repos with eww/waybar/AGS/polybar configs
- awesome-hyprland "Status Bar/Shell" section has 18 bar tools
- Prior research: `~/hairglasses-studio/docs/research/terminal/terminal-tui-rice-ecosystem.md` (eww/waybar/AGS comparison)

### Best Hyprland rices?
- `links/hyprland/` — 200 repos sorted by stars in INDEX.md
- Top: end-4/dots-hyprland (13.7k), caelestia-dots/shell (8.6k), prasanthrangan/hyprdots (8.5k)
- Prior research: `~/hairglasses-studio/docs/strategy/open-sourcing/competitive-analysis/dotfiles.md`

### Shader sources?
- `links/shaders/` — 8 repos with GLSL shader collections
- `ghostty/shaders/SOURCES.md` — 132 shaders from 27 contributors
- Prior research: `~/hairglasses-studio/docs/research/terminal/glsl-shader-ecosystem.md` (692 lines)
- Known shader repos: 0xhckr, fearlessgeekmedia, KroneCorylus, sahaj-b, thijskok, Crackerfracks, luiscarlospando, morvoso

### NixOS Hyprland configs?
- `links/nixos/` — 208 repos (largest category)
- Top: ryan4yin/nix-config (1.9k), fufexan/dotfiles (1.1k), gvolpe/nix-config (1.1k)

### Ricing tools?
- `links/tools/` — 118 repos
- Dotfile managers: chezmoi, stow, yadm, dotbot, rcm, dotter
- Color tools: pywal, base16, wpgtk, matugen
- Prior research: `~/hairglasses-studio/docs/research/terminal/dotfile-manager-comparison-2026.md`

### Component themes?
- `links/components/` — 41 repos (rofi, SDDM, Plymouth, GRUB, cursor, icon themes)
- `links/themes/` — 53 repos (catppuccin, nord, dracula, gruvbox, tokyo-night, rose-pine)

## Prior Research (docs repo)

These files in `~/hairglasses-studio/docs/research/terminal/` contain deep analysis:

| File | Lines | Topic |
|------|-------|-------|
| terminal-tui-rice-ecosystem.md | 727 | Terminal emulator survey, eww/waybar/AGS, Quickshell/Astal |
| glsl-shader-ecosystem.md | 692 | Shadertoy, shader platforms, performance tuning |
| dotfile-manager-comparison-2026.md | ~400 | 13 tools evaluated, dotter migration failed |
| community-dotfiles-index-2026-04-05.md | NEW | This index summary (written back by Round 10) |

Strategic research:
- `docs/strategy/open-sourcing/competitive-analysis/dotfiles.md` — Stars analysis, install script patterns
- `docs/research/competitive/awesome-list-submissions.md` — 6 submission targets with PR templates

## GitHub Search Queries Used

Round 1: `hyprland dotfiles`, `hyprland rice`, `dots-hyprland`, `hyprland config eww waybar`
Round 2: `sway dotfiles`, `sway rice wayland`, `i3 dotfiles linux`, `wayland tiling dotfiles`
Round 3: `eww bar linux`, `waybar config custom modules`, `ags hyprland shell`, `quickshell astal widget`
Round 4: `ghostty config shaders`, `ghostty theme`, `terminal shader glsl crt`, `kitty alacritty wezterm config theme`
Round 5: `nixos dotfiles flake`, `home-manager dotfiles`, `nixos hyprland rice`, `nix-darwin dotfiles`
Round 6: `awesome archlinux`, `arch linux rice tools`, `unixporn dotfiles`
Round 7: `dotfile manager stow chezmoi yadm`, `linux ricing tool pywal wpgtk`
Round 8: `catppuccin dotfiles rice`, `nord gruvbox theme dotfiles`, `dracula theme dotfiles linux`, `tokyo-night rose-pine dotfiles`, `pywal base16 flavor theme generator`
Round 9: `hyprlock swaylock config theme`, `rofi theme rice collection`, `mako dunst swaync notification theme`, `sddm theme plymouth grub theme`, `cursor theme icon theme linux`

All queries used `sort:stars` and `per_page=100`.
