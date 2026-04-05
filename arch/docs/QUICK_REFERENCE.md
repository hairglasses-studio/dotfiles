# Ubuntu → Arch Linux: Application Migration Guide

## Quick Reference Table

| Purpose | Ubuntu Default/Common | Arch Recommendation | Install Command |
|---------|----------------------|---------------------|-----------------|
| **Software Center** | Ubuntu Software Center | Pamac, GNOME Software | `yay -S pamac-aur` |
| **Package Manager** | apt/snap | pacman/yay | Built-in / `yay` from AUR |
| **System Backup** | Déjà Dup | Timeshift | `sudo pacman -S timeshift` |
| **App Launcher** | Ubuntu Dash | Ulauncher, Albert | `yay -S ulauncher` |
| **Screenshot** | GNOME Screenshot | Flameshot | `sudo pacman -S flameshot` |
| **System Monitor** | GNOME System Monitor | Stacer, htop, btop | `yay -S stacer` |
| **Office Suite** | LibreOffice (snap) | LibreOffice, OnlyOffice | `sudo pacman -S libreoffice-fresh` |
| **PDF Reader** | Evince | Okular, Zathura | `sudo pacman -S okular` |
| **Email Client** | Thunderbird (snap) | Thunderbird, Mailspring | `sudo pacman -S thunderbird` |
| **Video Editor** | OpenShot | Kdenlive, DaVinci Resolve | `sudo pacman -S kdenlive` |
| **Image Editor** | GIMP (snap) | GIMP, Krita | `sudo pacman -S gimp krita` |
| **Music Player** | Rhythmbox | Spotify, Clementine | `yay -S spotify` |
| **File Manager** | Nautilus | Dolphin, Thunar, Nemo | Depends on DE |
| **Terminal** | GNOME Terminal | Kitty, Alacritty | `sudo pacman -S kitty` |
| **Text Editor** | Gedit | VS Code, Sublime Text | `yay -S visual-studio-code-bin` |

## Essential Commands Comparison

### Package Management

| Task | Ubuntu | Arch |
|------|--------|------|
| Update package list | `sudo apt update` | `sudo pacman -Sy` |
| Upgrade all packages | `sudo apt upgrade` | `sudo pacman -Syu` |
| Install package | `sudo apt install [pkg]` | `sudo pacman -S [pkg]` |
| Remove package | `sudo apt remove [pkg]` | `sudo pacman -R [pkg]` |
| Search package | `apt search [pkg]` | `pacman -Ss [pkg]` |
| Install from AUR | N/A (use PPA) | `yay -S [pkg]` |
| Clean cache | `sudo apt clean` | `sudo pacman -Sc` |
| Fix broken packages | `sudo apt --fix-broken install` | `sudo pacman -Syu` |

## Top 20 Electron/Web Apps for Daily Use

### 🏆 Tier 1: Essential (Install First)
1. **VS Code** - Code editor
2. **Discord** - Gaming/community chat
3. **Slack** - Work communication
4. **Bitwarden** - Password manager
5. **Spotify** - Music streaming

### ⭐ Tier 2: Highly Recommended
6. **Obsidian** - Note-taking & knowledge base
7. **Signal** - Secure messaging
8. **GitHub Desktop** - Git GUI
9. **Postman** - API development
10. **OBS Studio** - Streaming/recording

### 💡 Tier 3: Productivity Boosters
11. **Notion** - All-in-one workspace
12. **Telegram** - Messaging with great file sharing
13. **FreeTube** - YouTube without ads/tracking
14. **Flameshot** - Advanced screenshots
15. **Ulauncher** - Quick app launcher

### 🎨 Tier 4: Nice to Have
16. **Figma** - Design tool
17. **draw.io** - Diagram editor
18. **Element** - Matrix client
19. **Stacer** - System optimizer
20. **Balena Etcher** - USB/SD flasher

## One-Line Installation Scripts

### Basic Desktop Setup
```bash
# Core productivity
sudo pacman -S --needed firefox thunderbird libreoffice-fresh vlc gimp && yay -S --needed google-chrome visual-studio-code-bin spotify discord slack-desktop
```

### Developer Setup
```bash
# Development tools
sudo pacman -S --needed git nodejs npm docker docker-compose dbeaver && yay -S --needed github-desktop-bin postman-bin sublime-text-4 cursor-ide-bin
```

### Content Creator Setup
```bash
# Media creation
sudo pacman -S --needed obs-studio kdenlive gimp krita inkscape audacity && yay -S --needed davinci-resolve spotify youtube-music-bin
```

### Privacy-Focused Setup
```bash
# Privacy tools
sudo pacman -S --needed firefox signal-desktop element-desktop bitwarden keepassxc tor torbrowser-launcher && yay -S --needed freetube-bin session-desktop-bin
```

## Performance Tips

### Make Electron Apps Faster
```bash
# Create config file for all Electron apps
cat << EOF > ~/.config/electron-flags.conf
--enable-features=UseOzonePlatform
--ozone-platform=wayland
--enable-accelerated-video-decode
--enable-gpu-rasterization
--enable-zero-copy
--ignore-gpu-blocklist
EOF
```

### Reduce RAM Usage
```bash
# Enable zram (compressed RAM)
sudo pacman -S zram-generator
sudo systemctl enable --now systemd-zram-setup@zram0.service
```

### Speed Up Package Downloads
```bash
# Use fastest mirrors
sudo pacman -S reflector
sudo reflector --latest 20 --protocol https --sort rate --save /etc/pacman.d/mirrorlist
```

## Troubleshooting Electron Apps

### Common Issues & Fixes

| Problem | Solution |
|---------|----------|
| App won't start | Run from terminal: `/usr/bin/appname` to see errors |
| Blank window | Delete app config: `rm -rf ~/.config/appname` |
| Slow performance | Add `--disable-gpu-sandbox` to desktop file |
| No system tray | Install AppIndicator extension (GNOME) |
| Blurry text | Force pixel scaling: `--force-device-scale-factor=1` |
| High CPU usage | Disable hardware acceleration: `--disable-gpu` |
| Can't play videos | Install codecs: `sudo pacman -S ffmpeg gst-plugins-good` |

## Quick Migration Checklist

- [ ] Install yay AUR helper
- [ ] Enable multilib repository (for Steam, Wine)
- [ ] Set up Timeshift for system backups
- [ ] Install essential fonts: `sudo pacman -S ttf-liberation ttf-dejavu noto-fonts`
- [ ] Configure hardware video acceleration
- [ ] Set up printer: `sudo pacman -S cups system-config-printer`
- [ ] Enable Bluetooth: `sudo pacman -S bluez bluez-utils`
- [ ] Install codecs: `sudo pacman -S gst-plugins-base gst-plugins-good gst-plugins-bad gst-plugins-ugly`
- [ ] Configure firewall: `sudo pacman -S ufw && sudo ufw enable`
- [ ] Set up file sharing: `sudo pacman -S samba`

## Final Tips

1. **Update frequently**: `yay -Syu` daily or weekly
2. **Read the Arch Wiki**: Best Linux documentation ever
3. **Join the community**: r/archlinux, Arch Forums
4. **Keep notes**: Document what you install and configure
5. **Backup before major changes**: Use Timeshift
6. **Learn pacman**: It's more powerful than apt
7. **Embrace the AUR**: Almost everything is available
8. **Don't fear the terminal**: It's your friend on Arch
9. **Customize everything**: That's the Arch way
10. **Have fun**: Arch gives you control - use it!

Remember: With great power comes great responsibility. Arch won't hold your hand, but it won't get in your way either!
