# Contributing

## Setup

```bash
git clone https://github.com/hairglasses-studio/dotfiles
cd dotfiles
./install.sh
```

## Structure

- `kitty/` — Kitty config, CRTty shader assets, and theme playlists
- `hyprland/` — Hyprland window manager config  
- `ironbar/` — Ironbar menubar config
- `zsh/` — Shell config + starship prompt
- `scripts/` — Helper scripts + shared libraries

## Shader Contributions

Kitty DarkWindow shaders live in `kitty/shaders/darkwindow/`. To add a new shader:
1. Place your `.glsl` file in the directory with a header comment (`// Shader attribution: <author>` + `// (<category>) — <one-line description>`)
2. Add the shader name to the relevant playlist(s) under `kitty/shaders/playlists/*.txt`
3. Test with `kitty/shaders/bin/shader-test.sh <name>.glsl` (requires glslangValidator)
4. Submit a PR

## Code Style

- Shell scripts: use `shellcheck`
- GLSL: self-contained (no external includes at runtime)
- Configs: follow existing formatting
