#!/usr/bin/env zsh
# Cross-platform aliases optimized for LLM agents
# Compatible with both macOS and Linux

# Helper function to check if command exists
function cmd_exists() {
    command -v "$1" > /dev/null 2>&1
}

# Cross-platform ls configuration
if cmd_exists eza; then
    alias ls='eza --color=always --group-directories-first --icons'
    alias ll='eza -l --color=always --group-directories-first --icons'
    alias la='eza -la --color=always --group-directories-first --icons'
    alias lt='eza -T --color=always --group-directories-first --icons'
elif cmd_exists exa; then
    alias ls='exa --color=always --group-directories-first --icons'
    alias ll='exa -l --color=always --group-directories-first --icons'
    alias la='exa -la --color=always --group-directories-first --icons'
    alias lt='exa -T --color=always --group-directories-first --icons'
elif cmd_exists lsd; then
    alias ls='lsd --color=always --group-dirs first'
    alias ll='lsd -l --color=always --group-dirs first'
    alias la='lsd -la --color=always --group-dirs first'
    alias lt='lsd --tree --color=always --group-dirs first'
else
    # Fallback to standard ls with cross-platform options
    if [[ "$OSTYPE" == "darwin"* ]]; then
        alias ls='ls -G'
        alias ll='ls -lG'
        alias la='ls -laG'
    else
        alias ls='ls --color=auto --group-directories-first'
        alias ll='ls -l --color=auto --group-directories-first'
        alias la='ls -la --color=auto --group-directories-first'
    fi
    alias lt='ls -ltrh'
fi

# File and directory navigation
alias ..='cd ..'
alias ...='cd ../..'
alias ....='cd ../../..'
alias .....='cd ../../../..'
alias ~='cd ~'
alias -- -='cd -'

# Platform-specific aliases loaded separately:
#   aliases.darwin.zsh — macOS (AeroSpace, SketchyBar, brew, etc.)
#   aliases.linux.zsh  — Linux (wl-copy, hyprctl, pacman/yay, etc.)

# File operations (cross-platform)
alias mkdir='mkdir -pv'
alias cp='cp -iv'
alias mv='mv -iv'
alias rm='rm -iv'
alias ln='ln -iv'

# Search and grep
alias grep='grep --color=auto'
alias grepi='grep -i --color=auto'
alias fgrep='fgrep --color=auto'
alias egrep='egrep --color=auto'

# Text processing
if cmd_exists bat; then
    alias cat='bat --paging=never'
    alias catn='bat --paging=never'
    alias less='bat --paging=always'
elif cmd_exists batcat; then
    alias cat='batcat --paging=never'
    alias catn='batcat --paging=never'
    alias less='batcat --paging=always'
fi

# Editor aliases
alias vim='nvim'
alias vi='nvim'
alias nano='nvim'
alias edit='nvim'

# Git aliases (LLM-friendly)
alias g='git'
alias ga='git add'
alias gaa='git add --all'
alias gap='git add --patch'
alias gb='git branch'
alias gba='git branch --all'
alias gbd='git branch --delete'
alias gbD='git branch --delete --force'
alias gc='git commit'
alias gcm='git commit --message'
alias gca='git commit --all'
alias gcam='git commit --all --message'
alias gco='git checkout'
alias gcb='git checkout -b'
alias gd='git diff'
alias gds='git diff --staged'
alias gf='git fetch'
alias gfa='git fetch --all'
alias gl='git log --oneline --graph --decorate'
alias gla='git log --oneline --graph --decorate --all'
alias gll='git log --graph --pretty=format:"%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset" --abbrev-commit'
alias gp='git push'
alias gpo='git push origin'
alias gpl='git pull'
alias gr='git remote'
alias grv='git remote --verbose'
alias gs='git status'
alias gss='git status --short'
alias gst='git stash'
alias gsta='git stash apply'
alias gstd='git stash drop'
alias gstl='git stash list'
alias gstp='git stash pop'
alias gsts='git stash show --text'
alias gsu='git stash && git pull && git stash pop'
alias gw='git show'
alias gwt='git worktree'

# Docker aliases (LLM-friendly)
if cmd_exists docker; then
    alias d='docker'
    alias dc='docker-compose'
    alias dcu='docker-compose up'
    alias dcd='docker-compose down'
    alias dcl='docker-compose logs'
    alias dps='docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"'
    alias dpsa='docker ps --all --format "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}"'
    alias di='docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"'
    alias dv='docker volume ls'
    alias dn='docker network ls'
    alias dlogs='docker logs --follow'
    alias dexec='docker exec --interactive --tty'
    alias dclean='docker system prune --all --force'
    alias dstop='docker stop $(docker ps -q)'
    alias drm='docker rm $(docker ps -aq)'
    alias drmi='docker rmi $(docker images -q)'
fi

# Kubernetes aliases (LLM-friendly)
if cmd_exists kubectl; then
    alias k='kubectl'
    alias kaf='kubectl apply --filename'
    alias kca='kubectl config current-context'
    alias kccc='kubectl config get-contexts'
    alias kcdc='kubectl config delete-context'
    alias kcsc='kubectl config set-context'
    alias kcuc='kubectl config use-context'
    alias kdel='kubectl delete'
    alias kdf='kubectl delete --filename'
    alias kdesc='kubectl describe'
    alias ke='kubectl exec --stdin --tty'
    alias kg='kubectl get'
    alias kga='kubectl get all'
    alias kgd='kubectl get deployment'
    alias kgi='kubectl get ingress'
    alias kgn='kubectl get nodes'
    alias kgp='kubectl get pods'
    alias kgr='kubectl get replicaset'
    alias kgs='kubectl get service'
    alias kl='kubectl logs'
    alias klf='kubectl logs --follow'
    alias kp='kubectl proxy'
    alias kpf='kubectl port-forward'
    alias kr='kubectl rollout'
    alias krs='kubectl rollout status'
    alias kru='kubectl rollout undo'
fi

# Python aliases (cross-platform)
if cmd_exists python3; then
    alias python='python3'
    alias pip='pip3'
fi
alias pv='python --version'
alias py='python'
alias serve='python -m http.server'
alias venv='python -m venv'
alias activate='source ./venv/bin/activate'

# Node.js aliases
if cmd_exists node; then
    alias nv='node --version'
    alias npmg='npm list --global --depth=0'
    alias npml='npm list --depth=0'
    alias npms='npm start'
    alias npmt='npm test'
    alias npmr='npm run'
    alias npmi='npm install'
    alias npmu='npm update'
fi

# System information (platform-specific aliases in aliases.{darwin,linux}.zsh)

# Network utilities (cross-platform)
alias ports='netstat -tuln'
alias myip='curl -s https://ipinfo.io/ip'
alias localip='hostname -I | cut -d" " -f1'
alias ping='ping -c 5'
alias fastping='ping -c 100 -s.2'

# Archive operations (cross-platform)
alias tarx='tar -xvf'
alias tarc='tar -cvf'
alias tarz='tar -czvf'
alias tarj='tar -cjvf'

# LLM-friendly development aliases
alias serve8000='python -m http.server 8000'
alias serve3000='python -m http.server 3000'
alias jsonpp='python -m json.tool'
alias urlencode='python -c "import sys, urllib.parse as ul; print(ul.quote_plus(sys.argv[1]))"'
alias urldecode='python -c "import sys, urllib.parse as ul; print(ul.unquote_plus(sys.argv[1]))"'

# Process management
alias psg='ps aux | grep'
alias psk='ps aux | grep'
alias killall='pkill'

# Quick directory access
alias dot='cd $HOME/dotfiles'
alias proj='cd $HOME/Projects'
alias docs='cd $HOME/Documents'
alias dl='cd $HOME/Downloads'
alias dt='cd $HOME/Desktop'

# Configuration file shortcuts
alias zshconfig='nvim ~/.zshrc'
alias zshreload='source ~/.zshrc'
alias aliasconfig='nvim ~/dotfiles/zsh/aliases.zsh'
alias nvimconfig='nvim ~/.config/nvim/init.vim'
alias gitconfig='nvim ~/.gitconfig'

# System utilities
alias h='history'
alias j='jobs'
alias path='echo -e ${PATH//:/\\n}'
alias now='date +"%T"'
alias nowtime='date +"%d-%m-%Y %T"'
alias nowdate='date +"%d-%m-%Y"'

# File finding and searching
alias ff='find . -type f -name'
alias fdd='find . -type d -name'
if cmd_exists rg; then
    alias search='rg --smart-case --follow --hidden'
elif cmd_exists ag; then
    alias search='ag --smart-case --follow --hidden'
else
    alias search='grep -r --exclude-dir=.git --exclude-dir=node_modules --exclude-dir=.env'
fi

# Clean up functions
alias cleanup='find . -type f -name "*.DS_Store" -ls -delete'
# emptytrash defined per-platform in aliases.{darwin,linux}.zsh

# Terraform aliases (if available)
if cmd_exists terraform; then
    alias tf='terraform'
    alias tfi='terraform init'
    alias tfp='terraform plan'
    alias tfa='terraform apply'
    alias tfd='terraform destroy'
    alias tfv='terraform validate'
    alias tff='terraform fmt'
    alias tfs='terraform state'
    alias tfo='terraform output'
fi

# Terragrunt aliases (if available)
if cmd_exists terragrunt; then
    alias tg='terragrunt'
    alias tgi='terragrunt init'
    alias tgp='terragrunt plan'
    alias tga='terragrunt apply'
    alias tgd='terragrunt destroy'
    alias tgv='terragrunt validate'
    alias tgf='terragrunt hclfmt'
fi

# AWS CLI aliases (if available)
if cmd_exists aws; then
    alias aws='aws --no-cli-pager'
    alias awswho='aws sts get-caller-identity'
    alias awsregion='aws configure get region'
fi

# tldr aliases (if available)
if cmd_exists tldr; then
    alias tldr='tldr --theme=base16'
    alias help='tldr'
fi

# Note: z/zoxide integration handled in zshrc via `eval "$(zoxide init zsh)"`

# Custom functions for LLM agents
llm_help() {
    echo "=== LLM-Friendly Commands ==="
    echo "llm_context - Show current directory context"
    echo "search <pattern> - Smart search across files"
    echo "serve <port> - Start HTTP server (default: 8000)"
    echo "myip - Show external IP address"
    echo "localip - Show local IP address"
    echo "ports - Show listening ports"
    echo "jsonpp - Pretty print JSON from stdin"
    echo "cleanup - Remove .DS_Store and temp files"
    echo "=== Git Shortcuts ==="
    echo "gs - git status"
    echo "ga - git add"
    echo "gc - git commit"
    echo "gp - git push"
    echo "gl - git log (pretty)"
    echo "gd - git diff"
    echo "=== Docker/K8s ==="
    echo "dps - docker ps (formatted)"
    echo "k - kubectl"
    echo "kgp - kubectl get pods"
    echo "=== Development ==="
    echo "py - python"
    echo "venv - create virtual environment"
    echo "activate - activate venv"
}

# Function to show file/directory information
info() {
    if [[ -f "$1" ]]; then
        echo "File: $1"
        file "$1"
        ls -lah "$1"
        if cmd_exists bat; then
            echo "--- Content Preview ---"
            bat --line-range=1:20 "$1"
        fi
    elif [[ -d "$1" ]]; then
        echo "Directory: $1"
        ls -lah "$1"
        echo "--- Contents ---"
        ls -la "$1"
    else
        echo "Path does not exist: $1"
    fi
}

# Quick backup function
backup() {
    if [[ -z "$1" ]]; then
        echo "Usage: backup <file_or_directory>"
        return 1
    fi
    cp -r "$1" "$1.backup.$(date +%Y%m%d_%H%M%S)"
    echo "Backup created: $1.backup.$(date +%Y%m%d_%H%M%S)"
}

# Extract function for various archive formats
extract() {
    if [[ -z "$1" ]]; then
        echo "Usage: extract <archive>"
        return 1
    fi
    if [[ -f "$1" ]]; then
        case "$1" in
            *.tar.bz2) tar xjf "$1" ;;
            *.tar.gz) tar xzf "$1" ;;
            *.bz2) bunzip2 "$1" ;;
            *.rar) unrar x "$1" ;;
            *.gz) gunzip "$1" ;;
            *.tar) tar xf "$1" ;;
            *.tbz2) tar xjf "$1" ;;
            *.tgz) tar xzf "$1" ;;
            *.zip) unzip "$1" ;;
            *.Z) uncompress "$1" ;;
            *.7z) 7z x "$1" ;;
            *) echo "'$1' cannot be extracted via extract()" ;;
        esac
    else
        echo "'$1' is not a valid file"
    fi
}

# ── Go development ─────────────────────────────
if cmd_exists go; then
    alias gorun='go run .'
    alias gobuild='go build -v ./...'
    alias gotest='go test -v ./...'
    alias gorace='go test -race -v ./...'
    alias gocover='go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out'
    alias golint='golangci-lint run ./...'
    alias gomod='go mod tidy && go mod verify'
    alias govet='go vet ./...'
    alias gogen='go generate ./...'
    alias gowork='go work sync'
fi

# ── Protobuf / gRPC ───────────────────────────
if cmd_exists protoc; then
    alias protogen='protoc --go_out=. --go-grpc_out=. *.proto'
fi
if cmd_exists grpcurl; then
    alias grpcls='grpcurl -plaintext localhost:50051 list'
    alias grpcdesc='grpcurl -plaintext localhost:50051 describe'
fi

# ── Enhanced AWS ───────────────────────────────
if cmd_exists aws; then
    alias ecsl='aws ecs list-clusters --output table'
    alias ecss='aws ecs list-services --output table --cluster'
    alias ecst='aws ecs list-tasks --output table --cluster'
    alias s3ls='aws s3 ls'
    alias lamls='aws lambda list-functions --output table'
    alias ssmsh='aws ssm start-session --target'
    alias ddbls='aws dynamodb list-tables --output table'
    alias cfnls='aws cloudformation list-stacks --stack-status-filter CREATE_COMPLETE UPDATE_COMPLETE --output table'
fi

# ── Helm ───────────────────────────────────────
if cmd_exists helm; then
    alias hls='helm list --all-namespaces'
    alias hup='helm upgrade --install'
    alias hdel='helm uninstall'
    alias hrepo='helm repo update'
    alias hval='helm template . --debug'
fi

# ── Container inspection ──────────────────────
if cmd_exists dive; then
    alias ddive='dive'
fi
if cmd_exists lazydocker; then
    alias lzd='lazydocker'
fi
cmd_exists lazygit && alias lg='lazygit'

# Modern CLI replacements
cmd_exists duf   && alias df='duf'
cmd_exists dust  && alias du='dust'
cmd_exists procs && alias psa='procs'

# Zoxide quick jump
alias z='__zoxide_z'
alias zi='__zoxide_zi'

if cmd_exists stern; then
    alias klog='stern'
fi
if cmd_exists kubectx; then
    alias kctx='kubectx'
    alias kns='kubens'
fi

# ── K8s extras ─────────────────────────────────
if cmd_exists kubectl; then
    alias kev='kubectl get events --sort-by=".lastTimestamp"'
    alias ktop='kubectl top pods --sort-by=cpu'
    alias knodetop='kubectl top nodes --sort-by=cpu'
    alias ksec='kubectl get secrets'
    alias kcm='kubectl get configmaps'
    alias kingress='kubectl get ingress --all-namespaces'
fi

# ── Window management (per-platform in aliases.{darwin,linux}.zsh) ──

# ── Eww dashboard ─────────────────────────────
if cmd_exists eww; then
  alias hud='eww open --toggle dashboard'
  alias hud-reload='eww reload'
  alias hud-kill='eww kill'
fi

# ── Email ─────────────────────────────────────
if cmd_exists aerc;     then alias mail='aerc'; fi
if cmd_exists himalaya; then
  alias inbox='himalaya list --folder INBOX | head -20'
  inbox-cyber() {
    if cmd_exists tte; then
      himalaya list --folder INBOX 2>/dev/null | head -20 | tte wipe \
        --final-gradient-stops 57c7ff ff6ac1 5af78e \
        --final-gradient-direction horizontal 2>/dev/null
    else
      himalaya list --folder INBOX | head -20
    fi
  }
fi

# ── Web browsing ──────────────────────────────
if cmd_exists browsh; then alias web='browsh'; fi
if cmd_exists w3m;    then alias www='w3m'; fi

# ── Finance ───────────────────────────────────
if cmd_exists hledger-ui; then alias budget='hledger-ui'; fi
if cmd_exists bagels;     then alias expenses='bagels'; fi
if cmd_exists wtfutil;    then alias dash='wtfutil'; fi

# ── Amazon orders ─────────────────────────────
if cmd_exists amazon-orders; then
  orders() {
    if cmd_exists tte && [[ -t 1 ]]; then
      amazon-orders --years 1 2>/dev/null | tte decrypt \
        --typing-speed 3 --ciphertext-colors 57c7ff ff6ac1 2>/dev/null
    else
      amazon-orders --years 1 2>/dev/null
    fi
  }
fi

# ── Claude Code ───────────────────────────────
if cmd_exists claude; then
  alias cc='claude'
  alias ccr='claude --resume'
  alias ccc='claude --continue'
  alias ccplan='claude --permission-mode plan'
  alias ccfast='claude --permission-mode acceptEdits'
fi

# ── New tools ─────────────────────────────────
alias top='btop'
alias yy='yazi'
alias viz='cava'
alias md='glow'

# ── Hacker aesthetic / fun ───────────────────
if cmd_exists pipes.sh;      then alias pipes='pipes.sh -t 2 -R -r 0 -p 5 -c 4'; fi
if cmd_exists cbonsai;        then alias bonsai='cbonsai -l -t 0.02 -c "  "'; fi
if cmd_exists tty-clock;      then alias clock='tty-clock -s -c -C 4 -b'; fi
if cmd_exists cmatrix;        then alias matrix='cmatrix -ab -C cyan'; fi
if cmd_exists unimatrix;      then alias umatrix='unimatrix -s 96 -l kKaA -c cyan'; fi
if cmd_exists figlet;         then alias banner='figlet -f slant'; fi
if cmd_exists toilet;         then alias cyberbanner='toilet -f future --filter border --filter gay'; fi
if cmd_exists lolcat;         then alias rainbow='lolcat'; fi
if cmd_exists nms;            then alias decrypt='nms -a -f cyan'; fi
if cmd_exists asciiquarium;   then alias aquarium='asciiquarium'; fi
if cmd_exists hollywood;      then alias hwood='hollywood'; fi
if cmd_exists onefetch;       then alias gitfetch='onefetch'; fi
if cmd_exists lavat;          then alias lava='lavat -c cyan -s 8 -r 2'; fi
if cmd_exists rusty-rain;     then alias rain='rusty-rain -C 0,255,255 -H 50 -S'; fi
if cmd_exists neo-matrix;     then alias neo='neo-matrix --color=cyan --charset=katakana --speed=6 --density=0.75'; fi
if cmd_exists tmatrix;        then alias tmatrix='tmatrix -c default -s 40'; fi
if cmd_exists genact;         then alias busy='genact'; fi
if cmd_exists mapscii;        then alias map='mapscii'; fi
if cmd_exists arttime;        then alias art='arttime'; fi
if cmd_exists durdraw;        then alias ansi='durdraw'; fi
if cmd_exists spotify_player; then alias spot='spotify_player'; fi
if cmd_exists tte;            then alias textfx='tte'; fi

# ── Wallpaper ─────────────────────────────────
if cmd_exists swww;           then
  alias wp='wallpaper-cycle.sh next'
  alias wpr='wallpaper-cycle.sh random'
  alias wps='wallpaper-cycle.sh set'
fi

# ── Network hacker tools ──────────────────────
if cmd_exists trip;           then alias trace='sudo trip'; fi
if cmd_exists bandwhich;      then alias netwatch='sudo bandwhich'; fi
if cmd_exists sniffnet;       then alias sniff='sniffnet'; fi

# ── Animated wrappers (duplicates removed — canonical versions below) ──

alias weather='curl -s "wttr.in?format=3"'
alias forecast='curl -s wttr.in'
cheat() { curl -s "cht.sh/$1"; }
alias colortest='for i in $(seq 0 255); do printf "\e[48;5;${i}m  %3s  \e[0m" "$i"; (( (i+1) % 16 == 0 )) && echo; done'

# ── Cyberpunk commands ──────────────────────
hack() {
  if cmd_exists tte; then
    echo "INITIATING..." | tte beams --beam-delay 2 --beam-gradient-stops 57c7ff ff6ac1 --final-gradient-stops 57c7ff 5af78e
  elif cmd_exists toilet && cmd_exists lolcat; then
    echo "" | toilet -f future "INITIATING..." --filter border 2>/dev/null | lolcat -f -S 50
  elif cmd_exists figlet; then
    figlet -f slant "INITIATING..."
  fi
  sleep 0.3
  local lines=(
    "Bypassing firewall ................. OK"
    "Decrypting AES-256 ................. OK"
    "Injecting payload .................. OK"
    "Elevating privileges ............... OK"
  )
  for line in "${lines[@]}"; do
    if cmd_exists tte; then
      echo "$line" | tte decrypt --typing-speed 4 --ciphertext-colors 57c7ff ff6ac1
    elif cmd_exists nms; then
      echo "$line" | nms -a -f cyan 2>/dev/null
    else
      echo "$line"
    fi
    sleep 0.1
  done
  echo
  if cmd_exists tte; then
    echo "ACCESS GRANTED" | tte scattered --final-gradient-stops 5af78e 57c7ff --movement-speed 0.5
  elif cmd_exists toilet && cmd_exists lolcat; then
    echo "" | toilet -f future "ACCESS GRANTED" --filter border 2>/dev/null | lolcat -f -S 100
  fi
  sleep 0.3
  cmd_exists fastfetch && fastfetch
}

dashboard() {
  tmux new-session -d -s cyber 2>/dev/null || { tmux switch-client -t cyber 2>/dev/null || tmux attach-session -t cyber; return; }
  # Left: system monitor (60%) — animated banner then btop
  tmux send-keys -t cyber 'echo "CYBER DASHBOARD" | tte synthgrid --grid-gradient-stops 57c7ff ff6ac1 --text-gradient-stops 57c7ff 5af78e --max-active-blocks 0.1 2>/dev/null; btop' C-m
  # Right column (40%)
  tmux split-window -t cyber -h -p 40
  # Top-right: audio visualizer
  tmux send-keys -t cyber 'cava 2>/dev/null || cmatrix -ab -C cyan' C-m
  # Mid-right: network monitor or matrix rain
  tmux split-window -t cyber -v -p 66
  if cmd_exists bandwhich; then
    tmux send-keys -t cyber 'sudo bandwhich 2>/dev/null || cmatrix -ab -C cyan' C-m
  elif cmd_exists rusty-rain; then
    tmux send-keys -t cyber 'rusty-rain -C 0,255,255 -H 50 -S' C-m
  else
    tmux send-keys -t cyber 'cmatrix -ab -C cyan' C-m
  fi
  # Bottom-right: lava lamp or screensaver
  tmux split-window -t cyber -v -p 50
  if cmd_exists lavat; then
    tmux send-keys -t cyber 'lavat -c cyan -s 8 -r 2' C-m
  elif cmd_exists pipes.sh; then
    tmux send-keys -t cyber 'pipes.sh -t 2 -R -r 0 -p 5 -c 4' C-m
  else
    tmux send-keys -t cyber 'tty-clock -s -c -C 4 -b 2>/dev/null || date' C-m
  fi
  tmux attach-session -t cyber 2>/dev/null || tmux switch-client -t cyber
}

screensaver() {
  local cmds=()
  cmd_exists cmatrix      && cmds+=("cmatrix -ab -C cyan")
  cmd_exists unimatrix    && cmds+=("unimatrix -s 96 -l kKaA -c cyan")
  cmd_exists pipes.sh     && cmds+=("pipes.sh -t 2 -R -r 0 -p 5 -c 4")
  cmd_exists asciiquarium && cmds+=("asciiquarium")
  cmd_exists cbonsai      && cmds+=("cbonsai -l -t 0.02")
  cmd_exists lavat        && cmds+=("lavat -c cyan -s 8 -r 2")
  cmd_exists rusty-rain   && cmds+=("rusty-rain -C 0,255,255 -H 50 -S")
  cmd_exists neo-matrix   && cmds+=("neo-matrix --color=cyan --charset=katakana --speed=6 --density=0.75")
  cmd_exists tty-clock    && cmds+=("tty-clock -s -c -C 4 -b")
  (( ${#cmds[@]} == 0 )) && { echo "No screensavers installed"; return 1; }
  eval "${cmds[$((RANDOM % ${#cmds[@]} + 1))]}"
}

# ── MOTD / fun combos ────────────────────────
alias motd='fortune -s | cowsay -f small | lolcat'
alias wisdom='fortune | toilet -f term --gay 2>/dev/null || fortune | lolcat'
alias hackermode='cmatrix -ab -C cyan'

# Cockpit — ambient dashboard layout in tmux (lightweight alternative to dashboard)
cockpit() {
  tmux new-session -d -s cockpit 'btop' \; \
    split-window -h 'cava' \; \
    split-window -v 'tty-clock -s -c -C 4 -b' \; \
    select-pane -t 0 \; \
    resize-pane -R 20 \; \
    attach
}

# ── TTE pipe presets ─────────────────────────────
if cmd_exists tte; then
  alias encrypt='tte decrypt --typing-speed 1 --ciphertext-colors 57c7ff ff6ac1'
  alias redact='tte burn --burn-colors ffffff f3f99d ff5c57 8A003C --final-gradient-stops ff5c57 ff6ac1'
  alias reveal='tte scattered --final-gradient-stops 5af78e 57c7ff --movement-speed 0.5'
  alias glitch='tte vhstape --glitch-line-colors 57c7ff ff6ac1 --noise-colors 686868 --final-gradient-stops 57c7ff ff6ac1'
  alias summon='tte blackhole --star-colors 57c7ff ff6ac1 5af78e --final-gradient-stops 57c7ff ff6ac1 5af78e'
  alias materialize='tte synthgrid --grid-gradient-stops 57c7ff ff6ac1 --text-gradient-stops 57c7ff 5af78e'
fi

# ── scan — network reconnaissance ───────────────
scan() {
  local target="${1:?Usage: scan <hostname|ip>}"
  local C='\033[38;2;87;199;255m' M='\033[38;2;255;106;193m' G='\033[38;2;90;247;142m' R='\033[0m'

  _scan_header() {
    if cmd_exists tte; then
      echo "$1" | tte decrypt --typing-speed 4 --ciphertext-colors 57c7ff ff6ac1 2>/dev/null
    else
      printf "${M}%s${R}\n" "$1"
    fi
  }

  _scan_header "▸ TARGET ACQUISITION: $target"
  echo

  _scan_header "▸ DNS RESOLUTION"
  if cmd_exists dig; then
    dig +short "$target" 2>/dev/null | while read -r line; do
      printf "  ${C}%s${R}\n" "$line"
    done
  else
    printf "  ${C}%s${R}\n" "$(host "$target" 2>/dev/null | head -3)"
  fi
  echo

  _scan_header "▸ ROUTE TRACE"
  if cmd_exists trip; then
    sudo trip "$target" --mode tui 2>/dev/null
  elif cmd_exists traceroute; then
    traceroute -m 15 "$target" 2>/dev/null
  else
    printf "  %s\n" "traceroute not available"
  fi
}

# ── deploy — dramatic git push ──────────────────
deploy() {
  local cmd="${*:-git push}"
  local C='57c7ff' M='ff6ac1' G='5af78e' R='ff5c57'

  if cmd_exists tte; then
    echo "DEPLOYING" | tte synthgrid \
      --grid-gradient-stops $C $M \
      --text-gradient-stops $C $G \
      --max-active-blocks 0.1 2>/dev/null
  elif cmd_exists figlet; then
    figlet -f slant "DEPLOYING" | lolcat -f 2>/dev/null
  fi

  local stages=(
    "Compiling artifacts ............... "
    "Running preflight checks .......... "
    "Authenticating deploy key ......... "
    "Pushing to remote ................. "
  )
  for stage in "${stages[@]}"; do
    if cmd_exists tte; then
      echo "$stage" | tte print --print-speed 4 \
        --final-gradient-stops $C $M 2>/dev/null
    else
      printf '\033[38;2;87;199;255m%s\033[0m\n' "$stage"
    fi
  done

  echo
  if eval "$cmd"; then
    echo
    if cmd_exists tte; then
      echo "DEPLOY SUCCESSFUL" | tte fireworks \
        --firework-colors $G $C $M \
        --final-gradient-stops $G $C \
        --explode-anywhere 2>/dev/null
    else
      printf '\033[38;2;90;247;142mDEPLOY SUCCESSFUL\033[0m\n'
    fi
  else
    echo
    if cmd_exists tte; then
      echo "DEPLOY FAILED" | tte burn \
        --burn-colors ffffff f3f99d $R 8A003C \
        --final-gradient-stops $R $M 2>/dev/null
    else
      printf '\033[38;2;255;92;87mDEPLOY FAILED\033[0m\n'
    fi
    return 1
  fi
}

# ── briefing — animated daily dashboard ─────────
briefing() {
  local C='57c7ff' M='ff6ac1' G='5af78e' Y='f3f99d'

  # Header
  if cmd_exists tte; then
    echo "DAILY BRIEFING // $(date '+%Y-%m-%d %H:%M')" | tte beams \
      --beam-gradient-stops $C $M \
      --final-gradient-stops $C $G \
      --beam-delay 2 2>/dev/null
  else
    printf '\033[38;2;87;199;255mDAILY BRIEFING // %s\033[0m\n' "$(date '+%Y-%m-%d %H:%M')"
  fi
  echo

  # System info
  if cmd_exists fastfetch; then
    if cmd_exists tte; then
      fastfetch 2>/dev/null | tte sweep \
        --final-gradient-stops $C $M $G \
        --final-gradient-direction horizontal 2>/dev/null
    else
      fastfetch 2>/dev/null
    fi
  fi
  echo

  # Git status of key repos
  local repos=("$HOME/hairglasses-studio" "$HOME/dotfiles")
  for repo in "${repos[@]}"; do
    [[ -d "$repo/.git" ]] || continue
    local repo_name="${repo##*/}"
    local branch=$(git -C "$repo" branch --show-current 2>/dev/null)
    local status=$(git -C "$repo" status --porcelain 2>/dev/null | wc -l | tr -d ' ')
    local line="▸ $repo_name ($branch) — $status uncommitted"
    if cmd_exists tte; then
      echo "$line" | tte decrypt --typing-speed 6 --ciphertext-colors $C $M 2>/dev/null
    else
      printf '\033[38;2;255;106;193m%s\033[0m\n' "$line"
    fi
  done
  echo

  # Weather
  if cmd_exists tte; then
    curl -s "wttr.in?format=3" 2>/dev/null | tte slide \
      --movement-speed 1.5 \
      --final-gradient-stops $Y $C 2>/dev/null
  else
    curl -s "wttr.in?format=3" 2>/dev/null
  fi
}

# ── cls — animated clear ────────────────────────
cls() {
  if cmd_exists tte && [[ -t 1 ]] && (( COLUMNS > 40 )); then
    printf '\033[2J\033[H'
    echo "BUFFER CLEARED // $(date '+%H:%M:%S')" | tte sweep \
      --final-gradient-stops 57c7ff 1a1a1a 000000 \
      --final-gradient-direction vertical 2>/dev/null
  else
    command clear
  fi
}

# ── SSH connection animation ───────────────────
ssh() {
  local target="${@: -1}"
  if [[ -t 1 ]] && cmd_exists tte; then
    echo "CONNECTING // $target" | tte decrypt \
      --typing-speed 4 --ciphertext-colors 57c7ff ff6ac1 2>/dev/null
  fi
  command ssh "$@"
}

# ── Command not found handler ──────────────────
command_not_found_handler() {
  if cmd_exists tte && [[ -t 1 ]]; then
    echo "UNKNOWN PROTOCOL: $1" | tte errorcorrect \
      --final-gradient-stops ff5c57 ff6ac1 2>/dev/null
  else
    printf '\033[38;2;255;92;87mUNKNOWN COMMAND: %s\033[0m\n' "$1"
  fi
  return 127
}

# ── Package management wrapper (in aliases.linux.zsh) ──

# ── CRT / Shader effects (macOS: aliases.darwin.zsh) ──

# ── MCP / Ralph ────────────────────────────────
alias hgs='cd ~/hairglasses-studio'
alias mcplog='tail -f /tmp/mcp-*.log 2>/dev/null || echo "No MCP logs found"'

# ── Ghostty shader switching ──────────────────
shader-crt() {
  local cfg="$HOME/.config/ghostty/config" tmp
  tmp="$(mktemp "${cfg}.XXXXXX")"
  sed -e "s|^custom-shader = .*|custom-shader = $HOME/.config/ghostty/shaders/green-crt.glsl|" \
      -e "s|^custom-shader-animation = .*|custom-shader-animation = true|" \
      "$cfg" > "$tmp"
  command mv -f "$tmp" "$cfg"
}
shader-none() {
  local cfg="$HOME/.config/ghostty/config" tmp
  tmp="$(mktemp "${cfg}.XXXXXX")"
  sed -e "s|^custom-shader = .*|# custom-shader = disabled|" \
      -e "s|^custom-shader-animation = .*|custom-shader-animation = false|" \
      "$cfg" > "$tmp"
  command mv -f "$tmp" "$cfg"
}
# Toggle shader on/off — remembers the last active shader
shader-toggle() {
  local cfg="$HOME/.config/ghostty/config"
  local state="$HOME/.local/state/ghostty/shader-toggle.last"
  mkdir -p "$(dirname "$state")"
  if grep -q '^custom-shader = ' "$cfg" && ! grep -q '^# custom-shader' "$cfg"; then
    # Shader is ON — save it, then disable
    grep '^custom-shader = ' "$cfg" | head -1 > "$state"
    grep '^custom-shader-animation = ' "$cfg" | head -1 >> "$state"
    shader-none
    echo "Shader OFF (saved state)"
  else
    # Shader is OFF — restore last shader
    if [[ -f "$state" ]] && [[ -s "$state" ]]; then
      local tmp; tmp="$(mktemp "${cfg}.XXXXXX")"
      local last_shader last_anim
      last_shader="$(grep '^custom-shader = ' "$state" | head -1)"
      last_anim="$(grep '^custom-shader-animation = ' "$state" | head -1)"
      sed -e "s|^# custom-shader.*|${last_shader}|" \
          -e "s|^custom-shader-animation = .*|${last_anim}|" \
          "$cfg" > "$tmp"
      command mv -f "$tmp" "$cfg"
      echo "Shader ON: ${last_shader#custom-shader = }"
    else
      echo "No previous shader saved — use a shader-* alias first"
      return 1
    fi
  fi
}
# Quick-switch helper (atomic write, auto-detects animation)
_shader-set() {
  local name="$1" cfg="$HOME/.config/ghostty/config"
  local path="$HOME/.config/ghostty/shaders/${name}.glsl"
  [[ -f "$path" ]] || { echo "Shader not found: $path"; return 1; }
  local anim=false
  grep -qE '(iTime|ghostty_time|u_time)' "$path" 2>/dev/null && anim=true
  local tmp; tmp="$(mktemp "${cfg}.XXXXXX")"
  sed -e "s|^custom-shader = .*|custom-shader = $path|" \
      -e "s|^# custom-shader.*|custom-shader = $path|" \
      -e "s|^custom-shader-animation = .*|custom-shader-animation = $anim|" \
      "$cfg" > "$tmp"
  command mv -f "$tmp" "$cfg"
  echo "Shader: $name (animation=$anim)"
}
# Cyberpunk collection quick-switches
alias shader-synthwave='_shader-set synthwave-horizon'
alias shader-holo='_shader-set holo-display'
alias shader-hexgrid='_shader-set neon-hex-grid'
alias shader-circuit='_shader-set circuit-trace'
alias shader-rain='_shader-set rain-on-glass'
alias shader-glitchholo='_shader-set cyber-glitch-holo'
# Interactive shader audition — test each shader one-by-one with keep/skip
alias shader-audit='bash ~/.config/ghostty/shaders/bin/shader-audit.sh'

# Benchmark shaders via glslViewer headless rendering
alias shader-bench='bash ~/.config/ghostty/shaders/bin/shader-benchmark.sh'

# Shader toolchain
alias shader-meta='bash ~/.config/ghostty/shaders/bin/shader-meta.sh'
alias shader-build='bash ~/.config/ghostty/shaders/bin/shader-build.sh'
alias shader-test='bash ~/.config/ghostty/shaders/bin/shader-test.sh'
alias shader-cycle='bash ~/.config/ghostty/shaders/bin/shader-cycle.sh'

# Font mixing — Monaspace multi-font profiles
alias font-mix='bash $DOTFILES_DIR/scripts/font-mix.sh'

# Peekaboo — macOS screen capture (in aliases.darwin.zsh)

# Shuffled playlist engine
source "$HOME/.config/ghostty/shaders/bin/shader-playlist.sh" 2>/dev/null

# Playlist-aware shader selection (called from zshrc on each new shell)
shader-next() {
  [[ -z "$GHOSTTY_RESOURCES_DIR" ]] && return
  if [[ "$GHOSTTY_QUICK_TERMINAL" = "1" ]]; then
    shader-playlist-next "immersive"
  else
    local active_pl="ambient"
    local pl_cfg="$HOME/.local/state/ghostty/auto-rotate-playlist"
    [[ -f "$pl_cfg" ]] && active_pl="$(< "$pl_cfg")"
    shader-playlist-next "$active_pl"
  fi
}

# Manually rotate shader on demand
alias shader-rotate='shader-playlist-next "ambient"'

# Random shader from all shaders (manual fallback, ignores playlists)
# Also available as AeroSpace keybind: alt-shift-s
shader-random() {
  bash "$HOME/.config/ghostty/shaders/bin/shader-random.sh"
}

# Playlist utilities
alias shader-reshuffle='rm -f ~/.local/state/ghostty/*.queue ~/.local/state/ghostty/*.idx && echo "Playlists reshuffled on next shell"'
shader-status() {
  local sd="$HOME/.local/state/ghostty"
  _ss_line() { local i t; i="$(cat "$sd/$1.idx" 2>/dev/null || echo 0)"; t="$(wc -l < "$sd/$1.queue" 2>/dev/null | tr -d ' ')" 2>/dev/null || t="?"; printf "%-20s %s / %s\n" "$2" "$i" "$t"; }
  echo "── Playlists ──"
  _ss_line "ambient"     "Ambient:"
  _ss_line "immersive"   "Immersive:"
  _ss_line "cyberpunk"   "Cyberpunk:"
  _ss_line "cursor-fx"   "Cursor FX:"
  _ss_line "retro"       "Retro:"
  _ss_line "showcase"    "Showcase:"
  echo "── Auto-rotate ──"
  local active_pl="ambient"
  [[ -f "$sd/auto-rotate-playlist" ]] && active_pl="$(< "$sd/auto-rotate-playlist")"
  printf "%-20s %s\n" "Active playlist:" "$active_pl"
  if systemctl --user is-active shader-rotate.timer &>/dev/null; then
    printf "%-20s %s\n" "Timer:" "running"
  else
    printf "%-20s %s\n" "Timer:" "stopped"
  fi
}

# Automatic timed shader rotation via systemd timer (Linux)
# shader-auto start [minutes]  — start rotating every N minutes (default 30)
# shader-auto stop             — stop automatic rotation
# shader-auto status           — check if timer is running
# Defined in aliases.linux.zsh

# Best-of shader showcase — curated playlist of the most impressive effects
# shader-best start [minutes]  — activate best-of rotation (default 15 min)
# shader-best stop             — revert to low-intensity and stop
# shader-best next             — manually advance to next best-of shader
shader-best() {
  local state_dir="$HOME/.local/state/ghostty"
  local playlist_cfg="$state_dir/auto-rotate-playlist"
  mkdir -p "$state_dir" 2>/dev/null

  case "${1:-next}" in
    start)
      printf '%s' "best-of" > "$playlist_cfg"
      shader-auto start "${2:-15}"
      echo "Best-of showcase active (every ${2:-15} minutes)"
      ;;
    stop)
      printf '%s' "low-intensity" > "$playlist_cfg"
      shader-auto stop
      echo "Best-of stopped, reverted to low-intensity"
      ;;
    next)
      printf '%s' "best-of" > "$playlist_cfg"
      shader-playlist-next "best-of"
      ;;
    *)
      echo "Usage: shader-best {start [minutes]|stop|next}"
      ;;
  esac
}

# FZF shader picker — reads metadata from shaders.toml manifest
shader-pick() {
  local dir="$HOME/.config/ghostty/shaders"
  local meta="$dir/bin/shader-meta.sh"
  [[ -d "$dir" ]] || { echo "No shaders directory found"; return 1; }
  [[ -x "$meta" ]] || { echo "shader-meta.sh not found"; return 1; }

  local pick
  pick="$("$meta" fzf-lines | fzf \
    --delimiter='\t' \
    --with-nth=2,1,3 \
    --header='Pick a shader (ESC to cancel)' \
    --no-multi)" || return
  local shader_name
  shader_name="$(echo "$pick" | cut -f1)"
  [[ -z "$shader_name" ]] && return

  local shader_path="$dir/$shader_name"
  local anim="false"
  grep -qE '(ghostty_time|iTime|u_time)' "$shader_path" 2>/dev/null && anim="true"

  local cfg="$HOME/.config/ghostty/config" tmp
  tmp="$(mktemp "${cfg}.XXXXXX")"
  sed -e "s|^custom-shader = .*|custom-shader = $shader_path|" \
      -e "s|^# custom-shader.*|custom-shader = $shader_path|" \
      -e "s|^custom-shader-animation = .*|custom-shader-animation = $anim|" \
      "$cfg" > "$tmp"
  mv "$tmp" "$cfg"
  echo "Shader set to: $shader_name (animation=$anim)"
}

# MCP audit dashboard
alias audit='$HOME/hairglasses-studio/dotfiles/scripts/audit-dashboard.sh'

# Load local aliases if they exist
[[ -f "$HOME/.aliases.local" ]] && source "$HOME/.aliases.local"
