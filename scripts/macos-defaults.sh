#!/usr/bin/env bash
set -euo pipefail

# macOS defaults for a clean rice
# Run: bash ~/dotfiles/scripts/macos-defaults.sh

echo "Applying macOS defaults..."

# ── Dock ──────────────────────────────────────
defaults write com.apple.dock tilesize -int 48
defaults write com.apple.dock autohide -bool true
defaults write com.apple.dock autohide-delay -float 0
defaults write com.apple.dock autohide-time-modifier -float 0.3
defaults write com.apple.dock show-recents -bool false
defaults write com.apple.dock mineffect -string "scale"
defaults write com.apple.dock minimize-to-application -bool true
defaults write com.apple.dock launchanim -bool false

# ── Mission Control ───────────────────────────
defaults write com.apple.dock mru-spaces -bool false

# ── Finder ────────────────────────────────────
defaults write com.apple.finder AppleShowAllFiles -bool true
defaults write com.apple.finder ShowPathbar -bool true
defaults write com.apple.finder ShowStatusBar -bool true
defaults write com.apple.finder _FXShowPosixPathInTitle -bool true
defaults write com.apple.finder FXDefaultSearchScope -string "SCcf"

# ── Keyboard ──────────────────────────────────
defaults write NSGlobalDomain KeyRepeat -int 2
defaults write NSGlobalDomain InitialKeyRepeat -int 15
defaults write NSGlobalDomain ApplePressAndHoldEnabled -bool false

# ── Screenshots ───────────────────────────────
mkdir -p "$HOME/Desktop/screenshots"
defaults write com.apple.screencapture location -string "$HOME/Desktop/screenshots"
defaults write com.apple.screencapture type -string "png"
defaults write com.apple.screencapture disable-shadow -bool true

# ── Restart affected services ─────────────────
killall Dock 2>/dev/null || true
killall Finder 2>/dev/null || true

echo "Done. Some changes may require logout/restart."
