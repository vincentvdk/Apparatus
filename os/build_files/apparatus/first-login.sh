#!/usr/bin/env bash
# Apparatus First Login
# Runs butler setup on first login

INIT_DONE="$HOME/.config/apparatus/init-done"
BOOTC_SWITCHED="$HOME/.config/apparatus/bootc-switched"

# Exit if already initialized
if [ -f "$INIT_DONE" ]; then
    exit 0
fi

# Switch bootc to remote registry (runs once, in background)
# This ensures rpm-ostree/bootc updates pull from GHCR instead of install source
if [ ! -f "$BOOTC_SWITCHED" ]; then
    (
        sleep 5  # Wait for network
        if bootc status 2>/dev/null | grep -q "ostree-unverified-image:oci:/run/install"; then
            pkexec bootc switch --transport registry ghcr.io/vincentvdk/apparatus-os:latest || true
        fi
        mkdir -p "$HOME/.config/apparatus"
        touch "$BOOTC_SWITCHED"
    ) &
fi

# Run butler init in a terminal
kitty -e butler init

# Start services and reload Hyprland if init completed successfully
if [ -f "$INIT_DONE" ]; then
    # Start services that were missed (exec-once doesn't re-run on reload)
    waybar &
    hyprpaper &
    mako &
    hyprctl reload
fi
