#!/bin/bash

# Electron Application Optimization Script
# Applies performance optimizations to all Electron-based applications

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}==>${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   print_warning "Running as root, will apply system-wide"
   CONFIG_DIR="/etc"
else
   CONFIG_DIR="$HOME/.config"
fi

print_status "Applying Electron optimizations..."

# Create electron flags configuration
cat > "$CONFIG_DIR/electron-flags.conf" << 'EOF'
# Electron Application Optimization Flags
# These flags improve performance and enable hardware acceleration

# Enable Wayland support (if using Wayland)
--enable-features=UseOzonePlatform
--ozone-platform=wayland

# Hardware acceleration flags
--enable-accelerated-video-decode
--enable-gpu-rasterization
--enable-zero-copy
--ignore-gpu-blocklist
--enable-accelerated-mjpeg-decode
--enable-accelerated-video
--disable-gpu-driver-bug-workarounds

# Performance improvements
--disable-frame-rate-limit
--disable-gpu-vsync
--max-active-webgl-contexts=100
--enable-features=VaapiVideoDecoder

# Memory optimizations
--max-old-space-size=4096
--js-flags="--max-old-space-size=4096"

# Security (optional - comment out if causes issues)
--enable-features=WebRTCPipeWireCapturer

# Disable telemetry and analytics
--disable-features=MediaSessionService
--disable-breakpad
--disable-component-update
--disable-features=CalculateNativeWinOcclusion
EOF

print_success "Created electron-flags.conf in $CONFIG_DIR"

# Create Chrome flags configuration (for Chrome/Chromium-based apps)
cat > "$CONFIG_DIR/chrome-flags.conf" << 'EOF'
# Chrome/Chromium Optimization Flags

# Hardware acceleration
--enable-gpu-rasterization
--enable-zero-copy
--ignore-gpu-blocklist
--enable-accelerated-video-decode
--enable-features=VaapiVideoDecoder

# Performance
--disable-gpu-driver-bug-workarounds
--enable-features=UseOzonePlatform
--ozone-platform=wayland
EOF

print_success "Created chrome-flags.conf in $CONFIG_DIR"

# Apply optimizations to specific applications
print_status "Applying app-specific optimizations..."

# VS Code
if [ -f "$CONFIG_DIR/Code/User/settings.json" ]; then
    print_status "Optimizing VS Code..."
    # Backup existing settings
    cp "$CONFIG_DIR/Code/User/settings.json" "$CONFIG_DIR/Code/User/settings.json.backup"
    
    # Add GPU acceleration settings (you might want to merge these manually)
    cat > "$CONFIG_DIR/Code/User/gpu-settings.json" << 'EOF'
{
    "window.titleBarStyle": "custom",
    "editor.smoothScrolling": true,
    "terminal.integrated.gpuAcceleration": "on",
    "workbench.list.smoothScrolling": true
}
EOF
    print_success "VS Code optimization settings saved to gpu-settings.json"
    print_warning "Manually merge gpu-settings.json with your settings.json"
fi

# Discord
if [ -d "$CONFIG_DIR/discord" ]; then
    print_status "Optimizing Discord..."
    mkdir -p "$CONFIG_DIR/discord"
    echo '{"SKIP_HOST_UPDATE":true,"DANGEROUS_ENABLE_DEVTOOLS_ONLY_ENABLE_IF_YOU_KNOW_WHAT_YOURE_DOING":false}' > "$CONFIG_DIR/discord/settings.json"
    print_success "Discord optimized"
fi

# Create desktop file modifications directory
mkdir -p "$HOME/.local/share/applications"

# Function to modify desktop files
modify_desktop_file() {
    local app_name=$1
    local desktop_file=$2
    local exec_flags=$3
    
    if [ -f "/usr/share/applications/$desktop_file" ]; then
        cp "/usr/share/applications/$desktop_file" "$HOME/.local/share/applications/"
        sed -i "s/^Exec=\(.*\)$/Exec=\1 $exec_flags/" "$HOME/.local/share/applications/$desktop_file"
        print_success "$app_name desktop file optimized"
    fi
}

# Optimize desktop files for common Electron apps
print_status "Modifying application launchers..."

# VS Code
modify_desktop_file "VS Code" "visual-studio-code.desktop" "--enable-features=UseOzonePlatform --ozone-platform=wayland"

# Discord
modify_desktop_file "Discord" "discord.desktop" "--enable-features=UseOzonePlatform --ozone-platform=wayland"

# Slack
modify_desktop_file "Slack" "slack.desktop" "--enable-features=UseOzonePlatform --ozone-platform=wayland"

# Signal
modify_desktop_file "Signal" "signal-desktop.desktop" "--enable-features=UseOzonePlatform --ozone-platform=wayland"

# Spotify
modify_desktop_file "Spotify" "spotify.desktop" "--enable-features=UseOzonePlatform --ozone-platform=wayland"

# System-wide optimizations
if [[ $EUID -eq 0 ]]; then
    print_status "Applying system-wide optimizations..."
    
    # Increase inotify watchers for VS Code and other IDEs
    echo "fs.inotify.max_user_watches=524288" > /etc/sysctl.d/40-max-user-watches.conf
    sysctl -p /etc/sysctl.d/40-max-user-watches.conf
    
    print_success "System-wide optimizations applied"
else
    print_warning "Run as root to apply system-wide optimizations"
fi

# Hardware video acceleration check
print_status "Checking hardware video acceleration..."

check_gpu_driver() {
    if lspci | grep -i nvidia > /dev/null; then
        print_status "NVIDIA GPU detected"
        print_warning "Ensure nvidia-utils is installed: sudo pacman -S nvidia-utils"
    elif lspci | grep -i amd > /dev/null; then
        print_status "AMD GPU detected"
        print_warning "Ensure mesa drivers are installed: sudo pacman -S mesa lib32-mesa"
    elif lspci | grep -i intel > /dev/null; then
        print_status "Intel GPU detected"
        print_warning "Ensure Intel drivers are installed: sudo pacman -S intel-media-driver libva-intel-driver"
    fi
}

check_gpu_driver

# Verify VA-API support
if command -v vainfo > /dev/null; then
    print_status "VA-API Info:"
    vainfo 2>/dev/null | grep "Driver version" || print_warning "VA-API not properly configured"
else
    print_warning "vainfo not installed. Install with: sudo pacman -S libva-utils"
fi

# Environment variables setup
print_status "Setting up environment variables..."

ENV_FILE="$HOME/.config/environment.d/electron.conf"
mkdir -p "$HOME/.config/environment.d"

cat > "$ENV_FILE" << 'EOF'
# Electron Environment Variables

# Force Wayland for Electron apps (if using Wayland)
ELECTRON_OZONE_PLATFORM_HINT=wayland

# Enable hardware acceleration
LIBVA_DRIVER_NAME=iHD  # For Intel, use 'radeonsi' for AMD, 'nvidia' for NVIDIA

# Node.js memory limit for Electron apps
NODE_OPTIONS="--max-old-space-size=4096"
EOF

print_success "Environment variables configured"

# Summary
echo ""
echo "════════════════════════════════════════════════════════"
print_success "Electron optimization complete!"
echo ""
echo "Applied optimizations:"
echo "  • Hardware acceleration flags"
echo "  • Wayland support (if applicable)"
echo "  • Memory optimizations"
echo "  • GPU acceleration"
echo "  • Application-specific tweaks"
echo ""
print_warning "Some changes require application restart or logout/login"
echo ""
echo "To verify optimizations are working:"
echo "  1. Launch an Electron app from terminal"
echo "  2. Check for messages about GPU acceleration"
echo "  3. Monitor CPU/GPU usage for improvements"
echo ""
echo "If you experience issues, remove:"
echo "  • $CONFIG_DIR/electron-flags.conf"
echo "  • $HOME/.local/share/applications/<app>.desktop"
echo "════════════════════════════════════════════════════════"
