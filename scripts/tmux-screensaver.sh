#!/usr/bin/env bash
# tmux-screensaver.sh — random screensaver for tmux lock-command

cmds=()
command -v cmatrix    &>/dev/null && cmds+=("cmatrix -ab -C cyan")
command -v neo-matrix &>/dev/null && cmds+=("neo-matrix --color=cyan --charset=katakana --speed=6 --density=0.75")
command -v pipes.sh   &>/dev/null && cmds+=("pipes.sh -t 2 -R -r 0 -p 5 -c 4")
command -v lavat      &>/dev/null && cmds+=("lavat -c cyan -s 8 -r 2")
command -v cbonsai    &>/dev/null && cmds+=("cbonsai -l -t 0.02")
command -v tty-clock  &>/dev/null && cmds+=("tty-clock -s -c -C 4 -b")

(( ${#cmds[@]} == 0 )) && exec cmatrix -ab -C cyan
exec bash -c "${cmds[$((RANDOM % ${#cmds[@]}))]}"
