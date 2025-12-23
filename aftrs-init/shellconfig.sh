#!/bin/bash

# Ensure zsh is installed
sudo apt update && sudo apt -y install zsh

# Install oh-my-zsh
sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"


# Install Antibody
curl -sfL git.io/antibody | sh -s - -b /usr/local/bin

# Install getantidote/use-omz - https://github.com/getantidote/use-omz
