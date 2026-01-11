#!/usr/bin/env zsh

set -e

# Path where inital config is stored
DEFAULT_CONFIG_PATH="/usr/share/apparatus"
NVM_DIR="$([ -z "${XDG_CONFIG_HOME-}" ] && printf %s "${HOME}/.nvm" || printf %s "${XDG_CONFIG_HOME}/nvm")"

# Initialize
echo "Initialising.."
mkdir -p "${HOME}/.config"


# ZSH
if [[ ! -f "{HOME}/.config/zsh/.zshrc" ]]; then
  mkdir -p "${ZDOTDIR}"
  cp ${DEFAULT_CONFIG_PATH}/zshrc "${ZDOTDIR}/.zshrc"
else
  echo 'zsh config already exists'
fi

# ZSH plugin manager
if [[ ! -f "${ZDOTDIR}/.antidote/antidote.zsh" ]]; then
  git clone --depth=1 https://github.com/mattmc3/antidote.git "${ZDOTDIR}/.antidote"
  cp ${DEFAULT_CONFIG_PATH}/zsh_plugins.txt ${ZDOTDIR}/.zsh_plugins.txt
  echo "Installing Antidote (zsh plugin manager).."
  echo '# Antidote' >> "${ZDOTDIR}/.zshrc"
  echo 'source ${ZDOTDIR}/.antidote/antidote.zsh' >> "${ZDOTDIR}/.zshrc"
  echo 'antidote load' >> "${ZDOTDIR}/.zshrc"
else
  echo 'Antidote already configured. Skipping..'
fi


# p10k Prompt
if [[ ! -f "${ZDOTDIR}/.p10k.zsh" ]]; then
  echo "Installing Powerlevel10k.."
  cp ${DEFAULT_CONFIG_PATH}/p10k.zsh "${ZDOTDIR}/.p10k.zsh"
  echo '# Powerlevel10k' >> ${ZDOTDIR}/.zshrc
  echo 'POWERLEVEL9K_DISABLE_CONFIGURATION_WIZARD=true' >> ${ZDOTDIR}/.zshrc
  echo 'source ${ZDOTDIR}/.p10k.zsh' >> ${ZDOTDIR}/.zshrc
else
  echo 'Powerlevel10k already configured. Skipping..'
fi

# NVM / Node
if [[ ! -f "${NVM_DIR}/nvm.sh" ]]; then
git clone https://github.com/nvm-sh/nvm.git "$NVM_DIR" && \
  cd "$NVM_DIR"
  git checkout `git describe --abbrev=0 --tags --match "v[0-9]*" $(git rev-list --tags --max-count=1)`
  ${NVM_DIR}/install.sh
  export NVM_DIR="${HOME}/.config/nvm"
  [ -s "${NVM_DIR}/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
  nvm install --lts ${NODE_VERSION}
  npm install -g typescript typescript-language-server
  npm install -g bash-language-server
else
  echo 'nvm already configured. Skipping..'
fi

## Zellij
if [ ! -f "${HOME}/.config/zellij/config.kdl" ]; then
  mkdir ~/.config/zellij
  zellij setup --dump-config > ~/.config/zellij/config.kdl
else
  echo 'Zellij already exists. Skipping..'
fi

# SSH
if [[ ! -d "${HOME}/.ssh" ]]; then
  echo "ssh config.."
  mkdir ${HOME}/.ssh
  chmod 0700 ${HOME}/.ssh
else
  echo '.ssh folder already exists. Skipping..'
fi

# Git
# set editor
git config --global core.editor /opt/nvim/nvim.appimage

