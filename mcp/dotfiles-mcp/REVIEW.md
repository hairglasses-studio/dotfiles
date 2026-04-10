# Review Guidelines — dotfiles-mcp

Inherits from org-wide [REVIEW.md](https://github.com/hairglasses-studio/.github/blob/main/REVIEW.md).

## Additional Focus
- **Compositor abstraction**: Tools must work on both Hyprland and Sway — check for compositor-specific assumptions
- **Shader file I/O**: Validate paths before read/write, prevent path traversal outside shader directories
- **Bluetooth/MIDI**: Handle device disconnect gracefully, don't panic on missing devices
- **Kitty visual state writes**: Keep CRTty shader and Kitty theme state updates atomic so new windows do not observe partial state
- **Public contract snapshots**: Any tool/resource/prompt change must keep `.well-known/mcp.json` and `snapshots/contract/*` in sync
