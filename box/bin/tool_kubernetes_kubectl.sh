#!/usr/bin/env bash
# Tool version manager for kubectl (using mise)
# $1: action to take
# $2: version (optional, for install_version)

TOOL="kubectl"
ACTION=$1

get_installed_versions() {
  mise ls "$TOOL" 2>/dev/null | awk '{print $2}' | tr '\n' ' '
}

main() {
  case "$1" in
    latest)
      version_latest
      ;;
    current)
      version_current
      ;;
    available)
      list_installed_versions
      ;;
    setversion)
      set_version
      ;;
    install_version)
      version_manual "$2"
      ;;
    *)
      echo "Usage: $0 {latest|current|available|setversion|install_version <version>}"
      ;;
  esac
}

# Install the latest version
version_latest() {
  echo "Installing/updating LATEST of: ${TOOL}"
  mise use --global "${TOOL}@latest"
}

# Show the current version
version_current() {
  CURRENT_VERSION=$(mise current "$TOOL" 2>/dev/null)
  if [[ -z "${CURRENT_VERSION}" ]]; then
    gum style --foreground "#f14e32" "${TOOL} is not installed. Please install it first.."
  else
    gum style --foreground "#f14e32" "${CURRENT_VERSION}"
  fi
}

# List all installed versions
list_installed_versions() {
  AVAILABLE_VERSIONS=$(get_installed_versions)
  if [[ -z "${AVAILABLE_VERSIONS}" ]]; then
    gum style --foreground "#f14e32" "${TOOL} is not installed. Please install it first.."
  else
    gum style --foreground "#f14e32" "${AVAILABLE_VERSIONS}"
  fi
}

# Install specific version
version_manual() {
  mise use --global "${TOOL}@$1"
}

# Set a default version from installed versions
set_version() {
  AVAILABLE_VERSIONS=($(get_installed_versions))
  if [[ ${#AVAILABLE_VERSIONS[@]} -eq 0 ]]; then
    gum style --foreground "#f14e32" "${TOOL} is not installed. Please install it first.."
    return 1
  fi
  VERSION=$(gum choose "${AVAILABLE_VERSIONS[@]}")
  mise use --global "${TOOL}@${VERSION}"
}

main "$1" "$2"
