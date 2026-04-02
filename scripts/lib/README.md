# Shared Shell Libraries

Source these from any `hg-*` script or shell environment.

## hg-core.sh

Foundational library. Provides Snazzy palette colors, formatted logging, and common paths.

```bash
source "$(dirname "$0")/lib/hg-core.sh"

hg_info "Starting process..."   # [info]  cyan
hg_ok "Done"                     # [ok]    green
hg_warn "Heads up"               # [warn]  yellow
hg_error "Something broke"       # [err]   red (to stderr)
hg_die "Fatal error" 1           # [err] + exit

hg_require git go make           # die if any command missing
```

**Exports:** `$HG_CYAN`, `$HG_GREEN`, `$HG_MAGENTA`, `$HG_YELLOW`, `$HG_RED`, `$HG_DIM`, `$HG_BOLD`, `$HG_RESET`, `$HG_DOTFILES`, `$HG_STATE_DIR`

## config.sh

Atomic config file operations. Use whenever modifying config files to prevent partial reads.

```bash
source "$(dirname "$0")/lib/config.sh"

config_atomic_write "$CONFIG" "$tmp"   # mktemp + mv pattern
config_sed_replace "$file" "s/old/new/"
config_backup "$file"                   # timestamped backup
config_reload_service "mako"            # reload via compositor
```

## compositor.sh

Cross-platform window manager abstraction. Detects Hyprland, Sway, or AeroSpace and routes IPC calls.

```bash
source "$(dirname "$0")/lib/compositor.sh"

compositor_type          # "hyprland", "sway", or "aerospace"
compositor_msg "reload"  # hyprctl reload / swaymsg reload
compositor_query "activewindow"
compositor_workspace 3   # switch to workspace 3
```
