# $1: action to take
#

TOOL="kubectl"
AVAILABLE_VERSIONS=($(asdf list $TOOL | tr -d '*\n' | xargs printf '%s '))
ACTION=$1

main() {
  if [[ $1 == "latest" ]]; then
    version_latest ${TOOL}
  elif [[ $1 == "current" ]]; then
    version_current ${TOOL}
  elif [[ $1 == "available" ]]; then
    list_installed_versions ${TOOL}
  elif [[ $1 == "setversion" ]]; then
    set_version ${TOOL}
  elif [[ $1 == "install_version" ]]; then
    version_manual ${TOOL} $2
  else
    echo "Action or option not supported.."
  fi
}

# Install the latest version
version_latest() {
  echo "Installing/updating LATEST of: ${TOOL}"
  asdf plugin add "${TOOL}"
  asdf install "${TOOL}" latest
  asdf global "${TOOL}" latest
}

# Show the current version
version_current() {
  CURRENT_VERSION=$(asdf current ${TOOL})
  if [[ -z "${CURRENT_VERSION}" ]]; then
    gum style --foreground "#f14e32" "${TOOL} is not installed. Please install it first.."
  else
    gum style --foreground "#f14e32" ${CURRENT_VERSION}
  fi
}

# List all installed versions
list_installed_versions() {
  if [[ -z "${AVAILABLE_VERSIONS}" ]]; then
    gum style --foreground "#f14e32" "${TOOL} is not installed. Please install it first.."
  else
    gum style --foreground "#f14e32" ${AVAILABLE_VERSIONS}
  fi
}

# Install specifig version
version_manual() {
  asdf install ${TOOL} $2
  asdf global ${TOOL} $2
}

# Set a default version
set_version() {
  VERSION=$(gum choose ${AVAILABLE_VERSIONS})
  asdf global ${TOOL} ${VERSION}
  asdf reshim ${TOOL}
}

main $1 $2
