#!/usr/bin/env bash
# Universal tool manager using mise
# Usage: tool-manager.sh <tool> <action> [version]

TOOL=$1
ACTION=$2
VERSION=$3

get_installed_versions() {
  mise ls "$TOOL" 2>/dev/null | awk '{print $2}' | tr '\n' ' '
}

version_latest() {
  echo "Installing/updating LATEST of: ${TOOL}"
  mise use --global "${TOOL}@latest"
}

version_current() {
  CURRENT_VERSION=$(mise current "$TOOL" 2>/dev/null)
  if [[ -z "${CURRENT_VERSION}" ]]; then
    echo "${TOOL} is not installed"
  else
    echo "${CURRENT_VERSION}"
  fi
}

list_installed_versions() {
  AVAILABLE_VERSIONS=$(get_installed_versions)
  if [[ -z "${AVAILABLE_VERSIONS}" ]]; then
    echo "${TOOL} is not installed"
  else
    echo "${AVAILABLE_VERSIONS}"
  fi
}

version_manual() {
  echo "Installing ${TOOL} version $1"
  mise use --global "${TOOL}@$1"
}

set_version() {
  AVAILABLE_VERSIONS=($(get_installed_versions))
  if [[ ${#AVAILABLE_VERSIONS[@]} -eq 0 ]]; then
    echo "${TOOL} is not installed"
    return 1
  fi
  # Output versions for selection (butler-tools will handle the UI)
  for v in "${AVAILABLE_VERSIONS[@]}"; do
    echo "$v"
  done
}

case "$ACTION" in
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
    if [[ -n "$VERSION" ]]; then
      mise use --global "${TOOL}@${VERSION}"
    else
      set_version
    fi
    ;;
  install_version)
    version_manual "$VERSION"
    ;;
  *)
    echo "Usage: $0 <tool> {latest|current|available|setversion|install_version} [version]"
    exit 1
    ;;
esac
