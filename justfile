# Apparatus OS build recipes
# Run `just --list` to see available commands

# Default recipe - show available commands
default:
    @just --list

# Get the short git hash for versioning
version := `git rev-parse --short HEAD 2>/dev/null || echo "dev"`

# Build the container image
build-container:
    podman build --build-arg APPARATUS_VERSION={{version}} -f ./os/Containerfile.bootc -t apparatus-os .

# Build the ISO (requires container to be built first)
build-iso: build-container
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p ./iso-output
    podman run --rm -it --privileged \
        --security-opt label=type:unconfined_t \
        -v ./iso-output:/output \
        -v /var/lib/containers/storage:/var/lib/containers/storage \
        -v ./disk_config/iso.toml:/config.toml:ro \
        quay.io/centos-bootc/bootc-image-builder:latest \
        --type anaconda-iso \
        --rootfs ext4 \
        --local \
        localhost/apparatus-os:latest
    echo "ISO available in: ./iso-output/"
    ls -lh ./iso-output/*.iso 2>/dev/null || true

# Build only the container (alias)
container: build-container

# Build the full ISO
iso: build-iso

# Clean build artifacts
clean:
    rm -rf ./iso-output
    podman rmi apparatus-os:latest 2>/dev/null || true

# Run the container for testing
run-container:
    podman run --rm -it apparatus-os:latest /bin/bash

# Check container image exists
check:
    @podman images | grep apparatus-os || echo "No apparatus-os image found. Run 'just build-container' first."

# Push container to GHCR
push:
    podman push apparatus-os:latest ghcr.io/vincentvdk/apparatus-os:latest

# Build and push container to GHCR
build-push: build-container push
