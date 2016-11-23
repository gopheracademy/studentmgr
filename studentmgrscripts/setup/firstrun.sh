#!/bin/bash

# First Time Server setup things

# install git, etc


set -e

DEBIAN_FRONTEND=noninteractive sudo apt-get update && sudo apt-get install -y openssh-server \
      ca-certificates curl unzip tar \
      zsh stow   
# Install development packages
DEBIAN_FRONTEND=noninteractive sudo apt-get install -y python-pip build-essential mercurial bzr python-dev ctags cmake software-properties-common python-software-properties 

DEBIAN_FRONTEND=noninteractive sudo add-apt-repository "deb http://archive.ubuntu.com/ubuntu $(lsb_release -sc) universe"
DEBIAN_FRONTEND=noninteractive sudo add-apt-repository -y ppa:neovim-ppa/unstable 
DEBIAN_FRONTEND=noninteractive sudo apt-get update && sudo apt-get -y install python3-dev python3-pip neovim 
pip2 install neovim
pip3 install neovim


# install go

mkdir -p /usr/local/go && curl -Ls https://storage.googleapis.com/golang/go1.7.1.linux-amd64.tar.gz | tar xvzf - -C /usr/local/go --strip-components=1

