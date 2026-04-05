# Troubleshooting Guide

## Common Installation Issues

### Script Permissions
**Problem**: `Permission denied` when running scripts
```bash
# Solution
chmod +x install.sh
chmod +x scripts/*.sh
```

### AUR Helper Not Found
**Problem**: `yay: command not found`
```bash
# Solution - Reinstall yay
cd /tmp
git clone https://aur.archlinux.org/yay.git
cd yay
makepkg -si
```

### Package Conflicts
**Problem**: `error: failed to prepare transaction (conflicting dependencies)`
```bash
# Solution - Remove conflicting package
sudo pacman -R conflicting-package
# Or force overwrite (use carefully)
sudo pacman -S package --overwrite '*'
```

## Application-Specific Issues

### Electron Apps

#### Blank Window or Won't Start
```bash
# Clear app cache
rm -rf ~/.config/[appname]

# Run with debugging
[appname] --enable-logging --v=1

# Disable GPU acceleration
[appname] --disable-gpu
```

#### High CPU/RAM Usage
```bash
# Add to ~/.config/electron-flags.conf
--max-old-space-size=2048
--js-flags="--max-old-space-size=2048"
```

#### Wayland Issues
```bash
# Force X11 mode
GDK_BACKEND=x11 [appname]

# Or add to .desktop file
Exec=env GDK_BACKEND=x11 /usr/bin/[appname]
```

### VS Code

#### Extensions Not Installing
```bash
# Clear extension cache
rm -rf ~/.vscode/extensions
code --list-extensions --show-versions

# Reinstall from command line
code --install-extension extension-id
```

#### Can't Open Files from Terminal
```bash
# Add to PATH
export PATH="/usr/bin:$PATH"
# Or create symlink
sudo ln -s /usr/bin/code /usr/local/bin/code
```

### Discord

#### No Audio
```bash
# Install PulseAudio support
sudo pacman -S pulseaudio pulseaudio-alsa

# Reset Discord audio settings
rm -rf ~/.config/discord/0.0.*/modules/discord_voice
```

#### Screen Share Black Screen
```bash
# Install pipewire
sudo pacman -S pipewire pipewire-pulse xdg-desktop-portal

# For X11
sudo pacman -S xdg-desktop-portal-gtk
# For Wayland
sudo pacman -S xdg-desktop-portal-wlr
```

### Steam

#### Won't Launch
```bash
# Install 32-bit libraries
sudo pacman -S lib32-mesa lib32-vulkan-icd-loader lib32-vulkan-radeon

# For NVIDIA
sudo pacman -S lib32-nvidia-utils

# Reset Steam
steam --reset
```

#### Controller Not Detected
```bash
# Add user to input group
sudo usermod -aG input $USER

# Install Steam controller support
sudo pacman -S steam-native-runtime
```

### Docker

#### Permission Denied
```bash
# Add user to docker group
sudo usermod -aG docker $USER
# Log out and back in

# Or use sudo (not recommended)
sudo docker run...
```

#### Cannot Connect to Docker Daemon
```bash
# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker

# Check status
sudo systemctl status docker
```

## System Issues

### Slow Boot

#### Analyze Boot Time
```bash
systemd-analyze
systemd-analyze blame
systemd-analyze critical-chain
```

#### Disable Unnecessary Services
```bash
# Disable NetworkManager-wait-online
sudo systemctl disable NetworkManager-wait-online.service

# Disable Plymouth (if using)
sudo systemctl disable plymouth
```

### High Memory Usage

#### Check Memory Consumers
```bash
# Top memory users
ps aux --sort=-%mem | head

# System memory
free -h

# Clear cache (safe)
sudo sync && echo 3 | sudo tee /proc/sys/vm/drop_caches
```

#### Enable ZRAM
```bash
sudo pacman -S zram-generator
sudo systemctl enable --now systemd-zram-setup@zram0.service
```

### No Sound

#### Check Audio System
```bash
# List audio devices
aplay -l

# Check PulseAudio
pulseaudio --check
pulseaudio --start

# Or use PipeWire
sudo pacman -S pipewire pipewire-pulse pipewire-alsa
systemctl --user enable --now pipewire pipewire-pulse
```

### Wi-Fi Issues

#### No Wi-Fi Adapter
```bash
# Check for hardware
lspci | grep -i wireless
lsusb | grep -i wireless

# Install drivers
# For Intel
sudo pacman -S linux-firmware

# For Broadcom
yay -S broadcom-wl
```

#### Can't Connect
```bash
# Restart NetworkManager
sudo systemctl restart NetworkManager

# Use nmcli
nmcli device wifi list
nmcli device wifi connect "SSID" password "password"
```

### Display Issues

#### Screen Tearing
```bash
# For Intel
echo 'options i915 enable_fbc=0' | sudo tee /etc/modprobe.d/i915.conf

# For NVIDIA
# Add to /etc/X11/xorg.conf.d/20-nvidia.conf
Section "Device"
    Identifier "NVIDIA Card"
    Driver "nvidia"
    Option "NoLogo" "true"
    Option "TripleBuffer" "true"
EndSection
```

#### Wrong Resolution
```bash
# List available resolutions
xrandr

# Set resolution
xrandr --output HDMI-1 --mode 1920x1080 --rate 60

# Make permanent - add to ~/.xprofile
xrandr --output HDMI-1 --mode 1920x1080 --rate 60
```

## Package Management

### Pacman Locked
```bash
# Remove lock file
sudo rm /var/lib/pacman/db.lck
```

### Corrupt Package Database
```bash
# Refresh package database
sudo pacman -Syy

# Force refresh
sudo pacman -Syyu
```

### Keyring Issues
```bash
# Reinitialize keyring
sudo rm -rf /etc/pacman.d/gnupg
sudo pacman-key --init
sudo pacman-key --populate archlinux
sudo pacman -Sy archlinux-keyring
```

### Mirror Issues
```bash
# Update mirrorlist
sudo pacman -S reflector
sudo reflector --latest 20 --protocol https --sort rate --save /etc/pacman.d/mirrorlist
```

## File System Issues

### Read-Only File System
```bash
# Remount as read-write
sudo mount -o remount,rw /

# Check file system
sudo fsck /dev/sdXY
```

### No Space Left
```bash
# Find large files
du -h / 2>/dev/null | grep '[0-9]G'

# Clean package cache
sudo pacman -Sc
yay -Sc

# Clean journal
sudo journalctl --vacuum-size=100M
```

## Desktop Environment Issues

### GNOME

#### Extensions Not Working
```bash
# Restart GNOME Shell
Alt+F2, type 'r', press Enter

# Reset extensions
gsettings reset org.gnome.shell enabled-extensions
```

### KDE

#### Plasma Crash
```bash
# Reset Plasma config
mv ~/.config/plasma* ~/.config/backup/
kquitapp5 plasmashell && kstart5 plasmashell
```

### XFCE

#### Panel Missing
```bash
# Reset panels
xfce4-panel --quit
pkill xfconfd
rm -rf ~/.config/xfce4/panel
xfce4-panel
```

## Performance Optimization

### Enable Faster Boots
```bash
# Disable unused services
systemctl list-unit-files --state=enabled
sudo systemctl disable [unnecessary-service]

# Enable readahead
sudo systemctl enable systemd-readahead-collect systemd-readahead-replay
```

### Reduce Swappiness
```bash
# Check current value
cat /proc/sys/vm/swappiness

# Set to 10 (for desktop use)
echo "vm.swappiness=10" | sudo tee /etc/sysctl.d/99-swappiness.conf
```

### Enable TRIM for SSDs
```bash
# Enable periodic TRIM
sudo systemctl enable --now fstrim.timer

# Check if enabled
systemctl status fstrim.timer
```

## Recovery Options

### Boot to TTY
```
If GUI fails:
1. Press Ctrl+Alt+F2 (or F3-F6)
2. Login with username and password
3. Fix issues from command line
4. Restart display manager:
   sudo systemctl restart gdm  # For GNOME
   sudo systemctl restart sddm # For KDE
   sudo systemctl restart lightdm # For XFCE
```

### Chroot from Live USB
```bash
# Boot from Arch USB
# Mount root partition
mount /dev/sdXY /mnt
# Mount boot partition (if separate)
mount /dev/sdXZ /mnt/boot
# Chroot
arch-chroot /mnt
# Fix issues, then exit and reboot
```

### Timeshift Restore
```bash
# From live USB
sudo timeshift --restore
# Select snapshot and follow prompts
```

## Getting Help

### Resources
- Arch Wiki: https://wiki.archlinux.org/
- Arch Forums: https://bbs.archlinux.org/
- Reddit: r/archlinux
- IRC: #archlinux on Libera.Chat

### Useful Commands for Debugging
```bash
# System logs
journalctl -xe
journalctl -b  # Current boot
journalctl -b -1  # Previous boot

# Hardware info
lspci
lsusb
lscpu
inxi -Fxz

# Service status
systemctl status [service]
systemctl --failed

# Package info
pacman -Qi [package]  # Installed package info
pacman -Ql [package]  # List package files
pacman -Qo [file]     # Which package owns file
```

## Prevention Tips

1. **Always backup before major changes**
   - Use Timeshift for system snapshots
   - Backup personal files separately

2. **Read update announcements**
   - Check Arch news before major updates
   - Review AUR comments for issues

3. **Test in VM first**
   - Try risky operations in virtual machine
   - Test new configurations safely

4. **Keep rescue tools ready**
   - Arch installation USB
   - Live USB with tools
   - Backup of important configs

5. **Document your setup**
   - Keep notes of customizations
   - Save working configurations
   - Use version control for dotfiles
