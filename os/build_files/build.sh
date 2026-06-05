# Tool versions
load_env() {
    local env_file="$1"
    if [ -f "$env_file" ]; then
        set -a
        . "$env_file"
        set +a
    else
        echo "Warning: $env_file not found."
    fi
}

load_env /delivery/build_files/apparatus.env

# Ensure essential versions are set
: "${FEDORA_RELEASE:?FEDORA_RELEASE must be defined in apparatus.env}"
: "${HYPRDYNAMICMONITORS_VERSION:?HYPRDYNAMICMONITORS_VERSION must be defined in apparatus.env}"
: "${WALKER_VERSION:?WALKER_VERSION must be defined in apparatus.env}"
: "${ELEPHANT_VERSION:?ELEPHANT_VERSION must be defined in apparatus.env}"

RELEASE="${FEDORA_RELEASE}"
VERSION="${APPARATUS_VERSION:-dev}"
