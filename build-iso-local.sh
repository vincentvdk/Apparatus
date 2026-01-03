#!/bin/bash
# Build Apparatus OS ISO locally
# Usage: sudo ./build-iso-local.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE_NAME="apparatus-os"
OUTPUT_DIR="${SCRIPT_DIR}/iso-output"
CONFIG_FILE="${SCRIPT_DIR}/disk_config/iso.toml"
BIB_IMAGE="quay.io/centos-bootc/bootc-image-builder:latest"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root (sudo)"
    exit 1
fi

echo "=== Building Apparatus OS ISO locally ==="
echo "Output directory: ${OUTPUT_DIR}"
echo ""

# Step 1: Build the container image
echo "=== Step 1/2: Building container image ==="
podman build -f "${SCRIPT_DIR}/os/Containerfile.bootc" -t "${IMAGE_NAME}" "${SCRIPT_DIR}"

# Step 2: Create output directory
mkdir -p "${OUTPUT_DIR}"

# Step 3: Build the ISO
echo ""
echo "=== Step 2/2: Building ISO ==="
podman run --rm -it --privileged --pull=newer \
    --security-opt label=type:unconfined_t \
    -v "${OUTPUT_DIR}:/output" \
    -v /var/lib/containers/storage:/var/lib/containers/storage \
    -v "${CONFIG_FILE}:/config.toml:ro" \
    "${BIB_IMAGE}" \
    --type anaconda-iso \
    --rootfs ext4 \
    --local \
    "localhost/${IMAGE_NAME}"

echo ""
echo "=== Build complete ==="
echo "ISO available in: ${OUTPUT_DIR}"
ls -lh "${OUTPUT_DIR}"/*.iso 2>/dev/null || echo "No ISO found in output directory"
