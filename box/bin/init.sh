#!/usr/bin/env zsh

set -e

# Path where inital config is stored
DEFAULT_CONFIG_PATH="/usr/share/apparatus"
CONFIG_MODE_FILE="${HOME}/.apparatus-config-mode"

# Initialize
echo "Initialising.."

# Handle shared home detection and user choice
if [[ "$APPARATUS_SHARED_HOME" == "1" ]]; then
    # Check if user has already made a choice
    if [[ -f "$CONFIG_MODE_FILE" ]]; then
        APPARATUS_CONFIG_MODE=$(cat "$CONFIG_MODE_FILE")
    else
        echo ""
        echo "╭─────────────────────────────────────────────────────────────╮"
        echo "│  Shared home directory detected                            │"
        echo "│  Your container shares the home directory with the host.   │"
        echo "╰─────────────────────────────────────────────────────────────╯"
        echo ""

        CHOICE=$(gum choose \
            "isolated  - Use container-specific configs (~/.config/apparatus-box/)" \
            "host      - Use existing host configs (no modifications)" \
            --header "How should this container handle configuration?")

        APPARATUS_CONFIG_MODE=$(echo "$CHOICE" | awk '{print $1}')
        echo "$APPARATUS_CONFIG_MODE" > "$CONFIG_MODE_FILE"
        echo ""
    fi

    if [[ "$APPARATUS_CONFIG_MODE" == "host" ]]; then
        echo "Using host configuration - skipping initialization."
        echo "Your existing shell, git, and other configs will be used as-is."
        echo ""
        echo "To change this later, remove: $CONFIG_MODE_FILE"
        exit 0
    fi

    echo "Using container-specific configs: $APPARATUS_CONFIG_HOME"
    echo "To change this later, remove: $CONFIG_MODE_FILE"
    echo ""
fi

# Set NVM_DIR based on XDG_CONFIG_HOME (already set by profile-custom.sh)
NVM_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/nvm"

mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}"


# ZSH
if [[ ! -f "${ZDOTDIR}/.zshrc" ]]; then
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
  [ -s "${NVM_DIR}/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
  nvm install --lts ${NODE_VERSION}
  npm install -g typescript typescript-language-server
  npm install -g bash-language-server
else
  echo 'nvm already configured. Skipping..'
fi

# SSH
# Skip when using shared home - host already has .ssh
if [[ "$APPARATUS_SHARED_HOME" != "1" ]]; then
  if [[ ! -d "${HOME}/.ssh" ]]; then
    echo "ssh config.."
    mkdir ${HOME}/.ssh
    chmod 0700 ${HOME}/.ssh
  else
    echo '.ssh folder already exists. Skipping..'
  fi
else
  echo 'Using host .ssh folder (shared home)'
fi

# Atuin (shell history)
if ! command -v atuin &>/dev/null; then
  echo "Installing Atuin.."
  curl --proto '=https' --tlsv1.2 -LsSf https://setup.atuin.sh | sh >/dev/null 2>&1
  mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/atuin"
  cp ${DEFAULT_CONFIG_PATH}/atuin_config.toml "${XDG_CONFIG_HOME:-$HOME/.config}/atuin/config.toml"
  echo '# Atuin' >> "${ZDOTDIR}/.zshrc"
  echo 'eval "$(atuin init zsh)"' >> "${ZDOTDIR}/.zshrc"
else
  echo 'Atuin already installed. Skipping..'
fi

# Git
# Set editor - use container-specific config when shared home
if [[ "$APPARATUS_SHARED_HOME" == "1" ]]; then
  # Ensure git config directory exists
  mkdir -p "$(dirname "$GIT_CONFIG_GLOBAL")"
  echo 'Using container-specific git config'
fi
git config --global core.editor /opt/nvim-linux-x86_64/bin/nvim

# Chezmoi (dotfiles manager)
# Disabled for now
# if [[ ! -d "${HOME}/.local/share/chezmoi" ]]; then
#   echo "Initializing chezmoi dotfiles.."
#   chezmoi init --apply https://github.com/vincentvdk/dotfiles.git
# else
#   echo 'Chezmoi already initialized. Skipping..'
# fi

