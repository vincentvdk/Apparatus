#!/bin/bash
# Build Apparatus OS ISO locally using titanoboa
set -euo pipefail
set -x

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
TITANOBOA_DIR="${REPO_ROOT}/.titanoboa"
IMAGE="${IMAGE:-ghcr.io/vincentvdk/apparatus-os:latest}"
OUTPUT="${OUTPUT:-apparatus-os.iso}"

echo "=== Apparatus OS ISO Builder ==="
echo "Image: $IMAGE"
echo "Output: $OUTPUT"
echo ""

# Check if running inside distrobox
if [ -f /run/.containerenv ]; then
    echo "Error: This script must run on the host, not inside distrobox"
    echo "Run: distrobox-host-exec $0"
    exit 1
fi

# Check dependencies
for cmd in just podman; do
    if ! command -v "$cmd" &>/dev/null; then
        echo "Error: $cmd is required but not installed"
        exit 1
    fi
done

# Shallow clone titanoboa if not present
if [ ! -d "$TITANOBOA_DIR" ]; then
    echo "Downloading titanoboa..."
    git clone --depth 1 https://github.com/ublue-os/titanoboa.git "$TITANOBOA_DIR"
fi

cd "$TITANOBOA_DIR"

echo ""
echo "Building ISO (this may take 10-20 minutes)..."
echo ""

# Detect privilege escalation command (pkexec for VanillaOS, sudo otherwise)
if command -v pkexec &>/dev/null && ! command -v sudo &>/dev/null; then
    PRIV_CMD="pkexec env PATH=$PATH"
else
    PRIV_CMD="sudo env PATH=$PATH"
fi

# Build with our hook
$PRIV_CMD \
    HOOK_post_rootfs="${REPO_ROOT}/os/iso_files/configure_live_session.sh" \
    TITANOBOA_BUILDER_DISTRO=fedora \
    just --justfile="${TITANOBOA_DIR}/Justfile" --working-directory="${TITANOBOA_DIR}" build "$IMAGE" 0

# Move output
if [ -f "output.iso" ]; then
    mv output.iso "${REPO_ROOT}/${OUTPUT}"
    echo ""
    echo "=== Build complete ==="
    echo "ISO: ${REPO_ROOT}/${OUTPUT}"
    echo ""
    echo "Test with QEMU: cd ${TITANOBOA_DIR} && just vm ${REPO_ROOT}/${OUTPUT}"
else
    echo "Error: output.iso not found"
    exit 1
fi
