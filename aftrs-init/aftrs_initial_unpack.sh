#!/bin/bash

# create the Docs directory if it doesn't exist
gum format -- "# Setting up Docs directory..."
mkdir -p /home/hg/Docs

GHORG_GITHUB_TOKEN="$(gh auth token)"
export GHORG_GITHUB_TOKEN

USER_NAME="hairglasses"
export USER_NAME

# Declare all GitHub organizations in a simple array (this list won't change much)
declare -a github_orgs_list=("aftrs-void" "secretstudios")

install_dependencies() {
  gum format -- "# Installing dependencies..."
  # Define command_exists function
  command_exists() {
    command -v "$1" >/dev/null 2>&1
  }

  sudo apt update && sudo apt -y upgrade && sudo apt install -y zsh git-extras build-essential make gcc python3 pipx lsd bat ripgrep fd-find fzf tmux openssh-server wget tldr unzip direnv && \
    gum log --level info "System packages installed."

  # install bpp - https://github.com/rail5/bpp
  if command_exists "bpp"; then
    gum log --level info "bpp already installed."
  else
    sudo curl -s -o /etc/apt/trusted.gpg.d/rail5-signing-key.gpg "https://deb.rail5.org/rail5-signing-key.gpg"
    sudo curl -s -o /etc/apt/sources.list.d/rail5.list "https://deb.rail5.org/rail5.list"
    sudo apt update
    sudo apt install bpp && gum log --level info "bpp installed."
  fi

  # Install uv (Universal Version Manager) - https://astral.sh/uv/
  if command_exists "uv"; then
    gum log --level info "uv already installed."
  else
    curl -LsSf https://astral.sh/uv/install.sh | sh && gum log --level info "uv installed."
  fi

  # use uv to install meta-package-manager - https://github.com/mpm-project/mpm
  if command_exists "mpm"; then
    gum log --level info "mpm already installed."
  else
    sudo cp /home/hg/Docs/aftrs-void/aftrs_init/bin/mpm-linux-x64.bin /usr/local/bin/mpm
    sudo chmod +x /usr/local/bin/mpm && gum log --level info "mpm installed."
  fi

  # install Homebrew - https://brew.sh/installation
  if command_exists "brew"; then
    gum log --level info "brew already installed."
  else
    NONINTERACTIVE=1 /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" && gum log --level info "brew installed."
  fi

  # install gum - https://github.com/charmbracelet/gum
  if command_exists "gum"; then
    gum log --level info "gum already installed."
  else
    # Use latest gum version (v0.16.2) for Linux x86_64
    curl -fsSL https://github.com/charmbracelet/gum/releases/download/v0.16.2/gum_0.16.2_Linux_x86_64.tar.gz | tar xz && \
    sudo mv gum_0.16.2_Linux_x86_64/gum /usr/local/bin/gum && \
    rm -rf gum_0.16.2_Linux_x86_64 && \
    gum log --level info "gum installed."
  fi

  # install GitHub cli (gh) - https://github.com/cli/cli
  if command_exists "gh"; then
    gum log --level info "gh already installed."
  else
    brew install gh && gum log --level info "gh installed."
    # Next, install the crguezl/gh-submodule-add extension for gh - https://github.com/crguezl/gh-submodule-add
    gh extension install crguezl/gh-submodule-add && gum log --level info "gh-submodule-add extension installed."
  fi

  # install ghorg - https://github.com/gabrie30/ghorg
  if command_exists "ghorg"; then
    gum log --level info "ghorg already installed."
  else
    brew install gabrie30/utils/ghorg && gum log --level info "ghorg installed."
  fi


  # copy updated ghorg config file
  # cp /home/hg/Docs/aftrs-void/aftrs_init/config/ghorg/conf.yaml /home/hg/Docs/.config/ghorg/conf.yaml

  # use meta-package-manager to upgrade all packages
  # cp /home/hg/Docs/aftrs-void/aftrs_init/config/mpm/config.toml /home/hg/Docs/.config/mpm/config.toml
  # sudo mpm upgrade --all
}

# Define command_exists function
command_exists() {
  command -v "$1" >/dev/null 2>&1
}

# Install Docker, Docker Engine, and Buildx
if command_exists "docker"; then
  gum log --level info "Docker already installed."
else
  gum format -- "# Installing Docker..."
  sudo apt-get remove -y containerd && \
  sudo apt-get install -y docker.io && gum log --level info "Docker installed."
fi

# Set up Docker Buildx
if ! docker buildx version >/dev/null 2>&1; then
  gum format -- "# Setting up Docker Buildx..."
  docker buildx create --name aftrs-builder --use && gum log --level info "Docker Buildx set as default builder."
else
  gum log --level info "Docker Buildx already set up."
fi

export GITHUB_TOKEN="$(gh auth token)"
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

clone_personal_repos() {
  gum format -- "# Cloning personal repositories..."
  mkdir -p /home/hg/Docs/"$USER_NAME" && cd /home/hg/Docs/"$USER_NAME"
  gum log --level info "Cloning repos for $USER_NAME..."
  ghorg clone "$USER_NAME" --clone-type=user --path="/home/hg/Docs/$USER_NAME" && gum log --level info "Personal repos cloned."
  shopt -s dotglob
  mv "/home/hg/Docs/$USER_NAME/$USER_NAME"/* "/home/hg/Docs/$USER_NAME/"
  shopt -u dotglob
  rm -rf "/home/hg/Docs/$USER_NAME/$USER_NAME"
  wait
}

clone_org_repos() {
  for i in "${github_orgs_list[@]}"
  do
    gum format -- "# Cloning $i organization repositories..."
    mkdir -p "/home/hg/Docs/$i"
    cd "/home/hg/Docs/$i" || exit
    temp_dir="/home/hg/Docs/$i/temp"
    mkdir -p "$temp_dir"
    gum log --level info "Cloning $i repos..."
    ghorg clone ${i} --path="$temp_dir" --skip-archived --concurrency=8 && gum log --level info "$i repos cloned."
    echo "[DEBUG] Contents of $temp_dir after ghorg clone:"
    ls -l "$temp_dir"
    echo "[DEBUG] Contents of $temp_dir/$i after ghorg clone:"
    ls -l "$temp_dir/$i"
    shopt -s dotglob
    mv "$temp_dir/$i"/* "/home/hg/Docs/$i/"
    echo "[DEBUG] Contents of /home/hg/Docs/$i after move:"
    ls -l "/home/hg/Docs/$i"
    shopt -u dotglob
    rm -rf "$temp_dir/$i"
    rm -rf "$temp_dir"
  done
}

clone_secretstudios_org() {
  gum format -- "# Cloning secretstudios organization repositories..."
  mkdir -p "/home/hg/Docs/secretstudios"
  cd "/home/hg/Docs/secretstudios" || exit
  temp_dir="/home/hg/Docs/secretstudios/temp"
  mkdir -p "$temp_dir"
  gum log --level info "Cloning secretstudios repos..."
  ghorg clone secretstudios --path="$temp_dir" --skip-archived --concurrency=8 && gum log --level info "secretstudios repos cloned."
  echo "[DEBUG] Contents of $temp_dir after ghorg clone:"
  ls -l "$temp_dir"
  echo "[DEBUG] Contents of $temp_dir/secretstudios after ghorg clone:"
  ls -l "$temp_dir/secretstudios"
  shopt -s dotglob
  mv "$temp_dir/secretstudios"/* "/home/hg/Docs/secretstudios/"
  echo "[DEBUG] Contents of /home/hg/Docs/secretstudios after move:"
  ls -l "/home/hg/Docs/secretstudios"
  shopt -u dotglob
  rm -rf "$temp_dir/secretstudios"
  rm -rf "$temp_dir"
}

git_bulk_register_workspaces() {
  git bulk --addworkspace "$USER_NAME" /home/hg/Docs/${USER_NAME}
  git bulk --addworkspace "aftrs-void" /home/hg/Docs/aftrs-void
  git bulk --addworkspace "secretstudios" /home/hg/Docs/secretstudios
}

git_bulk_pull_all() {
  git bulk -a pull
}

install_dependencies
clone_personal_repos
clone_org_repos
clone_secretstudios_org
# git_bulk_register_workspaces
# git_bulk_pull_all

# Ensure the BuildKit configuration directory exists
gum format -- "# Ensuring BuildKit configuration directory exists..."
sudo mkdir -p /etc/buildkit && \
  gum log --level info "BuildKit configuration directory ensured."

# Copy BuildKit configuration file
gum format -- "# Copying BuildKit configuration file..."
sudo cp /home/hg/Docs/aftrs-void/aftrs_init/config/buildkit/buildkitd.toml /etc/buildkit/buildkitd.toml && \
  gum log --level info "BuildKit configuration file copied to /etc/buildkit/buildkitd.toml."

# Restart Docker service
gum format -- "# Restarting Docker service..."
sudo systemctl restart docker && \
  gum log --level info "Docker service restarted."

# Delete each repo from the `temp/` folder after confirming it has been copied
for i in "${github_orgs_list[@]}"
do
  temp_dir="/home/hg/Docs/$i/temp"
  if [ -d "$temp_dir" ]; then
    gum format -- "# Deleting temporary directory for $i..."
    rm -rf "$temp_dir" && \
      gum log --level info "Temporary directory for $i deleted."
  fi
done

echo -e "| Task | Status |\n| --- | --- |\n| Install Dependencies | Completed |\n| Clone Personal Repos | Completed |\n| Clone Org Repos | Completed |" > /tmp/status_table.txt