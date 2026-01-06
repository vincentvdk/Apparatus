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
# This ensures bootc updates pull from GHCR instead of install media
if [ ! -f "$BOOTC_SWITCHED" ]; then
    (
        sleep 5  # Wait for network
        mkdir -p "$HOME/.config/apparatus"
        # Always attempt switch - will fail gracefully if already on correct source
        pkexec bootc switch --transport registry ghcr.io/vincentvdk/apparatus-os:latest 2>/dev/null || true
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
