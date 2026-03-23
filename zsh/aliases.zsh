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
    alias less='bat --paging=always'
elif cmd_exists batcat; then
    alias cat='batcat --paging=never'
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

# System information (cross-platform)
if [[ "$OSTYPE" == "darwin"* ]]; then
    alias top='top -o cpu'
    alias mem='memory_pressure'
    alias cpu='top -l 1 | grep "CPU usage"'
else
    alias mem='free -h'
    alias cpu='grep "cpu " /proc/stat | awk "{usage=(\$2+\$4)*100/(\$2+\$3+\$4+\$5)} END {print usage \"%\"}"'
fi

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
alias emptytrash='rm -rf ~/.Trash/*'

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

# ── Window management ─────────────────────────
alias aero-reload='aerospace reload-config'
alias bar-reload='sketchybar --reload'

# ── New tools ─────────────────────────────────
alias top='btop'
alias yy='yazi'
alias viz='cava'
alias md='glow'

# ── Hacker aesthetic / fun ───────────────────
if cmd_exists pipes.sh;  then alias pipes='pipes.sh'; fi
if cmd_exists cbonsai;   then alias bonsai='cbonsai -l -t 0.02'; fi
if cmd_exists tty-clock; then alias clock='tty-clock -s -c -C 4 -b'; fi
if cmd_exists cmatrix;   then alias matrix='cmatrix -ab -C cyan'; fi
if cmd_exists figlet;    then alias banner='figlet -f slant'; fi
if cmd_exists lolcat;    then alias rainbow='lolcat'; fi
if cmd_exists onefetch;  then alias gitfetch='onefetch'; fi
alias screensaver='pipes.sh -t 2 -R -r 0 -p 5'
alias weather='curl -s "wttr.in?format=3"'
alias forecast='curl -s wttr.in'
cheat() { curl -s "cht.sh/$1"; }
alias colortest='for i in $(seq 0 255); do printf "\e[48;5;${i}m  %3s  \e[0m" "$i"; (( (i+1) % 16 == 0 )) && echo; done'

# ── MCP / Ralph ────────────────────────────────
alias hgs='cd ~/hairglasses-studio'
alias mcplog='tail -f /tmp/mcp-*.log 2>/dev/null || echo "No MCP logs found"'

# ── Ghostty shader switching ──────────────────
alias shader-crt='sed -i "" "s|custom-shader = .*|custom-shader = $HOME/.config/ghostty/shaders/green-crt.glsl|" ~/.config/ghostty/config'
alias shader-none='sed -i "" "s|custom-shader = .*|# custom-shader = disabled|" ~/.config/ghostty/config'
# Interactive shader audition — test each shader one-by-one with keep/skip
alias shader-audit='bash ~/.config/ghostty/shaders/pick-shaders.sh'

# Random shader (called from zshrc on each new shell)
shader-random() {
  local dir="$HOME/.config/ghostty/shaders"
  local shaders=("$dir"/*.glsl(N))
  (( ${#shaders} == 0 )) && return
  local pick="${shaders[RANDOM % ${#shaders} + 1]}"
  sed -i "" "s|custom-shader = .*|custom-shader = $pick|" ~/.config/ghostty/config 2>/dev/null
}

# FZF shader picker with categories and preview
shader-pick() {
  local dir="$HOME/.config/ghostty/shaders"
  [[ -d "$dir" ]] || { echo "No shaders directory found"; return 1; }

  # Category → shader mapping (sorted by category)
  local entries=""
  local shader
  for shader in "$dir"/*.glsl(N); do
    local name="${shader:t}"
    local cat desc
    case "$name" in
      blue-crt*|green-crt*|crt.glsl|crt_glitch*|crt-chromatic*|bettercrt*|in-game-crt*|retro-terminal*|scanline*|amber-monitor*|vt320-amber*) cat="CRT" ;;
      cursor_*|cursor-*|*_cursor*|blaze_sparks*|last_letter_zoom*|manga_slash*|party_sparks*|sparks.glsl|cursor_explosion*|cursor_viberation*) cat="Cursor" ;;
      *-bg.glsl|graded-wash*|salt-bg*|splatter-bg*|variegated*|wet-on-wet*) cat="Watercolor" ;;
      animated-gradient*|clouds*|cubes*|electric*|galaxy*|gears*|gradient-background*|inside-the-matrix*|just-snow*|matrix-hallway*|sparks-from-fire*|splatter-fractal*|starfield*|water.glsl|underwater*) cat="Background" ;;
      dither*|drunkard*|flicker*|glitchy*|glow*|hexglitch*|mnoise*|pixels*|shake*|tft*|zoom_and_aberration*|chromatic-aberration*|vcr-*|vhs-*|vaporwave*|bloom*) cat="Post-FX" ;;
      *) cat="Other" ;;
    esac
    case "$name" in
      blue-crt*)           desc="Blue phosphor CRT with scanlines" ;;
      green-crt*)          desc="Green phosphor CRT with scanlines" ;;
      crt.glsl)            desc="Classic CRT scanlines and curvature" ;;
      crt_glitch*)         desc="CRT with glitch distortion" ;;
      in-game-crt-cursor*) desc="In-game CRT with cursor highlight" ;;
      in-game-crt.glsl)    desc="In-game style CRT monitor" ;;
      retro-terminal*)     desc="Retro green phosphor terminal" ;;
      scanline*)           desc="Simple scanline overlay" ;;
      crt-chromatic*)      desc="CRT with chromatic aberration + dot matrix" ;;
      bettercrt*)          desc="Enhanced CRT with barrel distortion" ;;
      in-game-crt-alt*)    desc="In-game CRT variant" ;;
      amber-monitor*)      desc="Amber phosphor monitor simulation" ;;
      vt320-amber*)        desc="VT320-style amber phosphor glow" ;;
      retro-terminal-soanvig*) desc="Retro terminal amalgam" ;;
      dither*)             desc="Dithering effect on output" ;;
      drunkard*)           desc="Wobbly distorted screen" ;;
      flicker*)            desc="Screen flicker effect" ;;
      glitchy*)            desc="Digital glitch/corruption" ;;
      glow.glsl)           desc="Simple bloom/glow effect" ;;
      glow-rgb*)           desc="RGB split with glow and twitch" ;;
      hexglitch*)          desc="Hex grid glitch distortion" ;;
      mnoise*)             desc="Perlin noise overlay" ;;
      pixels*)             desc="Pixel grid effect" ;;
      shake*)              desc="Screen shake effect" ;;
      tft*)                desc="TFT/LCD subpixel rendering" ;;
      zoom_and_aberration*) desc="Zoom with chromatic aberration" ;;
      chromatic-aberration*) desc="Radial chromatic aberration" ;;
      vcr-distortion*)     desc="VCR tape playback distortion" ;;
      vhs-tape*)           desc="VHS tape degradation effect" ;;
      vaporwave*)          desc="Vaporwave pink/cyan color grade" ;;
      bloom-classic*)      desc="Classic bloom/glow effect" ;;
      bloom-soft*)         desc="Soft bloom/glow effect" ;;
      bloom-warm*)         desc="Warm-toned bloom effect" ;;
      bloom025*)           desc="Bloom at 25% intensity" ;;
      bloom050*)           desc="Bloom at 50% intensity" ;;
      bloom060*)           desc="Bloom at 60% intensity" ;;
      bloom075*)           desc="Bloom at 75% intensity" ;;
      bloom1*)             desc="Bloom at full intensity" ;;
      blaze_sparks*)       desc="Sparking blaze cursor" ;;
      cursor_blaze_no_trail*) desc="Blaze cursor without trail" ;;
      cursor_blaze_tapered*) desc="Tapered fire trail cursor" ;;
      cursor_blaze*)       desc="Fire trail behind cursor" ;;
      cursor_border_1*)    desc="Cursor border glow" ;;
      cursor_frozen*)      desc="Frozen/ice cursor effect" ;;
      cursor-glitch*)      desc="Glitch distortion around cursor" ;;
      cursor_smear_fade*)  desc="Smear with fade cursor" ;;
      cursor_smear_gradient*) desc="Gradient smear cursor" ;;
      cursor_smear_rainbow*) desc="Rainbow smear cursor" ;;
      cursor_smear*)       desc="Smear trail cursor" ;;
      cursor_sweep*)       desc="Sweep trail behind cursor" ;;
      cursor_synesthaxia*) desc="Colorscheme-adaptive cursor" ;;
      cursor_tail*)        desc="Fading tail behind cursor" ;;
      cursor_warp*)        desc="Warp distortion around cursor" ;;
      last_letter_zoom*)   desc="Zoom on last typed character" ;;
      manga_slash*)        desc="Manga-style slash effect" ;;
      party_sparks*)       desc="Colorful party sparks" ;;
      ripple_cursor*)      desc="Ripple wave from cursor" ;;
      ripple_rectangle*)   desc="Rectangular ripple from cursor" ;;
      sparks.glsl)         desc="Spark particles from cursor" ;;
      cursor_explosion*)   desc="Particle explosion from cursor" ;;
      cursor_viberation*)  desc="Vibrating cursor with feedback" ;;
      cursor_blaze_alt*)   desc="Alternative fire trail cursor" ;;
      cursor_blaze_chardskarth*) desc="Original cursor blaze (chardskarth)" ;;
      cursor_blaze_no_trail_chardskarth*) desc="Blaze without trail (chardskarth)" ;;
      cursor_smear_alt*)   desc="Alternative smear trail cursor" ;;
      cursor_smear_*_original*|cursor_smear_original*) desc="Original smear variant (pre-edit)" ;;
      animated-gradient*)  desc="Animated color gradient" ;;
      clouds*)             desc="Parallax cloud background" ;;
      cubes*)              desc="Animated 3D cube grid" ;;
      electric-modes*)     desc="Electric effect with modes" ;;
      electric*)           desc="Electric/lightning effect" ;;
      galaxy*)             desc="Animated galaxy/nebula" ;;
      gears*)              desc="Mechanical gears animation" ;;
      gradient-background*) desc="Static color gradient" ;;
      inside-the-matrix*)  desc="Matrix rain animation" ;;
      just-snow*)          desc="Falling snow particles" ;;
      matrix-hallway*)     desc="Matrix code fly-through" ;;
      sparks-from-fire*)   desc="Fire sparks particles" ;;
      splatter-fractal*)   desc="Fractal paint splatter" ;;
      starfield-colors*)   desc="Colorful animated starfield" ;;
      starfield.glsl)      desc="Classic starfield fly-through" ;;
      starfield-alt*)      desc="Starfield variant" ;;
      starfield-sherwin*)  desc="Starfield with image overlay" ;;
      inside-the-matrix-alt*) desc="Matrix rain variant" ;;
      inside-the-matrix-sherwin*) desc="Matrix with custom blending" ;;
      sparks-from-fire-shadertoy*) desc="Shadertoy fire sparks" ;;
      underwater*)         desc="Underwater caustics and light" ;;
      water.glsl)          desc="Water ripple/wave" ;;
      graded-wash*)        desc="Graded color transition wash" ;;
      salt-bg*)            desc="Salt texture on watercolor" ;;
      splatter-bg*)        desc="Paint splatter effect" ;;
      variegated*)         desc="Multi-color variegated wash" ;;
      wet-on-wet*)         desc="Wet-on-wet watercolor blend" ;;
      *)                   desc="" ;;
    esac
    entries+="$(printf '%-12s │ %-36s %s' "$cat" "$name" "$desc")\n"
  done

  local pick
  pick="$(echo -e "$entries" | sort | fzf --header='Pick a shader (ESC to cancel)' --ansi --no-multi)" || return
  local shader_name
  shader_name="$(echo "$pick" | awk -F'│' '{print $2}' | awk '{print $1}')"
  [[ -z "$shader_name" ]] && return

  sed -i "" "s|custom-shader = .*|custom-shader = $dir/$shader_name|" ~/.config/ghostty/config 2>/dev/null
  echo "Shader set to: $shader_name (open a new terminal to see it)"
}

# Load local aliases if they exist
[[ -f "$HOME/.aliases.local" ]] && source "$HOME/.aliases.local"
