#!/usr/bin/env zsh
# AFTRS Shell - Consolidated Zsh Configuration
# Enhanced configuration using consolidated plugin setup with 307 zsh plugins integrated

# Set the directory we want to store plugins and configurations
export AFTRS_SHELL_HOME="${AFTRS_SHELL_HOME:-$HOME/Docs/aftrs-shell/aftrs-shell}"
export ANTIDOTE_HOME="$AFTRS_SHELL_HOME/submodules/antidote"
export ZSH="$AFTRS_SHELL_HOME/submodules/oh-my-zsh"

# Initialize antidote
source "$ANTIDOTE_HOME/antidote.zsh"

# Initialize antidote's Zsh completions and cache
autoload -Uz compinit && compinit

# 🎨 Load RICED OUT plugins for maximum visual impact!
if [[ -f "$AFTRS_SHELL_HOME/dotfiles/zsh_plugins_riced.txt" ]]; then
  antidote load "$AFTRS_SHELL_HOME/dotfiles/zsh_plugins_riced.txt"
elif [[ -f "$AFTRS_SHELL_HOME/dotfiles/zsh_plugins_fixed.txt" ]]; then
  echo "Warning: Using fixed plugin configuration as fallback"
  antidote load "$AFTRS_SHELL_HOME/dotfiles/zsh_plugins_fixed.txt"
elif [[ -f "$AFTRS_SHELL_HOME/dotfiles/zsh_plugins_minimal.txt" ]]; then
  echo "Warning: Using minimal plugin configuration as fallback"
  antidote load "$AFTRS_SHELL_HOME/dotfiles/zsh_plugins_minimal.txt"
else
  echo "Warning: Using standard plugin configuration as fallback"
  if [[ -f "$AFTRS_SHELL_HOME/dotfiles/zsh_plugins.txt" ]]; then
    antidote load "$AFTRS_SHELL_HOME/dotfiles/zsh_plugins.txt"
  fi
fi

# Oh-My-Zsh compatibility via use-omz
if [[ -f "$AFTRS_SHELL_HOME/submodules/use-omz/use-omz.plugin.zsh" ]]; then
  source "$AFTRS_SHELL_HOME/submodules/use-omz/use-omz.plugin.zsh"
fi

# History configuration (enhanced for large collections)
export HISTFILE="$HOME/.zsh_history"
export HISTSIZE=50000
export SAVEHIST=50000
setopt EXTENDED_HISTORY
setopt INC_APPEND_HISTORY
setopt SHARE_HISTORY
setopt HIST_IGNORE_DUPS
setopt HIST_IGNORE_ALL_DUPS
setopt HIST_REDUCE_BLANKS
setopt HIST_IGNORE_SPACE
setopt HIST_VERIFY
setopt HIST_EXPIRE_DUPS_FIRST

# General Zsh options (enhanced)
setopt AUTO_CD
setopt GLOB_DOTS
setopt EXTENDED_GLOB
setopt NUMERIC_GLOB_SORT
setopt CORRECT
setopt AUTO_PUSHD
setopt PUSHD_IGNORE_DUPS
setopt CDABLE_VARS

# Key bindings (enhanced)
bindkey '^R' history-incremental-search-backward
bindkey '^[[1;5C' forward-word
bindkey '^[[1;5D' backward-word
bindkey '^[[A' history-substring-search-up
bindkey '^[[B' history-substring-search-down

# 🎨 RICED OUT ALIASES - Maximum visual impact!
alias ll='colorls -lA --sd --gs' # Use colorls if available, fallback to ls
alias la='colorls -A --sd' 
alias l='colorls --sd'
alias ls='colorls --sd' # Override ls entirely with colorls
alias llt='colorls -lA --sd --gs --tree' # Tree view with colorls

# Colorful grep aliases
alias grep='grep --color=always'
alias fgrep='fgrep --color=always'
alias egrep='egrep --color=always'

# Navigation with style
alias ..='cd ..'
alias ...='cd ../..'
alias ....='cd ../../..'
alias tree='tree -C'
alias cat='colorize_via_pygmentize' # Syntax highlighted cat

# Fun visual aliases using lolcat (if available)
alias lolls='ls --color=always | lolcat'
alias loltree='tree -C | lolcat'
alias lolcat='lolcat'

# Git aliases (enhanced beyond oh-my-zsh defaults)
alias gst='git status'
alias gd='git diff'
alias gdc='git diff --cached'
alias gl='git pull'
alias gp='git push'
alias gcmsg='git commit -m'
alias gco='git checkout'
alias gcb='git checkout -b'
alias gb='git branch'
alias gba='git branch -a'
alias glol='git log --graph --pretty=format:"%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset" --abbrev-commit'

# Docker aliases (enhanced)
alias dps='docker ps'
alias dpsa='docker ps -a'
alias di='docker images'
alias drun='docker run -it --rm'
alias dexec='docker exec -it'

# 🎨 RICED OUT FUNCTIONS - Eye candy galore!
function mkcd() {
  mkdir -p "$1" && cd "$1"
  if command -v colorls >/dev/null 2>&1; then
    colorls --sd
  else
    ls --color=auto
  fi
}

# ASCII art welcome message
function show_welcome() {
  echo "
┌─┐┌─┐┌┬┐┬─┐┌─┐   ┌─┐┬ ┬┌─┐┬  ┬  
├─┤├┤  │ ├┬┘└─┐───└─┐├─┤├┤ │  │  
┴ ┴└   ┴ ┴└─└─┘   └─┘┴ ┴└─┘┴─┘┴─┘
   🎨 MAXIMUM RICE ACHIEVED 🎨   
" | lolcat 2>/dev/null || cat
  echo "Welcome to your beautiful shell! ✨"
}

# System info with style
function sysinfo() {
  if command -v screenfetch >/dev/null 2>&1; then
    screenfetch
  elif command -v neofetch >/dev/null 2>&1; then
    neofetch
  else
    echo "System: $(uname -s)"
    echo "Hostname: $(hostname)"
    echo "User: $(whoami)"
    echo "Date: $(date)"
  fi
}

# Colorful directory listing with info
function lls() {
  if command -v colorls >/dev/null 2>&1; then
    colorls -lA --sd --gs "$@"
  else
    ls -la --color=always "$@"
  fi
}

# Matrix-style directory listing (if lolcat available)
function matrix-ls() {
  if command -v lolcat >/dev/null 2>&1; then
    ls -la | lolcat
  else
    ls -la --color=always
  fi
}

function extract() {
  if [[ -f $1 ]]; then
    case $1 in
      *.tar.bz2)   tar xjf $1     ;;
      *.tar.gz)    tar xzf $1     ;;
      *.bz2)       bunzip2 $1     ;;
      *.rar)       unrar e $1     ;;
      *.gz)        gunzip $1      ;;
      *.tar)       tar xf $1      ;;
      *.tbz2)      tar xjf $1     ;;
      *.tgz)       tar xzf $1     ;;
      *.zip)       unzip $1       ;;
      *.Z)         uncompress $1  ;;
      *.7z)        7z x $1        ;;
      *)           echo "'$1' cannot be extracted via extract()" ;;
    esac
  else
    echo "'$1' is not a valid file"
  fi
}

# FZF configuration (if available)
if command -v fzf > /dev/null 2>&1; then
  export FZF_DEFAULT_OPTS="--height 40% --layout=reverse --border"
  export FZF_CTRL_T_OPTS="--preview 'cat {}' --preview-window down:3:hidden:wrap --bind '?:toggle-preview'"
  export FZF_ALT_C_OPTS="--preview 'tree -C {} | head -200'"
fi

# Starship prompt initialization (if using starship from local plugins)
if command -v starship > /dev/null 2>&1; then
  eval "$(starship init zsh)"
fi

# Load any local customizations
if [[ -f "$HOME/.zshrc.local" ]]; then
  source "$HOME/.zshrc.local"
fi

# Performance optimizations for large plugin collections
zstyle ':completion:*' use-cache on
zstyle ':completion:*' cache-path ~/.zsh/cache

# AFTRS Shell Information
echo "AFTRS Shell loaded with $(echo "$AFTRS_SHELL_HOME/dotfiles/zsh_plugins_consolidated.txt" | wc -l 2>/dev/null || echo "unknown") consolidated plugins"
