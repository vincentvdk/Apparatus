# Apparatus OS build recipes
# Run `just --list` to see available commands

# Default recipe - show available commands
default:
    @just --list

# Version for local builds: git describe (semver tag if present), see docs/releases.md
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`

# Show the version local builds will be stamped with
show-version:
    @echo {{version}}

# === Go tools ===

# Build, vet and test all Go tools (run from each module dir, like CI does)
test-tools:
    #!/usr/bin/env bash
    set -euo pipefail
    for mod in tools/butler tools/soundctl; do
        echo "==> $mod"
        (cd "$mod" && go build ./... && go vet ./... && go test ./...)
    done

# Build the container image
build-container:
    @SILVERBLUE_VERSION=$(grep 'SILVERBLUE_VERSION=' apparatus.env | cut -d'=' -f2) && \
    podman build --build-arg SILVERBLUE_VERSION=$$SILVERBLUE_VERSION --build-arg APPARATUS_VERSION={{version}} -f ./os/Containerfile.bootc -t apparatus-os:latest -t apparatus-os:{{version}} .



# Build the ISO (requires container to be built first)
build-iso: build-container
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p ./iso-output
    podman run --rm -it --privileged \
        --security-opt label=type:unconfined_t \
        -v ./iso-output:/output \
        -v /var/lib/containers/storage:/var/lib/containers/storage \
        -v ./os/disk_config/iso.toml:/config.toml:ro \
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

# === Distrobox recipes ===

# Build the distrobox image
build-box:
    podman build --build-arg APPARATUS_VERSION={{version}} -f ./box/Containerfile -t ghcr.io/vincentvdk/apparatus-box:latest -t ghcr.io/vincentvdk/apparatus-box:{{version}} .

# Push distrobox to GHCR
push-box:
    podman push ghcr.io/vincentvdk/apparatus-box:latest

# Build and push distrobox to GHCR
build-push-box: build-box push-box

# Create/recreate all distroboxes from the assemble manifest
assemble-boxes:
    distrobox-assemble create --file ~/.config/distrobox/distrobox.ini

# Create a new distrobox (usage: just create-box mybox)
# Adds the box to ~/.config/distrobox/distrobox.ini, then assembles it
create-box name:
    #!/usr/bin/env bash
    set -euo pipefail
    ini="$HOME/.config/distrobox/distrobox.ini"
    mkdir -p "$(dirname "$ini")"
    touch "$ini"
    if ! grep -q "^\[{{name}}\]" "$ini"; then
        printf '\n[%s]\nimage = ghcr.io/vincentvdk/apparatus-box:latest\npull = true\ninit_hooks = distrobox-host-exec butler sync %s\n' "{{name}}" "{{name}}" >> "$ini"
    fi
    distrobox-assemble create --file "$ini"
