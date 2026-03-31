#!/usr/bin/env bash
# Animated braille spinner for tmux status bar
chars=('⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏')
echo "${chars[$(($(date +%s) % 10))]}"
